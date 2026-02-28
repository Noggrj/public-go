//go:build integration

package infrastructure_test

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/noggrj/autorepair/internal/identity/domain"
	"github.com/noggrj/autorepair/internal/identity/infrastructure"
	"github.com/noggrj/autorepair/internal/platform/db"
)

func TestPostgresUserRepository(t *testing.T) {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		t.Skip("DB_URL not set")
	}

	pool, err := db.New(dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}
	defer pool.Close()

	repo := infrastructure.NewPostgresUserRepository(pool.Pool)

	t.Run("Save and GetByEmail", func(t *testing.T) {
		email := "itest_" + uuid.New().String() + "@example.com"
		user, _ := domain.NewUser("Test User", email, "secret", domain.RoleEmployee)

		err := repo.Save(user)
		if err != nil {
			t.Fatalf("Failed to save user: %v", err)
		}

		fetched, err := repo.GetByEmail(email)
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		if fetched.ID != user.ID {
			t.Errorf("Expected ID %v, got %v", user.ID, fetched.ID)
		}

		// Test GetByID
		fetchedByID, err := repo.GetByID(user.ID)
		if err != nil {
			t.Fatalf("Failed to get user by ID: %v", err)
		}
		if fetchedByID.Email != user.Email {
			t.Errorf("Expected Email %v, got %v", user.Email, fetchedByID.Email)
		}
	})

	t.Run("GetByEmail Not Found", func(t *testing.T) {
		_, err := repo.GetByEmail("nonexistent@example.com")
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if err.Error() != "user not found" {
			t.Errorf("Expected 'user not found', got '%v'", err)
		}
	})

	t.Run("GetByID Not Found", func(t *testing.T) {
		_, err := repo.GetByID(uuid.New())
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if err.Error() != "user not found" {
			t.Errorf("Expected 'user not found', got '%v'", err)
		}
	})

	t.Run("Save Duplicate Email", func(t *testing.T) {
		email := "duplicate_" + uuid.New().String() + "@example.com"
		user1, _ := domain.NewUser("User 1", email, "secret", domain.RoleEmployee)
		user2, _ := domain.NewUser("User 2", email, "secret", domain.RoleEmployee)

		err := repo.Save(user1)
		if err != nil {
			t.Fatalf("Failed to save user 1: %v", err)
		}

		err = repo.Save(user2)
		if err == nil {
			t.Error("Expected error due to duplicate email, got nil")
		}
	})
}
