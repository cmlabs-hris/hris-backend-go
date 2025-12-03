package invitation

import "github.com/cmlabs-hris/hris-backend-go/internal/pkg/validator"

// CreateRequest - used internally by EmployeeService when creating invitation
type CreateRequest struct {
	EmployeeID          string
	CompanyID           string
	InvitedByEmployeeID string
	Email               string
	Role                string  // "employee" or "manager" - assigned on acceptance
	EmployeeName        string  // For email template
	InviterName         string  // For email template
	CompanyName         string  // For email template
	PositionName        *string // For email template
}

func (r *CreateRequest) Validate() error {
	var errs validator.ValidationErrors

	if validator.IsEmpty(r.EmployeeID) {
		errs = append(errs, validator.ValidationError{
			Field:   "employee_id",
			Message: "employee_id is required",
		})
	}

	if validator.IsEmpty(r.CompanyID) {
		errs = append(errs, validator.ValidationError{
			Field:   "company_id",
			Message: "company_id is required",
		})
	}

	if validator.IsEmpty(r.Email) {
		errs = append(errs, validator.ValidationError{
			Field:   "email",
			Message: "email is required",
		})
	} else if !validator.IsValidEmail(r.Email) {
		errs = append(errs, validator.ValidationError{
			Field:   "email",
			Message: "email format is invalid",
		})
	}

	if validator.IsEmpty(r.EmployeeName) {
		errs = append(errs, validator.ValidationError{
			Field:   "employee_name",
			Message: "employee_name is required",
		})
	}

	if validator.IsEmpty(r.CompanyName) {
		errs = append(errs, validator.ValidationError{
			Field:   "company_name",
			Message: "company_name is required",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

// AcceptRequest for accepting an invitation
type AcceptRequest struct {
	Token  string `json:"-"` // From Chi URL param
	UserID string // From JWT - not from request body
}

func (r *AcceptRequest) Validate() error {
	var errs validator.ValidationErrors

	if validator.IsEmpty(r.Token) {
		errs = append(errs, validator.ValidationError{
			Field:   "token",
			Message: "token is required",
		})
	} else if !validator.IsValidUUID(r.Token) {
		errs = append(errs, validator.ValidationError{
			Field:   "token",
			Message: "token must be a valid UUID",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

// MyInvitationResponse - GET /invitations/my
type MyInvitationResponse struct {
	Token        string  `json:"token"`
	CompanyName  string  `json:"company_name"`
	CompanyLogo  *string `json:"company_logo,omitempty"`
	PositionName *string `json:"position_name,omitempty"`
	InviterName  string  `json:"inviter_name"`
	ExpiresAt    string  `json:"expires_at"`
	CreatedAt    string  `json:"created_at"`
}

// InvitationDetailResponse - GET /invitations/{token}
type InvitationDetailResponse struct {
	Token        string  `json:"token"`
	Email        string  `json:"email"`
	EmployeeName string  `json:"employee_name"`
	CompanyName  string  `json:"company_name"`
	CompanyLogo  *string `json:"company_logo,omitempty"`
	PositionName *string `json:"position_name,omitempty"`
	Role         string  `json:"role"`
	InviterName  string  `json:"inviter_name"`
	Status       string  `json:"status"`
	ExpiresAt    string  `json:"expires_at"`
	IsExpired    bool    `json:"is_expired"`
}

// AcceptResponse for invitation acceptance result
type AcceptResponse struct {
	Message     string `json:"message"`
	CompanyID   string `json:"company_id"`
	CompanyName string `json:"company_name"`
	EmployeeID  string `json:"employee_id"`
}
