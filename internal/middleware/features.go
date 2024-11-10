package middleware

import (
	"datingapp/internal"
	"datingapp/internal/repository"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func ActiveFeatures(repo repository.Repository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID := c.Get("user_id")
			if userID == nil {
				return next(c)
			}

			uidStr, ok := userID.(string)
			if !ok || uidStr == "" {
				return next(c)
			}

			uid, err := uuid.Parse(uidStr)
			if err != nil {
				return next(c)
			}

			features, err := repo.GetUserFeatures(c.Request().Context(), uid)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user features")
			}

			ctx := internal.SetActiveFeatures(c.Request().Context(), features)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}
