package seeder

import (
	"context"
	"datingapp/internal"
	"fmt"
	"math/rand"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type Seeder struct {
	db *sqlx.DB
}

func NewSeeder(db *sqlx.DB) *Seeder {
	return &Seeder{
		db: db,
	}
}

func (s *Seeder) SeedUsers(ctx context.Context, count int) error {
	query := `
		INSERT INTO users (
			id, email, password_hash, name, bio, birth_date, gender,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)`

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	password := "Password123!"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	genders := []string{"male", "female", "other"}
	now := time.Now()
	users := make([]uuid.UUID, 0, count)

	for i := 0; i < count; i++ {
		userID := uuid.New()
		users = append(users, userID)

		user := &internal.User{
			ID:           userID,
			Email:        gofakeit.Email(),
			PasswordHash: string(hashedPassword),
			Name:         gofakeit.Name(),
			Bio:          gofakeit.Sentence(20),
			BirthDate:    now.AddDate(-rand.Intn(42)-18, -rand.Intn(12), -rand.Intn(28)),
			Gender:       genders[rand.Intn(len(genders))],
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		_, err = stmt.ExecContext(ctx,
			user.ID,
			user.Email,
			user.PasswordHash,
			user.Name,
			user.Bio,
			user.BirthDate,
			user.Gender,
			user.CreatedAt,
			user.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("insert user %d: %w", i+1, err)
		}

		if (i+1)%100 == 0 {
			fmt.Printf("seeded %d users\n", i+1)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	fmt.Printf("successfully seeded %d users\n", count)

	if err := s.seedResponses(ctx, users); err != nil {
		return fmt.Errorf("seed responses: %w", err)
	}

	return nil
}

func (s *Seeder) seedResponses(ctx context.Context, users []uuid.UUID) error {
	query := `
		INSERT INTO profile_responses (
			from_user_id, to_user_id, response_type,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5
		)`

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	responseTypes := []string{"like", "pass"}
	responsesPerUser := 20 // TODO: make this dynamic
	totalResponses := len(users) * responsesPerUser
	now := time.Now()

	fmt.Printf("generating %d profile responses...\n", totalResponses)

	existingResponses := make(map[string]bool)

	for i, fromUserID := range users {
		respondTo := make([]uuid.UUID, 0, responsesPerUser)
		candidates := append(users[:i], users[i+1:]...) // All users except current
		rand.Shuffle(len(candidates), func(i, j int) {
			candidates[i], candidates[j] = candidates[j], candidates[i]
		})

		for _, toUserID := range candidates {
			key := fmt.Sprintf("%s-%s", fromUserID, toUserID)
			reverseKey := fmt.Sprintf("%s-%s", toUserID, fromUserID)

			if existingResponses[key] || existingResponses[reverseKey] {
				continue
			}

			respondTo = append(respondTo, toUserID)
			existingResponses[key] = true

			if len(respondTo) >= responsesPerUser {
				break
			}
		}

		for _, toUserID := range respondTo {
			if toUserID == fromUserID {
				continue
			}

			response := &internal.ProfileResponse{
				FromUserID:   fromUserID,
				ToUserID:     toUserID,
				ResponseType: responseTypes[rand.Intn(len(responseTypes))],
				CreatedAt:    now.Add(-time.Duration(rand.Intn(7*24)) * time.Hour),
				UpdatedAt:    now,
			}

			_, err = stmt.ExecContext(ctx,
				response.FromUserID,
				response.ToUserID,
				response.ResponseType,
				response.CreatedAt,
				response.UpdatedAt,
			)
			if err != nil {
				return fmt.Errorf("insert response: %w", err)
			}
		}

		if (i+1)%100 == 0 {
			fmt.Printf("generated responses for %d users\n", i+1)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	actualTotal := len(existingResponses)
	fmt.Printf("successfully generated %d unique profile responses\n", actualTotal)
	return nil
}
