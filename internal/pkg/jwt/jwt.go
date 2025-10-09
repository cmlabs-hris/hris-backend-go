package jwt

import (
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type Service interface {
	GenerateAccessToken(userID, email, companyID string, isAdmin bool) (token string, expiresAt int64, err error)
	GenerateRefreshToken(userID string) (token string, expiresAt int64, err error)
}

// type Claims struct {
// 	UserID          string `json:"user_id"`
// 	Email           string `json:"email"`
// 	CompanyID       string `json:"company_id"`
// 	CompanyUsername string `json:"company_username"`
// 	IsAdmin         bool   `json:"is_admin"`
// 	exp             int64  `json:"exp"`
// }

type JWTService struct {
	secretKey                  string
	accessTokenExpirationTime  string
	refreshTokenExpirationTime string
	tokenAuth                  *jwtauth.JWTAuth
}

func NewJWTService(secretKey string, accessTokenExpirationTime string, refreshTokenExpirationTime string) Service {
	return &JWTService{
		secretKey:                  secretKey,
		accessTokenExpirationTime:  accessTokenExpirationTime,
		refreshTokenExpirationTime: refreshTokenExpirationTime,
		tokenAuth:                  jwtauth.New("HS256", []byte(secretKey), nil, jwt.WithAcceptableSkew(30*time.Second)),
	}
}

// GenerateAccessToken implements Service.
func (j *JWTService) GenerateAccessToken(userID string, email string, companyID string, isAdmin bool) (string, int64, error) {
	accessClaims := map[string]interface{}{
		"user_id":    userID,
		"email":      email,
		"company_id": companyID,
		"is_admin":   isAdmin,
		"type":       "access",
	}
	expireTime, err := time.ParseDuration(j.accessTokenExpirationTime)
	if err != nil {
		return "", 0, err
	}
	expireAt := time.Now().Add(expireTime)
	jwtauth.SetExpiryIn(accessClaims, expireTime)
	_, accessToken, err := j.tokenAuth.Encode(accessClaims)
	if err != nil {
		return "", 0, err
	}
	return accessToken, expireAt.Unix(), nil
}

// GenerateRefreshToken implements Service.
func (j *JWTService) GenerateRefreshToken(userID string) (string, int64, error) {
	refreshClaims := map[string]interface{}{
		"user_id": userID,
		"type":    "refresh",
	}
	expireTime, err := time.ParseDuration(j.refreshTokenExpirationTime)
	if err != nil {
		return "", 0, err
	}
	expireAt := time.Now().Add(expireTime)
	jwtauth.SetExpiryIn(refreshClaims, expireTime)
	_, refreshToken, err := j.tokenAuth.Encode(refreshClaims)
	if err != nil {
		return "", 0, err
	}
	return refreshToken, expireAt.Unix(), nil
}
