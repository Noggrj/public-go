package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	inventoryDomain "github.com/noggrj/autorepair/internal/inventory/domain"
	serviceApplication "github.com/noggrj/autorepair/internal/service/application"
	serviceDomain "github.com/noggrj/autorepair/internal/service/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks (redefined for http package)
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

type MockServiceRepository struct {
	mock.Mock
}

func (m *MockServiceRepository) Save(service *serviceDomain.Service) error {
	args := m.Called(service)
	return args.Error(0)
}

func (m *MockServiceRepository) GetByID(id uuid.UUID) (*serviceDomain.Service, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*serviceDomain.Service), args.Error(1)
}

func (m *MockServiceRepository) List() ([]*serviceDomain.Service, error) {
	args := m.Called()
	return args.Get(0).([]*serviceDomain.Service), args.Error(1)
}

func (m *MockServiceRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
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

func TestOrderHandler_StartDiagnosis(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockPartRepo := new(MockPartRepository)
	mockServiceRepo := new(MockServiceRepository)
	mockClientRepo := new(MockClientRepository)
	mockNotifier := new(MockEmailService)

	orderService := serviceApplication.NewOrderService(mockOrderRepo, mockPartRepo, mockClientRepo, mockNotifier)
	handler := NewOrderHandler(mockOrderRepo, mockPartRepo, mockServiceRepo, orderService)

	t.Run("Success", func(t *testing.T) {
		orderID := uuid.New()
		clientID := uuid.New()
		order := &serviceDomain.Order{ID: orderID, ClientID: clientID, Status: serviceDomain.OrderStatusReceived}
		client := &serviceDomain.Client{ID: clientID, Email: "test@example.com"}

		mockOrderRepo.On("GetByID", orderID).Return(order, nil)
		mockOrderRepo.On("Save", mock.Anything).Return(nil)
		mockClientRepo.On("GetByID", clientID).Return(client, nil)
		mockNotifier.On("SendEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		r := httptest.NewRequest(http.MethodPost, "/admin/orders/{id}/diagnosis:start", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", orderID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		handler.StartDiagnosis(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestPartHandler_Create(t *testing.T) {
	mockRepo := new(MockPartRepository)
	handler := NewPartHandler(mockRepo)

	t.Run("Success", func(t *testing.T) {
		reqBody := CreatePartRequest{
			Name:        "Test Part",
			Description: "Test Description",
			Price:       100.0,
			StockQty:    10,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/admin/parts", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Part")).Return(nil)

		handler.Create(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})
}

func TestServiceHandler_Create(t *testing.T) {
	mockRepo := new(MockServiceRepository)
	handler := NewServiceHandler(mockRepo)

	t.Run("Success", func(t *testing.T) {
		reqBody := CreateServiceRequest{
			Name:        "Test Service",
			Description: "Test Description",
			Price:       50.0,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/admin/services", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockRepo.On("Save", mock.AnythingOfType("*domain.Service")).Return(nil)

		handler.Create(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})
}
