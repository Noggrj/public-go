package infrastructure_test

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/noggrj/autorepair/internal/service/domain"
	"github.com/noggrj/autorepair/internal/service/infrastructure"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestPostgresServiceRepository_Save(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	repo := infrastructure.NewPostgresServiceRepository(mock)
	service, _ := domain.NewService("Oil Change", "Desc", 100.0)

	// Success
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO services`)).
		WithArgs(service.ID, service.Name, service.Description, float64(service.Price), service.CreatedAt, service.UpdatedAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.Save(service)
	assert.NoError(t, err)

	// Error
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO services`)).
		WillReturnError(errors.New("db error"))

	err = repo.Save(service)
	assert.Error(t, err)
}

func TestPostgresServiceRepository_GetByID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	repo := infrastructure.NewPostgresServiceRepository(mock)
	id := uuid.New()
	now := time.Now()

	// Success
	rows := pgxmock.NewRows([]string{"id", "name", "description", "price", "created_at", "updated_at"}).
		AddRow(id, "Oil Change", "Desc", 100.0, now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, description, price, created_at, updated_at FROM services WHERE id = $1`)).
		WithArgs(id).
		WillReturnRows(rows)

	service, err := repo.GetByID(id)
	assert.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, id, service.ID)

	// Not Found
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WithArgs(id).
		WillReturnError(pgx.ErrNoRows)

	_, err = repo.GetByID(id)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrServiceNotFound, err)

	// DB Error
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WillReturnError(errors.New("connection error"))

	_, err = repo.GetByID(id)
	assert.Error(t, err)
}

func TestPostgresServiceRepository_List(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	repo := infrastructure.NewPostgresServiceRepository(mock)
	now := time.Now()

	// Success
	rows := pgxmock.NewRows([]string{"id", "name", "description", "price", "created_at", "updated_at"}).
		AddRow(uuid.New(), "S1", "D1", 100.0, now, now).
		AddRow(uuid.New(), "S2", "D2", 200.0, now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, description, price, created_at, updated_at FROM services`)).
		WillReturnRows(rows)

	services, err := repo.List()
	assert.NoError(t, err)
	assert.Len(t, services, 2)

	// Query Error
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WillReturnError(errors.New("db error"))

	_, err = repo.List()
	assert.Error(t, err)

	// Scan Error
	rowsScanErr := pgxmock.NewRows([]string{"id", "name", "description", "price", "created_at", "updated_at"}).
		AddRow(uuid.New(), "S1", "D1", "invalid-price", now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WillReturnRows(rowsScanErr)

	_, err = repo.List()
	assert.Error(t, err)
}
