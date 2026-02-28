package domain_test

import (
	"testing"

	"github.com/noggrj/autorepair/internal/inventory/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewPart(t *testing.T) {
	p, err := domain.NewPart("Tire", "Rubber tire", 10, 100.0) // name, desc, qty, price
	assert.NoError(t, err)
	assert.Equal(t, "Tire", p.Name)
	assert.Equal(t, 10, p.Quantity)

	// Invalid
	// Assuming NewPart validates negative quantity or price
	_, err = domain.NewPart("Name", "Desc", -5, 10.0)
	assert.Error(t, err)

	_, err = domain.NewPart("Name", "Desc", 5, -10.0)
	assert.Error(t, err)
}

func TestPart_RemoveStock(t *testing.T) {
	p, _ := domain.NewPart("Tire", "Desc", 10, 100.0)

	err := p.RemoveStock(5)
	assert.NoError(t, err)
	assert.Equal(t, 5, p.Quantity)
	
	err = p.RemoveStock(20) // More than stock
	assert.ErrorIs(t, err, domain.ErrInsufficientStock)
}
