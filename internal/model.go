package internal

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Name         string    `json:"name" db:"name"`
	Bio          string    `json:"bio" db:"bio"`
	BirthDate    time.Time `json:"birth_date" db:"birth_date"`
	Gender       string    `json:"gender" db:"gender"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type SubscriptionFeature struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type UserFeature struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	UserID             uuid.UUID  `json:"user_id" db:"user_id"`
	FeatureID          uuid.UUID  `json:"feature_id" db:"feature_id"`
	Value              int        `json:"value" db:"value"`
	StartDate          time.Time  `json:"start_date" db:"start_date"`
	EndDate            *time.Time `json:"end_date" db:"end_date"`
	Status             string     `json:"status" db:"status"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
	FeatureName        string     `json:"feature_name" db:"feature_name"`
	FeatureDescription string     `json:"feature_description" db:"feature_description"`
}

type ProfileResponse struct {
	ID           uuid.UUID `json:"id" db:"id"`
	FromUserID   uuid.UUID `json:"from_user_id" db:"from_user_id"`
	ToUserID     uuid.UUID `json:"to_user_id" db:"to_user_id"`
	ResponseType string    `json:"response_type" db:"response_type"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type DailyUsage struct {
	ID            uuid.UUID `json:"id" db:"id"`
	UserID        uuid.UUID `json:"user_id" db:"user_id"`
	UsageDate     time.Time `json:"usage_date" db:"usage_date"`
	ResponseCount int       `json:"response_count" db:"response_count"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}
