package branch

import (
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/validator"
)

// BranchResponse represents the response structure for a branch.
type BranchResponse struct {
	ID        string  `json:"id"`
	CompanyID string  `json:"company_id"`
	Name      string  `json:"name"`
	Address   *string `json:"address,omitempty"`
	Timezone  string  `json:"timezone"`
}

// CreateBranchRequest represents the request structure for creating a branch.
type CreateBranchRequest struct {
	CompanyID string  `json:"company_id"`
	Name      string  `json:"name"`
	Address   *string `json:"address,omitempty"`
}

func (r *CreateBranchRequest) Validate() error {
	var errs validator.ValidationErrors

	// CompanyID
	if validator.IsEmpty(r.CompanyID) {
		errs = append(errs, validator.ValidationError{
			Field:   "company_id",
			Message: "company_id is required",
		})
	}

	// Name
	if validator.IsEmpty(r.Name) {
		errs = append(errs, validator.ValidationError{
			Field:   "name",
			Message: "name is required",
		})
	}
	if len(r.Name) > 100 {
		errs = append(errs, validator.ValidationError{
			Field:   "name",
			Message: "name must not exceed 100 characters",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

// UpdateBranchRequest represents the request structure for updating a branch.
type UpdateBranchRequest struct {
	ID        string  `json:"id"`
	CompanyID string  `json:"-"` // From JWT
	Name      *string `json:"name,omitempty"`
	Address   *string `json:"address,omitempty"`
	Timezone  *string `json:"timezone,omitempty"`
}

func (r *UpdateBranchRequest) Validate() error {
	var errs validator.ValidationErrors

	// ID
	if validator.IsEmpty(r.ID) {
		errs = append(errs, validator.ValidationError{
			Field:   "id",
			Message: "id is required",
		})
	}

	// Name
	if r.Name != nil {
		if validator.IsEmpty(*r.Name) {
			errs = append(errs, validator.ValidationError{
				Field:   "name",
				Message: "name must not be empty",
			})
		}
		if len(*r.Name) > 100 {
			errs = append(errs, validator.ValidationError{
				Field:   "name",
				Message: "name must not exceed 100 characters",
			})
		}
	}

	// Address
	if r.Address != nil {
		if validator.IsEmpty(*r.Address) {
			errs = append(errs, validator.ValidationError{
				Field:   "address",
				Message: "address must not be empty",
			})
		}
	}

	// Timezone
	if r.Timezone != nil {
		if validator.IsEmpty(*r.Timezone) {
			errs = append(errs, validator.ValidationError{
				Field:   "timezone",
				Message: "timezone must not be empty if provided",
			})
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
