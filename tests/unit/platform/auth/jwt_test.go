package auth_test

import (
	"testing"

	"github.com/noggrj/autorepair/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestJWT(t *testing.T) {
	// auth package uses internal secret key, we don't pass it in constructor anymore
	// or we might need to check if we can set it.
	// Current implementation has hardcoded "secret" or similar.
	
	userID := uuid.New()
	role := "admin"

	// Test Generate Token
	accessToken, refreshToken, expiresIn, err := auth.GenerateToken(userID, role)
	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.True(t, expiresIn > 0)

	// Test Validate Token
	claims, err := auth.ValidateToken(accessToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, role, claims.Role)
}

func TestJWT_InvalidToken(t *testing.T) {
	_, err := auth.ValidateToken("invalid-token")
	assert.Error(t, err)
}
