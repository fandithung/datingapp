package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"datingapp/internal"
	"datingapp/internal/repository"

	"github.com/google/uuid"
)

type profileService struct {
	repo      repository.Repository
	jwtSecret []byte
}

func NewProfileService(repo repository.Repository, jwtSecret string) *profileService {
	return &profileService{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *profileService) CreateProfileResponse(ctx context.Context, fromUserID, toUserID uuid.UUID, responseType string) error {
	unlimitedResponses := internal.HasFeature(ctx, internal.FeatureDailyResponses)

	if !unlimitedResponses {
		since := time.Now().Truncate(24 * time.Hour)
		count, err := s.repo.GetDailyInteractionCount(ctx, fromUserID, since)
		if err != nil {
			return fmt.Errorf("get daily interaction count: %w", err)
		}

		if count >= 10 {
			return internal.ErrDailyInteractionLimitExceeded
		}
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	response := &internal.ProfileResponse{
		ID:           uuid.New(),
		FromUserID:   fromUserID,
		ToUserID:     toUserID,
		ResponseType: responseType,
	}

	if err := s.repo.CreateProfileResponse(ctx, tx, response); err != nil {
		errRollback := tx.Rollback()
		if errRollback != nil {
			log.Printf("failed to rollback transaction: %v", errRollback)
		}
		return fmt.Errorf("create profile response: %w", err)
	}

	return tx.Commit()
}

func (s *profileService) GetProfiles(ctx context.Context, userID uuid.UUID) ([]*internal.User, error) {
	unlimitedResponses := internal.HasFeature(ctx, internal.FeatureDailyResponses)

	if !unlimitedResponses {
		since := time.Now().Truncate(24 * time.Hour)
		count, err := s.repo.GetDailyInteractionCount(ctx, userID, since)
		if err != nil {
			return nil, fmt.Errorf("get daily interaction count: %w", err)
		}

		if count >= 10 {
			return nil, internal.ErrDailyInteractionLimitExceeded
		}
	}

	return s.repo.GetProfiles(ctx, userID, 1)
}
