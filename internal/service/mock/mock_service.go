// Code generated by MockGen. DO NOT EDIT.
// Source: internal/datingapp.go
//
// Generated by this command:
//
//	mockgen -source=internal/datingapp.go -destination=internal/service/mock/mock_service.go
//

// Package mock_internal is a generated GoMock package.
package mock_internal

import (
	context "context"
	internal "datingapp/internal"
	reflect "reflect"

	uuid "github.com/google/uuid"
	gomock "go.uber.org/mock/gomock"
)

// MockUserService is a mock of UserService interface.
type MockUserService struct {
	ctrl     *gomock.Controller
	recorder *MockUserServiceMockRecorder
	isgomock struct{}
}

// MockUserServiceMockRecorder is the mock recorder for MockUserService.
type MockUserServiceMockRecorder struct {
	mock *MockUserService
}

// NewMockUserService creates a new mock instance.
func NewMockUserService(ctrl *gomock.Controller) *MockUserService {
	mock := &MockUserService{ctrl: ctrl}
	mock.recorder = &MockUserServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserService) EXPECT() *MockUserServiceMockRecorder {
	return m.recorder
}

// Login mocks base method.
func (m *MockUserService) Login(ctx context.Context, email, password string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Login", ctx, email, password)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Login indicates an expected call of Login.
func (mr *MockUserServiceMockRecorder) Login(ctx, email, password any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Login", reflect.TypeOf((*MockUserService)(nil).Login), ctx, email, password)
}

// SignUp mocks base method.
func (m *MockUserService) SignUp(ctx context.Context, user *internal.User, password string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SignUp", ctx, user, password)
	ret0, _ := ret[0].(error)
	return ret0
}

// SignUp indicates an expected call of SignUp.
func (mr *MockUserServiceMockRecorder) SignUp(ctx, user, password any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SignUp", reflect.TypeOf((*MockUserService)(nil).SignUp), ctx, user, password)
}

// MockProfileService is a mock of ProfileService interface.
type MockProfileService struct {
	ctrl     *gomock.Controller
	recorder *MockProfileServiceMockRecorder
	isgomock struct{}
}

// MockProfileServiceMockRecorder is the mock recorder for MockProfileService.
type MockProfileServiceMockRecorder struct {
	mock *MockProfileService
}

// NewMockProfileService creates a new mock instance.
func NewMockProfileService(ctrl *gomock.Controller) *MockProfileService {
	mock := &MockProfileService{ctrl: ctrl}
	mock.recorder = &MockProfileServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProfileService) EXPECT() *MockProfileServiceMockRecorder {
	return m.recorder
}

// CreateProfileResponse mocks base method.
func (m *MockProfileService) CreateProfileResponse(ctx context.Context, fromUserID, toUserID uuid.UUID, responseType string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateProfileResponse", ctx, fromUserID, toUserID, responseType)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateProfileResponse indicates an expected call of CreateProfileResponse.
func (mr *MockProfileServiceMockRecorder) CreateProfileResponse(ctx, fromUserID, toUserID, responseType any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateProfileResponse", reflect.TypeOf((*MockProfileService)(nil).CreateProfileResponse), ctx, fromUserID, toUserID, responseType)
}

// GetProfiles mocks base method.
func (m *MockProfileService) GetProfiles(ctx context.Context, userID uuid.UUID) ([]*internal.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProfiles", ctx, userID)
	ret0, _ := ret[0].([]*internal.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProfiles indicates an expected call of GetProfiles.
func (mr *MockProfileServiceMockRecorder) GetProfiles(ctx, userID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProfiles", reflect.TypeOf((*MockProfileService)(nil).GetProfiles), ctx, userID)
}

// MockFeatureService is a mock of FeatureService interface.
type MockFeatureService struct {
	ctrl     *gomock.Controller
	recorder *MockFeatureServiceMockRecorder
	isgomock struct{}
}

// MockFeatureServiceMockRecorder is the mock recorder for MockFeatureService.
type MockFeatureServiceMockRecorder struct {
	mock *MockFeatureService
}

// NewMockFeatureService creates a new mock instance.
func NewMockFeatureService(ctrl *gomock.Controller) *MockFeatureService {
	mock := &MockFeatureService{ctrl: ctrl}
	mock.recorder = &MockFeatureServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFeatureService) EXPECT() *MockFeatureServiceMockRecorder {
	return m.recorder
}

// GetFeatures mocks base method.
func (m *MockFeatureService) GetFeatures(ctx context.Context) ([]*internal.SubscriptionFeature, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFeatures", ctx)
	ret0, _ := ret[0].([]*internal.SubscriptionFeature)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFeatures indicates an expected call of GetFeatures.
func (mr *MockFeatureServiceMockRecorder) GetFeatures(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFeatures", reflect.TypeOf((*MockFeatureService)(nil).GetFeatures), ctx)
}

// GetUserFeatures mocks base method.
func (m *MockFeatureService) GetUserFeatures(ctx context.Context, userID uuid.UUID) ([]*internal.UserFeature, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserFeatures", ctx, userID)
	ret0, _ := ret[0].([]*internal.UserFeature)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserFeatures indicates an expected call of GetUserFeatures.
func (mr *MockFeatureServiceMockRecorder) GetUserFeatures(ctx, userID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserFeatures", reflect.TypeOf((*MockFeatureService)(nil).GetUserFeatures), ctx, userID)
}

// SubscribeToFeature mocks base method.
func (m *MockFeatureService) SubscribeToFeature(ctx context.Context, feature *internal.UserFeature, period string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubscribeToFeature", ctx, feature, period)
	ret0, _ := ret[0].(error)
	return ret0
}

// SubscribeToFeature indicates an expected call of SubscribeToFeature.
func (mr *MockFeatureServiceMockRecorder) SubscribeToFeature(ctx, feature, period any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubscribeToFeature", reflect.TypeOf((*MockFeatureService)(nil).SubscribeToFeature), ctx, feature, period)
}
