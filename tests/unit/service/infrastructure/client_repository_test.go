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

func TestPostgresClientRepository_Save(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	repo := infrastructure.NewPostgresClientRepository(mock)
	client, _ := domain.NewClient("John Doe", "12345678909", "john@example.com", "123456789")

	// Success
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO clients`)).
		WithArgs(client.ID, client.Name, client.Document.String(), client.Email, client.Phone, client.CreatedAt, client.UpdatedAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.Save(client)
	assert.NoError(t, err)

	// Error
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO clients`)).
		WillReturnError(errors.New("db error"))

	err = repo.Save(client)
	assert.Error(t, err)
}

func TestPostgresClientRepository_GetByID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	repo := infrastructure.NewPostgresClientRepository(mock)
	id := uuid.New()
	now := time.Now()

	// Success
	rows := pgxmock.NewRows([]string{"id", "name", "document", "email", "phone", "created_at", "updated_at"}).
		AddRow(id, "John Doe", "12345678909", "john@example.com", "123456789", now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, document, email, phone, created_at, updated_at FROM clients WHERE id = $1`)).
		WithArgs(id).
		WillReturnRows(rows)

	client, err := repo.GetByID(id)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, id, client.ID)

	// Not Found
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WithArgs(id).
		WillReturnError(pgx.ErrNoRows)

	_, err = repo.GetByID(id)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrClientNotFound, err)

	// DB Error
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WillReturnError(errors.New("connection error"))

	_, err = repo.GetByID(id)
	assert.Error(t, err)

	// Scan Error (Invalid Document format in DB)
	// Although NewDocumentoBR validates, if DB has invalid data (e.g. legacy), scan might fail if we were validating there.
	// But scanClient calls NewDocumentoBR which validates length.
	// So let's return a string that fails NewDocumentoBR validation.
	rowsScanErr := pgxmock.NewRows([]string{"id", "name", "document", "email", "phone", "created_at", "updated_at"}).
		AddRow(id, "John Doe", "invalid", "john@example.com", "123456789", now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WithArgs(id).
		WillReturnRows(rowsScanErr)

	_, err = repo.GetByID(id)
	assert.Error(t, err)
}

func TestPostgresClientRepository_List(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	repo := infrastructure.NewPostgresClientRepository(mock)
	now := time.Now()

	// Success
	rows := pgxmock.NewRows([]string{"id", "name", "document", "email", "phone", "created_at", "updated_at"}).
		AddRow(uuid.New(), "C1", "12345678909", "e1@e.com", "123", now, now).
		AddRow(uuid.New(), "C2", "98765432100", "e2@e.com", "456", now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, document, email, phone, created_at, updated_at FROM clients`)).
		WillReturnRows(rows)

	clients, err := repo.List()
	assert.NoError(t, err)
	assert.Len(t, clients, 2)

	// Query Error
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WillReturnError(errors.New("db error"))

	_, err = repo.List()
	assert.Error(t, err)

	// Scan Error
	rowsScanErr := pgxmock.NewRows([]string{"id", "name", "document", "email", "phone", "created_at", "updated_at"}).
		AddRow(uuid.New(), "C1", "invalid", "e1@e.com", "123", now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WillReturnRows(rowsScanErr)

	_, err = repo.List()
	assert.Error(t, err)
}
