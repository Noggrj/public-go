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

type ClientHandler struct {
	repo domain.ClientRepository
}

func NewClientHandler(repo domain.ClientRepository) *ClientHandler {
	return &ClientHandler{repo: repo}
}

type CreateClientRequest struct {
	Name     string `json:"name"`
	Document string `json:"document"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

// @Summary Create Client
// @Description Register a new client
// @Tags clients
// @Accept json
// @Produce json
// @Param client body CreateClientRequest true "Client Details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} string "Invalid input"
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/clients [post]
func (h *ClientHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	client, err := domain.NewClient(req.Name, req.Document, req.Email, req.Phone)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.repo.Save(client); err != nil {
		http.Error(w, "Failed to save client", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(client); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// @Summary List Clients
// @Description List all clients
// @Tags clients
// @Accept json
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/clients [get]
func (h *ClientHandler) List(w http.ResponseWriter, r *http.Request) {
	clients, err := h.repo.List()
	if err != nil {
		http.Error(w, "Failed to list clients", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(clients); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// @Summary Update Client
// @Description Update an existing client
// @Tags clients
// @Accept json
// @Produce json
// @Param id path string true "Client ID"
// @Param client body CreateClientRequest true "Client Details"
// @Success 200 {object} domain.Client
// @Failure 400 {object} string "Invalid input"
// @Failure 404 {object} string "Client not found"
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/clients/{id} [put]
func (h *ClientHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	var req CreateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	client, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	// Update fields
	if req.Name != "" {
		client.Name = req.Name
	}
	if req.Document != "" {
		doc, err := sharedkernel.NewDocumentoBR(req.Document)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		client.Document = doc
	}
	if req.Email != "" {
		client.Email = req.Email
	}
	if req.Phone != "" {
		client.Phone = req.Phone
	}
	client.UpdatedAt = time.Now()

	if err := h.repo.Save(client); err != nil {
		http.Error(w, "Failed to update client", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(client); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// @Summary Delete Client
// @Description Delete a client by ID
// @Tags clients
// @Accept json
// @Produce json
// @Param id path string true "Client ID"
// @Success 204 {object} nil
// @Failure 400 {object} string "Invalid ID"
// @Failure 404 {object} string "Client not found"
// @Failure 500 {object} string "Internal Server Error"
// @Router /admin/clients/{id} [delete]
func (h *ClientHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	if err := h.repo.Delete(id); err != nil {
		if errors.Is(err, domain.ErrClientNotFound) {
			http.Error(w, "Client not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete client", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
