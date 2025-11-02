package auth

import (
	"context"
	"fmt"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/auth"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/company"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/employee"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/user"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/jwt"
	"github.com/cmlabs-hris/hris-backend-go/internal/repository/postgresql"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthServiceImpl struct {
	db *database.DB
	user.UserRepository
	company.CompanyRepository
	jwt.Service
	postgresql.JWTRepository
	employee.EmployeeRepository
}

func NewAuthService(db *database.DB, userRepository user.UserRepository, companyRepository company.CompanyRepository, jwtService jwt.Service, jwtRepository postgresql.JWTRepository) auth.AuthService {
	return &AuthServiceImpl{
		db:                db,
		UserRepository:    userRepository,
		CompanyRepository: companyRepository,
		Service:           jwtService,
		JWTRepository:     jwtRepository,
	}
}

func (a *AuthServiceImpl) hashPassword(password string) (string, error) {
	if password == "" {
		return "", nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	hashed := string(hash)
	return hashed, nil
}

// ForgotPassword implements auth.AuthService.
func (a *AuthServiceImpl) ForgotPassword(ctx context.Context, req auth.RefreshTokenRequest) error {
	panic("unimplemented")
}

// Login implements auth.AuthService.
func (a *AuthServiceImpl) Login(ctx context.Context, loginReq auth.LoginRequest, sessionTrackReq auth.SessionTrackingRequest) (auth.TokenResponse, error) {
	var tokenResponse auth.TokenResponse

	// Langsung ambil user, error jika tidak ada
	userData, err := a.UserRepository.GetByEmail(ctx, loginReq.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return auth.TokenResponse{}, auth.ErrInvalidCredentials
		}
		return auth.TokenResponse{}, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Cek password
	if userData.PasswordHash == nil {
		return auth.TokenResponse{}, auth.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*userData.PasswordHash), []byte(loginReq.Password)); err != nil {
		return auth.TokenResponse{}, auth.ErrInvalidCredentials
	}

	err = postgresql.WithTransaction(ctx, a.db, func(tx pgx.Tx) error {
		txCtx := context.WithValue(ctx, "tx", tx)

		tokenResponse.AccessToken, tokenResponse.AccessTokenExpiresIn, err = a.Service.GenerateAccessToken(userData.ID, userData.Email, userData.CompanyID, userData.Role)
		if err != nil {
			return fmt.Errorf("failed to create access token: %w", err)
		}
		tokenResponse.RefreshToken, tokenResponse.RefreshTokenExpiresIn, err = a.Service.GenerateRefreshToken(userData.ID)
		if err != nil {
			return fmt.Errorf("failed to create refresh token: %w", err)
		}

		err = a.CreateRefreshToken(txCtx, userData.ID, tokenResponse.RefreshToken, tokenResponse.RefreshTokenExpiresIn, sessionTrackReq)
		if err != nil {
			return fmt.Errorf("failed to save refresh token to database: %w", err)
		}
		return nil
	})

	if err != nil {
		return auth.TokenResponse{}, err
	}

	return tokenResponse, nil
}

// LoginWithEmployeeCode implements auth.AuthService.
func (a *AuthServiceImpl) LoginWithEmployeeCode(ctx context.Context, loginEmployeeCodeReq auth.LoginEmployeeCodeRequest, sessionTrackReq auth.SessionTrackingRequest) (auth.TokenResponse, error) {
	var tokenResponse auth.TokenResponse

	// Langsung ambil company, error jika tidak ada
	companyData, err := a.CompanyRepository.GetByUsername(ctx, loginEmployeeCodeReq.CompanyUsername)
	if err != nil {
		if err == pgx.ErrNoRows {
			return auth.TokenResponse{}, auth.ErrInvalidEmployeeCodeCredentials
		}
		return auth.TokenResponse{}, fmt.Errorf("failed to get company by username: %w", err)
	}

	// Langsung ambil employee, error jika tidak ada
	employeeData, err := a.EmployeeRepository.GetByEmployeeCode(ctx, companyData.ID, loginEmployeeCodeReq.EmployeeCode)
	if err != nil {
		if err == pgx.ErrNoRows {
			return auth.TokenResponse{}, auth.ErrInvalidEmployeeCodeCredentials
		}
		return auth.TokenResponse{}, fmt.Errorf("failed to get employee by code: %w", err)
	}

	// Ambil user, error jika tidak ada
	userData, err := a.UserRepository.GetByID(ctx, employeeData.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return auth.TokenResponse{}, auth.ErrInvalidEmployeeCodeCredentials
		}
		return auth.TokenResponse{}, fmt.Errorf("failed to get user by id: %w", err)
	}

	// Cek password
	if err := bcrypt.CompareHashAndPassword([]byte(*userData.PasswordHash), []byte(loginEmployeeCodeReq.Password)); err != nil {
		return auth.TokenResponse{}, auth.ErrInvalidEmployeeCodeCredentials
	}

	err = postgresql.WithTransaction(ctx, a.db, func(tx pgx.Tx) error {
		txCtx := context.WithValue(ctx, "tx", tx)

		tokenResponse.AccessToken, tokenResponse.AccessTokenExpiresIn, err = a.Service.GenerateAccessToken(userData.ID, userData.Email, userData.CompanyID, userData.Role)
		if err != nil {
			return fmt.Errorf("failed to create access token: %w", err)
		}
		tokenResponse.RefreshToken, tokenResponse.RefreshTokenExpiresIn, err = a.Service.GenerateRefreshToken(userData.ID)
		if err != nil {
			return fmt.Errorf("failed to create refresh token: %w", err)
		}

		err = a.CreateRefreshToken(txCtx, userData.ID, tokenResponse.RefreshToken, tokenResponse.RefreshTokenExpiresIn, sessionTrackReq)
		if err != nil {
			return fmt.Errorf("failed to save refresh token to database: %w", err)
		}
		return nil
	})

	if err != nil {
		return auth.TokenResponse{}, err
	}

	return tokenResponse, nil
}

// LoginWithGoogle implements auth.AuthService.
func (a *AuthServiceImpl) LoginWithGoogle(ctx context.Context, googleEmail string, googleID string, sessionTrackReq auth.SessionTrackingRequest) (auth.TokenResponse, error) {
	var tokenResponse auth.TokenResponse
	var userExists bool

	userData, err := a.UserRepository.GetByEmail(ctx, googleEmail)
	if err != nil {
		if err == pgx.ErrNoRows {
			userExists = false
		} else {
			return auth.TokenResponse{}, fmt.Errorf("failed to get user data by email: %w", err)
		}
	}

	if userData.ID != "" {
		userExists = true
	}

	// User does not exist so we create one
	if !userExists {
		newUser := user.User{
			CompanyID:               nil,
			Email:                   googleEmail,
			PasswordHash:            nil,
			Role:                    user.RolePending,
			OAuthProvider:           func(s string) *string { return &s }("google"),
			OAuthProviderID:         &googleID,
			EmailVerified:           true,
			EmailVerificationToken:  nil,
			EmailVerificationSentAt: nil,
		}
		userData, err = a.UserRepository.Create(ctx, newUser)
		if err != nil {
			return auth.TokenResponse{}, fmt.Errorf("failed to create user: %w", err)
		}

	}

	// If user exists, link google account
	if userData.OAuthProvider == nil || userData.OAuthProviderID == nil {
		_, err := a.UserRepository.LinkGoogleAccount(ctx, googleID, userData.Email)
		if err != nil {
			return auth.TokenResponse{}, err
		}
	}

	// Generate token
	err = postgresql.WithTransaction(ctx, a.db, func(tx pgx.Tx) error {
		txCtx := context.WithValue(ctx, "tx", tx)

		tokenResponse.AccessToken, tokenResponse.AccessTokenExpiresIn, err = a.Service.GenerateAccessToken(userData.ID, userData.Email, userData.CompanyID, userData.Role)
		if err != nil {
			return fmt.Errorf("failed to create access token: %w", err)
		}
		tokenResponse.RefreshToken, tokenResponse.RefreshTokenExpiresIn, err = a.Service.GenerateRefreshToken(userData.ID)
		if err != nil {
			return fmt.Errorf("failed to create refresh token: %w", err)
		}

		err = a.CreateRefreshToken(txCtx, userData.ID, tokenResponse.RefreshToken, tokenResponse.RefreshTokenExpiresIn, sessionTrackReq)
		if err != nil {
			return fmt.Errorf("failed to save refresh token to database: %w", err)
		}
		return nil
	})
	if err != nil {
		return auth.TokenResponse{}, err
	}

	return tokenResponse, nil
}

// Logout implements auth.AuthService.
func (a *AuthServiceImpl) Logout(ctx context.Context, token string) error {
	err := postgresql.WithTransaction(ctx, a.db, func(tx pgx.Tx) error {
		txCtx := context.WithValue(ctx, "tx", tx)

		_, isRevoked, err := a.JWTRepository.IsRefreshTokenRevoked(txCtx, token)
		if err != nil {
			return fmt.Errorf("failed to check if refresh token is revoked: %w", err)
		}
		if !isRevoked {
			if err := a.JWTRepository.RevokeRefreshToken(txCtx, token); err != nil {
				return fmt.Errorf("failed to revoke refresh token: %w", err)
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// OAuthCallbackGoogle implements auth.AuthService.
func (a *AuthServiceImpl) OAuthCallbackGoogle(ctx context.Context) (auth.TokenResponse, error) {
	// idk why this exists
	panic("unimplemented")
}

// RefreshToken implements auth.AuthService.
func (a *AuthServiceImpl) RefreshToken(ctx context.Context, req auth.RefreshTokenRequest) (auth.AccessTokenResponse, error) {
	var accessTokenResponse auth.AccessTokenResponse

	// 1. Verify JWT signature and expiry
	token, err := jwtauth.VerifyToken(a.JWTAuth(), req.RefreshToken)
	if err != nil {
		return auth.AccessTokenResponse{}, auth.ErrInvalidToken
	}

	// 2. Check token type is "refresh"
	claims, err := token.AsMap(ctx)
	if err != nil {
		return auth.AccessTokenResponse{}, auth.ErrInvalidToken
	}
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "refresh" {
		return auth.AccessTokenResponse{}, auth.ErrInvalidToken
	}

	// 3. Check DB for revocation/expiry (pass raw token, not hash)
	userID, isRevoked, err := a.JWTRepository.IsRefreshTokenRevoked(ctx, req.RefreshToken)
	if err != nil {
		return auth.AccessTokenResponse{}, auth.ErrInvalidToken
	}
	if isRevoked {
		return auth.AccessTokenResponse{}, auth.ErrRefreshTokenRevoked
	}

	// 4. Get user
	userData, err := a.UserRepository.GetByID(ctx, userID)
	if err != nil {
		return auth.AccessTokenResponse{}, auth.ErrUserNotFound
	}

	// 5. Generate new access token
	accessTokenResponse.AccessToken, accessTokenResponse.AccessTokenExpiresIn, err =
		a.Service.GenerateAccessToken(userData.ID, userData.Email, userData.CompanyID, userData.Role)
	if err != nil {
		return auth.AccessTokenResponse{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	return accessTokenResponse, nil
}

// Register implements auth.AuthService.
func (a *AuthServiceImpl) Register(ctx context.Context, registerReq auth.RegisterRequest, sessionTrackReq auth.SessionTrackingRequest) (auth.TokenResponse, error) {
	var tokenResponse auth.TokenResponse

	// Check user already exist or not
	userData, err := a.UserRepository.GetByEmail(ctx, registerReq.Email)
	if err != nil {
		if err != pgx.ErrNoRows {
			return auth.TokenResponse{}, fmt.Errorf("failed to get user data by email: %w", err)
		}
	}

	if userData.ID != "" {
		return auth.TokenResponse{}, auth.ErrEmailAlreadyExists
	}

	// if OAuthExists {
	// 	hashedPassword, err := a.hashPassword(registerReq.Password)
	// 	if err != nil {
	// 		return auth.TokenResponse{}, fmt.Errorf("failed to hash password: %w", err)
	// 	}
	// 	if _, err := a.UserRepository.LinkPasswordAccount(ctx, userData.ID, hashedPassword); err != nil {
	// 		return auth.TokenResponse{}, fmt.Errorf("failed to link password account: %w", err)
	// 	}
	// 	err = postgresql.WithTransaction(ctx, a.db, func(tx pgx.Tx) error {
	// 		txCtx := context.WithValue(ctx, "tx", tx)

	// 		tokenResponse.AccessToken, tokenResponse.AccessTokenExpiresIn, err = a.Service.GenerateAccessToken(userData.ID, userData.Email, userData.CompanyID, userData.IsAdmin)
	// 		if err != nil {
	// 			return fmt.Errorf("failed to create access token: %w", err)
	// 		}
	// 		tokenResponse.RefreshToken, tokenResponse.RefreshTokenExpiresIn, err = a.Service.GenerateRefreshToken(userData.ID)
	// 		if err != nil {
	// 			return fmt.Errorf("failed to create refresh token: %w", err)
	// 		}

	// 		err = a.CreateRefreshToken(txCtx, userData.ID, tokenResponse.RefreshToken, tokenResponse.RefreshTokenExpiresIn, sessionTrackReq)
	// 		if err != nil {
	// 			return fmt.Errorf("failed to save refresh token to database: %w", err)
	// 		}
	// 		return nil
	// 	})
	// 	if err != nil {
	// 		return auth.TokenResponse{}, err
	// 	}
	// 	return tokenResponse, nil
	// }

	// Hash the password before storing
	hashedPassword, err := a.hashPassword(registerReq.Password)
	if err != nil {
		return auth.TokenResponse{}, fmt.Errorf("failed to hash password: %w", err)
	}
	newUser := user.User{
		CompanyID:               nil,
		Email:                   registerReq.Email,
		PasswordHash:            &hashedPassword,
		Role:                    user.RolePending,
		OAuthProvider:           nil,
		OAuthProviderID:         nil,
		EmailVerified:           false,
		EmailVerificationToken:  nil,
		EmailVerificationSentAt: nil,
	}
	newUser, err = a.UserRepository.Create(ctx, newUser)
	if err != nil {
		return auth.TokenResponse{}, fmt.Errorf("failed to create user: %w", err)
	}

	err = postgresql.WithTransaction(ctx, a.db, func(tx pgx.Tx) error {
		txCtx := context.WithValue(ctx, "tx", tx)

		tokenResponse.AccessToken, tokenResponse.AccessTokenExpiresIn, err = a.Service.GenerateAccessToken(newUser.ID, newUser.Email, nil, newUser.Role)
		if err != nil {
			return fmt.Errorf("failed to create access token: %w", err)
		}
		tokenResponse.RefreshToken, tokenResponse.RefreshTokenExpiresIn, err = a.Service.GenerateRefreshToken(newUser.ID)
		if err != nil {
			return fmt.Errorf("failed to create refresh token: %w", err)
		}

		err = a.CreateRefreshToken(txCtx, newUser.ID, tokenResponse.RefreshToken, tokenResponse.RefreshTokenExpiresIn, sessionTrackReq)
		if err != nil {
			return fmt.Errorf("failed to save refresh token to database: %w", err)
		}
		return nil
	})
	if err != nil {
		return auth.TokenResponse{}, err
	}

	return tokenResponse, nil
}

// VerifyEmail implements auth.AuthService.
func (a *AuthServiceImpl) VerifyEmail(ctx context.Context, req auth.VerifyEmailRequest) error {
	panic("unimplemented")
}
