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
	if len(r.CompanyName) > 255 {
		errs = append(errs, validator.ValidationError{
			Field:   "company_name",
			Message: "company_name must not exceed 255 characters",
		})
	}
	if validator.IsEmpty(r.CompanyUsername) {
		errs = append(errs, validator.ValidationError{
			Field:   "company_username",
			Message: "company_username is required",
		})
	}
	if len(r.CompanyUsername) < 3 {
		errs = append(errs, validator.ValidationError{
			Field:   "company_username",
			Message: "company_username must be at least 3 characters long",
		})
	}
	if len(r.CompanyUsername) > 50 {
		errs = append(errs, validator.ValidationError{
			Field:   "company_username",
			Message: "company_username must not exceed 50 characters",
		})
	}
	if !validator.IsValidCompanyUsername(r.CompanyUsername) {
		errs = append(errs, validator.ValidationError{
			Field:   "company_username",
			Message: "company_username may only contain letters, numbers, dots, underscores, and hyphens",
		})
	}

	// Email
	if validator.IsEmpty(r.Email) {
		errs = append(errs, validator.ValidationError{
			Field:   "email",
			Message: "email is required",
		})
	}
	if len(r.Email) < 6 {
		errs = append(errs, validator.ValidationError{
			Field:   "email",
			Message: "email must be at least 6 characters long",
		})
	}
	if len(r.Email) > 254 {
		errs = append(errs, validator.ValidationError{
			Field:   "email",
			Message: "email must not exceed 254 characters",
		})
	}
	if !validator.IsValidEmail(r.Email) {
		errs = append(errs, validator.ValidationError{
			Field:   "email",
			Message: "email must be a valid email address (letters, numbers, ., _, %, +, - allowed before @; must contain @; domain must contain letters, numbers, ., -; must end with a dot and at least 2 letters, e.g. user@example.com)",
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
	} else if len(r.Password) > 255 {
		errs = append(errs, validator.ValidationError{
			Field:   "password",
			Message: "password must not exceed 255 characters",
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
	}
	if len(r.Email) < 6 {
		errs = append(errs, validator.ValidationError{
			Field:   "email",
			Message: "email must be at least 6 characters long",
		})
	}
	if len(r.Email) > 254 {
		errs = append(errs, validator.ValidationError{
			Field:   "email",
			Message: "email must not exceed 254 characters",
		})
	}
	if !validator.IsValidEmail(r.Email) {
		errs = append(errs, validator.ValidationError{
			Field:   "email",
			Message: "email must be a valid email address (letters, numbers, ., _, %, +, - allowed before @; must contain @; domain must contain letters, numbers, ., -; must end with a dot and at least 2 letters, e.g. user@example.com)",
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
	} else if len(r.Password) > 255 {
		errs = append(errs, validator.ValidationError{
			Field:   "password",
			Message: "password must not exceed 255 characters",
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
	if len(r.CompanyUsername) < 3 {
		errs = append(errs, validator.ValidationError{
			Field:   "company_username",
			Message: "company_username must be at least 3 characters long",
		})
	}
	if len(r.CompanyUsername) > 50 {
		errs = append(errs, validator.ValidationError{
			Field:   "company_username",
			Message: "company_username must not exceed 50 characters",
		})
	}
	if !validator.IsValidCompanyUsername(r.CompanyUsername) {
		errs = append(errs, validator.ValidationError{
			Field:   "company_username",
			Message: "company_username may only contain letters, numbers, dots, underscores, and hyphens",
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
	} else if len(r.Password) > 255 {
		errs = append(errs, validator.ValidationError{
			Field:   "password",
			Message: "password must not exceed 255 characters",
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
	if len(r.RefreshToken) > 255 {
		errs = append(errs, validator.ValidationError{
			Field:   "refresh_token",
			Message: "refresh_token must not exceed 255 characters",
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
	}
	if len(r.Email) < 6 {
		errs = append(errs, validator.ValidationError{
			Field:   "email",
			Message: "email must be at least 6 characters long",
		})
	}
	if len(r.Email) > 254 {
		errs = append(errs, validator.ValidationError{
			Field:   "email",
			Message: "email must not exceed 254 characters",
		})
	}
	if !validator.IsValidEmail(r.Email) {
		errs = append(errs, validator.ValidationError{
			Field:   "email",
			Message: "email must be a valid email address (letters, numbers, ., _, %, +, - allowed before @; must contain @; domain must contain letters, numbers, ., -; must end with a dot and at least 2 letters, e.g. user@example.com)",
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
	if len(r.Token) > 255 {
		errs = append(errs, validator.ValidationError{
			Field:   "token",
			Message: "token must not exceed 255 characters",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type SessionTrackingRequest struct {
	UserAgent string
	IPAddress string
}

type TokenResponse struct {
	AccessToken           string `json:"access_token"`
	AccessTokenExpiresIn  int64  `json:"access_token_expires_in"`
	RefreshToken          string `json:"refresh_token"`
	RefreshTokenExpiresIn int64  `json:"refresh_token_expires_in"`
}

type AccessTokenResponse struct {
	AccessToken          string `json:"access_token"`
	AccessTokenExpiresIn int64  `json:"access_token_expires_in"`
}
