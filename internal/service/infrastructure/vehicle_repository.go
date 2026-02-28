package infrastructure

// Repositories use pgxpool for database connection

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/noggrj/autorepair/internal/platform/db"
	"github.com/noggrj/autorepair/internal/service/domain"
	"github.com/noggrj/autorepair/internal/sharedkernel"
)

type PostgresVehicleRepository struct {
	db db.Connection
}

func NewPostgresVehicleRepository(db db.Connection) *PostgresVehicleRepository {
	return &PostgresVehicleRepository{db: db}
}

func (r *PostgresVehicleRepository) Save(vehicle *domain.Vehicle) error {
	query := `INSERT INTO vehicles (id, client_id, plate, brand, model, year, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			  ON CONFLICT (id) DO UPDATE SET
			  plate = EXCLUDED.plate,
			  brand = EXCLUDED.brand,
			  model = EXCLUDED.model,
			  year = EXCLUDED.year,
			  updated_at = EXCLUDED.updated_at`
	_, err := r.db.Exec(context.Background(), query,
		vehicle.ID, vehicle.ClientID, vehicle.Plate.String(), vehicle.Brand, vehicle.Model, vehicle.Year, vehicle.CreatedAt, vehicle.UpdatedAt)
	return err
}

func (r *PostgresVehicleRepository) GetByID(id uuid.UUID) (*domain.Vehicle, error) {
	query := `SELECT id, client_id, plate, brand, model, year, created_at, updated_at FROM vehicles WHERE id = $1`
	row := r.db.QueryRow(context.Background(), query, id)
	return scanVehicle(row)
}

func (r *PostgresVehicleRepository) ListByClientID(clientID uuid.UUID) ([]*domain.Vehicle, error) {
	query := `SELECT id, client_id, plate, brand, model, year, created_at, updated_at FROM vehicles WHERE client_id = $1`
	rows, err := r.db.Query(context.Background(), query, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicles []*domain.Vehicle
	for rows.Next() {
		v, err := scanVehicle(rows)
		if err != nil {
			return nil, err
		}
		vehicles = append(vehicles, v)
	}
	return vehicles, nil
}

func (r *PostgresVehicleRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM vehicles WHERE id = $1`
	result, err := r.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrVehicleNotFound
	}
	return nil
}

func scanVehicle(row pgx.Row) (*domain.Vehicle, error) {
	var v domain.Vehicle
	var plateStr string
	err := row.Scan(&v.ID, &v.ClientID, &plateStr, &v.Brand, &v.Model, &v.Year, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVehicleNotFound
		}
		return nil, err
	}
	plate, err := sharedkernel.NewPlacaBR(plateStr)
	if err != nil {
		return nil, err
	}
	v.Plate = plate
	return &v, nil
}
