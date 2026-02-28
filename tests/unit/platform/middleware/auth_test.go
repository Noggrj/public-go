package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/noggrj/autorepair/internal/platform/auth"
	"github.com/noggrj/autorepair/internal/platform/middleware"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	userID := uuid.New()
	token, _, _, _ := auth.GenerateToken(userID, "admin")

	mw := middleware.AuthMiddleware
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(middleware.UserContextKey).(*auth.Claims)
		assert.True(t, ok)
		assert.Equal(t, userID, claims.UserID)
		w.WriteHeader(http.StatusOK)
	})

	// Case 1: Valid Token
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	mw(nextHandler).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Case 2: No Header
	req, _ = http.NewRequest("GET", "/", nil)
	rr = httptest.NewRecorder()
	mw(nextHandler).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	// Case 3: Invalid Header Format
	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Token "+token)
	rr = httptest.NewRecorder()
	mw(nextHandler).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	// Case 4: Invalid Token
	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr = httptest.NewRecorder()
	mw(nextHandler).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRequireRole(t *testing.T) {
	mw := middleware.RoleMiddleware("admin")
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Case 1: Valid Role
	req, _ := http.NewRequest("GET", "/", nil)
	claims := &auth.Claims{Role: "admin"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	rr := httptest.NewRecorder()
	mw(nextHandler).ServeHTTP(rr, req.WithContext(ctx))
	assert.Equal(t, http.StatusOK, rr.Code)

	// Case 2: Invalid Role
	mwUser := middleware.RoleMiddleware("user") // requires 'user', but we pass 'guest' (if admin logic was strict, but admin overrides)
	// Actually logic is: if claims.Role != role && claims.Role != "admin"

	// Sub-case 2a: Wrong role, not admin
	req, _ = http.NewRequest("GET", "/", nil)
	claims = &auth.Claims{Role: "guest"}
	ctx = context.WithValue(req.Context(), middleware.UserContextKey, claims)
	rr = httptest.NewRecorder()
	mwUser(nextHandler).ServeHTTP(rr, req.WithContext(ctx))
	assert.Equal(t, http.StatusForbidden, rr.Code)

	// Sub-case 2b: Admin can access 'user' role route
	req, _ = http.NewRequest("GET", "/", nil)
	claims = &auth.Claims{Role: "admin"}
	ctx = context.WithValue(req.Context(), middleware.UserContextKey, claims)
	rr = httptest.NewRecorder()
	mwUser(nextHandler).ServeHTTP(rr, req.WithContext(ctx))
	assert.Equal(t, http.StatusOK, rr.Code)

	// Case 3: No Claims
	req, _ = http.NewRequest("GET", "/", nil)
	rr = httptest.NewRecorder()
	mw(nextHandler).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
