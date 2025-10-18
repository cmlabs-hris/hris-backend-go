package leave

import "github.com/cmlabs-hris/hris-backend-go/internal/pkg/validator"

type CreateLeaveTypeRequest struct {
	Name        string  `json:"leave_type_name"`
	Description *string `json:"leave_type_description,omitempty"`
}

func (r *CreateLeaveTypeRequest) Validate() error {
	var errs validator.ValidationErrors

	// Leave type name
	if validator.IsEmpty(r.Name) {
		errs = append(errs, validator.ValidationError{
			Field:   "leave_type_name",
			Message: "leave_type_name is required",
		})
	}
	if len(r.Name) > 255 {
		errs = append(errs, validator.ValidationError{
			Field:   "leave_type_name",
			Message: "leave_type_name must not exceed 255 characters",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type UpdateLeaveTypeRequest struct {
	ID          string  `json:"leave_type_id"`
	Name        *string `json:"leave_type_name,omitempty"`
	Description *string `json:"leave_type_description,omitempty"`
}

func (r *UpdateLeaveTypeRequest) Validate() error {
	var errs validator.ValidationErrors

	// Leave type id
	if validator.IsEmpty(r.ID) {
		errs = append(errs, validator.ValidationError{
			Field:   "leave_type_id",
			Message: "leave_type_id is required",
		})
	}

	// Leave type name
	if r.Name != nil {
		if validator.IsEmpty(*r.Name) {
			errs = append(errs, validator.ValidationError{
				Field:   "leave_type_name",
				Message: "leave_type_name must not be empty",
			})
		}
		if len(*r.Name) > 255 {
			errs = append(errs, validator.ValidationError{
				Field:   "leave_type_name",
				Message: "leave_type_name must not exceed 255 characters",
			})
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type CreateLeaveQuotaRequest struct {
	EmployeeID  string `json:"employee_id"`
	LeaveTypeID string `json:"leave_type_id"`
	Year        int    `json:"leave_type_year"`
	TotalQuota  int    `json:"total_quota"`
}

func (r *CreateLeaveQuotaRequest) Validate() error {
	var errs validator.ValidationErrors

	// Employee ID
	if validator.IsEmpty(r.EmployeeID) {
		errs = append(errs, validator.ValidationError{
			Field:   "employee_id",
			Message: "employee_id is required",
		})
	}

	// Leave Type ID
	if validator.IsEmpty(r.LeaveTypeID) {
		errs = append(errs, validator.ValidationError{
			Field:   "leave_type_id",
			Message: "leave_type_id is required",
		})
	}

	// Year
	if r.Year <= 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "leave_type_year",
			Message: "leave_type_year must be a positive integer",
		})
	}

	// Total Quota
	if r.TotalQuota < 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "total_quota",
			Message: "total_quota must not be negative",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type GetByEmployeeAndYearLeaveQuotaRequest struct {
	EmployeeID string `json:"employee_id"`
	Year       int    `json:"leave_quota_year"`
}

func (r *GetByEmployeeAndYearLeaveQuotaRequest) Validate() error {
	var errs validator.ValidationErrors

	// Employee ID
	if validator.IsEmpty(r.EmployeeID) {
		errs = append(errs, validator.ValidationError{
			Field:   "employee_id",
			Message: "employee_id is required",
		})
	}

	// Year
	if r.Year <= 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "leave_quota_year",
			Message: "leave_quota_year must be a positive integer",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type GetByEmployeeTypeYearLeaveQuotaRequest struct {
	EmployeeID  string `json:"employee_id"`
	LeaveTypeID string `json:"leave_type_id"`
	Year        int    `json:"leave_quota_year"`
}

func (r *GetByEmployeeTypeYearLeaveQuotaRequest) Validate() error {
	var errs validator.ValidationErrors

	// Employee ID
	if validator.IsEmpty(r.EmployeeID) {
		errs = append(errs, validator.ValidationError{
			Field:   "employee_id",
			Message: "employee_id is required",
		})
	}

	// Leave Type ID
	if validator.IsEmpty(r.LeaveTypeID) {
		errs = append(errs, validator.ValidationError{
			Field:   "leave_type_id",
			Message: "leave_type_id is required",
		})
	}

	// Year
	if r.Year <= 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "leave_quota_year",
			Message: "leave_quota_year must be a positive integer",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type UpdateLeaveQuotaRequest struct {
	ID          string  `json:"id"`
	EmployeeID  *string `json:"employee_id,omitempty"`
	LeaveTypeID *string `json:"leave_type_id,omitempty"`
	Year        *int    `json:"year,omitempty"`
	TotalQuota  *int    `json:"total_quota,omitempty"`
	TakenQuota  *int    `json:"taken_quota,omitempty"`
}

func (r *UpdateLeaveQuotaRequest) Validate() error {
	var errs validator.ValidationErrors

	// ID
	if validator.IsEmpty(r.ID) {
		errs = append(errs, validator.ValidationError{
			Field:   "id",
			Message: "id is required",
		})
	}

	// EmployeeID
	if r.EmployeeID != nil {
		if validator.IsEmpty(*r.EmployeeID) {
			errs = append(errs, validator.ValidationError{
				Field:   "employee_id",
				Message: "employee_id must not be empty",
			})
		}
	}

	// LeaveTypeID
	if r.LeaveTypeID != nil {
		if validator.IsEmpty(*r.LeaveTypeID) {
			errs = append(errs, validator.ValidationError{
				Field:   "leave_type_id",
				Message: "leave_type_id must not be empty",
			})
		}
	}

	// Year
	if r.Year != nil {
		if *r.Year <= 0 {
			errs = append(errs, validator.ValidationError{
				Field:   "year",
				Message: "year must be a positive integer",
			})
		}
	}

	// TotalQuota
	if r.TotalQuota != nil {
		if *r.TotalQuota < 0 {
			errs = append(errs, validator.ValidationError{
				Field:   "total_quota",
				Message: "total_quota must not be negative",
			})
		}
	}

	// TakenQuota
	if r.TakenQuota != nil {
		if *r.TakenQuota < 0 {
			errs = append(errs, validator.ValidationError{
				Field:   "taken_quota",
				Message: "taken_quota must not be negative",
			})
		}
	}

	// TakenQuota should not exceed TotalQuota if both are provided
	if r.TotalQuota != nil && r.TakenQuota != nil {
		if *r.TakenQuota > *r.TotalQuota {
			errs = append(errs, validator.ValidationError{
				Field:   "taken_quota",
				Message: "taken_quota must not exceed total_quota",
			})
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type DecrementQuotaRequest struct {
	QuotaID string `json:"quota_id"`
	Days    int    `json:"days"`
}

func (r *DecrementQuotaRequest) Validate() error {
	var errs validator.ValidationErrors

	// QuotaID
	if validator.IsEmpty(r.QuotaID) {
		errs = append(errs, validator.ValidationError{
			Field:   "quota_id",
			Message: "quota_id is required",
		})
	}

	// Days
	if r.Days <= 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "days",
			Message: "days must be a positive integer",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
