package jwt

import (
	"net/http"
	"sync"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/user"
	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// SubscriptionClaims contains subscription-related claims for the JWT
type SubscriptionClaims struct {
	Features              []string   // Feature codes the subscription has access to
	SubscriptionExpiresAt *time.Time // When the subscription period ends
}

type Service interface {
	GenerateAccessToken(userID string, email string, employeeID *string, companyID *string, role user.Role, subClaims *SubscriptionClaims) (token string, expiresAt int64, err error)
	GenerateRefreshToken(userID string) (token string, expiresAt int64, err error)
	GenerateSSEToken(userID string) (token string, expiresIn int, err error)
	ValidateSSEToken(tokenString string) (userID string, err error)
	JWTAuth() *jwtauth.JWTAuth
	RefreshTokenCookie(token string, expiresAt int64) *http.Cookie
	RevokeToken(token string)
	IsTokenRevoked(token string) bool
}

type JWTService struct {
	secretKey                  string
	accessTokenExpirationTime  string
	refreshTokenExpirationTime string
	tokenAuth                  *jwtauth.JWTAuth
	revokedTokens              map[string]int64
	mu                         sync.RWMutex
}

func (j *JWTService) JWTAuth() *jwtauth.JWTAuth {
	return j.tokenAuth
}

func NewJWTService(secretKey string, accessTokenExpirationTime string, refreshTokenExpirationTime string) Service {
	return &JWTService{
		secretKey:                  secretKey,
		accessTokenExpirationTime:  accessTokenExpirationTime,
		refreshTokenExpirationTime: refreshTokenExpirationTime,
		tokenAuth:                  jwtauth.New("HS256", []byte(secretKey), nil, jwt.WithAcceptableSkew(30*time.Second)),
		revokedTokens:              make(map[string]int64),
	}
}

func (j *JWTService) GenerateAccessToken(userID string, email string, employeeID *string, companyID *string, role user.Role, subClaims *SubscriptionClaims) (token string, expiresAt int64, err error) {
	expDuration, err := time.ParseDuration(j.accessTokenExpirationTime)
	if err != nil {
		return "", 0, err
	}
	expiresAt = time.Now().Add(expDuration).Unix()

	claims := map[string]interface{}{
		"user_id":     userID,
		"email":       email,
		"employee_id": j.returnValueOrNil(employeeID),
		"company_id":  j.returnValueOrNil(companyID),
		"role":        string(role),
		"type":        "access",
		"exp":         expiresAt,
	}

	// Add subscription claims if provided
	if subClaims != nil {
		if subClaims.Features != nil {
			claims["features"] = subClaims.Features
		}
		if subClaims.SubscriptionExpiresAt != nil {
			claims["subscription_expires_at"] = subClaims.SubscriptionExpiresAt.Unix()
		}
	}

	_, tokenString, err := j.tokenAuth.Encode(claims)
	return tokenString, expiresAt, err
}

func (j *JWTService) GenerateRefreshToken(userID string) (token string, expiresAt int64, err error) {
	expDuration, err := time.ParseDuration(j.refreshTokenExpirationTime)
	if err != nil {
		return "", 0, err
	}
	expiresAt = time.Now().Add(expDuration).Unix()
	_, tokenString, err := j.tokenAuth.Encode(map[string]interface{}{
		"user_id": userID,
		"exp":     expiresAt,
		"type":    "refresh",
	})
	return tokenString, expiresAt, err
}

func (j *JWTService) RefreshTokenCookie(token string, expiresAt int64) *http.Cookie {
	return &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/api/v1/auth",
		Expires:  time.Unix(expiresAt, 0),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	}
}

func (j *JWTService) RevokeToken(token string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.revokedTokens[token] = time.Now().Unix()
}

func (j *JWTService) IsTokenRevoked(token string) bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	_, revoked := j.revokedTokens[token]
	return revoked
}

func (j *JWTService) returnValueOrNil(value *string) interface{} {
	if value == nil {
		return nil
	} else {
		return *value
	}
}

// GenerateSSEToken generates a short-lived token for SSE connections
func (j *JWTService) GenerateSSEToken(userID string) (token string, expiresIn int, err error) {
	// SSE tokens are short-lived (5 minutes)
	expiresIn = 300 // 5 minutes in seconds
	expiresAt := time.Now().Add(5 * time.Minute).Unix()

	_, tokenString, err := j.tokenAuth.Encode(map[string]interface{}{
		"user_id": userID,
		"type":    "sse",
		"exp":     expiresAt,
	})
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresIn, nil
}

// ValidateSSEToken validates an SSE token and returns the user ID
func (j *JWTService) ValidateSSEToken(tokenString string) (userID string, err error) {
	token, err := j.tokenAuth.Decode(tokenString)
	if err != nil {
		return "", err
	}

	// Check token type
	tokenType, ok := token.Get("type")
	if !ok || tokenType != "sse" {
		return "", jwt.ErrInvalidJWT()
	}

	// Get user ID
	userIDVal, ok := token.Get("user_id")
	if !ok {
		return "", jwt.ErrInvalidJWT()
	}

	userID, ok = userIDVal.(string)
	if !ok {
		return "", jwt.ErrInvalidJWT()
	}

	return userID, nil
}
