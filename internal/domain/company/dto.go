package company

import "github.com/cmlabs-hris/hris-backend-go/internal/pkg/validator"

type CompanyResponse struct {
	Name     string  `json:"company_name"`
	Username string  `json:"company_username"`
	Address  *string `json:"company_address,omitempty"`
}

type CreateCompanyRequest struct {
	Name     string  `json:"company_name"`
	Username string  `json:"company_username"`
	Address  *string `json:"company_address"`
}

func (r *CreateCompanyRequest) Validate() error {
	var errs validator.ValidationErrors

	// Company
	if validator.IsEmpty(r.Name) {
		errs = append(errs, validator.ValidationError{
			Field:   "company_name",
			Message: "company_name is required",
		})
	}
	if validator.IsEmpty(r.Username) {
		errs = append(errs, validator.ValidationError{
			Field:   "company_username",
			Message: "company_username is required",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type UpdateCompanyRequest struct {
	Name    *string `json:"company_name"`
	Address *string `json:"company_address"`
}

func (r *UpdateCompanyRequest) Validate() error {
	var errs validator.ValidationErrors

	// Company
	if r.Name != nil {
		if len(*r.Name) > 255 {
			errs = append(errs, validator.ValidationError{
				Field:   "company_name",
				Message: "company_name must not exceed 255 characters",
			})
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
