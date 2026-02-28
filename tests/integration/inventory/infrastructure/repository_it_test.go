//go:build integration

package infrastructure_test

import (
	"context"
	"math/rand"
	"os"
	"testing"
	"time"

	inventoryDomain "github.com/noggrj/autorepair/internal/inventory/domain"
	"github.com/noggrj/autorepair/internal/inventory/infrastructure"
	"github.com/noggrj/autorepair/internal/platform/db"
	"github.com/stretchr/testify/assert"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func getDBURL() string {
	url := os.Getenv("DB_URL")
	if url == "" {
		return "postgres://admin:secret@localhost:5432/autorepair?sslmode=disable"
	}
	return url
}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func TestPostgresPartRepository(t *testing.T) {
	database, err := db.New(getDBURL())
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}
	defer database.Close()

	repo := infrastructure.NewPostgresPartRepository(database.Pool)

	// Create
	partName := "Test Part " + randomString(5)
	part, _ := inventoryDomain.NewPart(partName, "Desc", 50, 100.0) // name, desc, qty, price
	err = repo.Save(context.Background(), part)
	assert.NoError(t, err)

	// GetByID
	fetched, err := repo.GetByID(context.Background(), part.ID)
	assert.NoError(t, err)
	if fetched != nil {
		assert.Equal(t, part.Name, fetched.Name)
	}

	// List
	list, err := repo.List(context.Background())
	assert.NoError(t, err)
	assert.NotEmpty(t, list)

	// Decrease Stock (using domain method and update)
	fetched.RemoveStock(10)
	err = repo.Update(context.Background(), fetched)
	assert.NoError(t, err)

	fetchedAfter, _ := repo.GetByID(context.Background(), part.ID)
	assert.Equal(t, 40, fetchedAfter.Quantity)

	// Insufficient Stock
	err = fetchedAfter.RemoveStock(100)
	assert.ErrorIs(t, err, inventoryDomain.ErrInsufficientStock)
}
