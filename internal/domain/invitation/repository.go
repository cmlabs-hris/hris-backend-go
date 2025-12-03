package invitation

import (
	"context"
	"time"
)

// InvitationRepository defines the interface for invitation data access
type InvitationRepository interface {
	// Create creates a new invitation record
	Create(ctx context.Context, inv Invitation) (Invitation, error)

	// GetByTokenWithDetails retrieves an invitation by token with all related details (company, inviter names)
	GetByTokenWithDetails(ctx context.Context, token string) (InvitationWithDetails, error)

	// GetPendingByEmployeeID retrieves the latest pending invitation for an employee
	GetPendingByEmployeeID(ctx context.Context, employeeID, companyID string) (Invitation, error)

	// ExistsPendingByEmail checks if email has a pending non-expired invitation in the company
	ExistsPendingByEmail(ctx context.Context, email, companyID string) (bool, error)

	// ListPendingByEmail lists all pending non-expired invitations for an email (for user's "my invitations")
	ListPendingByEmail(ctx context.Context, email string) ([]InvitationWithDetails, error)

	// MarkAccepted marks an invitation as accepted
	MarkAccepted(ctx context.Context, id string) error

	// MarkRevoked marks an invitation as revoked
	MarkRevoked(ctx context.Context, id string) error

	// UpdateToken updates the token and expiry date (for resend)
	UpdateToken(ctx context.Context, id, newToken string, expiresAt time.Time) error
}
