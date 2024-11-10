package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"datingapp/internal/config"
	"datingapp/internal/handler"
	datingappMiddleware "datingapp/internal/middleware"
	"datingapp/internal/repository"
	"datingapp/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	db     *sqlx.DB
	echo   *echo.Echo
	config config.Config
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func NewServer(db *sqlx.DB, config config.Config) *Server {
	e := echo.New()
	v := validator.New()

	err := v.RegisterValidation("password", handler.PasswordValidator)
	if err != nil {
		log.Fatalf("failed to register password validator: %v", err)
	}

	e.Validator = &CustomValidator{validator: v}

	return &Server{
		config: config,
		db:     db,
		echo:   e,
	}
}

func (s *Server) Start(port string) error {
	s.setupRoutes()
	return s.echo.Start(fmt.Sprintf(":%s", port))
}

func (s *Server) setupRoutes() {
	repo := repository.NewRepository(s.db)
	userSvc := service.NewUserService(repo, s.config.JWTSecret)
	featureSvc := service.NewFeatureService(repo)
	profileSvc := service.NewProfileService(repo, s.config.JWTSecret)
	h := handler.NewHandler(userSvc, featureSvc, profileSvc)

	s.echo.Use(middleware.Logger())
	s.echo.Use(middleware.Recover())
	s.echo.Use(middleware.CORS())

	s.echo.Use(echoprometheus.NewMiddleware("dating_app"))

	go func() {
		metrics := echo.New()
		metrics.GET("/metrics", echoprometheus.NewHandler())
		if err := metrics.Start(":8081"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	v1 := s.echo.Group("/api/v1")

	v1.POST("/signup", h.SignUp)
	v1.POST("/login", h.Login)

	protected := v1.Group("")
	protected.Use(datingappMiddleware.JWTMiddleware(s.config.JWTSecret))
	protected.Use(datingappMiddleware.ActiveFeatures(repo))

	protected.GET("/profiles", h.GetProfiles)
	protected.POST("/profiles/:id/response", h.CreateProfileResponse)

	features := protected.Group("/features")
	features.GET("", h.GetFeatures)
	features.GET("/my", h.GetUserFeatures)
	features.POST("/:id/subscribe", h.SubscribeToFeature)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}
