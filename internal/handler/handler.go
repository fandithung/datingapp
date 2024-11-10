package handler

import (
	"errors"
	"net/http"
	"regexp"
	"time"

	"datingapp/internal"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type Handler struct {
	userSvc    internal.UserService
	featureSvc internal.FeatureService
	profileSvc internal.ProfileService
	log        echo.Logger
}

func NewHandler(userSvc internal.UserService, featureSvc internal.FeatureService, profileSvc internal.ProfileService) *Handler {
	return &Handler{
		userSvc:    userSvc,
		featureSvc: featureSvc,
		profileSvc: profileSvc,
		log:        log.New("handler"),
	}
}

type SignUpRequest struct {
	Email           string    `json:"email" validate:"required,email"`
	Password        string    `json:"password" validate:"required,min=8,password"`
	PasswordConfirm string    `json:"password_confirm" validate:"required,eqfield=Password"`
	Name            string    `json:"name" validate:"required"`
	Bio             string    `json:"bio"`
	BirthDate       time.Time `json:"birth_date" validate:"required"`
	Gender          string    `json:"gender" validate:"required,oneof=male female other"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

// Custom password validator
func PasswordValidator(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)

	return hasUpper && hasLower && hasNumber && hasSpecial
}

func (h *Handler) SignUp(c echo.Context) error {
	var req SignUpRequest
	if err := c.Bind(&req); err != nil {
		h.log.Errorf("failed to bind signup request: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		h.log.Errorf("failed to validate signup request: %v", err)
		var verr validator.ValidationErrors
		if errors.As(err, &verr) {
			for _, e := range verr {
				if e.Tag() == "password" {
					return echo.NewHTTPError(http.StatusBadRequest,
						"password must contain at least one uppercase letter, one lowercase letter, one number, and one special character")
				}
			}
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	user := &internal.User{
		Email:     req.Email,
		Name:      req.Name,
		Bio:       req.Bio,
		BirthDate: req.BirthDate,
		Gender:    req.Gender,
	}

	if err := h.userSvc.SignUp(c.Request().Context(), user, req.Password); err != nil {
		h.log.Errorf("failed to create user: %v", err)
		if errors.Is(err, internal.ErrEmailAlreadyExists) {
			return echo.NewHTTPError(http.StatusConflict, "email already exists")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user")
	}

	return c.JSON(http.StatusCreated, user)
}

func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		h.log.Errorf("failed to bind login request: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		h.log.Errorf("failed to validate login request: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	token, err := h.userSvc.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		h.log.Errorf("failed to login: %v", err)
		switch {
		case errors.Is(err, internal.ErrInvalidCredentials):
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
		case errors.Is(err, internal.ErrUserNotFound):
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to login")
		}
	}

	return c.JSON(http.StatusOK, LoginResponse{Token: token})
}

func (h *Handler) GetProfiles(c echo.Context) error {
	uidCtx := c.Get("user_id")
	if uidCtx == nil {
		h.log.Errorf("user ID is nil")
		return echo.NewHTTPError(http.StatusBadRequest, "user ID is required")
	}

	uid, ok := uidCtx.(string)
	if !ok || uid == "" {
		h.log.Errorf("user ID is empty")
		return echo.NewHTTPError(http.StatusBadRequest, "user ID is required")
	}

	userID, err := uuid.Parse(uid)
	if err != nil {
		h.log.Errorf("invalid user ID: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID")
	}

	profiles, err := h.profileSvc.GetProfiles(c.Request().Context(), userID)
	if err != nil {
		h.log.Errorf("failed to get profiles for user %s: %v", userID, err)
		switch {
		case errors.Is(err, internal.ErrDailyInteractionLimitExceeded):
			return echo.NewHTTPError(http.StatusTooManyRequests, "daily interaction limit exceeded")
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	return c.JSON(http.StatusOK, profiles)
}

func (h *Handler) CreateProfileResponse(c echo.Context) error {
	uidCtx := c.Get("user_id")
	if uidCtx == nil {
		h.log.Errorf("user ID is nil")
		return echo.NewHTTPError(http.StatusBadRequest, "user ID is required")
	}

	uid, ok := uidCtx.(string)
	if !ok || uid == "" {
		h.log.Errorf("user ID is empty")
		return echo.NewHTTPError(http.StatusBadRequest, "user ID is required")
	}

	fromUserID, err := uuid.Parse(uid)
	if err != nil {
		h.log.Errorf("invalid user ID: %+v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID")
	}

	toUserID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.log.Errorf("invalid target user ID: %+v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid target user ID")
	}

	var req struct {
		ResponseType string `json:"response_type" validate:"required,oneof=like pass"`
	}
	if err := c.Bind(&req); err != nil {
		h.log.Errorf("failed to bind response request: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(&req); err != nil {
		h.log.Errorf("failed to validate response request: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.profileSvc.CreateProfileResponse(c.Request().Context(), fromUserID, toUserID, req.ResponseType); err != nil {
		h.log.Errorf("failed to create profile response from %s to %s: %v", fromUserID, toUserID, err)
		switch {
		case errors.Is(err, internal.ErrDailyInteractionLimitExceeded):
			return echo.NewHTTPError(http.StatusTooManyRequests, "daily interaction limit exceeded")
		case errors.Is(err, internal.ErrConflictingResponse):
			return echo.NewHTTPError(http.StatusConflict, "you already responded to this profile")
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create response")
		}
	}

	return c.NoContent(http.StatusCreated)
}

func (h *Handler) GetFeatures(c echo.Context) error {
	features, err := h.featureSvc.GetFeatures(c.Request().Context())
	if err != nil {
		h.log.Errorf("failed to get features: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get features")
	}

	return c.JSON(http.StatusOK, features)
}

func (h *Handler) SubscribeToFeature(c echo.Context) error {
	uidCtx := c.Get("user_id")
	if uidCtx == nil {
		h.log.Errorf("user ID is nil")
		return echo.NewHTTPError(http.StatusBadRequest, "user ID is required")
	}

	uid, ok := uidCtx.(string)
	if !ok || uid == "" {
		h.log.Errorf("user ID is empty")
		return echo.NewHTTPError(http.StatusBadRequest, "user ID is required")
	}

	userID, err := uuid.Parse(uid)
	if err != nil {
		h.log.Errorf("invalid user ID: %+v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID")
	}

	featureID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.log.Errorf("invalid feature ID: %+v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid feature ID")
	}

	var req struct {
		Period string `json:"period" validate:"required,oneof=1_month 3_months 6_months 12_months"`
		Value  *int   `json:"value"`
	}
	if err := c.Bind(&req); err != nil {
		h.log.Errorf("failed to bind subscribe request: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(&req); err != nil {
		h.log.Errorf("failed to validate subscribe request: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	userFeature := &internal.UserFeature{
		UserID:    userID,
		FeatureID: featureID,
		Value:     0,
		Status:    "active",
	}

	if req.Value != nil {
		userFeature.Value = *req.Value
	}

	if err := h.featureSvc.SubscribeToFeature(c.Request().Context(), userFeature, req.Period); err != nil {
		h.log.Errorf("failed to subscribe to feature: %v", err)
		switch {
		case errors.Is(err, internal.ErrFeatureNotFound):
			return echo.NewHTTPError(http.StatusNotFound, "feature not found")
		case errors.Is(err, internal.ErrFeatureAlreadySubscribed):
			return echo.NewHTTPError(http.StatusConflict, "already subscribed to this feature")
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to subscribe to feature")
		}
	}

	return c.JSON(http.StatusCreated, userFeature)
}

func (h *Handler) GetUserFeatures(c echo.Context) error {
	uidCtx := c.Get("user_id")
	if uidCtx == nil {
		h.log.Errorf("user ID is nil")
		return echo.NewHTTPError(http.StatusBadRequest, "user ID is required")
	}

	uid, ok := uidCtx.(string)
	if !ok || uid == "" {
		h.log.Errorf("user ID is empty")
		return echo.NewHTTPError(http.StatusBadRequest, "user ID is required")
	}

	userID, err := uuid.Parse(uid)
	if err != nil {
		h.log.Errorf("invalid user ID: %+v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID")
	}

	features, err := h.featureSvc.GetUserFeatures(c.Request().Context(), userID)
	if err != nil {
		h.log.Errorf("failed to get user features: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user features")
	}

	return c.JSON(http.StatusOK, features)
}
