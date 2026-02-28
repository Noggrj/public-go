package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	serviceHttp "github.com/noggrj/autorepair/internal/service/delivery/http"
	serviceDomain "github.com/noggrj/autorepair/internal/service/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestVehicleHandler_Create(t *testing.T) {
	mockRepo := new(MockVehicleRepository)
	handler := serviceHttp.NewVehicleHandler(mockRepo)

	clientID := uuid.New()
	reqBody := map[string]interface{}{
		"client_id": clientID.String(),
		"plate":     "ABC-1234",
		"brand":     "Toyota",
		"model":     "Corolla",
		"year":      2020,
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/admin/vehicles", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	mockRepo.On("Save", mock.Anything).Return(nil)

	handler.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestVehicleHandler_ListByClient(t *testing.T) {
	mockRepo := new(MockVehicleRepository)
	handler := serviceHttp.NewVehicleHandler(mockRepo)

	clientID := uuid.New()
	vehicles := []*serviceDomain.Vehicle{
		{Brand: "Toyota"},
	}
	mockRepo.On("ListByClientID", clientID).Return(vehicles, nil)

	req, _ := http.NewRequest("GET", "/admin/vehicles?client_id="+clientID.String(), nil)
	rr := httptest.NewRecorder()

	handler.ListByClient(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestVehicleHandler_Create_InvalidJSON(t *testing.T) {
	handler := serviceHttp.NewVehicleHandler(nil)

	req, _ := http.NewRequest("POST", "/admin/vehicles", bytes.NewBuffer([]byte("{invalid")))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestVehicleHandler_Create_DomainError(t *testing.T) {
	mockRepo := new(MockVehicleRepository)
	handler := serviceHttp.NewVehicleHandler(mockRepo)

	clientID := uuid.New()
	reqBody := map[string]interface{}{
		"client_id": clientID.String(),
		"plate":     "ABC-1234",
		"brand":     "Toyota",
		"model":     "Corolla",
		"year":      1800, // Invalid year
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/admin/vehicles", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestVehicleHandler_Create_RepoError(t *testing.T) {
	mockRepo := new(MockVehicleRepository)
	handler := serviceHttp.NewVehicleHandler(mockRepo)

	clientID := uuid.New()
	reqBody := map[string]interface{}{
		"client_id": clientID.String(),
		"plate":     "ABC-1234",
		"brand":     "Toyota",
		"model":     "Corolla",
		"year":      2020,
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/admin/vehicles", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	mockRepo.On("Save", mock.Anything).Return(assert.AnError)

	handler.Create(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestVehicleHandler_ListByClient_RepoError(t *testing.T) {
	mockRepo := new(MockVehicleRepository)
	handler := serviceHttp.NewVehicleHandler(mockRepo)

	clientID := uuid.New()
	mockRepo.On("ListByClientID", clientID).Return(nil, assert.AnError)

	req, _ := http.NewRequest("GET", "/admin/vehicles?client_id="+clientID.String(), nil)
	rr := httptest.NewRecorder()

	handler.ListByClient(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestVehicleHandler_Create_InvalidUUID(t *testing.T) {
	handler := serviceHttp.NewVehicleHandler(nil)

	reqBody := map[string]interface{}{
		"client_id": "invalid-uuid",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/admin/vehicles", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestVehicleHandler_ListByClient_InvalidUUID(t *testing.T) {
	handler := serviceHttp.NewVehicleHandler(nil)

	req, _ := http.NewRequest("GET", "/admin/vehicles?client_id=invalid", nil)
	rr := httptest.NewRecorder()

	handler.ListByClient(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestVehicleHandler_ListByClient_MissingClientID(t *testing.T) {
	handler := serviceHttp.NewVehicleHandler(nil)

	req, _ := http.NewRequest("GET", "/admin/vehicles", nil)
	rr := httptest.NewRecorder()

	handler.ListByClient(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
