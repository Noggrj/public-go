package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	inventoryDomain "github.com/noggrj/autorepair/internal/inventory/domain"
	serviceApplication "github.com/noggrj/autorepair/internal/service/application"
	serviceDomain "github.com/noggrj/autorepair/internal/service/domain"
	"github.com/noggrj/autorepair/internal/sharedkernel"
)

type PartHandler struct {
	repo inventoryDomain.PartRepository
}

func NewPartHandler(repo inventoryDomain.PartRepository) *PartHandler {
	return &PartHandler{repo: repo}
}

type CreatePartRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	StockQty    int     `json:"stock_qty"`
}

// @Summary Create Part
// @Description Register a new part
// @Tags parts
// @Accept json
// @Produce json
// @Param part body CreatePartRequest true "Part Details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} string "Invalid input"
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/parts [post]
func (h *PartHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreatePartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	part, err := inventoryDomain.NewPart(req.Name, req.Description, req.StockQty, req.Price)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.repo.Save(r.Context(), part); err != nil {
		http.Error(w, "Failed to save part", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(part); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// @Summary List Parts
// @Description List all parts
// @Tags parts
// @Accept json
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/parts [get]
func (h *PartHandler) List(w http.ResponseWriter, r *http.Request) {
	parts, err := h.repo.List(r.Context())
	if err != nil {
		http.Error(w, "Failed to list parts", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(parts); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// @Summary Update Part
// @Description Update an existing part
// @Tags parts
// @Accept json
// @Produce json
// @Param id path string true "Part ID"
// @Param part body CreatePartRequest true "Part Details"
// @Success 200 {object} domain.Part
// @Failure 400 {object} string "Invalid input"
// @Failure 404 {object} string "Part not found"
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/parts/{id} [put]
func (h *PartHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	var req CreatePartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	part, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Part not found", http.StatusNotFound)
		return
	}

	// Update fields
	if req.Name != "" {
		part.Name = req.Name
	}
	if req.Description != "" {
		part.Description = req.Description
	}
	if req.Price > 0 {
		part.Price = req.Price
	}
	if req.StockQty >= 0 {
		part.Quantity = req.StockQty
	}

	if err := h.repo.Update(r.Context(), part); err != nil {
		http.Error(w, "Failed to update part", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(part); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// @Summary Delete Part
// @Description Delete a part by ID
// @Tags parts
// @Accept json
// @Produce json
// @Param id path string true "Part ID"
// @Success 204 {object} nil
// @Failure 400 {object} string "Invalid ID"
// @Failure 404 {object} string "Part not found"
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/parts/{id} [delete]
func (h *PartHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		http.Error(w, "Failed to delete part", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ---------------------------------------------------------

type ServiceHandler struct {
	repo serviceDomain.ServiceRepository
}

func NewServiceHandler(repo serviceDomain.ServiceRepository) *ServiceHandler {
	return &ServiceHandler{repo: repo}
}

type CreateServiceRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

// @Summary Create Service
// @Description Register a new service (labor)
// @Tags services
// @Accept json
// @Produce json
// @Param service body CreateServiceRequest true "Service Details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} string "Invalid input"
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/services [post]
func (h *ServiceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	service, err := serviceDomain.NewService(req.Name, req.Description, req.Price)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.repo.Save(service); err != nil {
		http.Error(w, "Failed to save service", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(service); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// @Summary List Services
// @Description List all services
// @Tags services
// @Accept json
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/services [get]
func (h *ServiceHandler) List(w http.ResponseWriter, r *http.Request) {
	services, err := h.repo.List()
	if err != nil {
		http.Error(w, "Failed to list services", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(services); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// @Summary Update Service
// @Description Update an existing service
// @Tags services
// @Accept json
// @Produce json
// @Param id path string true "Service ID"
// @Param service body CreateServiceRequest true "Service Details"
// @Success 200 {object} domain.Service
// @Failure 400 {object} string "Invalid input"
// @Failure 404 {object} string "Service not found"
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/services/{id} [put]
func (h *ServiceHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	var req CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	service, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	// Update fields
	if req.Name != "" {
		service.Name = req.Name
	}
	if req.Description != "" {
		service.Description = req.Description
	}
	if req.Price >= 0 {
		service.Price = sharedkernel.Money(req.Price)
	}
	service.UpdatedAt = time.Now()

	if err := h.repo.Save(service); err != nil {
		http.Error(w, "Failed to update service", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(service); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// @Summary Delete Service
// @Description Delete a service by ID
// @Tags services
// @Accept json
// @Produce json
// @Param id path string true "Service ID"
// @Success 204 {object} nil
// @Failure 400 {object} string "Invalid ID"
// @Failure 404 {object} string "Service not found"
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/services/{id} [delete]
func (h *ServiceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	if err := h.repo.Delete(id); err != nil {
		http.Error(w, "Failed to delete service", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ---------------------------------------------------------

type OrderHandler struct {
	orderRepo    serviceDomain.OrderRepository
	partRepo     inventoryDomain.PartRepository
	serviceRepo  serviceDomain.ServiceRepository
	orderService *serviceApplication.OrderService
}

func NewOrderHandler(
	orderRepo serviceDomain.OrderRepository,
	partRepo inventoryDomain.PartRepository,
	serviceRepo serviceDomain.ServiceRepository,
	orderService *serviceApplication.OrderService,
) *OrderHandler {
	return &OrderHandler{
		orderRepo:    orderRepo,
		partRepo:     partRepo,
		serviceRepo:  serviceRepo,
		orderService: orderService,
	}
}

// ... (Create method remains unchanged)

// @Summary Approve Order
// @Description Approve an order and reserve parts
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200
// @Failure 400 {object} string "Invalid order ID"
// @Failure 409 {object} string "Insufficient stock"
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/orders/{id}/approve [patch]
func (h *OrderHandler) Approve(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	if err := h.orderService.ApproveOrder(id); err != nil {
		if err == inventoryDomain.ErrInsufficientStock {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary Start Diagnosis
// @Description Start the diagnosis process for an order
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200
// @Failure 400 {object} string "Invalid order ID"
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/orders/{id}/diagnosis:start [post]
func (h *OrderHandler) StartDiagnosis(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	if err := h.orderService.StartDiagnosis(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary Send Budget
// @Description Send budget to the client for approval
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200
// @Failure 400 {object} string "Invalid order ID"
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/orders/{id}/budget:send [post]
func (h *OrderHandler) SendBudget(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	if err := h.orderService.SendBudget(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary Finish Order
// @Description Mark order as finished (Completed)
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200
// @Failure 400 {object} string "Invalid order ID"
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/orders/{id}/finish [post]
func (h *OrderHandler) FinishOrder(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	if err := h.orderService.FinishOrder(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary Deliver Order
// @Description Mark order as delivered to client
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200
// @Failure 400 {object} string "Invalid order ID"
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/orders/{id}/deliver [post]
func (h *OrderHandler) DeliverOrder(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	if err := h.orderService.DeliverOrder(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type UpdateStatusRequest struct {
	Status string `json:"status"`
}

// @Summary Update Order Status
// @Description Update status of an order
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param status body UpdateStatusRequest true "New Status"
// @Success 200
// @Failure 400 {object} string "Invalid input"
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/orders/{id}/status [patch]
func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	status := serviceDomain.OrderStatus(req.Status)
	// Ideally validate status enum here

	if err := h.orderService.UpdateStatus(id, status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type CreateOrderItemRequest struct {
	Type     string `json:"type"` // "service" or "part"
	RefID    string `json:"ref_id"`
	Quantity int    `json:"quantity"`
}

type CreateOrderRequest struct {
	ClientID  string                   `json:"client_id"`
	VehicleID string                   `json:"vehicle_id"`
	Items     []CreateOrderItemRequest `json:"items"`
}

// @Summary Create Order
// @Description Create a new service order
// @Tags orders
// @Accept json
// @Produce json
// @Param order body CreateOrderRequest true "Order Request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} string "Invalid input"
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/orders [post]
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	clientID, err := uuid.Parse(req.ClientID)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}
	vehicleID, err := uuid.Parse(req.VehicleID)
	if err != nil {
		http.Error(w, "Invalid vehicle ID", http.StatusBadRequest)
		return
	}

	order, err := serviceDomain.NewOrder(clientID, vehicleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, itemReq := range req.Items {
		refID, err := uuid.Parse(itemReq.RefID)
		if err != nil {
			http.Error(w, "Invalid item ref ID", http.StatusBadRequest)
			return
		}

		switch serviceDomain.OrderItemType(itemReq.Type) {
		case serviceDomain.ItemTypeService:
			svc, err := h.serviceRepo.GetByID(refID)
			if err != nil {
				http.Error(w, "Service not found: "+refID.String(), http.StatusBadRequest)
				return
			}
			err = order.AddItem(refID, serviceDomain.ItemTypeService, svc.Name, itemReq.Quantity, float64(svc.Price))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		case serviceDomain.ItemTypePart:
			part, err := h.partRepo.GetByID(r.Context(), refID)
			if err != nil {
				http.Error(w, "Part not found: "+refID.String(), http.StatusBadRequest)
				return
			}
			// Note: We check stock here but decrement only on approval (Sprint 3)
			// Requirements: "Automatically generate estimate/budget"
			err = order.AddItem(refID, serviceDomain.ItemTypePart, part.Name, itemReq.Quantity, float64(part.Price))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		default:
			http.Error(w, "Invalid item type", http.StatusBadRequest)
			return
		}
	}

	if err := h.orderRepo.Save(order); err != nil {
		http.Error(w, "Failed to save order", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// @Summary Get Order
// @Description Get details of a specific order
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} string "Invalid order ID"
// @Failure 404 {object} string "Order not found"
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/orders/{id} [get]
func (h *OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	order, err := h.orderRepo.GetByID(id)
	if err != nil {
		if err == serviceDomain.ErrOrderNotFound {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// @Summary Report Revenue
// @Description Get total revenue report
// @Tags reports
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/reports/revenue [get]
func (h *OrderHandler) ReportRevenue(w http.ResponseWriter, r *http.Request) {
	orders, err := h.orderRepo.List()
	if err != nil {
		http.Error(w, "Failed to list orders", http.StatusInternalServerError)
		return
	}

	var totalRevenue float64
	for _, o := range orders {
		totalRevenue += float64(o.Total)
	}

	response := map[string]interface{}{
		"period":        "all_time",
		"total_revenue": totalRevenue,
		"order_count":   len(orders),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// @Summary Report Avg Execution Time
// @Description Get average execution time of orders (In Execution -> Completed)
// @Tags reports
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/reports/avg-execution-time [get]
func (h *OrderHandler) ReportAvgExecutionTime(w http.ResponseWriter, r *http.Request) {
	orders, err := h.orderRepo.List()
	if err != nil {
		http.Error(w, "Failed to list orders", http.StatusInternalServerError)
		return
	}

	var totalDuration time.Duration
	var count int

	for _, o := range orders {
		if o.Status == serviceDomain.OrderStatusCompleted || o.Status == serviceDomain.OrderStatusDelivered {
			if o.StartedAt != nil && o.FinishedAt != nil {
				totalDuration += o.FinishedAt.Sub(*o.StartedAt)
				count++
			}
		}
	}

	var avgMinutes float64
	if count > 0 {
		avgMinutes = totalDuration.Minutes() / float64(count)
	}

	response := map[string]interface{}{
		"avg_execution_minutes": avgMinutes,
		"orders_counted":        count,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

type OrderTrackingItem struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Total    float64 `json:"total"`
}

type OrderTrackingResponse struct {
	ID        string              `json:"id"`
	Status    string              `json:"status"`
	Total     float64             `json:"total"`
	CreatedAt time.Time           `json:"created_at"`
	Items     []OrderTrackingItem `json:"items"`
}

// @Summary Track Order
// @Description Track the status of an order (Public)
// @Tags public
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} OrderTrackingResponse
// @Failure 400 {object} string "Invalid order ID"
// @Failure 404 {object} string "Order not found"
// @Failure 500 {object} string "Internal Server Error"
// @Router /orders/{id}/track [get]
func (h *OrderHandler) TrackOrder(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	order, err := h.orderRepo.GetByID(id)
	if err != nil {
		if err == serviceDomain.ErrOrderNotFound {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var items []OrderTrackingItem
	for _, item := range order.Items {
		items = append(items, OrderTrackingItem{
			Name:     item.Name,
			Quantity: item.Quantity,
			Total:    float64(item.Total),
		})
	}

	response := OrderTrackingResponse{
		ID:        order.ID.String(),
		Status:    string(order.Status),
		Total:     float64(order.Total),
		CreatedAt: order.CreatedAt,
		Items:     items,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// @Summary List Active Orders
// @Description List active service orders sorted by status priority (In Execution > Awaiting Approval > In Diagnosis > Received), oldest first. Excludes Completed and Delivered orders.
// @Tags orders
// @Accept json
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/orders [get]
func (h *OrderHandler) ListActive(w http.ResponseWriter, r *http.Request) {
	orders, err := h.orderRepo.ListActive()
	if err != nil {
		http.Error(w, "Failed to list orders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

type BudgetResponseRequest struct {
	Approved bool `json:"approved"`
}

// @Summary Respond to Budget
// @Description External endpoint for approving or rejecting an order budget (webhook-style)
// @Tags public
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param response body BudgetResponseRequest true "Approval/Rejection"
// @Success 200 {object} map[string]string
// @Failure 400 {object} string "Invalid input"
// @Failure 500 {object} string "Internal Server Error"
// @Router /orders/{id}/budget-response [post]
func (h *OrderHandler) ApproveBudget(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var req BudgetResponseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if req.Approved {
		if err := h.orderService.ApproveOrder(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err := h.orderService.RejectBudget(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	status := "approved"
	if !req.Approved {
		status = "rejected"
	}
	response := map[string]string{"status": status, "order_id": id.String()}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
