package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/auth"
	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/jwt"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/oauth"
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
	jwtService    jwt.Service
	authService   auth.AuthService
	googleService oauth.GoogleService
}

// ForgotPassword implements AuthHandler.
func (a *AuthHandlerImpl) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// Login implements AuthHandler.
func (a *AuthHandlerImpl) Login(w http.ResponseWriter, r *http.Request) {
	var loginReq auth.LoginRequest

	// 1. Decode JSON
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		slog.Error("Login decode error", "error", err)
		response.BadRequest(w, "Invalid request format", nil)
		return
	}

	// Validate DTO
	if err := loginReq.Validate(); err != nil {
		slog.Error("Login validate error", "error", err)
		response.HandleError(w, err)
		return
	}

	// Call service
	var sessionTrackReq auth.SessionTrackingRequest
	sessionTrackReq.IPAddress = r.RemoteAddr
	sessionTrackReq.UserAgent = r.UserAgent()
	tokenResponse, err := a.authService.Login(r.Context(), loginReq, sessionTrackReq)
	if err != nil {
		slog.Error("Login service error", "error", err)
		response.HandleError(w, err)
		return
	}

	// Success response
	refreshTokenCookie := a.jwtService.RefreshTokenCookie(tokenResponse.RefreshToken, tokenResponse.RefreshTokenExpiresIn)
	http.SetCookie(w, refreshTokenCookie)
	slog.Info("User logged in successfully")
	response.Created(w, "User logged in successfully", tokenResponse)
}

// LoginWithEmployeeCode implements AuthHandler.
func (a *AuthHandlerImpl) LoginWithEmployeeCode(w http.ResponseWriter, r *http.Request) {
	var loginEmployeeCodeReq auth.LoginEmployeeCodeRequest

	// 1. Decode JSON
	if err := json.NewDecoder(r.Body).Decode(&loginEmployeeCodeReq); err != nil {
		slog.Error("Login employee code decode error", "error", err)
		response.BadRequest(w, "Invalid request format", nil)
		return
	}

	// Validate DTO
	if err := loginEmployeeCodeReq.Validate(); err != nil {
		slog.Error("Login employee code validate error", "error", err)
		response.HandleError(w, err)
		return
	}

	// Call service
	var sessionTrackReq auth.SessionTrackingRequest
	sessionTrackReq.IPAddress = r.RemoteAddr
	sessionTrackReq.UserAgent = r.UserAgent()
	tokenResponse, err := a.authService.LoginWithEmployeeCode(r.Context(), loginEmployeeCodeReq, sessionTrackReq)
	if err != nil {
		slog.Error("Login service error", "error", err)
		response.HandleError(w, err)
		return
	}

	// Success response
	refreshTokenCookie := a.jwtService.RefreshTokenCookie(tokenResponse.RefreshToken, tokenResponse.RefreshTokenExpiresIn)
	http.SetCookie(w, refreshTokenCookie)
	slog.Info("User logged in successfully")
	response.Created(w, "User logged in successfully", tokenResponse)
}

// LoginWithGoogle implements AuthHandler.
func (a *AuthHandlerImpl) LoginWithGoogle(w http.ResponseWriter, r *http.Request) {
	state := a.googleService.GenerateState(r.UserAgent())
	cookie := &http.Cookie{
		Name:     "state",
		Value:    state,
		Path:     "/api/v1/auth/oauth/callback/google",
		Expires:  time.Now().Add(5 * time.Minute),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
	url := a.googleService.RedirectURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Logout implements AuthHandler.
func (a *AuthHandlerImpl) Logout(w http.ResponseWriter, r *http.Request) {
	refreshTokenCookieReq, err := r.Cookie("refresh_token")
	if err != nil {
		response.HandleError(w, auth.ErrRefreshTokenCookieNotFound)
		return
	}
	refreshToken := refreshTokenCookieReq.Value
	if refreshToken == "" {
		response.HandleError(w, auth.ErrRefreshTokenCookieEmpty)
		return
	}

	err = a.authService.Logout(r.Context(), refreshToken)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	// Clear the refresh token cookie
	clearedCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, clearedCookie)
	response.Success(w, "User logged out successfully")
}

// OAuthCallbackGoogle implements AuthHandler.
func (a *AuthHandlerImpl) OAuthCallbackGoogle(w http.ResponseWriter, r *http.Request) {
	stateReq, err := r.Cookie("state")
	if err != nil {
		response.HandleError(w, auth.ErrStateCookieNotFound)
		return
	}
	errorValue := r.URL.Query().Get("error")
	if errorValue == "access_denied" {
		response.HandleError(w, auth.ErrGoogleAccessDeniedByUser)
		return
	}
	if errorValue != "" {
		response.HandleError(w, err)
		return
	}

	stateCookie := stateReq.Value
	if stateCookie == "" {
		response.HandleError(w, auth.ErrStateCookieEmpty)
		return
	}

	stateParam := r.URL.Query().Get("state")
	if stateParam == "" {
		response.HandleError(w, auth.ErrStateParamEmpty)
		return
	}

	if stateParam != stateCookie {
		response.HandleError(w, auth.ErrStateMismatch)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		response.HandleError(w, auth.ErrCodeValueEmpty)
		return
	}

	token, err := a.googleService.VerifyToken(r.Context(), code)

	if err != nil {
		response.HandleError(w, err)
		return
	}

	userGoogle, err := a.googleService.VerifyUser(r.Context(), token)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	var sessionTrackReq auth.SessionTrackingRequest
	sessionTrackReq.IPAddress = r.RemoteAddr
	sessionTrackReq.UserAgent = r.UserAgent()
	tokenResponse, err := a.authService.LoginWithGoogle(r.Context(), userGoogle.Email, userGoogle.GoogleID, sessionTrackReq)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	refreshTokenCookie := a.jwtService.RefreshTokenCookie(tokenResponse.RefreshToken, tokenResponse.RefreshTokenExpiresIn)
	http.SetCookie(w, refreshTokenCookie)
	slog.Info("User logged in successfully")
	response.Created(w, "User logged in successfully", tokenResponse)
}

// RefreshToken implements AuthHandler.
func (a *AuthHandlerImpl) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// RefreshToken implements AuthHandler.
	var refreshTokenReq auth.RefreshTokenRequest

	// 1. Decode JSON
	if err := json.NewDecoder(r.Body).Decode(&refreshTokenReq); err != nil {
		slog.Error("Refresh Token decode error", "error", err)
		response.BadRequest(w, "Invalid request format", nil)
		return
	}

	// Validate DTO
	if err := refreshTokenReq.Validate(); err != nil {
		slog.Error("Refresh Token validate error", "error", err)
		response.HandleError(w, err)
		return
	}

	// Call service
	tokenResponse, err := a.authService.RefreshToken(r.Context(), refreshTokenReq)
	if err != nil {
		slog.Error("Refresh Token service error", "error", err)
		response.HandleError(w, err)
		return
	}

	// Success response
	slog.Info("Token refreshed successfully")
	response.Created(w, "Token refreshed successfully", tokenResponse)
}

// Register implements AuthHandler.
func (a *AuthHandlerImpl) Register(w http.ResponseWriter, r *http.Request) {
	var registerReq auth.RegisterRequest

	// 1. Decode JSON
	if err := json.NewDecoder(r.Body).Decode(&registerReq); err != nil {
		slog.Error("Register decode error", "error", err)
		response.BadRequest(w, "Invalid request format", nil)
		return
	}

	// Validate DTO
	if err := registerReq.Validate(); err != nil {
		slog.Error("Register validate error", "error", err)
		response.HandleError(w, err)
		return
	}

	// Call service
	var sessionTrackReq auth.SessionTrackingRequest
	sessionTrackReq.IPAddress = r.RemoteAddr
	sessionTrackReq.UserAgent = r.UserAgent()
	tokenResponse, err := a.authService.Register(r.Context(), registerReq, sessionTrackReq)
	if err != nil {
		slog.Error("Register service error", "error", err)
		response.HandleError(w, err)
		return
	}

	// Success response
	refreshTokenCookie := a.jwtService.RefreshTokenCookie(tokenResponse.RefreshToken, tokenResponse.RefreshTokenExpiresIn)
	http.SetCookie(w, refreshTokenCookie)
	slog.Info("User registered successfully")
	response.Created(w, "User created successfully", tokenResponse)
}

// VerifyEmail implements AuthHandler.
func (a *AuthHandlerImpl) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

func NewAuthHandler(jwtService jwt.Service, authService auth.AuthService, googleService oauth.GoogleService) AuthHandler {
	return &AuthHandlerImpl{
		jwtService:    jwtService,
		authService:   authService,
		googleService: googleService,
	}
}
