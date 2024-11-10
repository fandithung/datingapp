package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestJWTMiddleware(t *testing.T) {
	e := echo.New()
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	jwtSecret := "test-secret"

	tests := []struct {
		name           string
		setupAuth      func() string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid token",
			setupAuth: func() string {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"user_id": "123e4567-e89b-12d3-a456-426614174000",
					"email":   "test@example.com",
					"exp":     time.Now().Add(time.Hour).Unix(),
				})
				tokenString, _ := token.SignedString([]byte(jwtSecret))
				return "Bearer " + tokenString
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "expired token",
			setupAuth: func() string {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"user_id": "123e4567-e89b-12d3-a456-426614174000",
					"email":   "test@example.com",
					"exp":     time.Now().Add(-time.Hour).Unix(),
				})
				tokenString, _ := token.SignedString([]byte(jwtSecret))
				return "Bearer " + tokenString
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid token",
		},
		{
			name: "missing authorization header",
			setupAuth: func() string {
				return ""
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "missing authorization header",
		},
		{
			name: "invalid authorization format",
			setupAuth: func() string {
				return "InvalidFormat token123"
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid authorization header format",
		},
		{
			name: "invalid token format",
			setupAuth: func() string {
				return "Bearer invalid.token.format"
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid token",
		},
		{
			name: "wrong signing method",
			setupAuth: func() string {
				token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
					"user_id": "123e4567-e89b-12d3-a456-426614174000",
					"email":   "test@example.com",
					"exp":     time.Now().Add(time.Hour).Unix(),
				})
				tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
				return "Bearer " + tokenString
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if auth := tt.setupAuth(); auth != "" {
				req.Header.Set(echo.HeaderAuthorization, auth)
			}

			middleware := JWTMiddleware(jwtSecret)
			h := middleware(handler)

			err := h(c)

			if tt.expectedError != "" {
				if assert.Error(t, err) {
					he, ok := err.(*echo.HTTPError)
					if assert.True(t, ok) {
						assert.Equal(t, tt.expectedStatus, he.Code)
						assert.Equal(t, tt.expectedError, he.Message)
					}
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)

				if rec.Code == http.StatusOK {
					assert.NotNil(t, c.Get("user_id"))
					assert.NotNil(t, c.Get("email"))
				}
			}
		})
	}
}

func TestJWTMiddleware_ContextValues(t *testing.T) {
	e := echo.New()
	jwtSecret := "test-secret"
	expectedUserID := "123e4567-e89b-12d3-a456-426614174000"
	expectedEmail := "test@example.com"

	handler := func(c echo.Context) error {
		assert.Equal(t, expectedUserID, c.Get("user_id"))
		assert.Equal(t, expectedEmail, c.Get("email"))
		return c.String(http.StatusOK, "test")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": expectedUserID,
		"email":   expectedEmail,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(jwtSecret))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := JWTMiddleware(jwtSecret)
	h := middleware(handler)

	err := h(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
