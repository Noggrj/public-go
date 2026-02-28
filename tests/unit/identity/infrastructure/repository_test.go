package infrastructure_test

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/noggrj/autorepair/internal/identity/domain"
	"github.com/noggrj/autorepair/internal/identity/infrastructure"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestPostgresUserRepository_Save(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	repo := infrastructure.NewPostgresUserRepository(mock)
	user, _ := domain.NewUser("John Doe", "john@example.com", "password", domain.RoleAdmin)

	// Success
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO users (id, name, email, password_hash, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`)).
		WithArgs(user.ID, user.Name, user.Email, user.Password, user.Role, user.CreatedAt, user.UpdatedAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.Save(user)
	assert.NoError(t, err)

	// Error
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO users`)).
		WillReturnError(errors.New("db error"))

	err = repo.Save(user)
	assert.Error(t, err)
}

func TestPostgresUserRepository_GetByEmail(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	repo := infrastructure.NewPostgresUserRepository(mock)
	email := "john@example.com"
	id := uuid.New()
	now := time.Now()

	// Success
	rows := pgxmock.NewRows([]string{"id", "name", "email", "password_hash", "role", "created_at", "updated_at"}).
		AddRow(id, "John Doe", email, "hashed_pass", "admin", now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, email, password_hash, role, created_at, updated_at FROM users WHERE email = $1`)).
		WithArgs(email).
		WillReturnRows(rows)

	user, err := repo.GetByEmail(email)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, domain.RoleAdmin, user.Role)

	// Not Found
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WithArgs(email).
		WillReturnError(pgx.ErrNoRows)

	_, err = repo.GetByEmail(email)
	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())

	// DB Error
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WillReturnError(errors.New("connection error"))

	_, err = repo.GetByEmail(email)
	assert.Error(t, err)

	// Scan Error
	rowsScanErr := pgxmock.NewRows([]string{"id", "name", "email", "password_hash", "role", "created_at", "updated_at"}).
		AddRow(id, "John Doe", email, "hashed_pass", "admin", "invalid-time", now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WithArgs(email).
		WillReturnRows(rowsScanErr)

	_, err = repo.GetByEmail(email)
	assert.Error(t, err)
}

func TestPostgresUserRepository_GetByID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	repo := infrastructure.NewPostgresUserRepository(mock)
	id := uuid.New()
	email := "john@example.com"
	now := time.Now()

	// Success
	rows := pgxmock.NewRows([]string{"id", "name", "email", "password_hash", "role", "created_at", "updated_at"}).
		AddRow(id, "John Doe", email, "hashed_pass", "admin", now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, email, password_hash, role, created_at, updated_at FROM users WHERE id = $1`)).
		WithArgs(id).
		WillReturnRows(rows)

	user, err := repo.GetByID(id)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, id, user.ID)

	// Not Found
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WithArgs(id).
		WillReturnError(pgx.ErrNoRows)

	_, err = repo.GetByID(id)
	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())

	// DB Error
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WillReturnError(errors.New("connection error"))

	_, err = repo.GetByID(id)
	assert.Error(t, err)

	// Scan Error
	rowsScanErr := pgxmock.NewRows([]string{"id", "name", "email", "password_hash", "role", "created_at", "updated_at"}).
		AddRow(id, "John Doe", email, "hashed_pass", "admin", "invalid-time", now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WithArgs(id).
		WillReturnRows(rowsScanErr)

	_, err = repo.GetByID(id)
	assert.Error(t, err)
}
