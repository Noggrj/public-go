package application_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	inventoryDomain "github.com/noggrj/autorepair/internal/inventory/domain"
	"github.com/noggrj/autorepair/internal/service/application"
	serviceDomain "github.com/noggrj/autorepair/internal/service/domain"
	"github.com/stretchr/testify/assert"
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
	return args.Get(0).([]*inventoryDomain.Part), args.Error(1)
}

func (m *MockPartRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Removed DecreaseStock as it's not in interface anymore.

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
	return args.Get(0).([]*serviceDomain.Client), args.Error(1)
}

func (m *MockClientRepository) Delete(id uuid.UUID) error {
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

// --- Tests ---

func TestOrderService_StartDiagnosis(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockClientRepo := new(MockClientRepository)
	mockNotifier := new(MockNotifier)
	service := application.NewOrderService(mockOrderRepo, nil, mockClientRepo, mockNotifier)

	orderID := uuid.New()
	clientID := uuid.New()
	order := &serviceDomain.Order{
		ID:       orderID,
		ClientID: clientID,
		Status:   serviceDomain.OrderStatusReceived,
	}
	client := &serviceDomain.Client{ID: clientID, Email: "test@test.com"}

	mockOrderRepo.On("GetByID", orderID).Return(order, nil)
	mockOrderRepo.On("Save", mock.MatchedBy(func(o *serviceDomain.Order) bool {
		return o.Status == serviceDomain.OrderStatusInDiagnosis
	})).Return(nil)
	mockClientRepo.On("GetByID", clientID).Return(client, nil)
	mockNotifier.On("SendEmail", "test@test.com", mock.Anything, mock.Anything).Return(nil)

	err := service.StartDiagnosis(orderID)
	assert.NoError(t, err)
	mockOrderRepo.AssertExpectations(t)
}

func TestOrderService_SendBudget(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockClientRepo := new(MockClientRepository)
	mockNotifier := new(MockNotifier)
	service := application.NewOrderService(mockOrderRepo, nil, mockClientRepo, mockNotifier)

	orderID := uuid.New()
	clientID := uuid.New()
	order := &serviceDomain.Order{
		ID:       orderID,
		ClientID: clientID,
		Status:   serviceDomain.OrderStatusInDiagnosis,
		Total:    100.0,
	}
	client := &serviceDomain.Client{ID: clientID, Email: "test@test.com"}

	mockOrderRepo.On("GetByID", orderID).Return(order, nil)
	mockOrderRepo.On("Save", mock.MatchedBy(func(o *serviceDomain.Order) bool {
		return o.Status == serviceDomain.OrderStatusAwaitingApproval
	})).Return(nil)
	mockClientRepo.On("GetByID", clientID).Return(client, nil)
	mockNotifier.On("SendEmail", "test@test.com", mock.Anything, mock.Anything).Return(nil)

	err := service.SendBudget(orderID)
	assert.NoError(t, err)
	mockOrderRepo.AssertExpectations(t)
	mockNotifier.AssertExpectations(t)
}

func TestOrderService_ApproveOrder(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockPartRepo := new(MockPartRepository)
	mockClientRepo := new(MockClientRepository)
	mockNotifier := new(MockNotifier)
	service := application.NewOrderService(mockOrderRepo, mockPartRepo, mockClientRepo, mockNotifier)

	orderID := uuid.New()
	clientID := uuid.New()
	partID := uuid.New()

	order := &serviceDomain.Order{
		ID:       orderID,
		ClientID: clientID,
		Status:   serviceDomain.OrderStatusAwaitingApproval,
		Items: []*serviceDomain.OrderItem{
			{RefID: partID, Type: serviceDomain.ItemTypePart, Quantity: 2},
		},
	}
	client := &serviceDomain.Client{ID: clientID, Email: "test@test.com"}

	mockOrderRepo.On("GetByID", orderID).Return(order, nil)

	// Mock Part GetByID and Update (since DecreaseStock is now logical op on entity)
	part := &inventoryDomain.Part{ID: partID, Quantity: 10} // Enough stock
	mockPartRepo.On("GetByID", mock.Anything, partID).Return(part, nil)
	mockPartRepo.On("Update", mock.Anything, mock.MatchedBy(func(p *inventoryDomain.Part) bool {
		return p.ID == partID && p.Quantity == 8 // 10 - 2
	})).Return(nil)

	mockOrderRepo.On("Save", mock.MatchedBy(func(o *serviceDomain.Order) bool {
		return o.Status == serviceDomain.OrderStatusInExecution && o.StartedAt != nil
	})).Return(nil)
	mockClientRepo.On("GetByID", clientID).Return(client, nil)
	mockNotifier.On("SendEmail", "test@test.com", mock.Anything, mock.Anything).Return(nil)

	err := service.ApproveOrder(orderID)
	assert.NoError(t, err)
	mockOrderRepo.AssertExpectations(t)
	mockPartRepo.AssertExpectations(t)
}

func TestOrderService_ApproveOrder_InsufficientStock(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockPartRepo := new(MockPartRepository)
	service := application.NewOrderService(mockOrderRepo, mockPartRepo, nil, nil)

	orderID := uuid.New()
	partID := uuid.New()

	order := &serviceDomain.Order{
		ID:     orderID,
		Status: serviceDomain.OrderStatusReceived,
		Items: []*serviceDomain.OrderItem{
			{RefID: partID, Type: serviceDomain.ItemTypePart, Quantity: 10},
		},
	}

	mockOrderRepo.On("GetByID", orderID).Return(order, nil)

	// Mock Part GetByID returning low stock
	part := &inventoryDomain.Part{ID: partID, Quantity: 5} // Less than 10
	mockPartRepo.On("GetByID", mock.Anything, partID).Return(part, nil)

	err := service.ApproveOrder(orderID)
	assert.ErrorIs(t, err, inventoryDomain.ErrInsufficientStock)
	mockOrderRepo.AssertExpectations(t)
}

func TestOrderService_FinishOrder(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockClientRepo := new(MockClientRepository)
	mockNotifier := new(MockNotifier)
	service := application.NewOrderService(mockOrderRepo, nil, mockClientRepo, mockNotifier)

	orderID := uuid.New()
	clientID := uuid.New()
	order := &serviceDomain.Order{
		ID:       orderID,
		ClientID: clientID,
		Status:   serviceDomain.OrderStatusInExecution,
	}
	client := &serviceDomain.Client{ID: clientID, Email: "test@test.com"}

	mockOrderRepo.On("GetByID", orderID).Return(order, nil)
	mockOrderRepo.On("Save", mock.MatchedBy(func(o *serviceDomain.Order) bool {
		return o.Status == serviceDomain.OrderStatusCompleted && o.FinishedAt != nil
	})).Return(nil)
	mockClientRepo.On("GetByID", clientID).Return(client, nil)
	mockNotifier.On("SendEmail", "test@test.com", mock.Anything, mock.Anything).Return(nil)

	err := service.FinishOrder(orderID)
	assert.NoError(t, err)
}

func TestOrderService_DeliverOrder(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockClientRepo := new(MockClientRepository)
	mockNotifier := new(MockNotifier)
	service := application.NewOrderService(mockOrderRepo, nil, mockClientRepo, mockNotifier)

	orderID := uuid.New()
	clientID := uuid.New()
	order := &serviceDomain.Order{
		ID:       orderID,
		ClientID: clientID,
		Status:   serviceDomain.OrderStatusCompleted,
	}
	client := &serviceDomain.Client{ID: clientID, Email: "test@test.com"}

	mockOrderRepo.On("GetByID", orderID).Return(order, nil)
	mockOrderRepo.On("Save", mock.MatchedBy(func(o *serviceDomain.Order) bool {
		return o.Status == serviceDomain.OrderStatusDelivered
	})).Return(nil)
	mockClientRepo.On("GetByID", clientID).Return(client, nil)
	mockNotifier.On("SendEmail", "test@test.com", mock.Anything, mock.Anything).Return(nil)

	err := service.DeliverOrder(orderID)
	assert.NoError(t, err)
}

func TestOrderService_UpdateStatus(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockClientRepo := new(MockClientRepository)
	mockNotifier := new(MockNotifier)
	service := application.NewOrderService(mockOrderRepo, nil, mockClientRepo, mockNotifier)

	orderID := uuid.New()
	clientID := uuid.New()
	order := &serviceDomain.Order{ID: orderID, ClientID: clientID, Status: serviceDomain.OrderStatusReceived}
	client := &serviceDomain.Client{ID: clientID, Email: "test@test.com"}

	mockOrderRepo.On("GetByID", orderID).Return(order, nil)
	mockOrderRepo.On("Save", mock.MatchedBy(func(o *serviceDomain.Order) bool {
		return o.Status == serviceDomain.OrderStatusInExecution
	})).Return(nil)
	mockClientRepo.On("GetByID", clientID).Return(client, nil)
	mockNotifier.On("SendEmail", "test@test.com", mock.Anything, mock.Anything).Return(nil)

	err := service.UpdateStatus(orderID, serviceDomain.OrderStatusInExecution)
	assert.NoError(t, err)
}

func TestOrderService_StartDiagnosis_WrongStatus(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	service := application.NewOrderService(mockOrderRepo, nil, nil, nil)

	orderID := uuid.New()
	order := &serviceDomain.Order{ID: orderID, Status: serviceDomain.OrderStatusCompleted}

	mockOrderRepo.On("GetByID", orderID).Return(order, nil)

	err := service.StartDiagnosis(orderID)
	assert.Error(t, err)
	assert.Equal(t, "diagnosis can only be started from 'Received' status", err.Error())
}

// Add more edge cases as needed...
