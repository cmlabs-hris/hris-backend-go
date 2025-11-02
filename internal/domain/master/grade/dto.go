package grade

import "github.com/cmlabs-hris/hris-backend-go/internal/pkg/validator"

type CreateGradeRequest struct {
	Name string `json:"name" validate:"required,max=100"`
}

func (r *CreateGradeRequest) Validate() error {
	var errs validator.ValidationErrors

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

type UpdateGradeRequest struct {
	ID   string `json:"id" validate:"required"`
	Name string `json:"name" validate:"required,max=100"`
}

func (r *UpdateGradeRequest) Validate() error {
	var errs validator.ValidationErrors

	// ID
	if validator.IsEmpty(r.ID) {
		errs = append(errs, validator.ValidationError{
			Field:   "id",
			Message: "id is required",
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

type GradeResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
