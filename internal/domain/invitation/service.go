package invitation

import "context"

// InvitationService defines the interface for invitation business logic
type InvitationService interface {
	// CreateAndSend creates an invitation and sends email (called from EmployeeService)
	CreateAndSend(ctx context.Context, req CreateRequest) (Invitation, error)

	// GetByToken retrieves invitation details by token (public endpoint)
	GetByToken(ctx context.Context, token string) (InvitationDetailResponse, error)

	// ListMyInvitations lists pending invitations for user's email
	ListMyInvitations(ctx context.Context, email string) ([]MyInvitationResponse, error)

	// Accept accepts an invitation and links user to employee
	Accept(ctx context.Context, token, userID, userEmail string) (AcceptResponse, error)

	// Resend resends the invitation email with a new token
	Resend(ctx context.Context, employeeID, companyID string) error

	// Revoke revokes a pending invitation
	Revoke(ctx context.Context, employeeID, companyID string) error

	// ExistsPendingByEmail checks if email has pending invitation (for CreateEmployee validation)
	ExistsPendingByEmail(ctx context.Context, email, companyID string) (bool, error)
}
