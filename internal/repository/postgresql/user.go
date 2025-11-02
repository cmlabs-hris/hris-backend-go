package postgresql

import (
	"context"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/user"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
)

type userRepositoryImpl struct {
	db *database.DB
}

// LinkPasswordAccount implements user.UserRepository.
func (r *userRepositoryImpl) LinkPasswordAccount(ctx context.Context, id string, password string) (user.User, error) {
	q := GetQuerier(ctx, r.db)

	updateQuery := `
		UPDATE users
		SET password_hash = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, company_id, email, password_hash, role, oauth_provider, oauth_provider_id,
				  email_verified, email_verification_token, email_verification_sent_at,
				  created_at, updated_at
	`

	var updated user.User
	err := q.QueryRow(ctx, updateQuery, password, id).Scan(
		&updated.ID,
		&updated.CompanyID,
		&updated.Email,
		&updated.PasswordHash,
		&updated.Role,
		&updated.OAuthProvider,
		&updated.OAuthProviderID,
		&updated.EmailVerified,
		&updated.EmailVerificationToken,
		&updated.EmailVerificationSentAt,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		return user.User{}, err
	}

	return updated, nil
}

// LinkGoogleAccount implements user.UserRepository.
func (r *userRepositoryImpl) LinkGoogleAccount(ctx context.Context, GoogleID string, email string) (user.User, error) {
	q := GetQuerier(ctx, r.db)

	updateQuery := `
		UPDATE users
		SET oauth_provider = $1, oauth_provider_id = $2, updated_at = NOW()
		WHERE email = $3
		RETURNING id, company_id, email, password_hash, role, oauth_provider, oauth_provider_id,
				  email_verified, email_verification_token, email_verification_sent_at,
				  created_at, updated_at
	`

	var updated user.User
	err := q.QueryRow(ctx, updateQuery, "google", GoogleID, email).Scan(
		&updated.ID,
		&updated.CompanyID,
		&updated.Email,
		&updated.PasswordHash,
		&updated.Role,
		&updated.OAuthProvider,
		&updated.OAuthProviderID,
		&updated.EmailVerified,
		&updated.EmailVerificationToken,
		&updated.EmailVerificationSentAt,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		return user.User{}, err
	}

	return updated, nil
}

func NewUserRepository(db *database.DB) user.UserRepository {
	return &userRepositoryImpl{db: db}
}

// ExistsByIDOrEmail implements user.UserRepositoryImpl.
func (r *userRepositoryImpl) ExistsByIDOrEmail(ctx context.Context, id *string, email *string) (bool, error) {
	q := GetQuerier(ctx, r.db)

	var query string
	var arg interface{}

	switch {
	case id != nil:
		query = `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
		arg = *id
	case email != nil:
		query = `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
		arg = *email
	default:
		return false, nil
	}

	var exists bool
	err := q.QueryRow(ctx, query, arg).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// Create implements user.UserRepositoryImpl.
func (r *userRepositoryImpl) Create(ctx context.Context, newUser user.User) (user.User, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		INSERT INTO users (
			company_id, email, password_hash, role, oauth_provider, oauth_provider_id,
			email_verified, email_verification_token, email_verification_sent_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, company_id, email, password_hash, role, oauth_provider, oauth_provider_id,
				  email_verified, email_verification_token, email_verification_sent_at,
				  created_at, updated_at
	`

	var created user.User
	err := q.QueryRow(ctx, query,
		newUser.CompanyID,
		newUser.Email,
		newUser.PasswordHash,
		newUser.Role,
		newUser.OAuthProvider,
		newUser.OAuthProviderID,
		newUser.EmailVerified,
		newUser.EmailVerificationToken,
		newUser.EmailVerificationSentAt,
	).Scan(
		&created.ID,
		&created.CompanyID,
		&created.Email,
		&created.PasswordHash,
		&created.Role,
		&created.OAuthProvider,
		&created.OAuthProviderID,
		&created.EmailVerified,
		&created.EmailVerificationToken,
		&created.EmailVerificationSentAt,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return user.User{}, err
	}

	return created, nil
}

// GetByID implements user.UserRepositoryImpl.
func (r *userRepositoryImpl) GetByID(ctx context.Context, id string) (user.User, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, company_id, email, password_hash, role, oauth_provider, oauth_provider_id,
			   email_verified, email_verification_token, email_verification_sent_at,
			   created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var found user.User
	err := q.QueryRow(ctx, query, id).Scan(
		&found.ID,
		&found.CompanyID,
		&found.Email,
		&found.PasswordHash,
		&found.Role,
		&found.OAuthProvider,
		&found.OAuthProviderID,
		&found.EmailVerified,
		&found.EmailVerificationToken,
		&found.EmailVerificationSentAt,
		&found.CreatedAt,
		&found.UpdatedAt,
	)
	if err != nil {
		return user.User{}, err
	}

	return found, nil
}

// GetByEmail implements user.UserRepositoryImpl.
func (r *userRepositoryImpl) GetByEmail(ctx context.Context, email string) (user.User, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, company_id, email, password_hash, role, oauth_provider, oauth_provider_id,
			   email_verified, email_verification_token, email_verification_sent_at,
			   created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var found user.User
	err := q.QueryRow(ctx, query, email).Scan(
		&found.ID,
		&found.CompanyID,
		&found.Email,
		&found.PasswordHash,
		&found.Role,
		&found.OAuthProvider,
		&found.OAuthProviderID,
		&found.EmailVerified,
		&found.EmailVerificationToken,
		&found.EmailVerificationSentAt,
		&found.CreatedAt,
		&found.UpdatedAt,
	)
	if err != nil {
		return user.User{}, err
	}

	return found, nil
}
