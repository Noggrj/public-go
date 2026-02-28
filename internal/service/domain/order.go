package domain

import (
	"errors"
	"time"

	"github.com/noggrj/autorepair/internal/sharedkernel"
	"github.com/google/uuid"
)

var (
	ErrOrderNotFound = errors.New("order not found")
)

type OrderItemType string

const (
	ItemTypeService OrderItemType = "service"
	ItemTypePart    OrderItemType = "part"
)

type OrderItem struct {
	ID        uuid.UUID
	OrderID   uuid.UUID
	RefID     uuid.UUID // ID of the Service or Part
	Type      OrderItemType
	Name      string
	Quantity  int
	UnitPrice sharedkernel.Money
	Total     sharedkernel.Money
}

type Order struct {
	ID           uuid.UUID
	ClientID     uuid.UUID
	VehicleID    uuid.UUID
	Status       OrderStatus
	Items        []*OrderItem
	TotalService sharedkernel.Money
	TotalParts   sharedkernel.Money
	Total        sharedkernel.Money
	CreatedAt    time.Time
	UpdatedAt    time.Time
	StartedAt    *time.Time
	FinishedAt   *time.Time
}

func NewOrder(clientID, vehicleID uuid.UUID) (*Order, error) {
	if clientID == uuid.Nil || vehicleID == uuid.Nil {
		return nil, errors.New("client and vehicle are required")
	}

	return &Order{
		ID:        uuid.New(),
		ClientID:  clientID,
		VehicleID: vehicleID,
		Status:    OrderStatusReceived,
		Items:     []*OrderItem{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (o *Order) AddItem(refID uuid.UUID, itemType OrderItemType, name string, qty int, price float64) error {
	if qty <= 0 {
		return errors.New("quantity must be positive")
	}
	if price < 0 {
		return errors.New("price cannot be negative")
	}

	unitPrice := sharedkernel.Money(price)
	total := sharedkernel.Money(float64(qty) * price)

	item := &OrderItem{
		ID:        uuid.New(),
		OrderID:   o.ID,
		RefID:     refID,
		Type:      itemType,
		Name:      name,
		Quantity:  qty,
		UnitPrice: unitPrice,
		Total:     total,
	}

	o.Items = append(o.Items, item)
	o.CalculateTotal()
	return nil
}

func (o *Order) CalculateTotal() {
	var totalService, totalParts float64

	for _, item := range o.Items {
		if item.Type == ItemTypeService {
			totalService += float64(item.Total)
		} else {
			totalParts += float64(item.Total)
		}
	}

	o.TotalService = sharedkernel.Money(totalService)
	o.TotalParts = sharedkernel.Money(totalParts)
	o.Total = sharedkernel.Money(totalService + totalParts)
}
