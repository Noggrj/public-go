package infrastructure

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/noggrj/autorepair/internal/identity/domain"
	"github.com/noggrj/autorepair/internal/platform/db"
)

type PostgresUserRepository struct {
	db db.Connection
}

func NewPostgresUserRepository(db db.Connection) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Save(user *domain.User) error {
	query := `INSERT INTO users (id, name, email, password_hash, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(context.Background(), query, user.ID, user.Name, user.Email, user.Password, user.Role, user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *PostgresUserRepository) GetByEmail(email string) (*domain.User, error) {
	query := `SELECT id, name, email, password_hash, role, created_at, updated_at FROM users WHERE email = $1`
	row := r.db.QueryRow(context.Background(), query, email)

	var user domain.User
	var role string
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	user.Role = domain.Role(role)
	return &user, nil
}

func (r *PostgresUserRepository) GetByID(id uuid.UUID) (*domain.User, error) {
	query := `SELECT id, name, email, password_hash, role, created_at, updated_at FROM users WHERE id = $1`
	row := r.db.QueryRow(context.Background(), query, id)

	var user domain.User
	var role string
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	user.Role = domain.Role(role)
	return &user, nil
}
