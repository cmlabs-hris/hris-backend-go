package auth

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/auth"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/jwt"
	"github.com/cmlabs-hris/hris-backend-go/internal/repository/postgresql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

var (
	testAuthDB *database.DB
)

const (
	testAccessExp  = "1h"
	testRefreshExp = "24h"
	testSecret     = "test-secret-key-for-jwt"
)

func authTestInit() {
	if testAuthDB != nil {
		return
	}
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:root@localhost:5432/cmlabs_hris_test?sslmode=disable"
	}

	var err error
	testAuthDB, err = database.NewPostgreSQLDB(dsn)
	if err != nil {
		panic("Failed to connect to test database: " + err.Error())
	}
}

func truncateAuthTables(t *testing.T, ctx context.Context) {
	authTestInit()
	tables := []string{"refresh_tokens", "users", "companies", "employees"}

	for _, table := range tables {
		_, err := testAuthDB.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			// Some tables might not exist, skip
			continue
		}
	}
}

func createAuthTestCompany(t *testing.T, ctx context.Context) string {
	authTestInit()
	var companyID string
	// Generate unique username per test using high-precision time
	uniqueUsername := fmt.Sprintf("test-company-%d-%d", time.Now().Unix(), time.Now().Nanosecond())
	err := testAuthDB.QueryRow(ctx, `
		INSERT INTO companies (id, name, username, created_at, updated_at)
		VALUES (uuidv7(), 'Test Company', $1, NOW(), NOW())
		RETURNING id
	`, uniqueUsername).Scan(&companyID)
	require.NoError(t, err)
	return companyID
}

// createAuthTestUserWithEmail creates a test user and returns both userID and email
func createAuthTestUserWithEmail(t *testing.T, ctx context.Context, companyID string, email string) string {
	var userID string
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	hashedStr := string(hashedPassword)

	err := testAuthDB.QueryRow(ctx, `
		INSERT INTO users (company_id, email, password_hash, is_admin, email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, false, true, NOW(), NOW())
		RETURNING id
	`, companyID, email, hashedStr).Scan(&userID)
	require.NoError(t, err)
	return userID
}

// Test Login with valid credentials
func TestAuthService_Login_Success(t *testing.T) {
	ctx := context.Background()
	authTestInit()
	truncateAuthTables(t, ctx)

	// Setup
	companyID := createAuthTestCompany(t, ctx)
	testEmail := fmt.Sprintf("login-%d@example.com", time.Now().UnixNano())
	createAuthTestUserWithEmail(t, ctx, companyID, testEmail)

	// Create service
	userRepo := postgresql.NewUserRepository(testAuthDB)
	companyRepo := postgresql.NewCompanyRepository(testAuthDB)
	jwtService := jwt.NewJWTService(testSecret, testAccessExp, testRefreshExp)
	jwtRepo := postgresql.NewJWTRepository(testAuthDB)
	authService := NewAuthService(testAuthDB, userRepo, companyRepo, jwtService, jwtRepo)

	// Act
	loginReq := auth.LoginRequest{Email: testEmail, Password: "password123"}
	sessionReq := auth.SessionTrackingRequest{IPAddress: "127.0.0.1", UserAgent: "Mozilla/5.0"}
	response, err := authService.Login(ctx, loginReq, sessionReq)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Greater(t, response.AccessTokenExpiresIn, int64(0))
	assert.Greater(t, response.RefreshTokenExpiresIn, int64(0))
}

// Test Login with invalid password
func TestAuthService_Login_InvalidPassword(t *testing.T) {

	ctx := context.Background()
	authTestInit()
	truncateAuthTables(t, ctx)

	// Setup
	companyID := createAuthTestCompany(t, ctx)
	testEmail := fmt.Sprintf("invalidpass-%d@example.com", time.Now().UnixNano())
	createAuthTestUserWithEmail(t, ctx, companyID, testEmail)

	// Create service
	userRepo := postgresql.NewUserRepository(testAuthDB)
	companyRepo := postgresql.NewCompanyRepository(testAuthDB)
	jwtService := jwt.NewJWTService(testSecret, testAccessExp, testRefreshExp)
	jwtRepo := postgresql.NewJWTRepository(testAuthDB)
	authService := NewAuthService(testAuthDB, userRepo, companyRepo, jwtService, jwtRepo)

	// Act
	loginReq := auth.LoginRequest{Email: testEmail, Password: "wrongpassword"}
	sessionReq := auth.SessionTrackingRequest{IPAddress: "127.0.0.1", UserAgent: "Mozilla/5.0"}
	_, err := authService.Login(ctx, loginReq, sessionReq)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, auth.ErrInvalidCredentials, err)
}

// Test Login with non-existent user
func TestAuthService_Login_UserNotFound(t *testing.T) {

	ctx := context.Background()
	authTestInit()
	truncateAuthTables(t, ctx)

	// Create service
	userRepo := postgresql.NewUserRepository(testAuthDB)
	companyRepo := postgresql.NewCompanyRepository(testAuthDB)
	jwtService := jwt.NewJWTService(testSecret, testAccessExp, testRefreshExp)
	jwtRepo := postgresql.NewJWTRepository(testAuthDB)
	authService := NewAuthService(testAuthDB, userRepo, companyRepo, jwtService, jwtRepo)

	// Act
	loginReq := auth.LoginRequest{Email: "nonexistent@example.com", Password: "password123"}
	sessionReq := auth.SessionTrackingRequest{IPAddress: "127.0.0.1", UserAgent: "Mozilla/5.0"}
	_, err := authService.Login(ctx, loginReq, sessionReq)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, auth.ErrInvalidCredentials, err)
}

// Test LoginWithGoogle for new user
func TestAuthService_LoginWithGoogle_NewUser(t *testing.T) {

	ctx := context.Background()
	authTestInit()
	truncateAuthTables(t, ctx)

	// Create service
	userRepo := postgresql.NewUserRepository(testAuthDB)
	companyRepo := postgresql.NewCompanyRepository(testAuthDB)
	jwtService := jwt.NewJWTService(testSecret, testAccessExp, testRefreshExp)
	jwtRepo := postgresql.NewJWTRepository(testAuthDB)
	authService := NewAuthService(testAuthDB, userRepo, companyRepo, jwtService, jwtRepo)

	// Act
	googleEmail := "newgoogleuser@example.com"
	sessionReq := auth.SessionTrackingRequest{IPAddress: "127.0.0.1", UserAgent: "Mozilla/5.0"}
	response, err := authService.LoginWithGoogle(ctx, googleEmail, "google-id-123", sessionReq)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Greater(t, response.AccessTokenExpiresIn, int64(0))

	// Verify user was created
	createdUser, err := userRepo.GetByEmail(ctx, googleEmail)
	assert.NoError(t, err)
	assert.Equal(t, googleEmail, createdUser.Email)
	assert.NotNil(t, createdUser.OAuthProvider)
	assert.Equal(t, "google", *createdUser.OAuthProvider)
	assert.Equal(t, "google-id-123", *createdUser.OAuthProviderID)
	assert.True(t, createdUser.EmailVerified)
}

// Test LoginWithGoogle for existing user
func TestAuthService_LoginWithGoogle_ExistingUser(t *testing.T) {

	ctx := context.Background()
	authTestInit()
	truncateAuthTables(t, ctx)

	// Setup
	companyID := createAuthTestCompany(t, ctx)
	testEmail := "existinguser@example.com"
	_ = createAuthTestUserWithEmail(t, ctx, companyID, testEmail)

	// Create service
	userRepo := postgresql.NewUserRepository(testAuthDB)
	companyRepo := postgresql.NewCompanyRepository(testAuthDB)
	jwtService := jwt.NewJWTService(testSecret, testAccessExp, testRefreshExp)
	jwtRepo := postgresql.NewJWTRepository(testAuthDB)
	authService := NewAuthService(testAuthDB, userRepo, companyRepo, jwtService, jwtRepo)

	// Act - Link Google to existing account
	sessionReq := auth.SessionTrackingRequest{IPAddress: "127.0.0.1", UserAgent: "Mozilla/5.0"}
	response, err := authService.LoginWithGoogle(ctx, testEmail, "google-id-456", sessionReq)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
}

// Test Logout by revoking refresh token
func TestAuthService_RevokeRefreshToken_Success(t *testing.T) {

	ctx := context.Background()
	authTestInit()
	truncateAuthTables(t, ctx)

	// Setup
	companyID := createAuthTestCompany(t, ctx)
	testEmail := fmt.Sprintf("revoke-%d@example.com", time.Now().UnixNano())
	testUserID := createAuthTestUserWithEmail(t, ctx, companyID, testEmail)

	// Create service
	jwtService := jwt.NewJWTService(testSecret, testAccessExp, testRefreshExp)
	jwtRepo := postgresql.NewJWTRepository(testAuthDB)

	// Generate and save refresh token
	refreshToken, _, err := jwtService.GenerateRefreshToken(testUserID)
	require.NoError(t, err)

	sessionReq := auth.SessionTrackingRequest{IPAddress: "127.0.0.1", UserAgent: "Mozilla/5.0"}
	err = jwtRepo.CreateRefreshToken(ctx, testUserID, refreshToken, 86400, sessionReq)
	require.NoError(t, err)

	// Act - Revoke refresh token
	err = jwtRepo.RevokeRefreshToken(ctx, refreshToken)

	// Assert
	assert.NoError(t, err)

	// Verify token is revoked
	_, isRevoked, err := jwtRepo.IsRefreshTokenRevoked(ctx, refreshToken)
	assert.NoError(t, err)
	assert.True(t, isRevoked)
}

func TestAuthService_LoginWithEmployeeCode_Success(t *testing.T) {
	t.Skip("Skipping - Employee table requires complex setup with positions, grades, branches, etc.")
}

func TestAuthService_LoginWithEmployeeCode_InvalidCredentials(t *testing.T) {
	t.Skip("Skipping - Employee table requires complex setup with positions, grades, branches, etc.")
}

func TestAuthService_RefreshToken_Success(t *testing.T) {

	ctx := context.Background()
	authTestInit()
	truncateAuthTables(t, ctx)

	// Setup - First login to get a valid refresh token instead of manually creating
	companyID := createAuthTestCompany(t, ctx)
	testEmail := fmt.Sprintf("refresh-%d@example.com", time.Now().UnixNano())
	createAuthTestUserWithEmail(t, ctx, companyID, testEmail)

	// Create service
	userRepo := postgresql.NewUserRepository(testAuthDB)
	companyRepo := postgresql.NewCompanyRepository(testAuthDB)
	jwtService := jwt.NewJWTService(testSecret, testAccessExp, testRefreshExp)
	jwtRepo := postgresql.NewJWTRepository(testAuthDB)
	authService := NewAuthService(testAuthDB, userRepo, companyRepo, jwtService, jwtRepo)

	// Login to get a valid refresh token
	loginReq := auth.LoginRequest{Email: testEmail, Password: "password123"}
	sessionReq := auth.SessionTrackingRequest{IPAddress: "127.0.0.1", UserAgent: "Mozilla/5.0"}
	loginResp, err := authService.Login(ctx, loginReq, sessionReq)
	require.NoError(t, err)
	require.NotEmpty(t, loginResp.RefreshToken)

	// Act - Use the refresh token from login
	refreshReq := auth.RefreshTokenRequest{RefreshToken: loginResp.RefreshToken}
	resp, err := authService.RefreshToken(ctx, refreshReq)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.Greater(t, resp.AccessTokenExpiresIn, int64(0))
}

func TestAuthService_Logout_Success(t *testing.T) {

	ctx := context.Background()
	authTestInit()
	truncateAuthTables(t, ctx)

	// Setup
	companyID := createAuthTestCompany(t, ctx)
	testEmail := fmt.Sprintf("logout-%d@example.com", time.Now().UnixNano())
	createAuthTestUserWithEmail(t, ctx, companyID, testEmail)

	// Create service
	userRepo := postgresql.NewUserRepository(testAuthDB)
	companyRepo := postgresql.NewCompanyRepository(testAuthDB)
	jwtService := jwt.NewJWTService(testSecret, testAccessExp, testRefreshExp)
	jwtRepo := postgresql.NewJWTRepository(testAuthDB)
	authService := NewAuthService(testAuthDB, userRepo, companyRepo, jwtService, jwtRepo)

	// Login to get a token
	loginReq := auth.LoginRequest{Email: testEmail, Password: "password123"}
	sessionReq := auth.SessionTrackingRequest{IPAddress: "127.0.0.1", UserAgent: "Mozilla/5.0"}
	loginResp, err := authService.Login(ctx, loginReq, sessionReq)
	require.NoError(t, err)

	// Act - Logout (revoke token)
	err = authService.Logout(ctx, loginResp.RefreshToken)

	// Assert
	assert.NoError(t, err)

	// Verify token is now revoked
	_, isRevoked, err := jwtRepo.IsRefreshTokenRevoked(ctx, loginResp.RefreshToken)
	assert.NoError(t, err)
	assert.True(t, isRevoked)
}

func TestAuthService_Register_Success(t *testing.T) {

	ctx := context.Background()
	authTestInit()
	truncateAuthTables(t, ctx)

	// Setup
	testEmail := fmt.Sprintf("newuser-%d@example.com", time.Now().UnixNano())
	testPassword := "SecurePass123!"
	testConfirmPassword := testPassword

	// Create service
	userRepo := postgresql.NewUserRepository(testAuthDB)
	companyRepo := postgresql.NewCompanyRepository(testAuthDB)
	jwtService := jwt.NewJWTService(testSecret, testAccessExp, testRefreshExp)
	jwtRepo := postgresql.NewJWTRepository(testAuthDB)
	authService := NewAuthService(testAuthDB, userRepo, companyRepo, jwtService, jwtRepo)

	// Act
	registerReq := auth.RegisterRequest{
		Email:           testEmail,
		Password:        testPassword,
		ConfirmPassword: testConfirmPassword,
	}
	sessionReq := auth.SessionTrackingRequest{IPAddress: "127.0.0.1", UserAgent: "Mozilla/5.0"}
	resp, err := authService.Register(ctx, registerReq, sessionReq)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)

	// Verify user was created
	var userCount int
	err = testAuthDB.QueryRow(ctx,
		`SELECT COUNT(*) FROM users WHERE email = $1`,
		testEmail).Scan(&userCount)
	assert.NoError(t, err)
	assert.Equal(t, 1, userCount)
}

func TestAuthService_ForgotPassword_NotImplemented(t *testing.T) {
	t.Skip("Method not yet implemented in AuthService")
}

func TestAuthService_OAuthCallbackGoogle_NotImplemented(t *testing.T) {
	t.Skip("Method not yet implemented in AuthService")
}

func TestAuthService_VerifyEmail_NotImplemented(t *testing.T) {
	t.Skip("Method not yet implemented in AuthService")
}
