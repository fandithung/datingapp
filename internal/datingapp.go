package internal

import (
	"context"

	"github.com/google/uuid"
)

type UserService interface {
	SignUp(ctx context.Context, user *User, password string) error
	Login(ctx context.Context, email, password string) (string, error)
}

type ProfileService interface {
	GetProfiles(ctx context.Context, userID uuid.UUID) ([]*User, error)
	CreateProfileResponse(ctx context.Context, fromUserID, toUserID uuid.UUID, responseType string) error
}

type FeatureService interface {
	GetFeatures(ctx context.Context) ([]*SubscriptionFeature, error)
	SubscribeToFeature(ctx context.Context, feature *UserFeature, period string) error
	GetUserFeatures(ctx context.Context, userID uuid.UUID) ([]*UserFeature, error)
}
