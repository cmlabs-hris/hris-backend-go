package position

import "github.com/cmlabs-hris/hris-backend-go/internal/pkg/validator"

type CreatePositionRequest struct {
	Name string `json:"name"`
}

func (r *CreatePositionRequest) Validate() error {
	var errs validator.ValidationErrors

	if validator.IsEmpty(r.Name) {
		errs = append(errs, validator.ValidationError{
			Field:   "name",
			Message: "name is required",
		})
	} else if len(r.Name) > 100 {
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

type UpdatePositionRequest struct {
	ID        string `json:"id"`
	CompanyID string `json:"-"` // From JWT
	Name      string `json:"name"`
}

func (r *UpdatePositionRequest) Validate() error {
	var errs validator.ValidationErrors

	if validator.IsEmpty(r.ID) {
		errs = append(errs, validator.ValidationError{
			Field:   "id",
			Message: "id is required",
		})
	}

	if validator.IsEmpty(r.Name) {
		errs = append(errs, validator.ValidationError{
			Field:   "name",
			Message: "name is required",
		})
	} else if len(r.Name) > 100 {
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

type PositionResponse struct {
	ID        string `json:"id"`
	CompanyID string `json:"company_id"`
	Name      string `json:"name"`
}
