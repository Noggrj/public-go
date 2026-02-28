package infrastructure

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/noggrj/autorepair/internal/inventory/domain"
	"github.com/noggrj/autorepair/internal/platform/db"
)

type PostgresPartRepository struct {
	db db.Connection
}

func NewPostgresPartRepository(db db.Connection) *PostgresPartRepository {
	return &PostgresPartRepository{db: db}
}

func (r *PostgresPartRepository) Save(ctx context.Context, part *domain.Part) error {
	query := `INSERT INTO parts (id, name, description, stock_qty, price) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query, part.ID, part.Name, part.Description, part.Quantity, part.Price)
	return err
}

func (r *PostgresPartRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Part, error) {
	query := `SELECT id, name, description, stock_qty, price FROM parts WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)

	var part domain.Part
	err := row.Scan(&part.ID, &part.Name, &part.Description, &part.Quantity, &part.Price)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("part not found")
		}
		return nil, err
	}
	return &part, nil
}

func (r *PostgresPartRepository) Update(ctx context.Context, part *domain.Part) error {
	query := `UPDATE parts SET name = $1, description = $2, stock_qty = $3, price = $4 WHERE id = $5`
	_, err := r.db.Exec(ctx, query, part.Name, part.Description, part.Quantity, part.Price, part.ID)
	return err
}

func (r *PostgresPartRepository) List(ctx context.Context) ([]*domain.Part, error) {
	query := `SELECT id, name, description, stock_qty, price FROM parts`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parts []*domain.Part
	for rows.Next() {
		var p domain.Part
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Quantity, &p.Price); err != nil {
			return nil, err
		}
		parts = append(parts, &p)
	}
	return parts, nil
}

func (r *PostgresPartRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM parts WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("part not found")
	}
	return nil
}
