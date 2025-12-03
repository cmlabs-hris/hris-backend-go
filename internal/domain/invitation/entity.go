package invitation

import "time"

// Status represents the status of an invitation
type Status string

const (
	StatusPending  Status = "pending"
	StatusAccepted Status = "accepted"
	StatusRevoked  Status = "revoked"
)

// Invitation represents an employee invitation entity
type Invitation struct {
	ID                  string
	EmployeeID          string
	CompanyID           string
	InvitedByEmployeeID string
	Email               string
	Token               string
	Role                string // "employee" or "manager"
	Status              Status
	ExpiresAt           time.Time
	AcceptedAt          *time.Time
	RevokedAt           *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// InvitationWithDetails contains invitation data with joined related names
type InvitationWithDetails struct {
	Invitation
	EmployeeName string
	CompanyName  string
	CompanyLogo  *string
	PositionName *string
	InviterName  string
}

// IsExpired checks if the invitation has expired (query-time check)
func (i *Invitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

// CanBeAccepted checks if the invitation can be accepted
func (i *Invitation) CanBeAccepted() bool {
	return i.Status == StatusPending && !i.IsExpired()
}
