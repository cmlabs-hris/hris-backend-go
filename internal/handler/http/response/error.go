package response

import (
	"errors"
	"net/http"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/auth"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/employee"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/leave"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/validator"
)

// HandleError maps domain errors to HTTP responses
func HandleError(w http.ResponseWriter, err error) {
	// Check if it's a validation error
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		ValidationError(w, validationErrs.ToMap())
		return
	}

	// Auth domain errors
	switch {
	case errors.Is(err, auth.ErrInvalidCredentials):
		Unauthorized(w, err.Error())
	case errors.Is(err, auth.ErrTokenExpired):
		Unauthorized(w, "Token expired")
	case errors.Is(err, auth.ErrRefreshTokenRevoked):
		Unauthorized(w, "Refresh token revoked")
	case errors.Is(err, auth.ErrEmailNotVerified):
		Forbidden(w, "Email not verified")
	case errors.Is(err, auth.ErrUserNotFound):
		NotFound(w, "User not found")
	case errors.Is(err, auth.ErrCompanyNotFound):
		NotFound(w, "Company not found")

	// Employee domain errors
	case errors.Is(err, employee.ErrEmployeeNotFound):
		NotFound(w, "Employee not found")
	case errors.Is(err, employee.ErrEmployeeCodeExists):
		Conflict(w, "Employee code already exists")
	case errors.Is(err, employee.ErrNIKExists):
		Conflict(w, "NIK already registered")
	case errors.Is(err, employee.ErrEmailExists):
		Conflict(w, "Email already registered in this company")

	// Leave domain errors
	case errors.Is(err, leave.ErrLeaveRequestNotFound):
		NotFound(w, "Leave request not found")
	case errors.Is(err, leave.ErrInsufficientQuota):
		BadRequest(w, "Insufficient leave quota", nil)
	case errors.Is(err, leave.ErrLeaveRequestAlreadyProcessed):
		Conflict(w, "Leave request already processed")

	// Default
	default:
		InternalServerError(w, "An unexpected error occurred")
	}
}
