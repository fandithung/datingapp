package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"datingapp/internal"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Repository interface {
	BeginTx(ctx context.Context) (*sqlx.Tx, error)
	CreateProfileResponse(ctx context.Context, tx *sqlx.Tx, response *internal.ProfileResponse) error
	GetProfiles(ctx context.Context, userID uuid.UUID, limit int) ([]*internal.User, error)
	CreateUser(ctx context.Context, tx *sqlx.Tx, user *internal.User) (uuid.UUID, error)
	GetUserByEmail(ctx context.Context, email string) (*internal.User, error)
	GetDailyInteractionCount(ctx context.Context, userID uuid.UUID, since time.Time) (int, error)
	GetFeatures(ctx context.Context) ([]*internal.SubscriptionFeature, error)
	GetFeatureByID(ctx context.Context, featureID uuid.UUID) (*internal.SubscriptionFeature, error)
	CreateUserFeature(ctx context.Context, tx *sqlx.Tx, feature *internal.UserFeature) error
	GetUserFeatures(ctx context.Context, userID uuid.UUID) ([]*internal.UserFeature, error)
	HasActiveFeature(ctx context.Context, userID uuid.UUID, featureName string) (bool, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *repository {
	return &repository{db: db}
}

func (r *repository) CreateProfileResponse(ctx context.Context, tx *sqlx.Tx, response *internal.ProfileResponse) error {
	query := `
		INSERT INTO profile_responses (
			id, from_user_id, to_user_id, response_type,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id`

	var id uuid.UUID
	err := tx.QueryRowContext(ctx, query,
		response.ID,
		response.FromUserID,
		response.ToUserID,
		response.ResponseType,
	).Scan(&id)

	if err != nil {
		if isPgUniqueViolation(err) {
			return internal.ErrConflictingResponse
		}
		return fmt.Errorf("create profile response: %w", err)
	}

	return nil
}

func (r *repository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}

func (r *repository) GetDailyInteractionCount(ctx context.Context, userID uuid.UUID, since time.Time) (int, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM profile_responses
		WHERE from_user_id = $1
		AND created_at >= $2`

	err := r.db.GetContext(ctx, &count, query, userID, since)
	if err != nil {
		return 0, fmt.Errorf("get daily interaction count: %w", err)
	}

	return count, nil
}

func (r *repository) GetProfiles(ctx context.Context, userID uuid.UUID, limit int) ([]*internal.User, error) {
	query := `
		SELECT id, email, name, bio, birth_date, gender, created_at, updated_at
		FROM users
		WHERE id != $1
		AND id NOT IN (
			SELECT to_user_id
			FROM profile_responses
			WHERE from_user_id = $1
		)
		ORDER BY RANDOM()
		LIMIT $2`

	var users []*internal.User
	if err := r.db.SelectContext(ctx, &users, query, userID, limit); err != nil {
		return nil, fmt.Errorf("select candidates: %w", err)
	}

	return users, nil
}

func (r *repository) CreateUser(ctx context.Context, tx *sqlx.Tx, user *internal.User) (uuid.UUID, error) {
	query := `
		INSERT INTO users (email, password_hash, name, bio, birth_date, gender, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	var id uuid.UUID

	err := tx.QueryRowContext(ctx, query,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.Bio,
		user.BirthDate,
		user.Gender,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&id)
	if err != nil {
		if isPgUniqueViolation(err) {
			return uuid.Nil, internal.ErrEmailAlreadyExists
		}
		return uuid.Nil, fmt.Errorf("insert user: %w", err)
	}

	return id, nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*internal.User, error) {
	user := &internal.User{}
	query := `
		SELECT id, email, password_hash, name, bio, birth_date, gender, created_at, updated_at
		FROM users
		WHERE email = $1`

	err := r.db.GetContext(ctx, user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, internal.ErrUserNotFound
		}
		return nil, fmt.Errorf("select user: %w", err)
	}

	return user, nil
}

func (r *repository) GetFeatures(ctx context.Context) ([]*internal.SubscriptionFeature, error) {
	var features []*internal.SubscriptionFeature
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM subscription_features
		ORDER BY name`

	if err := r.db.SelectContext(ctx, &features, query); err != nil {
		return nil, fmt.Errorf("select features: %w", err)
	}

	return features, nil
}

func (r *repository) GetFeatureByID(ctx context.Context, featureID uuid.UUID) (*internal.SubscriptionFeature, error) {
	feature := &internal.SubscriptionFeature{}
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM subscription_features
		WHERE id = $1`

	if err := r.db.GetContext(ctx, feature, query, featureID); err != nil {
		return nil, fmt.Errorf("select feature: %w", err)
	}

	return feature, nil
}

func (r *repository) CreateUserFeature(ctx context.Context, tx *sqlx.Tx, feature *internal.UserFeature) error {
	query := `
		INSERT INTO user_features (
			user_id, feature_id, value, start_date, end_date,
			status, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`

	_, err := tx.ExecContext(ctx, query,
		feature.UserID,
		feature.FeatureID,
		feature.Value,
		feature.StartDate,
		feature.EndDate,
		feature.Status,
	)
	if err != nil {
		if isPgUniqueViolation(err) {
			return internal.ErrFeatureAlreadySubscribed
		}
		return fmt.Errorf("insert user feature: %w", err)
	}

	return nil
}

func (r *repository) GetUserFeatures(ctx context.Context, userID uuid.UUID) ([]*internal.UserFeature, error) {
	currentTime := time.Now()

	query := `
		SELECT
			uf.id,
			uf.user_id,
			uf.feature_id,
			uf.value,
			uf.start_date,
			uf.end_date,
			uf.status,
			uf.created_at,
			uf.updated_at,
			sf.name as feature_name,
			sf.description as feature_description
		FROM user_features uf
		JOIN subscription_features sf ON sf.id = uf.feature_id
		WHERE uf.user_id = $1
			AND uf.status = 'active'
			AND uf.start_date <= $2
			AND (uf.end_date IS NULL OR uf.end_date > $2)
		ORDER BY uf.created_at DESC`

	var features []*internal.UserFeature
	if err := r.db.SelectContext(ctx, &features, query, userID, currentTime); err != nil {
		return nil, fmt.Errorf("select user features: %w", err)
	}

	return features, nil
}

func (r *repository) HasActiveFeature(ctx context.Context, userID uuid.UUID, featureName string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM user_features uf
			JOIN subscription_features sf ON sf.id = uf.feature_id
			WHERE uf.user_id = $1
				AND sf.name = $2
				AND uf.start_date <= NOW()
				AND (uf.end_date IS NULL OR uf.end_date > NOW())
				AND uf.status = 'active'
		)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, userID, featureName)
	if err != nil {
		return false, fmt.Errorf("check active feature: %w", err)
	}

	fmt.Println("exists", exists)

	return exists, nil
}

func isPgUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}
