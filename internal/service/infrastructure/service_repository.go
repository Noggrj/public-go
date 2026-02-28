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

type PostgresServiceRepository struct {
	db db.Connection
}

func NewPostgresServiceRepository(db db.Connection) *PostgresServiceRepository {
	return &PostgresServiceRepository{db: db}
}

func (r *PostgresServiceRepository) Save(service *domain.Service) error {
	query := `INSERT INTO services (id, name, description, price, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6)
	          ON CONFLICT (id) DO UPDATE SET
	          name = EXCLUDED.name,
	          description = EXCLUDED.description,
	          price = EXCLUDED.price,
	          updated_at = EXCLUDED.updated_at`
	_, err := r.db.Exec(context.Background(), query,
		service.ID, service.Name, service.Description, float64(service.Price), service.CreatedAt, service.UpdatedAt)
	return err
}

func (r *PostgresServiceRepository) GetByID(id uuid.UUID) (*domain.Service, error) {
	query := `SELECT id, name, description, price, created_at, updated_at FROM services WHERE id = $1`
	row := r.db.QueryRow(context.Background(), query, id)
	return scanService(row)
}

func (r *PostgresServiceRepository) List() ([]*domain.Service, error) {
	query := `SELECT id, name, description, price, created_at, updated_at FROM services`
	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []*domain.Service
	for rows.Next() {
		s, err := scanService(rows)
		if err != nil {
			return nil, err
		}
		services = append(services, s)
	}
	return services, nil
}

func (r *PostgresServiceRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM services WHERE id = $1`
	result, err := r.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrServiceNotFound
	}
	return nil
}

func scanService(row pgx.Row) (*domain.Service, error) {
	var s domain.Service
	var price float64
	err := row.Scan(&s.ID, &s.Name, &s.Description, &price, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrServiceNotFound
		}
		return nil, err
	}
	s.Price = sharedkernel.Money(price)
	return &s, nil
}
