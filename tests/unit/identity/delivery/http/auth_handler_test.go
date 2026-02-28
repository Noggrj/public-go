package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	identityHttp "github.com/noggrj/autorepair/internal/identity/delivery/http"
	identityDomain "github.com/noggrj/autorepair/internal/identity/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Save(user *identityDomain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(email string) (*identityDomain.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*identityDomain.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*identityDomain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*identityDomain.User), args.Error(1)
}

// --- Tests ---

func TestAuthHandler_Login(t *testing.T) {
	mockRepo := new(MockUserRepository)
	// NewAuthHandler only takes repo now
	handler := identityHttp.NewAuthHandler(mockRepo)

	email := "test@test.com"
	password := "password123"
	user, _ := identityDomain.NewUser("Test User", email, password, "admin")

	mockRepo.On("GetByEmail", email).Return(user, nil)

	reqBody := map[string]string{
		"email":    email,
		"password": password,
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	handler.Login(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp["access_token"])
	assert.NotEmpty(t, resp["refresh_token"])
	assert.NotEmpty(t, resp["expires_in"])
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mockRepo := new(MockUserRepository)
	handler := identityHttp.NewAuthHandler(mockRepo)

	email := "test@test.com"
	password := "wrongpassword"
	user, _ := identityDomain.NewUser("Test User", email, "correctpassword", "admin")

	mockRepo.On("GetByEmail", email).Return(user, nil)

	reqBody := map[string]string{
		"email":    email,
		"password": password,
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	handler.Login(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
