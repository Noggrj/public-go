package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	inventoryDomain "github.com/noggrj/autorepair/internal/inventory/domain"
	serviceApplication "github.com/noggrj/autorepair/internal/service/application"
	serviceHttp "github.com/noggrj/autorepair/internal/service/delivery/http"
	serviceDomain "github.com/noggrj/autorepair/internal/service/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Helper ---

func setupOrderHandler() (*serviceHttp.OrderHandler, *MockOrderRepository, *MockPartRepository, *MockServiceRepository, *MockClientRepository) {
	mockOrderRepo := new(MockOrderRepository)
	mockPartRepo := new(MockPartRepository)
	mockServiceRepo := new(MockServiceRepository)
	mockClientRepo := new(MockClientRepository)
	mockNotifier := new(MockNotifier)

	orderService := serviceApplication.NewOrderService(mockOrderRepo, mockPartRepo, mockClientRepo, mockNotifier)
	handler := serviceHttp.NewOrderHandler(mockOrderRepo, mockPartRepo, mockServiceRepo, orderService)

	return handler, mockOrderRepo, mockPartRepo, mockServiceRepo, mockClientRepo
}

// --- Tests ---

func TestOrderHandler_Create(t *testing.T) {
	handler, mockOrderRepo, mockPartRepo, mockServiceRepo, _ := setupOrderHandler()

	clientID := uuid.New()
	vehicleID := uuid.New()
	partID := uuid.New()
	serviceID := uuid.New()

	reqBody := map[string]interface{}{
		"client_id":  clientID.String(),
		"vehicle_id": vehicleID.String(),
		"items": []map[string]interface{}{
			{"type": "part", "ref_id": partID.String(), "quantity": 1},
			{"type": "service", "ref_id": serviceID.String(), "quantity": 1},
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/admin/orders", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	// Update to use context match
	mockPartRepo.On("GetByID", mock.Anything, partID).Return(&inventoryDomain.Part{ID: partID, Name: "Tire", Price: 50.0}, nil)
	mockServiceRepo.On("GetByID", serviceID).Return(&serviceDomain.Service{ID: serviceID, Name: "Fix", Price: 100.0}, nil)
	mockOrderRepo.On("Save", mock.Anything).Return(nil)

	handler.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	mockOrderRepo.AssertExpectations(t)
}

func TestOrderHandler_Get(t *testing.T) {
	handler, mockOrderRepo, _, _, _ := setupOrderHandler()

	orderID := uuid.New()
	order := &serviceDomain.Order{ID: orderID, Total: 150.0}

	mockOrderRepo.On("GetByID", orderID).Return(order, nil)

	req, _ := http.NewRequest("GET", "/admin/orders/"+orderID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", orderID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.Get(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp serviceDomain.Order
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, orderID, resp.ID)
}

func TestOrderHandler_Get_NotFound(t *testing.T) {
	handler, mockOrderRepo, _, _, _ := setupOrderHandler()

	orderID := uuid.New()
	mockOrderRepo.On("GetByID", orderID).Return(nil, serviceDomain.ErrOrderNotFound)

	req, _ := http.NewRequest("GET", "/admin/orders/"+orderID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", orderID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.Get(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestOrderHandler_Approve(t *testing.T) {
	// Refactored to TestOrderHandler_Approve_Full below
}

// Refactor helper to return Notifier mock
func setupOrderHandlerWithNotifier() (*serviceHttp.OrderHandler, *MockOrderRepository, *MockPartRepository, *MockServiceRepository, *MockClientRepository, *MockNotifier) {
	mockOrderRepo := new(MockOrderRepository)
	mockPartRepo := new(MockPartRepository)
	mockServiceRepo := new(MockServiceRepository)
	mockClientRepo := new(MockClientRepository)
	mockNotifier := new(MockNotifier)

	orderService := serviceApplication.NewOrderService(mockOrderRepo, mockPartRepo, mockClientRepo, mockNotifier)
	handler := serviceHttp.NewOrderHandler(mockOrderRepo, mockPartRepo, mockServiceRepo, orderService)

	return handler, mockOrderRepo, mockPartRepo, mockServiceRepo, mockClientRepo, mockNotifier
}

func TestOrderHandler_Approve_Full(t *testing.T) {
	handler, mockOrderRepo, _, _, mockClientRepo, mockNotifier := setupOrderHandlerWithNotifier()

	orderID := uuid.New()
	clientID := uuid.New()
	order := &serviceDomain.Order{ID: orderID, ClientID: clientID, Status: serviceDomain.OrderStatusReceived}
	client := &serviceDomain.Client{ID: clientID, Email: "test@test.com"}

	mockOrderRepo.On("GetByID", orderID).Return(order, nil)
	mockOrderRepo.On("Save", mock.Anything).Return(nil)
	mockClientRepo.On("GetByID", clientID).Return(client, nil)
	mockNotifier.On("SendEmail", "test@test.com", mock.Anything, mock.Anything).Return(nil)

	req, _ := http.NewRequest("PATCH", "/admin/orders/"+orderID.String()+"/approve", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", orderID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.Approve(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestOrderHandler_ReportRevenue(t *testing.T) {
	handler, mockOrderRepo, _, _, _ := setupOrderHandler()

	orders := []*serviceDomain.Order{
		{Total: 100.0},
		{Total: 200.0},
	}
	mockOrderRepo.On("List").Return(orders, nil)

	req, _ := http.NewRequest("GET", "/admin/reports/revenue", nil)
	rr := httptest.NewRecorder()

	handler.ReportRevenue(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 300.0, resp["total_revenue"])
}

func TestOrderHandler_StartDiagnosis(t *testing.T) {
	handler, mockOrderRepo, _, _, mockClientRepo, mockNotifier := setupOrderHandlerWithNotifier()

	orderID := uuid.New()
	clientID := uuid.New()
	order := &serviceDomain.Order{ID: orderID, ClientID: clientID, Status: serviceDomain.OrderStatusReceived}
	client := &serviceDomain.Client{ID: clientID, Email: "test@test.com"}

	mockOrderRepo.On("GetByID", orderID).Return(order, nil)
	mockOrderRepo.On("Save", mock.MatchedBy(func(o *serviceDomain.Order) bool {
		return o.Status == serviceDomain.OrderStatusInDiagnosis
	})).Return(nil)
	mockClientRepo.On("GetByID", clientID).Return(client, nil)
	mockNotifier.On("SendEmail", "test@test.com", mock.Anything, mock.Anything).Return(nil)

	req, _ := http.NewRequest("POST", "/admin/orders/"+orderID.String()+"/diagnosis:start", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", orderID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.StartDiagnosis(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestOrderHandler_SendBudget(t *testing.T) {
	handler, mockOrderRepo, _, _, mockClientRepo, mockNotifier := setupOrderHandlerWithNotifier()

	orderID := uuid.New()
	clientID := uuid.New()
	order := &serviceDomain.Order{ID: orderID, ClientID: clientID, Status: serviceDomain.OrderStatusInDiagnosis}
	client := &serviceDomain.Client{ID: clientID, Email: "test@test.com"}

	mockOrderRepo.On("GetByID", orderID).Return(order, nil)
	mockOrderRepo.On("Save", mock.MatchedBy(func(o *serviceDomain.Order) bool {
		return o.Status == serviceDomain.OrderStatusAwaitingApproval
	})).Return(nil)
	mockClientRepo.On("GetByID", clientID).Return(client, nil)
	mockNotifier.On("SendEmail", "test@test.com", mock.Anything, mock.Anything).Return(nil)

	req, _ := http.NewRequest("POST", "/admin/orders/"+orderID.String()+"/budget:send", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", orderID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.SendBudget(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestOrderHandler_FinishOrder(t *testing.T) {
	handler, mockOrderRepo, _, _, mockClientRepo, mockNotifier := setupOrderHandlerWithNotifier()

	orderID := uuid.New()
	clientID := uuid.New()
	order := &serviceDomain.Order{ID: orderID, ClientID: clientID, Status: serviceDomain.OrderStatusInExecution}
	client := &serviceDomain.Client{ID: clientID, Email: "test@test.com"}

	mockOrderRepo.On("GetByID", orderID).Return(order, nil)
	mockOrderRepo.On("Save", mock.MatchedBy(func(o *serviceDomain.Order) bool {
		return o.Status == serviceDomain.OrderStatusCompleted
	})).Return(nil)
	mockClientRepo.On("GetByID", clientID).Return(client, nil)
	mockNotifier.On("SendEmail", "test@test.com", mock.Anything, mock.Anything).Return(nil)

	req, _ := http.NewRequest("POST", "/admin/orders/"+orderID.String()+"/finish", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", orderID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.FinishOrder(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestOrderHandler_DeliverOrder(t *testing.T) {
	handler, mockOrderRepo, _, _, mockClientRepo, mockNotifier := setupOrderHandlerWithNotifier()

	orderID := uuid.New()
	clientID := uuid.New()
	order := &serviceDomain.Order{ID: orderID, ClientID: clientID, Status: serviceDomain.OrderStatusCompleted}
	client := &serviceDomain.Client{ID: clientID, Email: "test@test.com"}

	mockOrderRepo.On("GetByID", orderID).Return(order, nil)
	mockOrderRepo.On("Save", mock.MatchedBy(func(o *serviceDomain.Order) bool {
		return o.Status == serviceDomain.OrderStatusDelivered
	})).Return(nil)
	mockClientRepo.On("GetByID", clientID).Return(client, nil)
	mockNotifier.On("SendEmail", "test@test.com", mock.Anything, mock.Anything).Return(nil)

	req, _ := http.NewRequest("POST", "/admin/orders/"+orderID.String()+"/deliver", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", orderID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.DeliverOrder(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestOrderHandler_UpdateStatus(t *testing.T) {
	handler, mockOrderRepo, _, _, mockClientRepo, mockNotifier := setupOrderHandlerWithNotifier()

	orderID := uuid.New()
	clientID := uuid.New()
	order := &serviceDomain.Order{ID: orderID, ClientID: clientID, Status: serviceDomain.OrderStatusReceived}
	client := &serviceDomain.Client{ID: clientID, Email: "test@test.com"}

	mockOrderRepo.On("GetByID", orderID).Return(order, nil)
	mockOrderRepo.On("Save", mock.Anything).Return(nil)
	mockClientRepo.On("GetByID", clientID).Return(client, nil)
	mockNotifier.On("SendEmail", "test@test.com", mock.Anything, mock.Anything).Return(nil)

	reqBody := map[string]string{"status": "in_execution"}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PATCH", "/admin/orders/"+orderID.String()+"/status", bytes.NewBuffer(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", orderID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.UpdateStatus(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestOrderHandler_ReportAvgExecutionTime(t *testing.T) {
	handler, mockOrderRepo, _, _, _ := setupOrderHandler()

	start := time.Now().Add(-2 * time.Hour)
	end := time.Now()
	orders := []*serviceDomain.Order{
		{Status: serviceDomain.OrderStatusCompleted, StartedAt: &start, FinishedAt: &end},
	}
	mockOrderRepo.On("List").Return(orders, nil)

	req, _ := http.NewRequest("GET", "/admin/reports/avg-execution-time", nil)
	rr := httptest.NewRecorder()

	handler.ReportAvgExecutionTime(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.InDelta(t, 120.0, resp["avg_execution_minutes"], 0.1)
}

func TestOrderHandler_StartDiagnosis_InvalidID(t *testing.T) {
	handler, _, _, _, _ := setupOrderHandler()

	req, _ := http.NewRequest("POST", "/admin/orders/invalid-id/diagnosis:start", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.StartDiagnosis(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_Create_InvalidJSON(t *testing.T) {
	handler, _, _, _, _ := setupOrderHandler()

	req, _ := http.NewRequest("POST", "/admin/orders", bytes.NewBuffer([]byte("{invalid-json")))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_TrackOrder(t *testing.T) {
	handler, mockOrderRepo, _, _, _ := setupOrderHandler()

	orderID := uuid.New()
	order := &serviceDomain.Order{
		ID:        orderID,
		Status:    serviceDomain.OrderStatusReceived,
		Total:     150.0,
		CreatedAt: time.Now(),
		Items: []*serviceDomain.OrderItem{
			{Name: "Item 1", Quantity: 1, Total: 50.0},
			{Name: "Item 2", Quantity: 2, Total: 100.0},
		},
	}

	mockOrderRepo.On("GetByID", orderID).Return(order, nil)

	req, _ := http.NewRequest("GET", "/orders/"+orderID.String()+"/track", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", orderID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.TrackOrder(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, orderID.String(), resp["id"])
	assert.Equal(t, "Received", resp["status"])
	assert.Equal(t, 150.0, resp["total"])
	assert.Len(t, resp["items"], 2)
}

func TestOrderHandler_TrackOrder_NotFound(t *testing.T) {
	handler, mockOrderRepo, _, _, _ := setupOrderHandler()

	orderID := uuid.New()
	mockOrderRepo.On("GetByID", orderID).Return(nil, serviceDomain.ErrOrderNotFound)

	req, _ := http.NewRequest("GET", "/orders/"+orderID.String()+"/track", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", orderID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.TrackOrder(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestOrderHandler_TrackOrder_InvalidID(t *testing.T) {
	handler, _, _, _, _ := setupOrderHandler()

	req, _ := http.NewRequest("GET", "/orders/invalid-id/track", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.TrackOrder(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_Create_SaveError(t *testing.T) {
	handler, mockOrderRepo, mockPartRepo, mockServiceRepo, _ := setupOrderHandler()

	clientID := uuid.New()
	vehicleID := uuid.New()
	partID := uuid.New()
	serviceID := uuid.New()

	reqBody := map[string]interface{}{
		"client_id":  clientID.String(),
		"vehicle_id": vehicleID.String(),
		"items": []map[string]interface{}{
			{"type": "part", "ref_id": partID.String(), "quantity": 1},
			{"type": "service", "ref_id": serviceID.String(), "quantity": 1},
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/admin/orders", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	mockPartRepo.On("GetByID", mock.Anything, partID).Return(&inventoryDomain.Part{ID: partID, Name: "Tire", Price: 50.0}, nil)
	mockServiceRepo.On("GetByID", serviceID).Return(&serviceDomain.Service{ID: serviceID, Name: "Fix", Price: 100.0}, nil)
	mockOrderRepo.On("Save", mock.Anything).Return(errors.New("db error"))

	handler.Create(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestOrderHandler_Create_InvalidUUID(t *testing.T) {
	handler, _, _, _, _ := setupOrderHandler()

	reqBody := map[string]interface{}{
		"client_id":  "invalid-uuid",
		"vehicle_id": uuid.New().String(),
		"items":      []interface{}{},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/admin/orders", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_Approve_InvalidID(t *testing.T) {
	handler, _, _, _, _ := setupOrderHandler()

	req, _ := http.NewRequest("PATCH", "/admin/orders/invalid-id/approve", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.Approve(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_UpdateStatus_InvalidID(t *testing.T) {
	handler, _, _, _, _ := setupOrderHandler()

	req, _ := http.NewRequest("PATCH", "/admin/orders/invalid-id/status", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.UpdateStatus(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_UpdateStatus_InvalidBody(t *testing.T) {
	handler, _, _, _, _ := setupOrderHandler()

	orderID := uuid.New()
	req, _ := http.NewRequest("PATCH", "/admin/orders/"+orderID.String()+"/status", bytes.NewBuffer([]byte("invalid")))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", orderID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.UpdateStatus(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_Approve_InsufficientStock(t *testing.T) {
	handler, mockOrderRepo, mockPartRepo, _, _ := setupOrderHandler()

	orderID := uuid.New()
	partID := uuid.New()
	order := &serviceDomain.Order{
		ID:     orderID,
		Status: serviceDomain.OrderStatusReceived,
		Items: []*serviceDomain.OrderItem{
			{RefID: partID, Type: serviceDomain.ItemTypePart, Quantity: 10},
		},
	}

	mockOrderRepo.On("GetByID", orderID).Return(order, nil)

	// Mock part behavior for insufficient stock
	part := &inventoryDomain.Part{ID: partID, Quantity: 5} // Less than 10
	mockPartRepo.On("GetByID", mock.Anything, partID).Return(part, nil)
	// RemoveStock will be called on domain object inside service, we don't mock it here directly if we return a real part object.
	// But since we return a pointer, the service modifies it.
	// The service will call part.RemoveStock(10) which returns error.
	// We don't need to mock DecreaseStock on repo anymore.

	req, _ := http.NewRequest("PATCH", "/admin/orders/"+orderID.String()+"/approve", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", orderID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.Approve(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestOrderHandler_Create_ServiceNotFound(t *testing.T) {
	handler, _, _, mockServiceRepo, _ := setupOrderHandler()

	clientID := uuid.New()
	vehicleID := uuid.New()
	serviceID := uuid.New()

	reqBody := map[string]interface{}{
		"client_id":  clientID.String(),
		"vehicle_id": vehicleID.String(),
		"items": []map[string]interface{}{
			{"type": "service", "ref_id": serviceID.String(), "quantity": 1},
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/admin/orders", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	mockServiceRepo.On("GetByID", serviceID).Return(nil, errors.New("not found"))

	handler.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_Create_PartNotFound(t *testing.T) {
	handler, _, mockPartRepo, _, _ := setupOrderHandler()

	clientID := uuid.New()
	vehicleID := uuid.New()
	partID := uuid.New()

	reqBody := map[string]interface{}{
		"client_id":  clientID.String(),
		"vehicle_id": vehicleID.String(),
		"items": []map[string]interface{}{
			{"type": "part", "ref_id": partID.String(), "quantity": 1},
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/admin/orders", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	// Update to use context match
	mockPartRepo.On("GetByID", mock.Anything, partID).Return(nil, errors.New("not found"))

	handler.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_StartDiagnosis_ServiceError(t *testing.T) {
	handler, mockOrderRepo, _, _, _ := setupOrderHandler()

	orderID := uuid.New()
	mockOrderRepo.On("GetByID", orderID).Return(nil, errors.New("db error"))

	req, _ := http.NewRequest("POST", "/admin/orders/"+orderID.String()+"/diagnosis:start", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", orderID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.StartDiagnosis(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestOrderHandler_Create_InvalidInput(t *testing.T) {
	handler, _, _, _, _ := setupOrderHandler()

	req, _ := http.NewRequest("POST", "/admin/orders", bytes.NewBuffer([]byte("invalid json")))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
