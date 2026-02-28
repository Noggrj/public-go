package application_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	inventoryDomain "github.com/noggrj/autorepair/internal/inventory/domain"
	"github.com/noggrj/autorepair/internal/service/application"
	serviceDomain "github.com/noggrj/autorepair/internal/service/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOrderService_SendBudget_Errors(t *testing.T) {
	t.Run("GetByID Error", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		service := application.NewOrderService(mockOrderRepo, nil, nil, nil)
		orderID := uuid.New()
		mockOrderRepo.On("GetByID", orderID).Return(nil, errors.New("repo error"))

		err := service.SendBudget(orderID)
		assert.Error(t, err)
		assert.Equal(t, "repo error", err.Error())
	})

	t.Run("Wrong Status", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		service := application.NewOrderService(mockOrderRepo, nil, nil, nil)
		orderID := uuid.New()
		order := &serviceDomain.Order{ID: orderID, Status: serviceDomain.OrderStatusReceived}
		mockOrderRepo.On("GetByID", orderID).Return(order, nil)

		err := service.SendBudget(orderID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can only be sent from 'In diagnosis' status")
	})

	t.Run("Save Error", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		service := application.NewOrderService(mockOrderRepo, nil, nil, nil)
		orderID := uuid.New()
		order := &serviceDomain.Order{ID: orderID, Status: serviceDomain.OrderStatusInDiagnosis}
		mockOrderRepo.On("GetByID", orderID).Return(order, nil)
		mockOrderRepo.On("Save", mock.Anything).Return(errors.New("save error"))

		err := service.SendBudget(orderID)
		assert.Error(t, err)
		assert.Equal(t, "save error", err.Error())
	})

	t.Run("Client Repo Error (Should Log and Continue)", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		mockClientRepo := new(MockClientRepository)
		service := application.NewOrderService(mockOrderRepo, nil, mockClientRepo, nil)
		orderID := uuid.New()
		clientID := uuid.New()
		order := &serviceDomain.Order{ID: orderID, ClientID: clientID, Status: serviceDomain.OrderStatusInDiagnosis}

		mockOrderRepo.On("GetByID", orderID).Return(order, nil)
		mockOrderRepo.On("Save", mock.Anything).Return(nil)
		mockClientRepo.On("GetByID", clientID).Return(nil, errors.New("client error"))

		err := service.SendBudget(orderID)
		assert.NoError(t, err)
		// Status should still be updated
		assert.Equal(t, serviceDomain.OrderStatusAwaitingApproval, order.Status)
	})

	t.Run("Notifier Error (Should Log and Continue)", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		mockClientRepo := new(MockClientRepository)
		mockNotifier := new(MockNotifier)
		service := application.NewOrderService(mockOrderRepo, nil, mockClientRepo, mockNotifier)
		orderID := uuid.New()
		clientID := uuid.New()
		order := &serviceDomain.Order{ID: orderID, ClientID: clientID, Status: serviceDomain.OrderStatusInDiagnosis}
		client := &serviceDomain.Client{ID: clientID, Email: "test@test.com"}

		mockOrderRepo.On("GetByID", orderID).Return(order, nil)
		mockOrderRepo.On("Save", mock.Anything).Return(nil)
		mockClientRepo.On("GetByID", clientID).Return(client, nil)
		mockNotifier.On("SendEmail", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("email error"))

		err := service.SendBudget(orderID)
		assert.NoError(t, err)
		assert.Equal(t, serviceDomain.OrderStatusAwaitingApproval, order.Status)
	})
}

func TestOrderService_ApproveOrder_Errors(t *testing.T) {
	t.Run("GetByID Error", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		service := application.NewOrderService(mockOrderRepo, nil, nil, nil)
		orderID := uuid.New()
		mockOrderRepo.On("GetByID", orderID).Return(nil, errors.New("repo error"))

		err := service.ApproveOrder(orderID)
		assert.Error(t, err)
	})

	t.Run("Wrong Status", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		service := application.NewOrderService(mockOrderRepo, nil, nil, nil)
		orderID := uuid.New()
		order := &serviceDomain.Order{ID: orderID, Status: serviceDomain.OrderStatusInExecution}
		mockOrderRepo.On("GetByID", orderID).Return(order, nil)

		err := service.ApproveOrder(orderID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can only be approved from")
	})

	t.Run("Part GetByID Error", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		mockPartRepo := new(MockPartRepository)
		service := application.NewOrderService(mockOrderRepo, mockPartRepo, nil, nil)
		orderID := uuid.New()
		partID := uuid.New()
		order := &serviceDomain.Order{
			ID:     orderID,
			Status: serviceDomain.OrderStatusReceived,
			Items:  []*serviceDomain.OrderItem{{RefID: partID, Type: serviceDomain.ItemTypePart, Quantity: 1}},
		}

		mockOrderRepo.On("GetByID", orderID).Return(order, nil)
		mockPartRepo.On("GetByID", mock.Anything, partID).Return(nil, errors.New("part error"))

		err := service.ApproveOrder(orderID)
		assert.Error(t, err)
	})

	t.Run("Part RemoveStock Error (Not reachable via normal flow if logic is correct, but let's test insufficient stock again or other logic)", func(t *testing.T) {
		// Insufficient stock is already tested in existing file
	})

	t.Run("Part Update Error", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		mockPartRepo := new(MockPartRepository)
		service := application.NewOrderService(mockOrderRepo, mockPartRepo, nil, nil)
		orderID := uuid.New()
		partID := uuid.New()
		order := &serviceDomain.Order{
			ID:     orderID,
			Status: serviceDomain.OrderStatusReceived,
			Items:  []*serviceDomain.OrderItem{{RefID: partID, Type: serviceDomain.ItemTypePart, Quantity: 1}},
		}
		part := &inventoryDomain.Part{ID: partID, Quantity: 10}

		mockOrderRepo.On("GetByID", orderID).Return(order, nil)
		mockPartRepo.On("GetByID", mock.Anything, partID).Return(part, nil)
		mockPartRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("update error"))

		err := service.ApproveOrder(orderID)
		assert.Error(t, err)
		assert.Equal(t, "update error", err.Error())
	})

	t.Run("Save Error", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		service := application.NewOrderService(mockOrderRepo, nil, nil, nil)
		orderID := uuid.New()
		order := &serviceDomain.Order{ID: orderID, Status: serviceDomain.OrderStatusReceived}

		mockOrderRepo.On("GetByID", orderID).Return(order, nil)
		mockOrderRepo.On("Save", mock.Anything).Return(errors.New("save error"))

		err := service.ApproveOrder(orderID)
		assert.Error(t, err)
	})
}

func TestOrderService_FinishOrder_Errors(t *testing.T) {
	t.Run("GetByID Error", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		service := application.NewOrderService(mockOrderRepo, nil, nil, nil)
		orderID := uuid.New()
		mockOrderRepo.On("GetByID", orderID).Return(nil, errors.New("repo error"))
		err := service.FinishOrder(orderID)
		assert.Error(t, err)
	})

	t.Run("Wrong Status", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		service := application.NewOrderService(mockOrderRepo, nil, nil, nil)
		orderID := uuid.New()
		order := &serviceDomain.Order{ID: orderID, Status: serviceDomain.OrderStatusReceived}
		mockOrderRepo.On("GetByID", orderID).Return(order, nil)
		err := service.FinishOrder(orderID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can only be finished from 'In execution'")
	})

	t.Run("Save Error", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		service := application.NewOrderService(mockOrderRepo, nil, nil, nil)
		orderID := uuid.New()
		order := &serviceDomain.Order{ID: orderID, Status: serviceDomain.OrderStatusInExecution}
		mockOrderRepo.On("GetByID", orderID).Return(order, nil)
		mockOrderRepo.On("Save", mock.Anything).Return(errors.New("save error"))
		err := service.FinishOrder(orderID)
		assert.Error(t, err)
	})
}

func TestOrderService_DeliverOrder_Errors(t *testing.T) {
	t.Run("GetByID Error", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		service := application.NewOrderService(mockOrderRepo, nil, nil, nil)
		orderID := uuid.New()
		mockOrderRepo.On("GetByID", orderID).Return(nil, errors.New("repo error"))
		err := service.DeliverOrder(orderID)
		assert.Error(t, err)
	})

	t.Run("Wrong Status", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		service := application.NewOrderService(mockOrderRepo, nil, nil, nil)
		orderID := uuid.New()
		order := &serviceDomain.Order{ID: orderID, Status: serviceDomain.OrderStatusReceived}
		mockOrderRepo.On("GetByID", orderID).Return(order, nil)
		err := service.DeliverOrder(orderID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can only be delivered from 'Completed'")
	})

	t.Run("Save Error", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		service := application.NewOrderService(mockOrderRepo, nil, nil, nil)
		orderID := uuid.New()
		order := &serviceDomain.Order{ID: orderID, Status: serviceDomain.OrderStatusCompleted}
		mockOrderRepo.On("GetByID", orderID).Return(order, nil)
		mockOrderRepo.On("Save", mock.Anything).Return(errors.New("save error"))
		err := service.DeliverOrder(orderID)
		assert.Error(t, err)
	})
}

func TestOrderService_UpdateStatus_Errors(t *testing.T) {
	t.Run("GetByID Error", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		service := application.NewOrderService(mockOrderRepo, nil, nil, nil)
		orderID := uuid.New()
		mockOrderRepo.On("GetByID", orderID).Return(nil, errors.New("repo error"))
		err := service.UpdateStatus(orderID, serviceDomain.OrderStatusCompleted)
		assert.Error(t, err)
	})

	t.Run("Save Error", func(t *testing.T) {
		mockOrderRepo := new(MockOrderRepository)
		service := application.NewOrderService(mockOrderRepo, nil, nil, nil)
		orderID := uuid.New()
		order := &serviceDomain.Order{ID: orderID, Status: serviceDomain.OrderStatusReceived}
		mockOrderRepo.On("GetByID", orderID).Return(order, nil)
		mockOrderRepo.On("Save", mock.Anything).Return(errors.New("save error"))
		err := service.UpdateStatus(orderID, serviceDomain.OrderStatusCompleted)
		assert.Error(t, err)
	})
}
