package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/noggrj/autorepair/internal/service/domain"
	"github.com/noggrj/autorepair/internal/sharedkernel"
)

type VehicleHandler struct {
	repo domain.VehicleRepository
}

func NewVehicleHandler(repo domain.VehicleRepository) *VehicleHandler {
	return &VehicleHandler{repo: repo}
}

type CreateVehicleRequest struct {
	ClientID string `json:"client_id"`
	Plate    string `json:"plate"`
	Brand    string `json:"brand"`
	Model    string `json:"model"`
	Year     int    `json:"year"`
}

// @Summary Create Vehicle
// @Description Register a new vehicle
// @Tags vehicles
// @Accept json
// @Produce json
// @Param vehicle body CreateVehicleRequest true "Vehicle Details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} string "Invalid input"
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/vehicles [post]
func (h *VehicleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	clientID, err := uuid.Parse(req.ClientID)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	vehicle, err := domain.NewVehicle(clientID, req.Plate, req.Brand, req.Model, req.Year)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.repo.Save(vehicle); err != nil {
		http.Error(w, "Failed to save vehicle", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(vehicle); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// @Summary List Vehicles
// @Description List vehicles by client
// @Tags vehicles
// @Accept json
// @Produce json
// @Param client_id query string true "Client ID"
// @Success 200 {array} map[string]interface{}
// @Failure 400 {object} string "Invalid client ID"
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/vehicles [get]
func (h *VehicleHandler) ListByClient(w http.ResponseWriter, r *http.Request) {
	clientIDStr := r.URL.Query().Get("client_id")
	if clientIDStr == "" {
		http.Error(w, "client_id is required", http.StatusBadRequest)
		return
	}

	clientID, err := uuid.Parse(clientIDStr)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	vehicles, err := h.repo.ListByClientID(clientID)
	if err != nil {
		http.Error(w, "Failed to list vehicles", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(vehicles); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// @Summary Update Vehicle
// @Description Update an existing vehicle
// @Tags vehicles
// @Accept json
// @Produce json
// @Param id path string true "Vehicle ID"
// @Param vehicle body CreateVehicleRequest true "Vehicle Details"
// @Success 200 {object} domain.Vehicle
// @Failure 400 {object} string "Invalid input"
// @Failure 404 {object} string "Vehicle not found"
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/vehicles/{id} [put]
func (h *VehicleHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	var req CreateVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	vehicle, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, "Vehicle not found", http.StatusNotFound)
		return
	}

	// Update fields
	if req.ClientID != "" {
		clientID, err := uuid.Parse(req.ClientID)
		if err == nil {
			vehicle.ClientID = clientID
		}
	}
	if req.Plate != "" {
		plate, err := sharedkernel.NewPlacaBR(req.Plate)
		if err == nil {
			vehicle.Plate = plate
		}
	}
	if req.Brand != "" {
		vehicle.Brand = req.Brand
	}
	if req.Model != "" {
		vehicle.Model = req.Model
	}
	if req.Year > 0 {
		vehicle.Year = req.Year
	}
	vehicle.UpdatedAt = time.Now()

	if err := h.repo.Save(vehicle); err != nil {
		http.Error(w, "Failed to update vehicle", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(vehicle); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// @Summary Delete Vehicle
// @Description Delete a vehicle by ID
// @Tags vehicles
// @Accept json
// @Produce json
// @Param id path string true "Vehicle ID"
// @Success 204 {object} nil
// @Failure 400 {object} string "Invalid ID"
// @Failure 404 {object} string "Vehicle not found"
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/vehicles/{id} [delete]
func (h *VehicleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	if err := h.repo.Delete(id); err != nil {
		if errors.Is(err, domain.ErrVehicleNotFound) {
			http.Error(w, "Vehicle not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete vehicle", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
