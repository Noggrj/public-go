package domain

import (
	"errors"

	"github.com/google/uuid"
)

var ErrInsufficientStock = errors.New("insufficient stock")

type Part struct {
	ID          uuid.UUID
	Name        string
	Description string
	Quantity    int
	Price       float64
}

func NewPart(name, description string, quantity int, price float64) (*Part, error) {
	if quantity < 0 {
		return nil, errors.New("quantity cannot be negative")
	}
	if price < 0 {
		return nil, errors.New("price cannot be negative")
	}

	return &Part{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Quantity:    quantity,
		Price:       price,
	}, nil
}

func (p *Part) RemoveStock(qty int) error {
	if p.Quantity < qty {
		return ErrInsufficientStock
	}
	p.Quantity -= qty
	return nil
}

func (p *Part) AddStock(qty int) {
	p.Quantity += qty
}
