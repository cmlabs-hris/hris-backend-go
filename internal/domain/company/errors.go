package company

import "errors"

var (
	ErrCompanyNotFound              = errors.New("company not found")
	ErrCompanyUsernameExists        = errors.New("company username already exists")
	ErrInvalidCompanyUsernameFormat = errors.New("invalid company username format")
	ErrInvalidCompanyName           = errors.New("company name cannot be empty")
	ErrUpdatedAtBeforeCreatedAt     = errors.New("updated_at cannot be before created_at")
)
