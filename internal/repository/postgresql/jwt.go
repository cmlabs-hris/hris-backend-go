package postgresql

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/auth"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
)

type JWTRepository interface {
	CreateRefreshToken(ctx context.Context, userID string, token string, expiresAt int64, sessionReq auth.SessionTrackingRequest) error
	IsRefreshTokenRevoked(ctx context.Context, token string) (bool, error)
	RevokeRefreshToken(ctx context.Context, token string) error
}

type jwtRepositoryImpl struct {
	db *database.DB
}

// NewJWTRepository creates a new instance of JWTRepository.
func NewJWTRepository(db *database.DB) JWTRepository {
	return &jwtRepositoryImpl{db: db}
}

// hashToken hashes the input string using SHA256 and encodes the result in base64.
func (j *jwtRepositoryImpl) hashToken(input string) string {
	hash := sha256.Sum256([]byte(input))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func (j *jwtRepositoryImpl) CreateRefreshToken(ctx context.Context, userID string, token string, expiresAt int64, sessionReq auth.SessionTrackingRequest) error {
	q := GetQuerier(ctx, j.db)
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, user_agent, ip_address)
		VALUES ($1, $2, $3, $4, $5)
	`
	tokenHash := j.hashToken(token)
	_, err := q.Exec(ctx, query, userID, tokenHash, time.Unix(expiresAt, 0).UTC(), sessionReq.UserAgent, sessionReq.IPAddress)
	return err
}

func (j *jwtRepositoryImpl) IsRefreshTokenRevoked(ctx context.Context, token string) (bool, error) {
	q := GetQuerier(ctx, j.db)

	query := `
		SELECT revoked_at, expires_at
		FROM refresh_tokens
		WHERE token_hash = $1
		ORDER BY expires_at DESC
		LIMIT 1
	`
	tokenHash := j.hashToken(token)

	var revokedAt *time.Time
	var expiresAt time.Time

	err := q.QueryRow(ctx, query, tokenHash).Scan(&revokedAt, &expiresAt)
	if err != nil {
		return false, err
	}

	now := time.Now()
	if revokedAt != nil || !expiresAt.After(now) {
		return true, nil
	}
	return false, nil
}

func (j *jwtRepositoryImpl) RevokeRefreshToken(ctx context.Context, token string) error {
	q := GetQuerier(ctx, j.db)

	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`
	tokenHash := j.hashToken(token)
	_, err := q.Exec(ctx, query, tokenHash)
	return err
}
