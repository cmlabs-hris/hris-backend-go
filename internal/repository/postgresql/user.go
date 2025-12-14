package postgresql

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/user"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/jackc/pgx/v5"
)

type userRepositoryImpl struct {
	db *database.DB
}

// Update implements user.UserRepository.
func (r *userRepositoryImpl) Update(ctx context.Context, req user.UpdateUserRequest) error {
	q := GetQuerier(ctx, r.db)

	updates := make([]string, 0)
	args := make([]interface{}, 0)
	argIdx := 1

	if req.Email != nil {
		updates = append(updates, fmt.Sprintf("email = $%d", argIdx))
		args = append(args, *req.Email)
		argIdx++
	}
	if req.Password != nil {
		updates = append(updates, fmt.Sprintf("password_hash = $%d", argIdx))
		args = append(args, *req.Password)
		argIdx++
	}
	if req.Role != nil {
		updates = append(updates, fmt.Sprintf("role = $%d", argIdx))
		args = append(args, *req.Role)
		argIdx++
	}
	if req.CompanyID != nil {
		updates = append(updates, fmt.Sprintf("company_id = $%d", argIdx))
		args = append(args, *req.CompanyID)
		argIdx++
	}

	// Always update the updated_at field
	updates = append(updates, fmt.Sprintf("updated_at = $%d", argIdx))
	args = append(args, time.Now())
	argIdx++

	args = append(args, req.ID)
	idIdx := argIdx

	query := "UPDATE users SET " + strings.Join(updates, ", ") +
		fmt.Sprintf(" WHERE id = $%d RETURNING id", idIdx)

	var updatedID string
	if err := q.QueryRow(ctx, query, args...).Scan(&updatedID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// UpdateRole implements user.UserRepository.
func (r *userRepositoryImpl) UpdateRole(ctx context.Context, req user.UpdateUserRoleRequest) error {
	q := GetQuerier(ctx, r.db)

	updateQuery := `
		UPDATE users
		SET role = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := q.Exec(ctx, updateQuery, req.Role, req.ID)
	if err != nil {
		return err
	}

	return nil
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
		SELECT u.id, u.company_id, u.email, u.password_hash, u.role, u.oauth_provider, u.oauth_provider_id,
			   u.email_verified, u.email_verification_token, u.email_verification_sent_at,
			   u.created_at, u.updated_at, e.id AS employee_id
		FROM users u
		LEFT JOIN employees e ON u.id = e.user_id
		WHERE u.id = $1
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
		&found.EmployeeID,
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
		SELECT u.id, u.company_id, u.email, u.password_hash, u.role, u.oauth_provider, u.oauth_provider_id,
			   u.email_verified, u.email_verification_token, u.email_verification_sent_at,
			   u.created_at, u.updated_at, e.id AS employee_id
		FROM users u
		LEFT JOIN employees e ON u.id = e.user_id
		WHERE u.email = $1
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
		&found.EmployeeID, // This will be nil if no employee record exists
	)
	if err != nil {
		return user.User{}, err
	}

	return found, nil
}

// UpdateCompanyAndRole implements user.UserRepository.
func (r *userRepositoryImpl) UpdateCompanyAndRole(ctx context.Context, userID, companyID, role string) error {
	q := GetQuerier(ctx, r.db)

	query := `
		UPDATE users
		SET company_id = $1, role = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id
	`

	var updatedID string
	err := q.QueryRow(ctx, query, companyID, role, userID).Scan(&updatedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("user not found: %w", err)
		}
		return fmt.Errorf("failed to update user company and role: %w", err)
	}

	return nil
}

// UpdatePassword implements user.UserRepository.
func (r *userRepositoryImpl) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	q := GetQuerier(ctx, r.db)

	query := `
		UPDATE users
		SET password_hash = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id
	`

	var updatedID string
	err := q.QueryRow(ctx, query, passwordHash, userID).Scan(&updatedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("user not found: %w", err)
		}
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// VerifyEmail implements user.UserRepository.
func (r *userRepositoryImpl) VerifyEmail(ctx context.Context, userID string) error {
	q := GetQuerier(ctx, r.db)

	query := `
		UPDATE users
		SET email_verified = true, email_verification_token = NULL, updated_at = NOW()
		WHERE id = $1
		RETURNING id
	`

	var updatedID string
	err := q.QueryRow(ctx, query, userID).Scan(&updatedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("user not found: %w", err)
		}
		return fmt.Errorf("failed to verify email: %w", err)
	}

	return nil
}

// GetByEmailVerificationToken implements user.UserRepository.
func (r *userRepositoryImpl) GetByEmailVerificationToken(ctx context.Context, token string) (user.User, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT 
			u.id,
			u.company_id,
			u.email,
			u.password_hash,
			u.role,
			u.oauth_provider,
			u.oauth_provider_id,
			u.email_verified,
			u.email_verification_token,
			u.email_verification_sent_at,
			u.created_at,
			u.updated_at,
			e.id AS employee_id
		FROM users u
		LEFT JOIN employees e ON e.user_id = u.id AND e.deleted_at IS NULL
		WHERE u.email_verification_token = $1
		LIMIT 1
	`

	var u user.User
	err := q.QueryRow(ctx, query, token).Scan(
		&u.ID,
		&u.CompanyID,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.OAuthProvider,
		&u.OAuthProviderID,
		&u.EmailVerified,
		&u.EmailVerificationToken,
		&u.EmailVerificationSentAt,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.EmployeeID,
	)
	if err != nil {
		return user.User{}, err
	}

	return u, nil
}
