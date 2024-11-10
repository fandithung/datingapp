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

type featureService struct {
	repo repository.Repository
}

func NewFeatureService(repo repository.Repository) *featureService {
	return &featureService{
		repo: repo,
	}
}

func (s *featureService) GetFeatures(ctx context.Context) ([]*internal.SubscriptionFeature, error) {
	return s.repo.GetFeatures(ctx)
}

func (s *featureService) SubscribeToFeature(ctx context.Context, feature *internal.UserFeature, period string) error {
	if _, err := s.repo.GetFeatureByID(ctx, feature.FeatureID); err != nil {
		return internal.ErrFeatureNotFound
	}

	now := time.Now()
	feature.StartDate = now

	months := map[string]int{
		"1_month":   1,
		"3_months":  3,
		"6_months":  6,
		"12_months": 12,
	}

	endDate := now.AddDate(0, months[period], 0)
	feature.EndDate = &endDate

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	if err := s.repo.CreateUserFeature(ctx, tx, feature); err != nil {
		errRollback := tx.Rollback()
		if errRollback != nil {
			log.Printf("failed to rollback transaction: %v", errRollback)
		}
		return fmt.Errorf("create user feature: %w", err)
	}

	return tx.Commit()
}

func (s *featureService) GetUserFeatures(ctx context.Context, userID uuid.UUID) ([]*internal.UserFeature, error) {
	return s.repo.GetUserFeatures(ctx, userID)
}
