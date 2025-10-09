package auth

import (
	"context"
	"fmt"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/auth"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/company"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/user"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/jwt"
	"github.com/cmlabs-hris/hris-backend-go/internal/repository/postgresql"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthServiceImpl struct {
	db *database.DB
	user.UserRepository
	company.CompanyRepository
	jwt.Service
	postgresql.JWTRepository
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

	userExists, err := a.UserRepository.ExistsByIDOrEmail(ctx, nil, &loginReq.Email)
	if err != nil {
		return auth.TokenResponse{}, fmt.Errorf("check user exists: %w", err)
	}
	if !userExists {
		return auth.TokenResponse{}, auth.ErrInvalidCredentials
	}

	userData, err := a.UserRepository.GetByEmail(ctx, loginReq.Email)
	if err != nil {
		return auth.TokenResponse{}, fmt.Errorf("failed to get user by email: %w", err)
	}

	hashedPassword, err := a.hashPassword(loginReq.Password)
	if err != nil {
		return auth.TokenResponse{}, fmt.Errorf("failed to hash password: %w", err)
	}

	if hashedPassword != *userData.PasswordHash {
		return auth.TokenResponse{}, auth.ErrInvalidCredentials
	}

	err = postgresql.WithTransaction(ctx, a.db, func(tx pgx.Tx) error {
		txCtx := context.WithValue(ctx, "tx", tx)

		tokenResponse.AccessToken, tokenResponse.AccessTokenExpiresIn, err = a.Service.GenerateAccessToken(userData.ID, userData.Email, userData.CompanyID, userData.IsAdmin)
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
func (a *AuthServiceImpl) LoginWithEmployeeCode(ctx context.Context, req auth.LoginEmployeeCodeRequest) (auth.TokenResponse, error) {
	panic("unimplemented")
}

// LoginWithGoogle implements auth.AuthService.
func (a *AuthServiceImpl) LoginWithGoogle(ctx context.Context, email string, id string) error {
	panic("unimplemented")
}

// Logout implements auth.AuthService.
func (a *AuthServiceImpl) Logout(ctx context.Context) error {
	panic("unimplemented")
}

// OAuthCallbackGoogle implements auth.AuthService.
func (a *AuthServiceImpl) OAuthCallbackGoogle(ctx context.Context) (auth.TokenResponse, error) {
	panic("unimplemented")
}

// RefreshToken implements auth.AuthService.
func (a *AuthServiceImpl) RefreshToken(ctx context.Context, req auth.RefreshTokenRequest) (auth.AccessTokenResponse, error) {
	panic("unimplemented")
}

// Register implements auth.AuthService.
func (a *AuthServiceImpl) Register(ctx context.Context, registerReq auth.RegisterRequest, sessionTrackReq auth.SessionTrackingRequest) (auth.TokenResponse, error) {
	var tokenResponse auth.TokenResponse
	err := postgresql.WithTransaction(ctx, a.db, func(tx pgx.Tx) error {
		txCtx := context.WithValue(ctx, "tx", tx)

		// Check company already exist or not
		companyExists, err := a.CompanyRepository.ExistsByIDOrUsername(txCtx, nil, &registerReq.CompanyUsername)
		if err != nil {
			return fmt.Errorf("check company exists: %w", err)
		}
		if companyExists {
			return company.ErrCompanyUsernameExists
		}
		newCompany := company.Company{
			Name:     registerReq.CompanyName,
			Username: registerReq.CompanyUsername,
		}
		newCompany, err = a.CompanyRepository.Create(txCtx, newCompany)
		if err != nil {
			return fmt.Errorf("failed to create company: %w", err)
		}

		// Check user already exist or not
		userExists, err := a.UserRepository.ExistsByIDOrEmail(txCtx, nil, &registerReq.Email)
		if err != nil {
			return fmt.Errorf("check user exists: %w", err)
		}
		if userExists {
			return user.ErrUserEmailExists
		}

		// Hash the password before storing
		hashedPassword, err := a.hashPassword(registerReq.Password)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		newUser := user.User{
			CompanyID:               newCompany.ID,
			Email:                   registerReq.Email,
			PasswordHash:            &hashedPassword,
			IsAdmin:                 true,
			OAuthProvider:           nil,
			OAuthProviderID:         nil,
			EmailVerified:           false,
			EmailVerificationToken:  nil,
			EmailVerificationSentAt: nil,
		}
		newUser, err = a.UserRepository.Create(txCtx, newUser)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		tokenResponse.AccessToken, tokenResponse.AccessTokenExpiresIn, err = a.Service.GenerateAccessToken(newUser.ID, newUser.Email, newCompany.ID, newUser.IsAdmin)
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
