package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	serviceHttp "github.com/noggrj/autorepair/internal/service/delivery/http"
	serviceDomain "github.com/noggrj/autorepair/internal/service/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClientHandler_Create(t *testing.T) {
	mockRepo := new(MockClientRepository)
	handler := serviceHttp.NewClientHandler(mockRepo)

	reqBody := map[string]string{
		"name":     "John Doe",
		"document": "52998224725", // Generated valid CPF
		"email":    "john@example.com",
		"phone":    "123456789",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/admin/clients", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	mockRepo.On("Save", mock.Anything).Return(nil)

	handler.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestClientHandler_List(t *testing.T) {
	mockRepo := new(MockClientRepository)
	handler := serviceHttp.NewClientHandler(mockRepo)

	clients := []*serviceDomain.Client{
		{Name: "John Doe"},
	}
	mockRepo.On("List").Return(clients, nil)

	req, _ := http.NewRequest("GET", "/admin/clients", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestClientHandler_Create_InvalidJSON(t *testing.T) {
	handler := serviceHttp.NewClientHandler(nil)

	req, _ := http.NewRequest("POST", "/admin/clients", bytes.NewBuffer([]byte("{invalid")))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestClientHandler_List_Error(t *testing.T) {
	mockRepo := new(MockClientRepository)
	handler := serviceHttp.NewClientHandler(mockRepo)

	mockRepo.On("List").Return(nil, assert.AnError)

	req, _ := http.NewRequest("GET", "/admin/clients", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
