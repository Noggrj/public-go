//go:build integration

package infrastructure_test

import (
	"context"
	"errors"
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

func TestOrderApprovalFlow(t *testing.T) {
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

	// App Service
	notifier := &notificationInfra.MockEmailService{}
	notifier.On("SendEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	orderService := serviceApplication.NewOrderService(orderRepo, partRepo, clientRepo, notifier)

	// 1. Create Dependencies
	clientID := uuid.New()
	client, err := serviceDomain.NewClient("Test Client 2", "12345678901", "test2@test.com", "123")
	// Ensure unique doc
	randomDoc := "2" + time.Now().Format("0405000000")
	doc, _ := sharedkernel.NewDocumentoBR(randomDoc)
	client.Document = doc
	client.ID = clientID
	if err := clientRepo.Save(client); err != nil {
		t.Fatalf("Failed to save client: %v", err)
	}

	vehicleID := uuid.New()
	vehicle, err := serviceDomain.NewVehicle(clientID, "ABC1234", "Brand", "Model", 2020)
	// Ensure unique plate
	plateStr := "ABC" + time.Now().Format("0405")
	plate, _ := sharedkernel.NewPlacaBR(plateStr)
	vehicle.Plate = plate
	vehicle.ID = vehicleID
	if err := vehicleRepo.Save(vehicle); err != nil {
		t.Fatalf("Failed to save vehicle: %v", err)
	}

	// 2. Create Part with Stock
	partID := uuid.New()
	part, _ := inventoryDomain.NewPart("Test Part", "Desc", 10, 50.0) // 10 in stock, args: qty, price
	part.ID = partID
	if err := partRepo.Save(context.Background(), part); err != nil {
		t.Fatalf("Failed to save part: %v", err)
	}

	// 3. Create Order
	order, _ := serviceDomain.NewOrder(clientID, vehicleID)
	// Add 5 parts (Stock 10 -> 5)
	err = order.AddItem(partID, serviceDomain.ItemTypePart, part.Name, 5, 50.0)
	if err != nil {
		t.Fatalf("Failed to add item: %v", err)
	}
	if err := orderRepo.Save(order); err != nil {
		t.Fatalf("Failed to save order: %v", err)
	}

	// 4. Approve Order
	err = orderService.ApproveOrder(order.ID)
	if err != nil {
		t.Fatalf("Failed to approve order: %v", err)
	}

	// 5. Verify Order Status
	updatedOrder, _ := orderRepo.GetByID(order.ID)
	if updatedOrder.Status != serviceDomain.OrderStatusInExecution {
		t.Errorf("Expected InExecution, got %v", updatedOrder.Status)
	}

	// 6. Verify Part Stock
	updatedPart, _ := partRepo.GetByID(context.Background(), partID)
	if updatedPart.Quantity != 5 {
		t.Errorf("Expected stock 5, got %d", updatedPart.Quantity)
	}

	// 7. Try to approve again (should fail)
	err = orderService.ApproveOrder(order.ID)
	if err == nil {
		t.Error("Expected error approving already approved order")
	}

	// 8. Test Insufficient Stock
	// Create another order for 6 parts (Stock is 5)
	order2, _ := serviceDomain.NewOrder(clientID, vehicleID)
	order2.AddItem(partID, serviceDomain.ItemTypePart, part.Name, 6, 50.0)
	orderRepo.Save(order2)

	err = orderService.ApproveOrder(order2.ID)
	if err == nil {
		t.Error("Expected insufficient stock error")
	}
	if !errors.Is(err, inventoryDomain.ErrInsufficientStock) {
		t.Errorf("Expected ErrInsufficientStock, got %v", err)
	}
}
