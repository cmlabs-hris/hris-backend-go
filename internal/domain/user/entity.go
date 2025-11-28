package user

import "time"

type Role string

const (
	RoleOwner    Role = "owner"    // Company owner - full access
	RoleManager  Role = "manager"  // Can approve leave/attendance
	RoleEmployee Role = "employee" // Regular employee
	RolePending  Role = "pending"  // Still in onboarding
)

type User struct {
	ID                      string
	CompanyID               *string
	Email                   string
	PasswordHash            *string
	Role                    Role
	OAuthProvider           *string
	OAuthProviderID         *string
	EmailVerified           bool
	EmailVerificationToken  *string
	EmailVerificationSentAt *time.Time
	CreatedAt               time.Time
	UpdatedAt               time.Time

	// DTO / Join
	EmployeeID *string
}

// IsOwner checks if user is company owner
func (u *User) IsOwner() bool {
	return u.Role == RoleOwner
}

// IsManager checks if user is manager or owner
func (u *User) IsManager() bool {
	return u.Role == RoleManager || u.Role == RoleOwner
}

// IsPending checks if user is still in onboarding
func (u *User) IsPending() bool {
	return u.Role == RolePending
}

// CanApprove checks if user can approve requests
func (u *User) CanApprove() bool {
	return u.IsManager()
}

// CanManageCompany checks if user can manage company settings
func (u *User) CanManageCompany() bool {
	return u.IsOwner()
}
