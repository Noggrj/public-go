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

func TestServiceHandler_Create(t *testing.T) {
	mockRepo := new(MockServiceRepository)
	handler := serviceHttp.NewServiceHandler(mockRepo)

	reqBody := map[string]interface{}{
		"name":        "Oil Change",
		"description": "Standard oil change",
		"price":       100.0,
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/admin/services", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	mockRepo.On("Save", mock.Anything).Return(nil)

	handler.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	mockRepo.AssertExpectations(t)
}

func TestServiceHandler_List(t *testing.T) {
	mockRepo := new(MockServiceRepository)
	handler := serviceHttp.NewServiceHandler(mockRepo)

	services := []*serviceDomain.Service{
		{Name: "Oil Change"},
	}
	mockRepo.On("List").Return(services, nil)

	req, _ := http.NewRequest("GET", "/admin/services", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestServiceHandler_Create_InvalidJSON(t *testing.T) {
	handler := serviceHttp.NewServiceHandler(nil)

	req, _ := http.NewRequest("POST", "/admin/services", bytes.NewBuffer([]byte("{invalid")))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestServiceHandler_List_Error(t *testing.T) {
	mockRepo := new(MockServiceRepository)
	handler := serviceHttp.NewServiceHandler(mockRepo)

	mockRepo.On("List").Return(nil, assert.AnError)

	req, _ := http.NewRequest("GET", "/admin/services", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
