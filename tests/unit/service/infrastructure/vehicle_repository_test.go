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

func TestPostgresVehicleRepository_Save(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	repo := infrastructure.NewPostgresVehicleRepository(mock)
	clientID := uuid.New()
	vehicle, _ := domain.NewVehicle(clientID, "ABC1234", "Ford", "Fiesta", 2020)

	// Success
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO vehicles`)).
		WithArgs(vehicle.ID, vehicle.ClientID, vehicle.Plate.String(), vehicle.Brand, vehicle.Model, vehicle.Year, vehicle.CreatedAt, vehicle.UpdatedAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.Save(vehicle)
	assert.NoError(t, err)

	// Error
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO vehicles`)).
		WillReturnError(errors.New("db error"))

	err = repo.Save(vehicle)
	assert.Error(t, err)
}

func TestPostgresVehicleRepository_GetByID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	repo := infrastructure.NewPostgresVehicleRepository(mock)
	id := uuid.New()
	clientID := uuid.New()
	now := time.Now()

	// Success
	rows := pgxmock.NewRows([]string{"id", "client_id", "plate", "brand", "model", "year", "created_at", "updated_at"}).
		AddRow(id, clientID, "ABC-1234", "Ford", "Fiesta", 2020, now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, client_id, plate, brand, model, year, created_at, updated_at FROM vehicles WHERE id = $1`)).
		WithArgs(id).
		WillReturnRows(rows)

	vehicle, err := repo.GetByID(id)
	assert.NoError(t, err)
	assert.NotNil(t, vehicle)
	assert.Equal(t, id, vehicle.ID)

	// Not Found
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WithArgs(id).
		WillReturnError(pgx.ErrNoRows)

	_, err = repo.GetByID(id)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrVehicleNotFound, err)

	// DB Error
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WillReturnError(errors.New("connection error"))

	_, err = repo.GetByID(id)
	assert.Error(t, err)

	// Scan Error (Invalid Plate)
	rowsScanErr := pgxmock.NewRows([]string{"id", "client_id", "plate", "brand", "model", "year", "created_at", "updated_at"}).
		AddRow(id, clientID, "invalid", "Ford", "Fiesta", 2020, now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WithArgs(id).
		WillReturnRows(rowsScanErr)

	_, err = repo.GetByID(id)
	assert.Error(t, err)
}

func TestPostgresVehicleRepository_ListByClientID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	repo := infrastructure.NewPostgresVehicleRepository(mock)
	clientID := uuid.New()
	now := time.Now()

	// Success
	rows := pgxmock.NewRows([]string{"id", "client_id", "plate", "brand", "model", "year", "created_at", "updated_at"}).
		AddRow(uuid.New(), clientID, "ABC-1234", "Ford", "Fiesta", 2020, now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, client_id, plate, brand, model, year, created_at, updated_at FROM vehicles WHERE client_id = $1`)).
		WithArgs(clientID).
		WillReturnRows(rows)

	vehicles, err := repo.ListByClientID(clientID)
	assert.NoError(t, err)
	assert.Len(t, vehicles, 1)

	// Query Error
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WillReturnError(errors.New("db error"))

	_, err = repo.ListByClientID(clientID)
	assert.Error(t, err)

	// Scan Error
	rowsScanErr := pgxmock.NewRows([]string{"id", "client_id", "plate", "brand", "model", "year", "created_at", "updated_at"}).
		AddRow(uuid.New(), clientID, "invalid", "Ford", "Fiesta", 2020, now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
		WillReturnRows(rowsScanErr)

	_, err = repo.ListByClientID(clientID)
	assert.Error(t, err)
}
