package domain

import (
	"errors"
	"time"

	"github.com/noggrj/autorepair/internal/sharedkernel"
	"github.com/google/uuid"
)

var (
	ErrClientNotFound = errors.New("client not found")
)

type Client struct {
	ID        uuid.UUID
	Name      string
	Document  sharedkernel.DocumentoBR
	Email     string
	Phone     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewClient(name, doc, email, phone string) (*Client, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}

	d, err := sharedkernel.NewDocumentoBR(doc)
	if err != nil {
		return nil, err
	}

	return &Client{
		ID:        uuid.New(),
		Name:      name,
		Document:  d,
		Email:     email,
		Phone:     phone,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// ClientRepository interface is moved to repository.go
