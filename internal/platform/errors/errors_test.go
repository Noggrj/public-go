package errors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBadRequest(t *testing.T) {
	recorder := httptest.NewRecorder()
	BadRequest(recorder, "Invalid request")

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Invalid request")
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
}

func TestUnauthorized(t *testing.T) {
	recorder := httptest.NewRecorder()
	Unauthorized(recorder, "Unauthorized access")

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Unauthorized access")
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
}

func TestForbidden(t *testing.T) {
	recorder := httptest.NewRecorder()
	Forbidden(recorder, "Access forbidden")

	assert.Equal(t, http.StatusForbidden, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Access forbidden")
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
}

func TestNotFound(t *testing.T) {
	recorder := httptest.NewRecorder()
	NotFound(recorder, "Resource not found")

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Resource not found")
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
}

func TestInternalServerError(t *testing.T) {
	recorder := httptest.NewRecorder()
	InternalServerError(recorder, "Internal server error")

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Internal server error")
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
}

func TestJSON(t *testing.T) {
	recorder := httptest.NewRecorder()
	JSON(recorder, http.StatusOK, "Success message")

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Success message")
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
}

func TestJSON_WithEmptyMessage(t *testing.T) {
	recorder := httptest.NewRecorder()
	JSON(recorder, http.StatusNoContent, "")

	assert.Equal(t, http.StatusNoContent, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "")
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
}