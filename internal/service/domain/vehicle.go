package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/noggrj/autorepair/internal/sharedkernel"
)

var (
	ErrVehicleNotFound = errors.New("vehicle not found")
)

type Vehicle struct {
	ID        uuid.UUID
	ClientID  uuid.UUID
	Plate     sharedkernel.PlacaBR
	Brand     string
	Model     string
	Year      int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewVehicle(clientID uuid.UUID, plate, brand, model string, year int) (*Vehicle, error) {
	if clientID == uuid.Nil {
		return nil, errors.New("client id is required")
	}
	if brand == "" || model == "" {
		return nil, errors.New("brand and model are required")
	}
	if year < 1900 || year > time.Now().Year()+1 {
		return nil, errors.New("invalid year")
	}

	p, err := sharedkernel.NewPlacaBR(plate)
	if err != nil {
		return nil, err
	}

	return &Vehicle{
		ID:        uuid.New(),
		ClientID:  clientID,
		Plate:     p,
		Brand:     brand,
		Model:     model,
		Year:      year,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

type VehicleRepository interface {
	Save(vehicle *Vehicle) error
	GetByID(id uuid.UUID) (*Vehicle, error)
	ListByClientID(clientID uuid.UUID) ([]*Vehicle, error)
	Delete(id uuid.UUID) error
}
