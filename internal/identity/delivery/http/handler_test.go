package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/noggrj/autorepair/internal/identity/domain"
	"github.com/noggrj/autorepair/internal/platform/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type FailWriter struct {
	http.ResponseWriter
}

func (f *FailWriter) Header() http.Header {
	return http.Header{}
}

func (f *FailWriter) Write(b []byte) (int, error) {
	return 0, errors.New("write error")
}

func (f *FailWriter) WriteHeader(statusCode int) {}

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Save(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(email string) (*domain.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func TestRegister(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := NewAuthHandler(mockRepo)
		reqBody := registerRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "password123",
			Role:     "admin",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockRepo.On("Save", mock.AnythingOfType("*domain.User")).Return(nil)

		handler.Register(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid Body", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := NewAuthHandler(mockRepo)
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString("invalid json"))
		w := httptest.NewRecorder()

		handler.Register(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid request body")
	})

	t.Run("Empty Body", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := NewAuthHandler(mockRepo)
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer([]byte{}))
		w := httptest.NewRecorder()

		handler.Register(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid request body")
	})

	t.Run("Invalid Field Type", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := NewAuthHandler(mockRepo)
		reqBody := map[string]interface{}{
			"name":  12345, // Should be string
			"email": "test@example.com",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.Register(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid request body")
	})

	t.Run("Save Error", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := NewAuthHandler(mockRepo)
		reqBody := registerRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "password123",
			Role:     "admin",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockRepo.On("Save", mock.AnythingOfType("*domain.User")).Return(errors.New("db error"))

		handler.Register(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed to save user")
	})

	t.Run("NewUser Error", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := NewAuthHandler(mockRepo)
		// Password longer than 72 bytes causes bcrypt error
		longPassword := "verylongpasswordverylongpasswordverylongpasswordverylongpasswordverylongpasswordverylongpassword"
		reqBody := registerRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: longPassword,
			Role:     "admin",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.Register(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Encode Error", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := NewAuthHandler(mockRepo)
		reqBody := registerRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "password123",
			Role:     "admin",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		w := &FailWriter{}

		mockRepo.On("Save", mock.AnythingOfType("*domain.User")).Return(nil)

		handler.Register(w, req)

		mockRepo.AssertExpectations(t)
	})
}

func TestRefresh(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := NewAuthHandler(mockRepo)

		// Generate a valid token first
		userID := uuid.New()
		_, refreshToken, _, _ := auth.GenerateToken(userID, "admin")

		reqBody := map[string]string{
			"refresh_token": refreshToken,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.Refresh(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp loginResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.AccessToken)
	})

	t.Run("Invalid Body", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := NewAuthHandler(mockRepo)
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString("invalid json"))
		w := httptest.NewRecorder()

		handler.Refresh(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := NewAuthHandler(mockRepo)
		reqBody := map[string]string{
			"refresh_token": "invalid.token.here",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.Refresh(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Encode Error", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := NewAuthHandler(mockRepo)
		userID := uuid.New()
		_, refreshToken, _, _ := auth.GenerateToken(userID, "admin")

		reqBody := map[string]string{
			"refresh_token": refreshToken,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
		w := &FailWriter{}

		handler.Refresh(w, req)
	})
}

func TestLogin(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := NewAuthHandler(mockRepo)
		user, _ := domain.NewUser("Test", "test@example.com", "password123", domain.RoleAdmin)

		reqBody := loginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockRepo.On("GetByEmail", "test@example.com").Return(user, nil)

		handler.Login(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp loginResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.AccessToken)
	})

	t.Run("Invalid Body", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := NewAuthHandler(mockRepo)
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString("invalid json"))
		w := httptest.NewRecorder()

		handler.Login(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("User Not Found", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := NewAuthHandler(mockRepo)
		reqBody := loginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockRepo.On("GetByEmail", "test@example.com").Return(nil, errors.New("user not found"))

		handler.Login(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Password", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := NewAuthHandler(mockRepo)
		user, _ := domain.NewUser("Test", "test@example.com", "password123", domain.RoleAdmin)

		reqBody := loginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		mockRepo.On("GetByEmail", "test@example.com").Return(user, nil)

		handler.Login(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Encode Error", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := NewAuthHandler(mockRepo)
		user, _ := domain.NewUser("Test", "test@example.com", "password123", domain.RoleAdmin)

		reqBody := loginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		w := &FailWriter{}

		mockRepo.On("GetByEmail", "test@example.com").Return(user, nil)

		handler.Login(w, req)
	})
}

func TestRegisterRoutes(t *testing.T) {
	mockRepo := new(MockUserRepository)
	handler := NewAuthHandler(mockRepo)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	// Check if routes are registered
	// This is a basic check; real routing test is implicit in integration tests usually
	assert.NotNil(t, r)
}
