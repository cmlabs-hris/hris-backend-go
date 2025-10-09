package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/auth"
	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/jwt"
)

type AuthHandler interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	LoginWithEmployeeCode(w http.ResponseWriter, r *http.Request)
	LoginWithGoogle(w http.ResponseWriter, r *http.Request)
	OAuthCallbackGoogle(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	RefreshToken(w http.ResponseWriter, r *http.Request)
	ForgotPassword(w http.ResponseWriter, r *http.Request)
	VerifyEmail(w http.ResponseWriter, r *http.Request)
}

type AuthHandlerImpl struct {
	jwtService  jwt.Service
	authService auth.AuthService
}

// ForgotPassword implements AuthHandler.
func (a *AuthHandlerImpl) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// Login implements AuthHandler.
func (a *AuthHandlerImpl) Login(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// LoginWithEmployeeCode implements AuthHandler.
func (a *AuthHandlerImpl) LoginWithEmployeeCode(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// LoginWithGoogle implements AuthHandler.
func (a *AuthHandlerImpl) LoginWithGoogle(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// Logout implements AuthHandler.
func (a *AuthHandlerImpl) Logout(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// OAuthCallbackGoogle implements AuthHandler.
func (a *AuthHandlerImpl) OAuthCallbackGoogle(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// RefreshToken implements AuthHandler.
func (a *AuthHandlerImpl) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// RefreshToken implements AuthHandler.
	panic("unimplemented")
}

// Register implements AuthHandler.
func (a *AuthHandlerImpl) Register(w http.ResponseWriter, r *http.Request) {
	var req auth.RegisterRequest

	// 1. Decode JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Register decode error", "error", err)
		response.BadRequest(w, "Invalid request format", nil)
		return
	}

	// Validate DTO
	if err := req.Validate(); err != nil {
		slog.Error("Register validate error", "error", err)
		response.HandleError(w, err)
		return
	}

	// Call service
	tokenResponse, err := a.authService.Register(r.Context(), req)
	if err != nil {
		slog.Error("Register service error", "error", err)
		response.HandleError(w, err)
		return
	}

	// Success response
	slog.Info("User registered successfully")
	response.Created(w, "User created successfully", tokenResponse)
}

// VerifyEmail implements AuthHandler.
func (a *AuthHandlerImpl) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

func NewAuthHandler(jwtService jwt.Service, authService auth.AuthService) AuthHandler {
	return &AuthHandlerImpl{
		jwtService:  jwtService,
		authService: authService,
	}
}
