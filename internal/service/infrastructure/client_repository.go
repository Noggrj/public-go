package infrastructure

import (
	"context"
	"errors"

	"github.com/noggrj/autorepair/internal/service/domain"
	"github.com/noggrj/autorepair/internal/sharedkernel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/noggrj/autorepair/internal/platform/db"
)

type PostgresClientRepository struct {
	db db.Connection
}

func NewPostgresClientRepository(db db.Connection) *PostgresClientRepository {
	return &PostgresClientRepository{db: db}
}

func (r *PostgresClientRepository) Save(client *domain.Client) error {
	query := `INSERT INTO clients (id, name, document, email, phone, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7)
	          ON CONFLICT (id) DO UPDATE SET
	          name = EXCLUDED.name,
	          document = EXCLUDED.document,
	          email = EXCLUDED.email,
	          phone = EXCLUDED.phone,
	          updated_at = EXCLUDED.updated_at`
	_, err := r.db.Exec(context.Background(), query,
		client.ID, client.Name, client.Document.String(), client.Email, client.Phone, client.CreatedAt, client.UpdatedAt)
	return err
}

func (r *PostgresClientRepository) GetByID(id uuid.UUID) (*domain.Client, error) {
	query := `SELECT id, name, document, email, phone, created_at, updated_at FROM clients WHERE id = $1`
	row := r.db.QueryRow(context.Background(), query, id)
	return scanClient(row)
}

func (r *PostgresClientRepository) List() ([]*domain.Client, error) {
	query := `SELECT id, name, document, email, phone, created_at, updated_at FROM clients`
	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []*domain.Client
	for rows.Next() {
		client, err := scanClient(rows)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}
	return clients, nil
}

func (r *PostgresClientRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM clients WHERE id = $1`
	result, err := r.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrClientNotFound
	}
	return nil
}

func scanClient(row pgx.Row) (*domain.Client, error) {
	var client domain.Client
	var docStr string
	err := row.Scan(&client.ID, &client.Name, &docStr, &client.Email, &client.Phone, &client.CreatedAt, &client.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrClientNotFound
		}
		return nil, err
	}
	doc, err := sharedkernel.NewDocumentoBR(docStr)
	if err != nil {
		return nil, err // Should not happen if DB is consistent
	}
	client.Document = doc
	return &client, nil
}
