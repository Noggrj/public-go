package config_test

import (
	"os"
	"testing"

	"github.com/noggrj/autorepair/internal/platform/config"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Set env vars
	os.Setenv("PORT", "9090")
	os.Setenv("DB_URL", "postgres://user:pass@localhost:5432/db")

	cfg, err := config.Load()
	assert.NoError(t, err)

	assert.Equal(t, "9090", cfg.Port)
	assert.Equal(t, "postgres://user:pass@localhost:5432/db", cfg.DBURL)

	// Clean up
	os.Unsetenv("PORT")
	os.Unsetenv("DB_URL")
}

func TestLoad_Defaults(t *testing.T) {
	// Ensure env vars are unset
	os.Unsetenv("PORT")
	os.Unsetenv("DB_URL") // DB_URL is required, so Load should fail if not set, or we set it for test

	// Wait, previous code had defaults for DB_URL?
	// Let's check config.go again.
	// config.go:
	// func Load() (*Config, error) {
	// 	dbURL := os.Getenv("DB_URL")
	// 	if dbURL == "" {
	// 		return nil, fmt.Errorf("DB_URL is required")
	// 	}
    // ...
    // }
    
    // So TestLoad_Defaults will fail if DB_URL is not set.
    // I should set DB_URL for this test or expect error if that's the intent.
    // The previous test asserted a default value for DB_URL, implying older version had default.
    // Current version requires it.
    
    os.Setenv("DB_URL", "postgres://test:test@localhost:5432/test")
	cfg, err := config.Load()
	assert.NoError(t, err)

	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "postgres://test:test@localhost:5432/test", cfg.DBURL)
    
    os.Unsetenv("DB_URL")
}

func TestLoad_MissingDBURL(t *testing.T) {
    os.Unsetenv("DB_URL")
    _, err := config.Load()
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "DB_URL is required")
}
