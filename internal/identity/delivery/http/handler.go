package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/noggrj/autorepair/internal/identity/domain"
	"github.com/noggrj/autorepair/internal/platform/auth"
	"github.com/noggrj/autorepair/internal/platform/errors"
)

type AuthHandler struct {
	repo domain.UserRepository
}

func NewAuthHandler(repo domain.UserRepository) *AuthHandler {
	return &AuthHandler{repo: repo}
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with name, email, password and role
// @Tags auth
// @Accept json
// @Produce json
// @Param request body registerRequest true "Register Request"
// @Success 201 {object} domain.User
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.BadRequest(w, "invalid request body")
		return
	}

	user, err := domain.NewUser(req.Name, req.Email, req.Password, domain.Role(req.Role))
	if err != nil {
		errors.InternalServerError(w, err.Error())
		return
	}

	if err := h.repo.Save(user); err != nil {
		errors.InternalServerError(w, "failed to save user")
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		errors.InternalServerError(w, "failed to encode response")
		return
	}
}

type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Login godoc
// @Summary Login
// @Description Login with email and password to get a JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body loginRequest true "Login Request"
// @Success 200 {object} loginResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.BadRequest(w, "invalid request body")
		return
	}

	user, err := h.repo.GetByEmail(req.Email)
	if err != nil {
		errors.Unauthorized(w, "invalid credentials")
		return
	}

	if !user.CheckPassword(req.Password) {
		errors.Unauthorized(w, "invalid credentials")
		return
	}

	accessToken, refreshToken, expiresIn, err := auth.GenerateToken(user.ID, string(user.Role))
	if err != nil {
		errors.InternalServerError(w, "failed to generate token")
		return
	}

	if err := json.NewEncoder(w).Encode(loginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}); err != nil {
		errors.InternalServerError(w, "failed to encode response")
		return
	}
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Refresh godoc
// @Summary Refresh Access Token
// @Description Use a refresh token to get a new access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body refreshTokenRequest true "Refresh Token Request"
// @Success 200 {object} loginResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.BadRequest(w, "invalid request body")
		return
	}

	claims, err := auth.ValidateToken(req.RefreshToken)
	if err != nil {
		errors.Unauthorized(w, "invalid refresh token")
		return
	}

	// In a real app, we might check if the user is still active or if token is revoked here.

	accessToken, refreshToken, expiresIn, err := auth.GenerateToken(claims.UserID, claims.Role)
	if err != nil {
		errors.InternalServerError(w, "failed to generate token")
		return
	}

	if err := json.NewEncoder(w).Encode(loginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}); err != nil {
		errors.InternalServerError(w, "failed to encode response")
		return
	}
}

func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.Refresh)
}
