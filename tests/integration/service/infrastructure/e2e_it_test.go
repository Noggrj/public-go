//go:build integration

package infrastructure_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	inventoryDomain "github.com/noggrj/autorepair/internal/inventory/domain"
	inventoryInfra "github.com/noggrj/autorepair/internal/inventory/infrastructure"
	notificationInfra "github.com/noggrj/autorepair/internal/notification/infrastructure"
	"github.com/noggrj/autorepair/internal/platform/db"
	serviceApplication "github.com/noggrj/autorepair/internal/service/application"
	serviceDomain "github.com/noggrj/autorepair/internal/service/domain"
	"github.com/noggrj/autorepair/internal/service/infrastructure"
	"github.com/noggrj/autorepair/internal/sharedkernel"
	"github.com/stretchr/testify/mock"
)

func TestFullOrderLifecycleE2E(t *testing.T) {
	// This test simulates a full flow from Order Creation -> Approval -> Notification -> Revenue Reporting

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		t.Skip("DB_URL not set")
	}

	pool, err := db.New(dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}
	defer pool.Close()

	// Repos
	orderRepo := infrastructure.NewPostgresOrderRepository(pool.Pool)
	partRepo := inventoryInfra.NewPostgresPartRepository(pool.Pool)
	clientRepo := infrastructure.NewPostgresClientRepository(pool.Pool)
	vehicleRepo := infrastructure.NewPostgresVehicleRepository(pool.Pool)
	serviceRepo := infrastructure.NewPostgresServiceRepository(pool.Pool)

	// Services
	notifier := &notificationInfra.MockEmailService{}
	notifier.On("SendEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	orderService := serviceApplication.NewOrderService(orderRepo, partRepo, clientRepo, notifier)

	// 1. Setup Data
	// Client
	clientID := uuid.New()
	client, _ := serviceDomain.NewClient("E2E Client", "12345678901", "e2e@test.com", "123")
	docSuffix := time.Now().Format("0405000000")
	doc, _ := sharedkernel.NewDocumentoBR("3" + docSuffix)
	client.Document = doc
	client.ID = clientID
	clientRepo.Save(client)

	// Vehicle
	vehicleID := uuid.New()
	vehicle, _ := serviceDomain.NewVehicle(clientID, "ABC1234", "Toyota", "Corolla", 2022)
	plateSuffix := time.Now().Format("0405.000000")[5:9]
	plate, _ := sharedkernel.NewPlacaBR("EZE" + plateSuffix)
	vehicle.Plate = plate
	vehicle.ID = vehicleID
	vehicleRepo.Save(vehicle)

	// Service
	svcID := uuid.New()
	svc, _ := serviceDomain.NewService("E2E Service", "Desc", 100.0)
	svc.ID = svcID
	serviceRepo.Save(svc)

	// Part
	partID := uuid.New()
	part, _ := inventoryDomain.NewPart("E2E Part", "Desc", 10, 50.0) // qty, price
	part.ID = partID
	partRepo.Save(context.Background(), part)

	// 2. Create Order
	order, _ := serviceDomain.NewOrder(clientID, vehicleID)
	order.AddItem(svcID, serviceDomain.ItemTypeService, svc.Name, 1, float64(svc.Price))
	order.AddItem(partID, serviceDomain.ItemTypePart, part.Name, 2, float64(part.Price)) // 2 * 50 = 100
	// Total should be 100 + 100 = 200

	if err := orderRepo.Save(order); err != nil {
		t.Fatalf("Failed to save order: %v", err)
	}

	// 3. Approve Order
	err = orderService.ApproveOrder(order.ID)
	if err != nil {
		t.Fatalf("Failed to approve order: %v", err)
	}

	// 4. Verify Status and Stock
	updatedOrder, _ := orderRepo.GetByID(order.ID)
	if updatedOrder.Status != serviceDomain.OrderStatusInExecution {
		t.Errorf("Order status mismatch: %v", updatedOrder.Status)
	}

	updatedPart, _ := partRepo.GetByID(context.Background(), partID)
	if updatedPart.Quantity != 8 { // 10 - 2
		t.Errorf("Stock mismatch: expected 8, got %d", updatedPart.Quantity)
	}

	// 5. Update Status to Completed
	err = orderService.UpdateStatus(order.ID, serviceDomain.OrderStatusCompleted)
	if err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	// 6. Verify Reporting (Revenue)
	// We can reuse the List() method on repo to simulate reporting aggregation
	orders, _ := orderRepo.List()
	var totalRevenue float64
	var found bool
	for _, o := range orders {
		totalRevenue += float64(o.Total)
		if o.ID == order.ID {
			found = true
		}
	}

	if !found {
		t.Error("Created order not found in list")
	}
	if totalRevenue < 200.0 {
		t.Errorf("Total revenue too low: %f", totalRevenue)
	}
}
