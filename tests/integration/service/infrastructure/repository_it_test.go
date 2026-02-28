//go:build integration

package infrastructure_test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/noggrj/autorepair/internal/platform/db"
	serviceDomain "github.com/noggrj/autorepair/internal/service/domain"
	"github.com/noggrj/autorepair/internal/service/infrastructure"
	"github.com/stretchr/testify/assert"
)

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func randomString(n int) string {
	var letters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func randomDigit(n int) string {
	var digits = []rune("0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = digits[rand.Intn(len(digits))]
	}
	return string(b)
}

func generateValidCPF() string {
	d := make([]int, 9)
	for i := 0; i < 9; i++ {
		d[i] = rand.Intn(10)
	}

	sum := 0
	for i := 0; i < 9; i++ {
		sum += d[i] * (10 - i)
	}
	rem := sum % 11
	d1 := 0
	if rem >= 2 {
		d1 = 11 - rem
	}

	sum = 0
	for i := 0; i < 9; i++ {
		sum += d[i] * (11 - i)
	}
	sum += d1 * 2
	rem = sum % 11
	d2 := 0
	if rem >= 2 {
		d2 = 11 - rem
	}

	return fmt.Sprintf("%d%d%d%d%d%d%d%d%d%d%d", d[0], d[1], d[2], d[3], d[4], d[5], d[6], d[7], d[8], d1, d2)
}

func getDBURL() string {
	url := os.Getenv("DB_URL")
	if url == "" {
		return "postgres://admin:secret@localhost:5432/autorepair?sslmode=disable"
	}
	return url
}

func TestPostgresServiceRepository(t *testing.T) {
	database, err := db.New(getDBURL())
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}
	defer database.Close()

	repo := infrastructure.NewPostgresServiceRepository(database.Pool)

	// Create
	svcName := "Test Service " + randomString(5)
	svc, _ := serviceDomain.NewService(svcName, "Desc", 150.0)
	err = repo.Save(svc)
	assert.NoError(t, err)

	// GetByID
	fetched, err := repo.GetByID(svc.ID)
	assert.NoError(t, err)
	if fetched != nil {
		assert.Equal(t, svc.Name, fetched.Name)
	}

	// List
	list, err := repo.List()
	assert.NoError(t, err)
	assert.NotEmpty(t, list)

	found := false
	for _, s := range list {
		if s.ID == svc.ID {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestPostgresVehicleRepository(t *testing.T) {
	database, err := db.New(getDBURL())
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}
	defer database.Close()

	// Need a client first
	clientRepo := infrastructure.NewPostgresClientRepository(database.Pool)
	clientEmail := "v" + randomString(5) + "@test.com"
	clientDoc := generateValidCPF()
	client, err := serviceDomain.NewClient("Vehicle Owner", clientDoc, clientEmail, "123")
	assert.NoError(t, err)

	err = clientRepo.Save(client)
	assert.NoError(t, err)

	repo := infrastructure.NewPostgresVehicleRepository(database.Pool)

	// Create Unique Plate: AAA1111 (Old format)
	plate := randomString(3) + randomDigit(4)

	v, err := serviceDomain.NewVehicle(client.ID, plate, "Ford", "Fiesta", 2019)
	assert.NoError(t, err)

	err = repo.Save(v)
	assert.NoError(t, err)

	// GetByID
	fetched, err := repo.GetByID(v.ID)
	assert.NoError(t, err)
	if fetched != nil {
		assert.Equal(t, v.Plate.String(), fetched.Plate.String())
	}

	// ListByClientID
	list, err := repo.ListByClientID(client.ID)
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	if len(list) > 0 {
		assert.Equal(t, v.ID, list[0].ID)
	}
}

func TestPostgresClientRepository_Full(t *testing.T) {
	database, err := db.New(getDBURL())
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}
	defer database.Close()

	repo := infrastructure.NewPostgresClientRepository(database.Pool)

	// Create
	email := "list" + randomString(5) + "@test.com"
	doc := generateValidCPF()
	c, _ := serviceDomain.NewClient("List Test", doc, email, "123")
	err = repo.Save(c)
	assert.NoError(t, err)

	// List
	list, err := repo.List()
	assert.NoError(t, err)
	assert.NotEmpty(t, list)
}
