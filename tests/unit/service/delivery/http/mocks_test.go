package http_test

import (
	"context"

	"github.com/google/uuid"
	inventoryDomain "github.com/noggrj/autorepair/internal/inventory/domain"
	serviceDomain "github.com/noggrj/autorepair/internal/service/domain"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Save(order *serviceDomain.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockOrderRepository) GetByID(id uuid.UUID) (*serviceDomain.Order, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*serviceDomain.Order), args.Error(1)
}

func (m *MockOrderRepository) List() ([]*serviceDomain.Order, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*serviceDomain.Order), args.Error(1)
}

func (m *MockOrderRepository) ListActive() ([]*serviceDomain.Order, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*serviceDomain.Order), args.Error(1)
}

type MockPartRepository struct {
	mock.Mock
}

func (m *MockPartRepository) Save(ctx context.Context, part *inventoryDomain.Part) error {
	args := m.Called(ctx, part)
	return args.Error(0)
}

func (m *MockPartRepository) GetByID(ctx context.Context, id uuid.UUID) (*inventoryDomain.Part, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*inventoryDomain.Part), args.Error(1)
}

func (m *MockPartRepository) Update(ctx context.Context, part *inventoryDomain.Part) error {
	args := m.Called(ctx, part)
	return args.Error(0)
}

func (m *MockPartRepository) List(ctx context.Context) ([]*inventoryDomain.Part, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*inventoryDomain.Part), args.Error(1)
}

func (m *MockPartRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Removed DecreaseStock as it's not in the interface anymore.

type MockServiceRepository struct {
	mock.Mock
}

func (m *MockServiceRepository) Save(service *serviceDomain.Service) error {
	args := m.Called(service)
	return args.Error(0)
}

func (m *MockServiceRepository) GetByID(id uuid.UUID) (*serviceDomain.Service, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*serviceDomain.Service), args.Error(1)
}

func (m *MockServiceRepository) List() ([]*serviceDomain.Service, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*serviceDomain.Service), args.Error(1)
}

func (m *MockServiceRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

type MockClientRepository struct {
	mock.Mock
}

func (m *MockClientRepository) Save(client *serviceDomain.Client) error {
	args := m.Called(client)
	return args.Error(0)
}

func (m *MockClientRepository) GetByID(id uuid.UUID) (*serviceDomain.Client, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*serviceDomain.Client), args.Error(1)
}

func (m *MockClientRepository) List() ([]*serviceDomain.Client, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*serviceDomain.Client), args.Error(1)
}

func (m *MockClientRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockClientRepository) Update(client *serviceDomain.Client) error {
	args := m.Called(client)
	return args.Error(0)
}

type MockVehicleRepository struct {
	mock.Mock
}

func (m *MockVehicleRepository) Save(vehicle *serviceDomain.Vehicle) error {
	args := m.Called(vehicle)
	return args.Error(0)
}

func (m *MockVehicleRepository) GetByID(id uuid.UUID) (*serviceDomain.Vehicle, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*serviceDomain.Vehicle), args.Error(1)
}

func (m *MockVehicleRepository) ListByClientID(clientID uuid.UUID) ([]*serviceDomain.Vehicle, error) {
	args := m.Called(clientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*serviceDomain.Vehicle), args.Error(1)
}

func (m *MockVehicleRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

type MockNotifier struct {
	mock.Mock
}

func (m *MockNotifier) SendEmail(to, subject, body string) error {
	args := m.Called(to, subject, body)
	return args.Error(0)
}
