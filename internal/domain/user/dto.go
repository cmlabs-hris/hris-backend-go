package user

import (
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/validator"
)

// UserResponse represents user data in API responses
type UserResponse struct {
	ID            string  `json:"id"`
	CompanyID     *string `json:"company_id,omitempty"`
	Email         string  `json:"email"`
	Role          string  `json:"role"`
	OAuthProvider *string `json:"oauth_provider,omitempty"`
	EmailVerified bool    `json:"email_verified"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// CreateUserRequest represents request to create a new user
type CreateUserRequest struct {
	CompanyID *string `json:"company_id,omitempty"`
	Email     string  `json:"email"`
	Password  string  `json:"password"`
	Role      string  `json:"role"`
}

func (r *CreateUserRequest) Validate() error {
	var errs validator.ValidationErrors

	if validator.IsEmpty(r.Email) {
		errs = append(errs, validator.ValidationError{
			Field:   "email",
			Message: "email is required",
		})
	} else if !validator.IsValidEmail(r.Email) {
		errs = append(errs, validator.ValidationError{
			Field:   "email",
			Message: "invalid email format",
		})
	}

	if validator.IsEmpty(r.Password) {
		errs = append(errs, validator.ValidationError{
			Field:   "password",
			Message: "password is required",
		})
	} else if len(r.Password) < 8 {
		errs = append(errs, validator.ValidationError{
			Field:   "password",
			Message: "password must be at least 8 characters",
		})
	}

	if validator.IsEmpty(r.Role) {
		errs = append(errs, validator.ValidationError{
			Field:   "role",
			Message: "role is required",
		})
	} else {
		validRoles := []string{string(RoleOwner), string(RoleManager), string(RoleEmployee), string(RolePending)}
		if !validator.IsInSlice(r.Role, validRoles) {
			errs = append(errs, validator.ValidationError{
				Field:   "role",
				Message: "invalid role",
			})
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

// UpdateUserRequest represents request to update user
type UpdateUserRequest struct {
	ID        string  `json:"id"`
	Email     *string `json:"email,omitempty"`
	Password  *string `json:"password,omitempty"`
	Role      *string `json:"role,omitempty"`
	CompanyID *string `json:"company_id,omitempty"`
}

func (r *UpdateUserRequest) Validate() error {
	var errs validator.ValidationErrors

	if validator.IsEmpty(r.ID) {
		errs = append(errs, validator.ValidationError{
			Field:   "id",
			Message: "id is required",
		})
	}

	if r.Email != nil {
		if validator.IsEmpty(*r.Email) {
			errs = append(errs, validator.ValidationError{
				Field:   "email",
				Message: "email must not be empty",
			})
		} else if !validator.IsValidEmail(*r.Email) {
			errs = append(errs, validator.ValidationError{
				Field:   "email",
				Message: "invalid email format",
			})
		}
	}

	if r.Password != nil {
		if validator.IsEmpty(*r.Password) {
			errs = append(errs, validator.ValidationError{
				Field:   "password",
				Message: "password must not be empty",
			})
		} else if len(*r.Password) < 8 {
			errs = append(errs, validator.ValidationError{
				Field:   "password",
				Message: "password must be at least 8 characters",
			})
		}
	}

	if r.Role != nil {
		if validator.IsEmpty(*r.Role) {
			errs = append(errs, validator.ValidationError{
				Field:   "role",
				Message: "role must not be empty",
			})
		} else {
			validRoles := []string{string(RoleOwner), string(RoleManager), string(RoleEmployee), string(RolePending)}
			if !validator.IsInSlice(*r.Role, validRoles) {
				errs = append(errs, validator.ValidationError{
					Field:   "role",
					Message: "invalid role",
				})
			}
		}
	}

	if r.CompanyID != nil && validator.IsEmpty(*r.CompanyID) {
		errs = append(errs, validator.ValidationError{
			Field:   "company_id",
			Message: "company_id must not be empty",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

// UpdateUserRoleRequest represents request to update user role
type UpdateUserRoleRequest struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

func (r *UpdateUserRoleRequest) Validate() error {
	var errs validator.ValidationErrors

	if validator.IsEmpty(r.ID) {
		errs = append(errs, validator.ValidationError{
			Field:   "id",
			Message: "id is required",
		})
	}

	if validator.IsEmpty(r.Role) {
		errs = append(errs, validator.ValidationError{
			Field:   "role",
			Message: "role is required",
		})
	} else {
		validRoles := []string{string(RoleOwner), string(RoleManager), string(RoleEmployee), string(RolePending)}
		if !validator.IsInSlice(r.Role, validRoles) {
			errs = append(errs, validator.ValidationError{
				Field:   "role",
				Message: "invalid role",
			})
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

// VerifyEmailRequest represents email verification request
type VerifyEmailRequest struct {
	Token string `json:"token"`
}

func (r *VerifyEmailRequest) Validate() error {
	var errs validator.ValidationErrors

	if validator.IsEmpty(r.Token) {
		errs = append(errs, validator.ValidationError{
			Field:   "token",
			Message: "token is required",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
