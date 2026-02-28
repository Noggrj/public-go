package application

import (
	"context"
	"testing"

	"github.com/google/uuid"
	inventoryDomain "github.com/noggrj/autorepair/internal/inventory/domain"
	serviceDomain "github.com/noggrj/autorepair/internal/service/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks
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

func (m *MockPartRepository) List(ctx context.Context) ([]*inventoryDomain.Part, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*inventoryDomain.Part), args.Error(1)
}

func (m *MockPartRepository) Update(ctx context.Context, part *inventoryDomain.Part) error {
	args := m.Called(ctx, part)
	return args.Error(0)
}

func (m *MockPartRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
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

func (m *MockClientRepository) GetByCPF(cpf string) (*serviceDomain.Client, error) {
	args := m.Called(cpf)
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

type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendEmail(to, subject, body string) error {
	args := m.Called(to, subject, body)
	return args.Error(0)
}

func TestStartDiagnosis(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockPartRepo := new(MockPartRepository)
	mockClientRepo := new(MockClientRepository)
	mockNotifier := new(MockEmailService)

	service := NewOrderService(mockOrderRepo, mockPartRepo, mockClientRepo, mockNotifier)

	orderID := uuid.New()
	clientID := uuid.New()
	order := &serviceDomain.Order{
		ID:       orderID,
		ClientID: clientID,
		Status:   serviceDomain.OrderStatusReceived,
	}

	client := &serviceDomain.Client{
		ID:    clientID,
		Name:  "Test Client",
		Email: "client@test.com",
	}

	t.Run("Success", func(t *testing.T) {
		mockOrderRepo.On("GetByID", orderID).Return(order, nil)
		mockOrderRepo.On("Save", mock.Anything).Return(nil)
		mockClientRepo.On("GetByID", clientID).Return(client, nil)
		mockNotifier.On("SendEmail", client.Email, mock.Anything, mock.Anything).Return(nil)

		err := service.StartDiagnosis(orderID)

		assert.NoError(t, err)
		assert.Equal(t, serviceDomain.OrderStatusInDiagnosis, order.Status)
	})
}

func TestApproveOrder(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockPartRepo := new(MockPartRepository)
	mockClientRepo := new(MockClientRepository)
	mockNotifier := new(MockEmailService)

	service := NewOrderService(mockOrderRepo, mockPartRepo, mockClientRepo, mockNotifier)

	orderID := uuid.New()
	clientID := uuid.New()
	partID := uuid.New()

	order := &serviceDomain.Order{
		ID:       orderID,
		ClientID: clientID,
		Status:   serviceDomain.OrderStatusAwaitingApproval,
		Items: []*serviceDomain.OrderItem{
			{
				Type:     serviceDomain.ItemTypePart,
				RefID:    partID,
				Quantity: 2,
			},
		},
	}

	part := &inventoryDomain.Part{
		ID:       partID,
		Quantity: 10,
	}

	client := &serviceDomain.Client{
		ID:    clientID,
		Name:  "Test Client",
		Email: "client@test.com",
	}

	t.Run("Success", func(t *testing.T) {
		mockOrderRepo.On("GetByID", orderID).Return(order, nil)
		mockPartRepo.On("GetByID", context.Background(), partID).Return(part, nil)
		mockPartRepo.On("Update", context.Background(), part).Return(nil)
		mockOrderRepo.On("Save", mock.Anything).Return(nil)
		mockClientRepo.On("GetByID", clientID).Return(client, nil)
		mockNotifier.On("SendEmail", client.Email, mock.Anything, mock.Anything).Return(nil)

		err := service.ApproveOrder(orderID)

		assert.NoError(t, err)
		assert.Equal(t, serviceDomain.OrderStatusInExecution, order.Status)
		assert.Equal(t, 8, part.Quantity)
	})
}

func TestSendBudget(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockPartRepo := new(MockPartRepository)
	mockClientRepo := new(MockClientRepository)
	mockNotifier := new(MockEmailService)

	service := NewOrderService(mockOrderRepo, mockPartRepo, mockClientRepo, mockNotifier)

	orderID := uuid.New()
	clientID := uuid.New()
	order := &serviceDomain.Order{
		ID:       orderID,
		ClientID: clientID,
		Status:   serviceDomain.OrderStatusInDiagnosis,
		Total:    100.0,
	}

	client := &serviceDomain.Client{
		ID:    clientID,
		Name:  "Test Client",
		Email: "client@test.com",
	}

	t.Run("Success", func(t *testing.T) {
		mockOrderRepo.On("GetByID", orderID).Return(order, nil)
		mockOrderRepo.On("Save", mock.Anything).Return(nil)
		mockClientRepo.On("GetByID", clientID).Return(client, nil)
		mockNotifier.On("SendEmail", client.Email, mock.Anything, mock.Anything).Return(nil)

		err := service.SendBudget(orderID)

		assert.NoError(t, err)
		assert.Equal(t, serviceDomain.OrderStatusAwaitingApproval, order.Status)
	})
}

func TestFinishOrder(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockPartRepo := new(MockPartRepository)
	mockClientRepo := new(MockClientRepository)
	mockNotifier := new(MockEmailService)

	service := NewOrderService(mockOrderRepo, mockPartRepo, mockClientRepo, mockNotifier)

	orderID := uuid.New()
	clientID := uuid.New()
	order := &serviceDomain.Order{
		ID:       orderID,
		ClientID: clientID,
		Status:   serviceDomain.OrderStatusInExecution,
	}

	client := &serviceDomain.Client{
		ID:    clientID,
		Name:  "Test Client",
		Email: "client@test.com",
	}

	t.Run("Success", func(t *testing.T) {
		mockOrderRepo.On("GetByID", orderID).Return(order, nil)
		mockOrderRepo.On("Save", mock.Anything).Return(nil)
		mockClientRepo.On("GetByID", clientID).Return(client, nil)
		mockNotifier.On("SendEmail", client.Email, mock.Anything, mock.Anything).Return(nil)

		err := service.FinishOrder(orderID)

		assert.NoError(t, err)
		assert.Equal(t, serviceDomain.OrderStatusCompleted, order.Status)
		assert.NotNil(t, order.FinishedAt)
	})
}

func TestDeliverOrder(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockPartRepo := new(MockPartRepository)
	mockClientRepo := new(MockClientRepository)
	mockNotifier := new(MockEmailService)

	service := NewOrderService(mockOrderRepo, mockPartRepo, mockClientRepo, mockNotifier)

	orderID := uuid.New()
	clientID := uuid.New()
	order := &serviceDomain.Order{
		ID:       orderID,
		ClientID: clientID,
		Status:   serviceDomain.OrderStatusCompleted,
	}

	client := &serviceDomain.Client{
		ID:    clientID,
		Name:  "Test Client",
		Email: "client@test.com",
	}

	t.Run("Success", func(t *testing.T) {
		mockOrderRepo.On("GetByID", orderID).Return(order, nil)
		mockOrderRepo.On("Save", mock.Anything).Return(nil)
		mockClientRepo.On("GetByID", clientID).Return(client, nil)
		mockNotifier.On("SendEmail", client.Email, mock.Anything, mock.Anything).Return(nil)

		err := service.DeliverOrder(orderID)

		assert.NoError(t, err)
		assert.Equal(t, serviceDomain.OrderStatusDelivered, order.Status)
	})
}
