package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"datingapp/internal"
	mock_service "datingapp/internal/service/mock"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func TestHandler_SignUp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_service.NewMockUserService(ctrl)
	featureSvc := mock_service.NewMockFeatureService(ctrl)
	profileSvc := mock_service.NewMockProfileService(ctrl)
	h := NewHandler(mockSvc, featureSvc, profileSvc)

	e := echo.New()
	v := validator.New()
	v.RegisterValidation("password", PasswordValidator)
	e.Validator = &CustomValidator{validator: v}

	validUser := &internal.User{
		Email:     "test@example.com",
		Name:      "Test User",
		Bio:       "Test Bio",
		BirthDate: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		Gender:    "male",
	}

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMock      func()
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful signup",
			requestBody: map[string]interface{}{
				"email":            "test@example.com",
				"password":         "Password123!",
				"password_confirm": "Password123!",
				"name":             "Test User",
				"bio":              "Test Bio",
				"birth_date":       "1990-01-01T00:00:00Z",
				"gender":           "male",
			},
			setupMock: func() {
				mockSvc.EXPECT().
					SignUp(gomock.Any(), gomock.Any(), "Password123!").
					DoAndReturn(func(ctx context.Context, user *internal.User, password string) error {
						assert.Equal(t, validUser.Email, user.Email)
						assert.Equal(t, validUser.Name, user.Name)
						assert.Equal(t, validUser.Bio, user.Bio)
						assert.Equal(t, validUser.Gender, user.Gender)
						return nil
					})
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "email already exists",
			requestBody: map[string]interface{}{
				"email":            "test@example.com",
				"password":         "Password123!",
				"password_confirm": "Password123!",
				"name":             "Test User",
				"bio":              "Test Bio",
				"birth_date":       "1990-01-01T00:00:00Z",
				"gender":           "male",
			},
			setupMock: func() {
				mockSvc.EXPECT().
					SignUp(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(internal.ErrEmailAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "email already exists",
		},
		{
			name: "invalid password - no uppercase",
			requestBody: map[string]interface{}{
				"email":            "test@example.com",
				"password":         "password123!",
				"password_confirm": "password123!",
				"name":             "Test User",
				"bio":              "Test Bio",
				"birth_date":       "1990-01-01T00:00:00Z",
				"gender":           "male",
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "password must contain at least one uppercase letter, one lowercase letter, one number, and one special character",
		},
		{
			name: "password mismatch",
			requestBody: map[string]interface{}{
				"email":            "test@example.com",
				"password":         "Password123!",
				"password_confirm": "Password124!",
				"name":             "Test User",
				"bio":              "Test Bio",
				"birth_date":       "1990-01-01T00:00:00Z",
				"gender":           "male",
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid email",
			requestBody: map[string]interface{}{
				"email":            "invalid-email",
				"password":         "Password123!",
				"password_confirm": "Password123!",
				"name":             "Test User",
				"bio":              "Test Bio",
				"birth_date":       "1990-01-01T00:00:00Z",
				"gender":           "male",
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid gender",
			requestBody: map[string]interface{}{
				"email":            "test@example.com",
				"password":         "Password123!",
				"password_confirm": "Password123!",
				"name":             "Test User",
				"bio":              "Test Bio",
				"birth_date":       "1990-01-01T00:00:00Z",
				"gender":           "invalid",
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := h.SignUp(c)

			if err == nil {
				assert.Equal(t, tt.expectedStatus, rec.Code)
				if rec.Code == http.StatusCreated {
					var response internal.User
					err := json.Unmarshal(rec.Body.Bytes(), &response)
					assert.NoError(t, err)
					assert.Equal(t, validUser.Email, response.Email)
					assert.Equal(t, validUser.Name, response.Name)
				}
			} else {
				var httpError *echo.HTTPError
				assert.ErrorAs(t, err, &httpError)
				assert.Equal(t, tt.expectedStatus, httpError.Code)
				if tt.expectedError != "" {
					assert.Equal(t, tt.expectedError, httpError.Message)
				}
			}
		})
	}
}

func TestHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_service.NewMockUserService(ctrl)
	featureSvc := mock_service.NewMockFeatureService(ctrl)
	profileSvc := mock_service.NewMockProfileService(ctrl)
	h := NewHandler(mockSvc, featureSvc, profileSvc)

	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMock      func()
		expectedStatus int
		expectedError  string
		expectedToken  string
	}{
		{
			name: "successful login",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "Password123!",
			},
			setupMock: func() {
				mockSvc.EXPECT().
					Login(gomock.Any(), "test@example.com", "Password123!").
					Return("valid.jwt.token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedToken:  "valid.jwt.token",
		},
		{
			name: "invalid credentials",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "WrongPassword123!",
			},
			setupMock: func() {
				mockSvc.EXPECT().
					Login(gomock.Any(), "test@example.com", "WrongPassword123!").
					Return("", internal.ErrInvalidCredentials)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid credentials",
		},
		{
			name: "missing email",
			requestBody: map[string]interface{}{
				"password": "Password123!",
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing password",
			requestBody: map[string]interface{}{
				"email": "test@example.com",
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid email format",
			requestBody: map[string]interface{}{
				"email":    "invalid-email",
				"password": "Password123!",
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "server error",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "Password123!",
			},
			setupMock: func() {
				mockSvc.EXPECT().
					Login(gomock.Any(), "test@example.com", "Password123!").
					Return("", errors.New("unexpected error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to login",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := h.Login(c)

			if err == nil {
				assert.Equal(t, tt.expectedStatus, rec.Code)
				if rec.Code == http.StatusOK {
					var response LoginResponse
					err := json.Unmarshal(rec.Body.Bytes(), &response)
					assert.NoError(t, err)
					assert.Equal(t, tt.expectedToken, response.Token)
				}
			} else {
				if tt.expectedError != "" {
					var httpError *echo.HTTPError
					if assert.Error(t, err) {
						assert.ErrorAs(t, err, &httpError)
						assert.Equal(t, tt.expectedStatus, httpError.Code)
						assert.Equal(t, tt.expectedError, httpError.Message)
					}
				}
			}
		})
	}
}

func TestHandler_GetProfiles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_service.NewMockUserService(ctrl)
	featureSvc := mock_service.NewMockFeatureService(ctrl)
	profileSvc := mock_service.NewMockProfileService(ctrl)
	h := NewHandler(mockSvc, featureSvc, profileSvc)

	e := echo.New()

	validUserID := uuid.New()
	mockProfiles := []*internal.User{
		{
			ID:        uuid.New(),
			Email:     "candidate1@example.com",
			Name:      "Candidate 1",
			Bio:       "Bio 1",
			BirthDate: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
			Gender:    "female",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Email:     "candidate2@example.com",
			Name:      "Candidate 2",
			Bio:       "Bio 2",
			BirthDate: time.Date(1992, 1, 1, 0, 0, 0, 0, time.UTC),
			Gender:    "male",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	tests := []struct {
		name           string
		setupContext   func(echo.Context)
		setupMock      func()
		expectedStatus int
		expectedError  string
		expectedLen    int
	}{
		{
			name: "successful get candidates",
			setupContext: func(c echo.Context) {
				c.Set("user_id", validUserID.String())
			},
			setupMock: func() {
				profileSvc.EXPECT().
					GetProfiles(gomock.Any(), validUserID).
					Return(mockProfiles, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLen:    2,
		},
		{
			name: "no candidates available",
			setupContext: func(c echo.Context) {
				c.Set("user_id", validUserID.String())
			},
			setupMock: func() {
				profileSvc.EXPECT().
					GetProfiles(gomock.Any(), validUserID).
					Return([]*internal.User{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLen:    0,
		},
		{
			name: "invalid user ID in context",
			setupContext: func(c echo.Context) {
				c.Set("user_id", "invalid-uuid")
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid user ID",
		},
		{
			name:           "missing user ID in context",
			setupContext:   func(c echo.Context) {},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "user ID is required",
		},
		{
			name: "invalid user ID in context",
			setupContext: func(c echo.Context) {
				c.Set("user_id", "invalid-uuid")
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid user ID",
		},
		{
			name: "service error",
			setupContext: func(c echo.Context) {
				c.Set("user_id", validUserID.String())
			},
			setupMock: func() {
				profileSvc.EXPECT().
					GetProfiles(gomock.Any(), validUserID).
					Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/candidates", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.setupContext != nil {
				tt.setupContext(c)
			}
			tt.setupMock()

			err := h.GetProfiles(c)

			if tt.expectedError != "" {
				var httpError *echo.HTTPError
				if assert.Error(t, err) {
					assert.ErrorAs(t, err, &httpError)
					assert.Equal(t, tt.expectedStatus, httpError.Code)
					assert.Equal(t, tt.expectedError, httpError.Message)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)

				var response []*internal.User
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, tt.expectedLen)

				if tt.expectedLen > 0 {
					for _, profile := range response {
						assert.NotEmpty(t, profile.ID)
						assert.NotEmpty(t, profile.Email)
						assert.NotEmpty(t, profile.Name)
						assert.NotZero(t, profile.BirthDate)
						assert.NotEmpty(t, profile.Gender)
						assert.Empty(t, profile.PasswordHash)
					}
				}
			}
		})
	}
}

func TestHandler_CreateProfileResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_service.NewMockUserService(ctrl)
	featureSvc := mock_service.NewMockFeatureService(ctrl)
	profileSvc := mock_service.NewMockProfileService(ctrl)
	h := NewHandler(mockSvc, featureSvc, profileSvc)

	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	validUserID := uuid.New()
	targetUserID := uuid.New()

	tests := []struct {
		name           string
		setupContext   func(echo.Context)
		targetID       string
		requestBody    map[string]interface{}
		setupMock      func()
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful response",
			setupContext: func(c echo.Context) {
				c.Set("user_id", validUserID.String())
			},
			targetID: targetUserID.String(),
			requestBody: map[string]interface{}{
				"response_type": "like",
			},
			setupMock: func() {
				profileSvc.EXPECT().
					CreateProfileResponse(gomock.Any(), validUserID, targetUserID, "like").
					Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "daily limit exceeded",
			setupContext: func(c echo.Context) {
				c.Set("user_id", validUserID.String())
			},
			targetID: targetUserID.String(),
			requestBody: map[string]interface{}{
				"response_type": "like",
			},
			setupMock: func() {
				profileSvc.EXPECT().
					CreateProfileResponse(gomock.Any(), validUserID, targetUserID, "like").
					Return(internal.ErrDailyInteractionLimitExceeded)
			},
			expectedStatus: http.StatusTooManyRequests,
			expectedError:  "daily interaction limit exceeded",
		},
		{
			name: "invalid response type",
			setupContext: func(c echo.Context) {
				c.Set("user_id", validUserID.String())
			},
			targetID: targetUserID.String(),
			requestBody: map[string]interface{}{
				"response_type": "invalid",
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing response type",
			setupContext: func(c echo.Context) {
				c.Set("user_id", validUserID.String())
			},
			targetID:       targetUserID.String(),
			requestBody:    map[string]interface{}{},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "already responded",
			setupContext: func(c echo.Context) {
				c.Set("user_id", validUserID.String())
			},
			targetID: targetUserID.String(),
			requestBody: map[string]interface{}{
				"response_type": "like",
			},
			setupMock: func() {
				profileSvc.EXPECT().
					CreateProfileResponse(gomock.Any(), validUserID, targetUserID, "like").
					Return(internal.ErrConflictingResponse)
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "you already responded to this profile",
		},
		{
			name: "service error",
			setupContext: func(c echo.Context) {
				c.Set("user_id", validUserID.String())
			},
			targetID: targetUserID.String(),
			requestBody: map[string]interface{}{
				"response_type": "like",
			},
			setupMock: func() {
				profileSvc.EXPECT().
					CreateProfileResponse(gomock.Any(), validUserID, targetUserID, "like").
					Return(errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			c.SetParamNames("id")
			c.SetParamValues(tt.targetID)

			if tt.setupContext != nil {
				tt.setupContext(c)
			}
			tt.setupMock()

			err := h.CreateProfileResponse(c)

			if err != nil {
				assert.Error(t, err)
				var httpError *echo.HTTPError
				assert.ErrorAs(t, err, &httpError)
				assert.Equal(t, tt.expectedStatus, httpError.Code)
				if tt.expectedError != "" {
					assert.Equal(t, tt.expectedError, httpError.Message)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestHandler_GetFeatures(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(svc *mock_service.MockFeatureService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			setupMock: func(svc *mock_service.MockFeatureService) {
				features := []*internal.SubscriptionFeature{
					{
						ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
						Name:        "Test Feature",
						Description: "Test Description",
					},
				}
				svc.EXPECT().
					GetFeatures(gomock.Any()).
					Return(features, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":"550e8400-e29b-41d4-a716-446655440000","name":"Test Feature","description":"Test Description","created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}]`,
		},
		{
			name: "internal error",
			setupMock: func(svc *mock_service.MockFeatureService) {
				svc.EXPECT().
					GetFeatures(gomock.Any()).
					Return(nil, errors.New("unexpected error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"message":"failed to get features"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mock_service.NewMockFeatureService(ctrl)
			tt.setupMock(mockSvc)

			userSvc := mock_service.NewMockUserService(ctrl)
			profileSvc := mock_service.NewMockProfileService(ctrl)
			h := NewHandler(userSvc, mockSvc, profileSvc)
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/features", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := h.GetFeatures(c)
			if err != nil {
				he, ok := err.(*echo.HTTPError)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedStatus, he.Code)
				assert.Equal(t, tt.expectedBody, fmt.Sprintf(`{"message":"%v"}`, he.Message))
				return
			}

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.JSONEq(t, tt.expectedBody, rec.Body.String())
		})
	}
}

func TestHandler_SubscribeToFeature(t *testing.T) {
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	featureID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")

	tests := []struct {
		name           string
		setupMock      func(svc *mock_service.MockFeatureService)
		requestBody    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			setupMock: func(svc *mock_service.MockFeatureService) {
				svc.EXPECT().
					SubscribeToFeature(gomock.Any(), gomock.Any(), "1_month").
					DoAndReturn(func(_ context.Context, uf *internal.UserFeature, _ string) error {
						assert.Equal(t, userID, uf.UserID)
						assert.Equal(t, featureID, uf.FeatureID)
						assert.Equal(t, 5, uf.Value)
						return nil
					})
			},
			requestBody:    `{"period":"1_month","value":5}`,
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id":"*","user_id":"550e8400-e29b-41d4-a716-446655440001","feature_id":"550e8400-e29b-41d4-a716-446655440002","value":5,"status":"active"}`,
		},
		{
			name: "feature not found",
			setupMock: func(svc *mock_service.MockFeatureService) {
				svc.EXPECT().
					SubscribeToFeature(gomock.Any(), gomock.Any(), "1_month").
					Return(internal.ErrFeatureNotFound)
			},
			requestBody:    `{"period":"1_month","value":5}`,
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"message":"feature not found"}`,
		},
		{
			name: "already subscribed",
			setupMock: func(svc *mock_service.MockFeatureService) {
				svc.EXPECT().
					SubscribeToFeature(gomock.Any(), gomock.Any(), "1_month").
					Return(internal.ErrFeatureAlreadySubscribed)
			},
			requestBody:    `{"period":"1_month","value":5}`,
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"message":"already subscribed to this feature"}`,
		},
		{
			name:           "invalid period",
			setupMock:      func(svc *mock_service.MockFeatureService) {},
			requestBody:    `{"period":"invalid"}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"message":"Key: 'Period' Error:Field validation for 'Period' failed on the 'oneof' tag"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mock_service.NewMockFeatureService(ctrl)
			tt.setupMock(mockSvc)

			userSvc := mock_service.NewMockUserService(ctrl)
			profileSvc := mock_service.NewMockProfileService(ctrl)
			h := NewHandler(userSvc, mockSvc, profileSvc)
			e := echo.New()
			e.Validator = &CustomValidator{validator: validator.New()}

			req := httptest.NewRequest(http.MethodPost, "/features/"+featureID.String(), strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user_id", userID.String())
			c.SetParamNames("id")
			c.SetParamValues(featureID.String())

			err := h.SubscribeToFeature(c)
			if err != nil {
				he, ok := err.(*echo.HTTPError)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedStatus, he.Code)
				assert.Equal(t, tt.expectedBody, fmt.Sprintf(`{"message":"%v"}`, he.Message))
				return
			}

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedStatus == http.StatusCreated {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, userID.String(), response["user_id"])
				assert.Equal(t, featureID.String(), response["feature_id"])
				assert.Equal(t, float64(5), response["value"])
				assert.Equal(t, "active", response["status"])
			} else {
				assert.JSONEq(t, tt.expectedBody, rec.Body.String())
			}
		})
	}
}
