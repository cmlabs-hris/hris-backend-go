package postgresql

import (
	"context"
	"fmt"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/invitation"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/jackc/pgx/v5"
)

type invitationRepositoryImpl struct {
	db *database.DB
}

// NewInvitationRepository creates a new invitation repository instance
func NewInvitationRepository(db *database.DB) invitation.InvitationRepository {
	return &invitationRepositoryImpl{db: db}
}

// Create implements invitation.InvitationRepository.
func (r *invitationRepositoryImpl) Create(ctx context.Context, inv invitation.Invitation) (invitation.Invitation, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		INSERT INTO employee_invitations (
			employee_id, company_id, invited_by_employee_id, email, role, status, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, employee_id, company_id, invited_by_employee_id, email, token, role, status, 
				  expires_at, accepted_at, revoked_at, created_at, updated_at
	`

	var created invitation.Invitation
	err := q.QueryRow(ctx, query,
		inv.EmployeeID, inv.CompanyID, inv.InvitedByEmployeeID,
		inv.Email, inv.Role, inv.Status, inv.ExpiresAt,
	).Scan(
		&created.ID, &created.EmployeeID, &created.CompanyID, &created.InvitedByEmployeeID,
		&created.Email, &created.Token, &created.Role, &created.Status, &created.ExpiresAt,
		&created.AcceptedAt, &created.RevokedAt, &created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return invitation.Invitation{}, fmt.Errorf("failed to create invitation: %w", err)
	}

	return created, nil
}

// GetByTokenWithDetails implements invitation.InvitationRepository.
func (r *invitationRepositoryImpl) GetByTokenWithDetails(ctx context.Context, token string) (invitation.InvitationWithDetails, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT 
			ei.id, ei.employee_id, ei.company_id, ei.invited_by_employee_id, 
			ei.email, ei.token, ei.role, ei.status, ei.expires_at, 
			ei.accepted_at, ei.revoked_at, ei.created_at, ei.updated_at,
			e.full_name AS employee_name,
			c.name AS company_name, c.logo_url AS company_logo,
			p.name AS position_name,
			inviter.full_name AS inviter_name
		FROM employee_invitations ei
		JOIN employees e ON e.id = ei.employee_id
		JOIN companies c ON c.id = ei.company_id
		JOIN employees inviter ON inviter.id = ei.invited_by_employee_id
		LEFT JOIN positions p ON p.id = e.position_id
		WHERE ei.token = $1
	`

	var inv invitation.InvitationWithDetails
	var positionName *string

	err := q.QueryRow(ctx, query, token).Scan(
		&inv.ID, &inv.EmployeeID, &inv.CompanyID, &inv.InvitedByEmployeeID,
		&inv.Email, &inv.Token, &inv.Role, &inv.Status, &inv.ExpiresAt,
		&inv.AcceptedAt, &inv.RevokedAt, &inv.CreatedAt, &inv.UpdatedAt,
		&inv.EmployeeName, &inv.CompanyName, &inv.CompanyLogo,
		&positionName, &inv.InviterName,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return inv, invitation.ErrInvitationNotFound
		}
		return inv, fmt.Errorf("failed to get invitation by token: %w", err)
	}

	inv.PositionName = positionName
	return inv, nil
}

// GetPendingByEmployeeID implements invitation.InvitationRepository.
func (r *invitationRepositoryImpl) GetPendingByEmployeeID(ctx context.Context, employeeID, companyID string) (invitation.Invitation, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, employee_id, company_id, invited_by_employee_id, email, token, role, status,
			   expires_at, accepted_at, revoked_at, created_at, updated_at
		FROM employee_invitations
		WHERE employee_id = $1 AND company_id = $2 AND status = 'pending'
		ORDER BY created_at DESC
		LIMIT 1
	`

	var inv invitation.Invitation
	err := q.QueryRow(ctx, query, employeeID, companyID).Scan(
		&inv.ID, &inv.EmployeeID, &inv.CompanyID, &inv.InvitedByEmployeeID,
		&inv.Email, &inv.Token, &inv.Role, &inv.Status, &inv.ExpiresAt,
		&inv.AcceptedAt, &inv.RevokedAt, &inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return inv, invitation.ErrNoPendingInvitation
		}
		return inv, fmt.Errorf("failed to get pending invitation: %w", err)
	}

	return inv, nil
}

// ExistsPendingByEmail implements invitation.InvitationRepository.
func (r *invitationRepositoryImpl) ExistsPendingByEmail(ctx context.Context, email, companyID string) (bool, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT EXISTS(
			SELECT 1 FROM employee_invitations 
			WHERE email = $1 AND company_id = $2 AND status = 'pending' AND expires_at > NOW()
		)
	`

	var exists bool
	err := q.QueryRow(ctx, query, email, companyID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check pending invitation: %w", err)
	}

	return exists, nil
}

// ListPendingByEmail implements invitation.InvitationRepository.
func (r *invitationRepositoryImpl) ListPendingByEmail(ctx context.Context, email string) ([]invitation.InvitationWithDetails, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT 
			ei.id, ei.employee_id, ei.company_id, ei.invited_by_employee_id, 
			ei.email, ei.token, ei.role, ei.status, ei.expires_at, 
			ei.accepted_at, ei.revoked_at, ei.created_at, ei.updated_at,
			e.full_name AS employee_name,
			c.name AS company_name, c.logo_url AS company_logo,
			p.name AS position_name,
			inviter.full_name AS inviter_name
		FROM employee_invitations ei
		JOIN employees e ON e.id = ei.employee_id
		JOIN companies c ON c.id = ei.company_id
		JOIN employees inviter ON inviter.id = ei.invited_by_employee_id
		LEFT JOIN positions p ON p.id = e.position_id
		WHERE ei.email = $1 AND ei.status = 'pending' AND ei.expires_at > NOW()
		ORDER BY ei.created_at DESC
	`

	rows, err := q.Query(ctx, query, email)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending invitations: %w", err)
	}
	defer rows.Close()

	var invitations []invitation.InvitationWithDetails
	for rows.Next() {
		var inv invitation.InvitationWithDetails
		var positionName *string

		err := rows.Scan(
			&inv.ID, &inv.EmployeeID, &inv.CompanyID, &inv.InvitedByEmployeeID,
			&inv.Email, &inv.Token, &inv.Role, &inv.Status, &inv.ExpiresAt,
			&inv.AcceptedAt, &inv.RevokedAt, &inv.CreatedAt, &inv.UpdatedAt,
			&inv.EmployeeName, &inv.CompanyName, &inv.CompanyLogo,
			&positionName, &inv.InviterName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invitation: %w", err)
		}

		inv.PositionName = positionName
		invitations = append(invitations, inv)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return invitations, nil
}

// MarkAccepted implements invitation.InvitationRepository.
func (r *invitationRepositoryImpl) MarkAccepted(ctx context.Context, id string) error {
	q := GetQuerier(ctx, r.db)

	query := `
		UPDATE employee_invitations 
		SET status = 'accepted', accepted_at = NOW(), updated_at = NOW()
		WHERE id = $1
		RETURNING id
	`

	var updatedID string
	err := q.QueryRow(ctx, query, id).Scan(&updatedID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return invitation.ErrInvitationNotFound
		}
		return fmt.Errorf("failed to mark invitation as accepted: %w", err)
	}

	return nil
}

// MarkRevoked implements invitation.InvitationRepository.
func (r *invitationRepositoryImpl) MarkRevoked(ctx context.Context, id string) error {
	q := GetQuerier(ctx, r.db)

	query := `
		UPDATE employee_invitations 
		SET status = 'revoked', revoked_at = NOW(), updated_at = NOW()
		WHERE id = $1
		RETURNING id
	`

	var updatedID string
	err := q.QueryRow(ctx, query, id).Scan(&updatedID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return invitation.ErrInvitationNotFound
		}
		return fmt.Errorf("failed to mark invitation as revoked: %w", err)
	}

	return nil
}

// UpdateToken implements invitation.InvitationRepository.
func (r *invitationRepositoryImpl) UpdateToken(ctx context.Context, id, newToken string, expiresAt time.Time) error {
	q := GetQuerier(ctx, r.db)

	query := `
		UPDATE employee_invitations 
		SET token = $1, expires_at = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id
	`

	var updatedID string
	err := q.QueryRow(ctx, query, newToken, expiresAt, id).Scan(&updatedID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return invitation.ErrInvitationNotFound
		}
		return fmt.Errorf("failed to update invitation token: %w", err)
	}

	return nil
}
