package employee

import "errors"

var (
	ErrEmployeeNotFound        = errors.New("employee not found")
	ErrEmployeeCodeExists      = errors.New("employee code already exists")
	ErrNIKExists               = errors.New("NIK already registered")
	ErrEmailExists             = errors.New("email already registered in this company")
	ErrInvalidEmployeeCode     = errors.New("invalid employee code format")
	ErrInvalidNIK              = errors.New("NIK must be exactly 16 digits")
	ErrInvalidPhoneNumber      = errors.New("phone number must be 10-13 digits")
	ErrInvalidGender           = errors.New("gender must be Male or Female")
	ErrMinimumAge              = errors.New("employee must be at least 17 years old")
	ErrFutureDateNotAllowed    = errors.New("date cannot be in the future")
	ErrUnauthorized            = errors.New("unauthorized to access this employee")
	ErrEmployeeAlreadyActive   = errors.New("employee is already active")
	ErrEmployeeAlreadyInactive = errors.New("employee is already inactive")
	ErrCannotDeleteSelf        = errors.New("cannot delete your own employee record")
)
