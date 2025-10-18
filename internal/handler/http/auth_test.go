package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/auth"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/jwt"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/oauth"
	"github.com/cmlabs-hris/hris-backend-go/internal/repository/postgresql"
	authService "github.com/cmlabs-hris/hris-backend-go/internal/service/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

var (
	testHandlerDB *database.DB
)

const (
	handlerTestAccessExp  = "1h"
	handlerTestRefreshExp = "24h"
	handlerTestSecret     = "test-secret-key-for-jwt"
)

func handlerTestInit() {
	if testHandlerDB != nil {
		return
	}
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:root@localhost:5432/cmlabs_hris_test?sslmode=disable"
	}

	var err error
	testHandlerDB, err = database.NewPostgreSQLDB(dsn)
	if err != nil {
		panic("Failed to connect to test database: " + err.Error())
	}
}

func truncateHandlerTables(t *testing.T, ctx context.Context) {
	handlerTestInit()
	tables := []string{"refresh_tokens", "users", "companies", "employees"}

	for _, table := range tables {
		_, err := testHandlerDB.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			// Some tables might not exist, skip
			continue
		}
	}
}

func createHandlerTestCompany(t *testing.T, ctx context.Context) string {
	handlerTestInit()
	var companyID string
	uniqueUsername := fmt.Sprintf("test-company-%d-%d", time.Now().Unix(), time.Now().Nanosecond())
	err := testHandlerDB.QueryRow(ctx, `
		INSERT INTO companies (id, name, username, created_at, updated_at)
		VALUES (uuidv7(), 'Test Company', $1, NOW(), NOW())
		RETURNING id
	`, uniqueUsername).Scan(&companyID)
	require.NoError(t, err)
	return companyID
}

func createHandlerTestUser(t *testing.T, ctx context.Context, companyID string, email string) string {
	var userID string
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	hashedStr := string(hashedPassword)

	err := testHandlerDB.QueryRow(ctx, `
		INSERT INTO users (company_id, email, password_hash, is_admin, email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, false, true, NOW(), NOW())
		RETURNING id
	`, companyID, email, hashedStr).Scan(&userID)
	require.NoError(t, err)
	return userID
}

func createAuthHandler(t *testing.T, ctx context.Context) AuthHandler {
	userRepo := postgresql.NewUserRepository(testHandlerDB)
	companyRepo := postgresql.NewCompanyRepository(testHandlerDB)
	jwtSvc := jwt.NewJWTService(handlerTestSecret, handlerTestAccessExp, handlerTestRefreshExp)
	jwtRepo := postgresql.NewJWTRepository(testHandlerDB)
	authSvc := authService.NewAuthService(testHandlerDB, userRepo, companyRepo, jwtSvc, jwtRepo)

	// Use real GoogleService - OAuth endpoints will fail but that's OK for handler tests
	googleSvc := oauth.NewGoogleService("test-client-id", "test-client-secret", "http://localhost:3000/callback", []string{"email"})

	return NewAuthHandler(jwtSvc, authSvc, googleSvc)
}

// ===== HANDLER TESTS =====

// Test Register - Success
func TestAuthHandler_Register_Success(t *testing.T) {
	ctx := context.Background()
	handlerTestInit()
	truncateHandlerTables(t, ctx)

	handler := createAuthHandler(t, ctx)

	// Create request
	testEmail := fmt.Sprintf("register-%d@example.com", time.Now().UnixNano())
	registerReq := auth.RegisterRequest{
		Email:           testEmail,
		Password:        "SecurePass123!",
		ConfirmPassword: "SecurePass123!",
	}
	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Act
	handler.Register(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.True(t, resp["success"].(bool))
	assert.NotNil(t, resp["data"])

	// Verify response contains tokens
	data := resp["data"].(map[string]interface{})
	assert.NotEmpty(t, data["access_token"])
	assert.NotEmpty(t, data["refresh_token"])
}

// Test Register - Invalid Password Mismatch
func TestAuthHandler_Register_PasswordMismatch(t *testing.T) {
	ctx := context.Background()
	handlerTestInit()
	truncateHandlerTables(t, ctx)

	handler := createAuthHandler(t, ctx)

	// Create request with mismatched passwords
	testEmail := fmt.Sprintf("register-mismatch-%d@example.com", time.Now().UnixNano())
	registerReq := auth.RegisterRequest{
		Email:           testEmail,
		Password:        "SecurePass123!",
		ConfirmPassword: "DifferentPass123!",
	}
	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Act
	handler.Register(w, req)

	// Assert - Should get error
	assert.NotEqual(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// Test Register - Invalid JSON
func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	handlerTestInit()

	handler := createAuthHandler(t, ctx)

	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader([]byte("invalid json")))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Act
	handler.Register(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test Login - Success
func TestAuthHandler_Login_Success(t *testing.T) {
	ctx := context.Background()
	handlerTestInit()
	truncateHandlerTables(t, ctx)

	// Setup
	companyID := createHandlerTestCompany(t, ctx)
	testEmail := fmt.Sprintf("login-%d@example.com", time.Now().UnixNano())
	createHandlerTestUser(t, ctx, companyID, testEmail)

	handler := createAuthHandler(t, ctx)

	// Create request
	loginReq := auth.LoginRequest{
		Email:    testEmail,
		Password: "password123",
	}
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Act
	handler.Login(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.True(t, resp["success"].(bool))

	// Verify tokens in response
	data := resp["data"].(map[string]interface{})
	assert.NotEmpty(t, data["access_token"])
	assert.NotEmpty(t, data["refresh_token"])

	// Verify refresh token cookie is set
	cookies := w.Result().Cookies()
	var refreshTokenCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" {
			refreshTokenCookie = cookie
			break
		}
	}
	assert.NotNil(t, refreshTokenCookie)
	assert.NotEmpty(t, refreshTokenCookie.Value)
}

// Test Login - Invalid Credentials
func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	ctx := context.Background()
	handlerTestInit()
	truncateHandlerTables(t, ctx)

	// Setup
	companyID := createHandlerTestCompany(t, ctx)
	testEmail := fmt.Sprintf("login-invalid-%d@example.com", time.Now().UnixNano())
	createHandlerTestUser(t, ctx, companyID, testEmail)

	handler := createAuthHandler(t, ctx)

	// Create request with wrong password
	loginReq := auth.LoginRequest{
		Email:    testEmail,
		Password: "wrongpassword",
	}
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Act
	handler.Login(w, req)

	// Assert
	assert.NotEqual(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// Test Login - User Not Found
func TestAuthHandler_Login_UserNotFound(t *testing.T) {
	ctx := context.Background()
	handlerTestInit()
	truncateHandlerTables(t, ctx)

	handler := createAuthHandler(t, ctx)

	// Create request
	loginReq := auth.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Act
	handler.Login(w, req)

	// Assert
	assert.NotEqual(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// Test Login - Invalid JSON
func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	handlerTestInit()

	handler := createAuthHandler(t, ctx)

	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte("invalid json")))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Act
	handler.Login(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test LoginWithEmployeeCode - Skipped
func TestAuthHandler_LoginWithEmployeeCode_Success(t *testing.T) {
	t.Skip("Skipping - Employee table requires complex setup with positions, grades, branches, etc.")
}

// Test LoginWithGoogle - Redirect
func TestAuthHandler_LoginWithGoogle_Redirect(t *testing.T) {
	ctx := context.Background()
	handlerTestInit()

	handler := createAuthHandler(t, ctx)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/oauth/login/google", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Act
	handler.LoginWithGoogle(w, req)

	// Assert
	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)

	// Verify state cookie is set
	cookies := w.Result().Cookies()
	var stateCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "state" {
			stateCookie = cookie
			break
		}
	}
	assert.NotNil(t, stateCookie)
	assert.NotEmpty(t, stateCookie.Value)

	// Verify redirect location
	assert.NotEmpty(t, w.Header().Get("Location"))
}

// Test OAuthCallbackGoogle - Not Implemented
func TestAuthHandler_OAuthCallbackGoogle_NotImplemented(t *testing.T) {
	t.Skip("OAuthCallbackGoogle implementation requires full Google OAuth flow testing")
}

// Test Logout - Success
func TestAuthHandler_Logout_Success(t *testing.T) {
	ctx := context.Background()
	handlerTestInit()
	truncateHandlerTables(t, ctx)

	// Setup - First login to get token
	companyID := createHandlerTestCompany(t, ctx)
	testEmail := fmt.Sprintf("logout-%d@example.com", time.Now().UnixNano())
	createHandlerTestUser(t, ctx, companyID, testEmail)

	handler := createAuthHandler(t, ctx)

	// Login first
	loginReq := auth.LoginRequest{
		Email:    testEmail,
		Password: "password123",
	}
	loginBody, _ := json.Marshal(loginReq)
	loginReqHttp := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginReqHttp = loginReqHttp.WithContext(ctx)
	loginW := httptest.NewRecorder()
	handler.Login(loginW, loginReqHttp)

	// Extract refresh token from login response
	var loginResp map[string]interface{}
	json.NewDecoder(loginW.Body).Decode(&loginResp)
	refreshToken := loginResp["data"].(map[string]interface{})["refresh_token"].(string)

	// Create logout request with refresh token cookie
	logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	logoutReq = logoutReq.WithContext(ctx)
	logoutReq.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: refreshToken,
	})
	logoutW := httptest.NewRecorder()

	// Act
	handler.Logout(logoutW, logoutReq)

	// Assert
	assert.Equal(t, http.StatusOK, logoutW.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(logoutW.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.True(t, resp["success"].(bool))

	// Verify refresh token cookie is cleared
	cookies := logoutW.Result().Cookies()
	var refreshTokenCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" {
			refreshTokenCookie = cookie
			break
		}
	}
	assert.NotNil(t, refreshTokenCookie)
	assert.Empty(t, refreshTokenCookie.Value)
}

// Test Logout - No Cookie
func TestAuthHandler_Logout_NoCookie(t *testing.T) {
	ctx := context.Background()
	handlerTestInit()

	handler := createAuthHandler(t, ctx)

	// Create logout request without refresh token cookie
	logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	logoutReq = logoutReq.WithContext(ctx)
	logoutW := httptest.NewRecorder()

	// Act
	handler.Logout(logoutW, logoutReq)

	// Assert - Should get error
	assert.NotEqual(t, http.StatusOK, logoutW.Code)
}

// Test RefreshToken - Success
func TestAuthHandler_RefreshToken_Success(t *testing.T) {
	ctx := context.Background()
	handlerTestInit()
	truncateHandlerTables(t, ctx)

	// Setup - First login to get token
	companyID := createHandlerTestCompany(t, ctx)
	testEmail := fmt.Sprintf("refresh-%d@example.com", time.Now().UnixNano())
	createHandlerTestUser(t, ctx, companyID, testEmail)

	handler := createAuthHandler(t, ctx)

	// Login first
	loginReq := auth.LoginRequest{
		Email:    testEmail,
		Password: "password123",
	}
	loginBody, _ := json.Marshal(loginReq)
	loginReqHttp := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginReqHttp = loginReqHttp.WithContext(ctx)
	loginW := httptest.NewRecorder()
	handler.Login(loginW, loginReqHttp)

	// Extract refresh token from login response
	var loginResp map[string]interface{}
	json.NewDecoder(loginW.Body).Decode(&loginResp)
	refreshToken := loginResp["data"].(map[string]interface{})["refresh_token"].(string)

	// Create refresh token request
	refreshReq := auth.RefreshTokenRequest{RefreshToken: refreshToken}
	refreshBody, _ := json.Marshal(refreshReq)
	refreshReqHttp := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(refreshBody))
	refreshReqHttp = refreshReqHttp.WithContext(ctx)
	refreshW := httptest.NewRecorder()

	// Act
	handler.RefreshToken(refreshW, refreshReqHttp)

	// Assert
	assert.Equal(t, http.StatusCreated, refreshW.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(refreshW.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.True(t, resp["success"].(bool))

	// Verify new access token in response
	data := resp["data"].(map[string]interface{})
	assert.NotEmpty(t, data["access_token"])
}

// Test RefreshToken - Invalid Token
func TestAuthHandler_RefreshToken_InvalidToken(t *testing.T) {
	ctx := context.Background()
	handlerTestInit()

	handler := createAuthHandler(t, ctx)

	// Create refresh token request with invalid token
	refreshReq := auth.RefreshTokenRequest{RefreshToken: "invalid-token"}
	refreshBody, _ := json.Marshal(refreshReq)
	refreshReqHttp := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(refreshBody))
	refreshReqHttp = refreshReqHttp.WithContext(ctx)
	refreshW := httptest.NewRecorder()

	// Act
	handler.RefreshToken(refreshW, refreshReqHttp)

	// Assert - Should get error
	assert.NotEqual(t, http.StatusCreated, refreshW.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(refreshW.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// Test RefreshToken - Invalid JSON
func TestAuthHandler_RefreshToken_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	handlerTestInit()

	handler := createAuthHandler(t, ctx)

	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader([]byte("invalid json")))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Act
	handler.RefreshToken(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test ForgotPassword - Not Implemented
func TestAuthHandler_ForgotPassword_NotImplemented(t *testing.T) {
	t.Skip("Method not yet implemented in AuthHandler")
}

// Test VerifyEmail - Not Implemented
func TestAuthHandler_VerifyEmail_NotImplemented(t *testing.T) {
	t.Skip("Method not yet implemented in AuthHandler")
}

// ===== RESPONSE HELPER TESTS =====

// Test that responses are properly formatted
func TestAuthHandler_ResponseFormat_Success(t *testing.T) {
	ctx := context.Background()
	handlerTestInit()
	truncateHandlerTables(t, ctx)

	handler := createAuthHandler(t, ctx)

	// Create request
	testEmail := fmt.Sprintf("response-%d@example.com", time.Now().UnixNano())
	registerReq := auth.RegisterRequest{
		Email:           testEmail,
		Password:        "SecurePass123!",
		ConfirmPassword: "SecurePass123!",
	}
	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Act
	handler.Register(w, req)

	// Assert - Check Content-Type
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Verify response structure
	var resp map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Contains(t, resp, "success")
	assert.Contains(t, resp, "data")
}

// Test that error responses are properly formatted
func TestAuthHandler_ResponseFormat_Error(t *testing.T) {
	ctx := context.Background()
	handlerTestInit()

	handler := createAuthHandler(t, ctx)

	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte("invalid")))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Act
	handler.Login(w, req)

	// Assert - Check Content-Type
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Verify error response structure
	var resp map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Contains(t, resp, "success")
	assert.False(t, resp["success"].(bool))
}

// ===== SESSION TRACKING TESTS =====

// Test that session tracking info is captured (IP and User-Agent)
func TestAuthHandler_SessionTracking_IPAndUserAgent(t *testing.T) {
	ctx := context.Background()
	handlerTestInit()
	truncateHandlerTables(t, ctx)

	// Setup
	companyID := createHandlerTestCompany(t, ctx)
	testEmail := fmt.Sprintf("session-%d@example.com", time.Now().UnixNano())
	createHandlerTestUser(t, ctx, companyID, testEmail)

	handler := createAuthHandler(t, ctx)

	// Create request with IP and User-Agent
	loginReq := auth.LoginRequest{
		Email:    testEmail,
		Password: "password123",
	}
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req = req.WithContext(ctx)
	req.Header.Set("User-Agent", "Mozilla/5.0 Test Browser")
	req.RemoteAddr = "192.168.1.100"
	w := httptest.NewRecorder()

	// Act
	handler.Login(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	// Session info is captured inside handler (verify at database level in service tests)
	var resp map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.True(t, resp["success"].(bool))
}
