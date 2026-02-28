package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/noggrj/autorepair/internal/service/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	// Valid
	c, err := domain.NewClient("Name", "12345678909", "test@test.com", "123")
	assert.NoError(t, err)
	assert.NotNil(t, c)

	// Invalid Document
	_, err = domain.NewClient("Name", "invalid", "test@test.com", "123")
	assert.Error(t, err)

	// Invalid Name
	_, err = domain.NewClient("", "12345678909", "test@test.com", "123")
	assert.Error(t, err)
}

func TestNewVehicle(t *testing.T) {
	clientID := uuid.New()
	// Valid
	v, err := domain.NewVehicle(clientID, "ABC1234", "Ford", "Fiesta", 2020)
	assert.NoError(t, err)
	assert.NotNil(t, v)

	// Invalid Plate
	_, err = domain.NewVehicle(clientID, "invalid", "Ford", "Fiesta", 2020)
	assert.Error(t, err)

	// Invalid Brand/Model
	_, err = domain.NewVehicle(clientID, "ABC1234", "", "Fiesta", 2020)
	assert.Error(t, err)
	_, err = domain.NewVehicle(clientID, "ABC1234", "Ford", "", 2020)
	assert.Error(t, err)

	// Invalid ClientID
	_, err = domain.NewVehicle(uuid.Nil, "ABC1234", "Ford", "Fiesta", 2020)
	assert.Error(t, err)

	// Invalid Year
	_, err = domain.NewVehicle(clientID, "ABC1234", "Ford", "Fiesta", 1800)
	assert.Error(t, err)
}

func TestNewService(t *testing.T) {
	// Valid
	s, err := domain.NewService("Oil Change", "Desc", 100.0)
	assert.NoError(t, err)
	assert.NotNil(t, s)

	// Invalid Price
	_, err = domain.NewService("Oil Change", "Desc", -10.0)
	assert.Error(t, err)

	// Invalid Name
	_, err = domain.NewService("", "Desc", 100.0)
	assert.Error(t, err)
}

func TestNewOrder(t *testing.T) {
	clientID := uuid.New()
	vehicleID := uuid.New()

	// Valid
	o, err := domain.NewOrder(clientID, vehicleID)
	assert.NoError(t, err)
	assert.NotNil(t, o)

	// Invalid IDs
	_, err = domain.NewOrder(uuid.Nil, vehicleID)
	assert.Error(t, err)

	_, err = domain.NewOrder(clientID, uuid.Nil)
	assert.Error(t, err)
}

func TestOrder_AddItem(t *testing.T) {
	o, _ := domain.NewOrder(uuid.New(), uuid.New())
	refID := uuid.New()

	// Valid
	err := o.AddItem(refID, domain.ItemTypeService, "Service", 1, 100.0)
	assert.NoError(t, err)

	// Invalid Qty
	err = o.AddItem(refID, domain.ItemTypeService, "Service", 0, 100.0)
	assert.Error(t, err)

	// Invalid Price
	err = o.AddItem(refID, domain.ItemTypeService, "Service", 1, -10.0)
	assert.Error(t, err)
}
