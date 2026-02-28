package domain_test

import (
	"testing"

	"github.com/noggrj/autorepair/internal/identity/domain"
)

func TestNewUser(t *testing.T) {
	user, err := domain.NewUser("Test User", "test@example.com", "password", domain.RoleAdmin)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", user.Email)
	}

	if !user.CheckPassword("password") {
		t.Error("Expected password validation to pass")
	}

	if user.CheckPassword("wrong") {
		t.Error("Expected password validation to fail")
	}
}
