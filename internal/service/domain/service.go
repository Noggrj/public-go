package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/noggrj/autorepair/internal/sharedkernel"
)

var (
	ErrServiceNotFound = errors.New("service not found")
)

type Service struct {
	ID          uuid.UUID
	Name        string
	Description string
	Price       sharedkernel.Money
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewService(name, description string, price float64) (*Service, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	if price < 0 {
		return nil, errors.New("price cannot be negative")
	}

	return &Service{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Price:       sharedkernel.Money(price),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

type ServiceRepository interface {
	Save(service *Service) error
	GetByID(id uuid.UUID) (*Service, error)
	List() ([]*Service, error)
	Delete(id uuid.UUID) error
}
