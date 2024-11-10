package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"datingapp/internal"
	"datingapp/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repo      repository.Repository
	jwtSecret []byte
}

func NewUserService(repo repository.Repository, jwtSecret string) *userService {
	return &userService{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *userService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", internal.ErrInvalidCredentials
	}

	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // 24 hours
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signedToken, nil
}

func (s *userService) SignUp(ctx context.Context, user *internal.User, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	user.PasswordHash = string(hashedPassword)

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	id, err := s.repo.CreateUser(ctx, tx, user)
	if err != nil {
		errRollback := tx.Rollback()
		if errRollback != nil {
			log.Printf("failed to rollback transaction: %v", errRollback)
		}
		return fmt.Errorf("create user: %w", err)
	}

	user.ID = id

	return tx.Commit()
}
