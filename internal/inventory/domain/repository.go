package domain

import (
	"context"

	"github.com/google/uuid"
)

type PartRepository interface {
	Save(ctx context.Context, part *Part) error
	GetByID(ctx context.Context, id uuid.UUID) (*Part, error)
	Update(ctx context.Context, part *Part) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*Part, error)
}
