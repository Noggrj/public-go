package domain

import (
	"github.com/google/uuid"
)

type ClientRepository interface {
	Save(client *Client) error
	GetByID(id uuid.UUID) (*Client, error)
	List() ([]*Client, error)
	Delete(id uuid.UUID) error
}

type OrderRepository interface {
	Save(order *Order) error
	GetByID(id uuid.UUID) (*Order, error)
	List() ([]*Order, error)
	// ListActive returns orders excluding Completed and Delivered,
	// sorted by status priority (In Execution > Awaiting Approval > In Diagnosis > Received)
	// and then by creation date (oldest first).
	ListActive() ([]*Order, error)
}
