//go:build integration

package infrastructure_test

import (
	"os"
	"testing"
	"time"

	"github.com/noggrj/autorepair/internal/platform/db"
	serviceDomain "github.com/noggrj/autorepair/internal/service/domain"
	"github.com/noggrj/autorepair/internal/service/infrastructure"
	"github.com/noggrj/autorepair/internal/sharedkernel"
	"github.com/google/uuid"
)

func TestPostgresOrderRepository(t *testing.T) {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		t.Skip("DB_URL not set")
	}

	pool, err := db.New(dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}
	defer pool.Close()

	// Setup Repos
	orderRepo := infrastructure.NewPostgresOrderRepository(pool.Pool)
	clientRepo := infrastructure.NewPostgresClientRepository(pool.Pool)
	vehicleRepo := infrastructure.NewPostgresVehicleRepository(pool.Pool)

	// Create dependencies
	clientID := uuid.New()
	// Use a valid fixed doc for initial creation (will be overwritten below)
	client, err := serviceDomain.NewClient("Test Client", "12345678901", "test@test.com", "123")
	if err != nil {
		t.Fatalf("Failed to create client domain obj: %v", err)
	}
	// Make sure we use a unique doc since previous runs might have conflicted.
	// DocumentoBR must be 11 or 14 digits.
	// Generate random 11 digit string.
	// We need to be very sure it's unique even if test runs fast.
	// Use microsecond to ensure uniqueness.
	randomDoc := time.Now().Format("150405.000000")[0:11]
	// Remove dot if any
	randomDoc = "1" + time.Now().Format("0405000000") // 1 + 10 digits = 11 digits

	doc, err := sharedkernel.NewDocumentoBR(randomDoc)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}
	client.Document = doc
	client.ID = clientID
	if err := clientRepo.Save(client); err != nil {
		t.Fatalf("Failed to save client: %v", err)
	}

	vehicleID := uuid.New()
	vehicle, err := serviceDomain.NewVehicle(clientID, "ABC1234", "Brand", "Model", 2020)
	if err != nil {
		t.Fatalf("Failed to create vehicle domain obj: %v", err)
	}
	// NewVehicle might fail if plate is invalid format.
	// PlacaBR logic: LLLNLNN or LLLNNNN.
	// Let's force a valid one.
	// 3 letters, 4 numbers
	// We need a unique plate here too.
	// Use microsecond precision for uniqueness
	plateSuffix := time.Now().Format("0405.000000")[5:9] // 4 digits
	plateStr := "ABC" + plateSuffix
	plate, _ := sharedkernel.NewPlacaBR(plateStr)
	vehicle.Plate = plate
	vehicle.ID = vehicleID
	if err := vehicleRepo.Save(vehicle); err != nil {
		t.Fatalf("Failed to save vehicle: %v", err)
	}

	t.Run("Create and Get Order", func(t *testing.T) {
		order, _ := serviceDomain.NewOrder(clientID, vehicleID)

		err := order.AddItem(uuid.New(), serviceDomain.ItemTypePart, "Oil Filter", 1, 50.0)
		if err != nil {
			t.Fatalf("Failed to add item: %v", err)
		}

		err = order.AddItem(uuid.New(), serviceDomain.ItemTypeService, "Oil Change", 1, 100.0)
		if err != nil {
			t.Fatalf("Failed to add item: %v", err)
		}

		if err := orderRepo.Save(order); err != nil {
			t.Fatalf("Failed to save order: %v", err)
		}

		fetched, err := orderRepo.GetByID(order.ID)
		if err != nil {
			t.Fatalf("Failed to get order: %v", err)
		}

		if fetched.Status != serviceDomain.OrderStatusReceived {
			t.Errorf("Expected status Received, got %v", fetched.Status)
		}

		if len(fetched.Items) != 2 {
			t.Errorf("Expected 2 items, got %d", len(fetched.Items))
		}

		if float64(fetched.Total) != 150.0 {
			t.Errorf("Expected total 150.0, got %f", fetched.Total)
		}
	})
}
