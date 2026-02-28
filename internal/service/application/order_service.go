package application

import (
	"errors"
	"fmt"
	"time"

	"context"

	"github.com/google/uuid"
	inventoryDomain "github.com/noggrj/autorepair/internal/inventory/domain"
	notificationDomain "github.com/noggrj/autorepair/internal/notification/domain"
	serviceDomain "github.com/noggrj/autorepair/internal/service/domain"
)

type OrderService struct {
	orderRepo  serviceDomain.OrderRepository
	partRepo   inventoryDomain.PartRepository
	clientRepo serviceDomain.ClientRepository
	notifier   notificationDomain.EmailService
}

func NewOrderService(
	orderRepo serviceDomain.OrderRepository,
	partRepo inventoryDomain.PartRepository,
	clientRepo serviceDomain.ClientRepository,
	notifier notificationDomain.EmailService,
) *OrderService {
	return &OrderService{
		orderRepo:  orderRepo,
		partRepo:   partRepo,
		clientRepo: clientRepo,
		notifier:   notifier,
	}
}

func (s *OrderService) StartDiagnosis(orderID uuid.UUID) error {
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return err
	}

	if order.Status != serviceDomain.OrderStatusReceived {
		return errors.New("diagnosis can only be started from 'Received' status")
	}

	order.Status = serviceDomain.OrderStatusInDiagnosis
	if err := s.orderRepo.Save(order); err != nil {
		return err
	}
	s.notifyStatusChange(order)
	return nil
}

func (s *OrderService) SendBudget(orderID uuid.UUID) error {
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return err
	}

	if order.Status != serviceDomain.OrderStatusInDiagnosis {
		return errors.New("budget can only be sent from 'In diagnosis' status")
	}

	order.Status = serviceDomain.OrderStatusAwaitingApproval
	if err := s.orderRepo.Save(order); err != nil {
		return err
	}

	// Notify specifically about budget
	client, err := s.clientRepo.GetByID(order.ClientID)
	if err != nil {
		// Log error but continue with order status update
		return nil
	}
	if err := s.notifier.SendEmail(client.Email, "Order Budget Ready", fmt.Sprintf("Your budget for order %s is ready. Total: %.2f", order.ID, order.Total)); err != nil {
		// Log error but continue with order status update
		_ = err // ignore error
	}

	return nil
}

func (s *OrderService) ApproveOrder(orderID uuid.UUID) error {
	// 1. Get Order
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return err
	}

	// 2. Validate Status
	// Allow approval from Received (direct) or Awaiting Approval (flow)
	if order.Status != serviceDomain.OrderStatusReceived && order.Status != serviceDomain.OrderStatusAwaitingApproval {
		return errors.New("order can only be approved from 'Received' or 'Awaiting approval' status")
	}

	// 3. Process Parts (Decrease Stock)
	for _, item := range order.Items {
		if item.Type == serviceDomain.ItemTypePart {
			// Get Part to check stock and decrease it
			part, err := s.partRepo.GetByID(context.Background(), item.RefID)
			if err != nil {
				return err
			}

			if err := part.RemoveStock(item.Quantity); err != nil {
				return err
			}

			if err := s.partRepo.Update(context.Background(), part); err != nil {
				return err
			}
		}
	}

	// 4. Update Status
	order.Status = serviceDomain.OrderStatusInExecution
	now := time.Now()
	order.StartedAt = &now

	// 5. Save
	if err := s.orderRepo.Save(order); err != nil {
		return err
	}

	// 6. Notify
	s.notifyStatusChange(order)

	return nil
}

func (s *OrderService) FinishOrder(orderID uuid.UUID) error {
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return err
	}

	if order.Status != serviceDomain.OrderStatusInExecution {
		return errors.New("order can only be finished from 'In execution' status")
	}

	order.Status = serviceDomain.OrderStatusCompleted // "Finished" in requirements
	now := time.Now()
	order.FinishedAt = &now

	if err := s.orderRepo.Save(order); err != nil {
		return err
	}
	s.notifyStatusChange(order)
	return nil
}

func (s *OrderService) DeliverOrder(orderID uuid.UUID) error {
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return err
	}

	if order.Status != serviceDomain.OrderStatusCompleted {
		return errors.New("order can only be delivered from 'Completed' status")
	}

	order.Status = serviceDomain.OrderStatusDelivered
	if err := s.orderRepo.Save(order); err != nil {
		return err
	}
	s.notifyStatusChange(order)
	return nil
}

func (s *OrderService) UpdateStatus(orderID uuid.UUID, status serviceDomain.OrderStatus) error {
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return err
	}

	order.Status = status
	if err := s.orderRepo.Save(order); err != nil {
		return err
	}

	s.notifyStatusChange(order)

	return nil
}

func (s *OrderService) RejectBudget(orderID uuid.UUID) error {
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return err
	}

	if order.Status != serviceDomain.OrderStatusAwaitingApproval {
		return errors.New("budget can only be rejected from 'Awaiting approval' status")
	}

	order.Status = serviceDomain.OrderStatusReceived
	order.UpdatedAt = time.Now()

	if err := s.orderRepo.Save(order); err != nil {
		return err
	}

	// Notify client about rejection
	client, err := s.clientRepo.GetByID(order.ClientID)
	if err != nil {
		return nil
	}
	subject := "Order Budget Rejected"
	body := fmt.Sprintf("Hello %s, the budget for order %s has been rejected. The order has been returned to Received status.", client.Name, order.ID)
	if err := s.notifier.SendEmail(client.Email, subject, body); err != nil {
		_ = err // Log but don't fail
	}

	return nil
}

func (s *OrderService) notifyStatusChange(order *serviceDomain.Order) {
	client, err := s.clientRepo.GetByID(order.ClientID)
	if err != nil {
		// Log error but don't fail flow
		return
	}

	subject := fmt.Sprintf("Order Update: %s", order.Status)
	body := fmt.Sprintf("Hello %s, your order %s status has been updated to: %s", client.Name, order.ID, order.Status)

	if err := s.notifier.SendEmail(client.Email, subject, body); err != nil {
		// Log error but continue
		_ = err // ignore error
	}
}
