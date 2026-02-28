package errors

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func JSON(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(ErrorResponse{Message: message}); err != nil {
		// Log encoding error but cannot change response at this point
		_ = err // ignore error
	}
}

func BadRequest(w http.ResponseWriter, message string) {
	JSON(w, http.StatusBadRequest, message)
}

func Unauthorized(w http.ResponseWriter, message string) {
	JSON(w, http.StatusUnauthorized, message)
}

func Forbidden(w http.ResponseWriter, message string) {
	JSON(w, http.StatusForbidden, message)
}

func NotFound(w http.ResponseWriter, message string) {
	JSON(w, http.StatusNotFound, message)
}

func InternalServerError(w http.ResponseWriter, message string) {
	JSON(w, http.StatusInternalServerError, message)
}
