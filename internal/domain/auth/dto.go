package auth

import "github.com/cmlabs-hris/hris-backend-go/internal/pkg/validator"

type RegisterRequest struct {
	CompanyName     string `json:"company_name"`
	CompanyUsername string `json:"company_username"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

func (r *RegisterRequest) Validate() error {
	var errs validator.ValidationErrors

	// Company
	if validator.IsEmpty(r.CompanyName) {
		errs = append(errs, validator.ValidationError{
			Field:   "company_name",
			Message: "company_name is required",
		})
	}
	if validator.IsEmpty(r.CompanyUsername) {
		errs = append(errs, validator.ValidationError{
			Field:   "company_username",
			Message: "company_username is required",
		})
	}

	// Email
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

	// Password
	if validator.IsEmpty(r.Password) {
		errs = append(errs, validator.ValidationError{
			Field:   "password",
			Message: "password is required",
		})
	} else if len(r.Password) < 8 {
		errs = append(errs, validator.ValidationError{
			Field:   "password",
			Message: "password must be at least 8 characters long",
		})
	}
	if validator.IsEmpty(r.ConfirmPassword) {
		errs = append(errs, validator.ValidationError{
			Field:   "confirm_password",
			Message: "confirm_password is required",
		})
	} else if r.ConfirmPassword != r.Password {
		errs = append(errs, validator.ValidationError{
			Field:   "confirm_password",
			Message: "password and confirm_password do not match",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *LoginRequest) Validate() error {
	var errs validator.ValidationErrors

	// Email
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

	// Password
	if validator.IsEmpty(r.Password) {
		errs = append(errs, validator.ValidationError{
			Field:   "password",
			Message: "password is required",
		})
	} else if len(r.Password) < 8 {
		errs = append(errs, validator.ValidationError{
			Field:   "password",
			Message: "password must be at least 8 characters long",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type LoginEmployeeCodeRequest struct {
	CompanyUsername string `json:"company_username"`
	EmployeeCode    string `json:"employee_code"`
	Password        string `json:"password"`
}

func (r *LoginEmployeeCodeRequest) Validate() error {
	var errs validator.ValidationErrors

	// Company
	if validator.IsEmpty(r.CompanyUsername) {
		errs = append(errs, validator.ValidationError{
			Field:   "company_username",
			Message: "company_username is required",
		})
	}

	// Password
	if validator.IsEmpty(r.Password) {
		errs = append(errs, validator.ValidationError{
			Field:   "password",
			Message: "password is required",
		})
	} else if len(r.Password) < 8 {
		errs = append(errs, validator.ValidationError{
			Field:   "password",
			Message: "password must be at least 8 characters long",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (r *RefreshTokenRequest) Validate() error {
	var errs validator.ValidationErrors

	// Refresh Token
	if validator.IsEmpty(r.RefreshToken) {
		errs = append(errs, validator.ValidationError{
			Field:   "refresh_token",
			Message: "refresh_token is required",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

func (r *ForgotPasswordRequest) Validate() error {
	var errs validator.ValidationErrors

	// Email
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

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type VerifyEmailRequest struct {
	Token string `json:"token"`
}

func (r *VerifyEmailRequest) Validate() error {
	var errs validator.ValidationErrors

	// Token
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
