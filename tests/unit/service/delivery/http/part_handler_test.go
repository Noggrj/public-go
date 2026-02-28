package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	inventoryDomain "github.com/noggrj/autorepair/internal/inventory/domain"
	serviceHttp "github.com/noggrj/autorepair/internal/service/delivery/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPartHandler_Create(t *testing.T) {
	mockRepo := new(MockPartRepository)
	handler := serviceHttp.NewPartHandler(mockRepo)

	reqBody := map[string]interface{}{
		"name":        "Oil Filter",
		"description": "Filter for Toyota",
		"price":       50.0,
		"stock_qty":   10,
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/admin/parts", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	// Update to use context match
	mockRepo.On("Save", mock.Anything, mock.Anything).Return(nil)

	handler.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestPartHandler_List(t *testing.T) {
	mockRepo := new(MockPartRepository)
	handler := serviceHttp.NewPartHandler(mockRepo)

	parts := []*inventoryDomain.Part{
		{Name: "Oil Filter"},
	}
	// Update to use context match
	mockRepo.On("List", mock.Anything).Return(parts, nil)

	req, _ := http.NewRequest("GET", "/admin/parts", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestPartHandler_Create_InvalidJSON(t *testing.T) {
	handler := serviceHttp.NewPartHandler(nil)

	req, _ := http.NewRequest("POST", "/admin/parts", bytes.NewBuffer([]byte("{invalid")))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestPartHandler_List_Error(t *testing.T) {
	mockRepo := new(MockPartRepository)
	handler := serviceHttp.NewPartHandler(mockRepo)

	// Update to use context match
	mockRepo.On("List", mock.Anything).Return(nil, assert.AnError)

	req, _ := http.NewRequest("GET", "/admin/parts", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
