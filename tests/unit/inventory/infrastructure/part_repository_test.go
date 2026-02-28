package infrastructure_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/noggrj/autorepair/internal/inventory/domain"
	"github.com/noggrj/autorepair/internal/inventory/infrastructure"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestPostgresPartRepository_Save(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	repo := infrastructure.NewPostgresPartRepository(mock)
	part, _ := domain.NewPart("Test Part", "Description", 10, 100.0)

	// Success
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO parts (id, name, description, stock_qty, price) VALUES ($1, $2, $3, $4, $5)`)).
		WithArgs(part.ID, part.Name, part.Description, part.Quantity, part.Price).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.Save(context.Background(), part)
	assert.NoError(t, err)

	// Error
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO parts`)).
		WillReturnError(errors.New("db error"))

	err = repo.Save(context.Background(), part)
	assert.Error(t, err)
}

func TestPostgresPartRepository_GetByID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	repo := infrastructure.NewPostgresPartRepository(mock)
	id := uuid.New()

	// Success
	rows := pgxmock.NewRows([]string{"id", "name", "description", "stock_qty", "price"}).
		AddRow(id, "Part 1", "Desc 1", 5, 50.0)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, description, stock_qty, price FROM parts WHERE id = $1`)).
		WithArgs(id).
		WillReturnRows(rows)

	part, err := repo.GetByID(context.Background(), id)
	assert.NoError(t, err)
	assert.NotNil(t, part)
	assert.Equal(t, id, part.ID)

	// Not Found
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WithArgs(id).
		WillReturnError(pgx.ErrNoRows)

	_, err = repo.GetByID(context.Background(), id)
	assert.Error(t, err)
	assert.Equal(t, "part not found", err.Error())

	// DB Error
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WillReturnError(errors.New("connection error"))

	_, err = repo.GetByID(context.Background(), id)
	assert.Error(t, err)
}

func TestPostgresPartRepository_Update(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	repo := infrastructure.NewPostgresPartRepository(mock)
	part, _ := domain.NewPart("Updated", "Desc", 20, 200.0)

	// Success
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE parts SET name = $1, description = $2, stock_qty = $3, price = $4 WHERE id = $5`)).
		WithArgs(part.Name, part.Description, part.Quantity, part.Price, part.ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = repo.Update(context.Background(), part)
	assert.NoError(t, err)

	// Error
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE parts`)).
		WillReturnError(errors.New("db error"))

	err = repo.Update(context.Background(), part)
	assert.Error(t, err)
}

func TestPostgresPartRepository_List(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	repo := infrastructure.NewPostgresPartRepository(mock)

	// Success
	rows := pgxmock.NewRows([]string{"id", "name", "description", "stock_qty", "price"}).
		AddRow(uuid.New(), "Part 1", "Desc 1", 10, 100.0).
		AddRow(uuid.New(), "Part 2", "Desc 2", 20, 200.0)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, description, stock_qty, price FROM parts`)).
		WillReturnRows(rows)

	parts, err := repo.List(context.Background())
	assert.NoError(t, err)
	assert.Len(t, parts, 2)

	// Query Error
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WillReturnError(errors.New("db error"))

	_, err = repo.List(context.Background())
	assert.Error(t, err)

	// Scan Error
	rowsScanErr := pgxmock.NewRows([]string{"id", "name", "description", "stock_qty", "price"}).
		AddRow(uuid.New(), "Part 1", "Desc 1", "invalid", 100.0)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WillReturnRows(rowsScanErr)

	_, err = repo.List(context.Background())
	assert.Error(t, err)
}
