package employee

import (
	"mime/multipart"
	"strings"

	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/validator"
	"github.com/shopspring/decimal"
)

// ========================================
// EMPLOYEE DTOs
// ========================================

// CreateEmployeeRequest for creating a new employee
type CreateEmployeeRequest struct {
	WorkScheduleID        string                `json:"work_schedule_id"`
	PositionID            string                `json:"position_id"`
	GradeID               string                `json:"grade_id"`
	BranchID              string                `json:"branch_id,omitempty"`
	EmployeeCode          string                `json:"employee_code"`
	FullName              string                `json:"full_name"`
	Email                 string                `json:"email"` // Required for invitation
	Role                  string                `json:"role"`  // "employee" (default) or "manager"
	NIK                   *string               `json:"nik,omitempty"`
	Gender                string                `json:"gender"`
	PhoneNumber           string                `json:"phone_number"`
	Address               *string               `json:"address,omitempty"`
	PlaceOfBirth          *string               `json:"place_of_birth,omitempty"`
	DOB                   *string               `json:"dob,omitempty"`
	Education             *string               `json:"education,omitempty"`
	HireDate              string                `json:"hire_date"`
	EmploymentType        string                `json:"employment_type"`
	WarningLetter         *string               `json:"warning_letter,omitempty"`
	BankName              *string               `json:"bank_name,omitempty"`
	BankAccountHolderName *string               `json:"bank_account_holder_name,omitempty"`
	BankAccountNumber     *string               `json:"bank_account_number,omitempty"`
	BaseSalary            *decimal.Decimal      `json:"base_salary,omitempty"`
	File                  multipart.File        `json:"-"`
	FileHeader            *multipart.FileHeader `json:"-"`
}

func (r *CreateEmployeeRequest) Validate() error {
	var errs validator.ValidationErrors

	if validator.IsEmpty(r.WorkScheduleID) {
		errs = append(errs, validator.ValidationError{
			Field:   "work_schedule_id",
			Message: "work_schedule_id is required",
		})
	}

	if validator.IsEmpty(r.PositionID) {
		errs = append(errs, validator.ValidationError{
			Field:   "position_id",
			Message: "position_id is required",
		})
	}

	if validator.IsEmpty(r.GradeID) {
		errs = append(errs, validator.ValidationError{
			Field:   "grade_id",
			Message: "grade_id is required",
		})
	}

	if validator.IsEmpty(r.GradeID) {
		errs = append(errs, validator.ValidationError{
			Field:   "branch_id",
			Message: "branch_id is required",
		})
	}

	if validator.IsEmpty(r.EmployeeCode) {
		errs = append(errs, validator.ValidationError{
			Field:   "employee_code",
			Message: "employee_code is required",
		})
	}

	if validator.IsEmpty(r.FullName) {
		errs = append(errs, validator.ValidationError{
			Field:   "full_name",
			Message: "full_name is required",
		})
	}

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

	// Role validation: default to "employee" if empty, only allow "employee" or "manager"
	if validator.IsEmpty(r.Role) {
		r.Role = "employee" // Default value
	} else {
		validRoles := []string{"employee", "manager"}
		if !validator.IsInSlice(strings.ToLower(r.Role), validRoles) {
			errs = append(errs, validator.ValidationError{
				Field:   "role",
				Message: "role must be either 'employee' or 'manager'",
			})
		} else {
			r.Role = strings.ToLower(r.Role) // Normalize to lowercase
		}
	}

	if validator.IsEmpty(r.Gender) {
		errs = append(errs, validator.ValidationError{
			Field:   "gender",
			Message: "gender is required",
		})
	} else if r.Gender != "Male" && r.Gender != "Female" {
		errs = append(errs, validator.ValidationError{
			Field:   "gender",
			Message: "gender must be Male or Female",
		})
	}

	if validator.IsEmpty(r.PhoneNumber) {
		errs = append(errs, validator.ValidationError{
			Field:   "phone_number",
			Message: "phone_number is required",
		})
	} else if len(r.PhoneNumber) < 10 || len(r.PhoneNumber) > 13 {
		errs = append(errs, validator.ValidationError{
			Field:   "phone_number",
			Message: "phone_number must be between 10 and 13 digits",
		})
	}

	if r.NIK != nil && *r.NIK != "" && len(*r.NIK) != 16 {
		errs = append(errs, validator.ValidationError{
			Field:   "nik",
			Message: "NIK must be exactly 16 digits",
		})
	}

	if validator.IsEmpty(r.HireDate) {
		errs = append(errs, validator.ValidationError{
			Field:   "hire_date",
			Message: "hire_date is required",
		})
	} else if _, valid := validator.IsValidDate(r.HireDate); !valid {
		errs = append(errs, validator.ValidationError{
			Field:   "hire_date",
			Message: "hire_date must be in YYYY-MM-DD format",
		})
	}

	if r.DOB != nil && *r.DOB != "" {
		if _, valid := validator.IsValidDate(*r.DOB); !valid {
			errs = append(errs, validator.ValidationError{
				Field:   "dob",
				Message: "dob must be in YYYY-MM-DD format",
			})
		}
	}

	if validator.IsEmpty(r.EmploymentType) {
		errs = append(errs, validator.ValidationError{
			Field:   "employment_type",
			Message: "employment_type is required",
		})
	} else {
		validTypes := []string{"permanent", "probation", "contract", "internship", "freelance"}
		if !validator.IsInSlice(strings.ToLower(r.EmploymentType), validTypes) {
			errs = append(errs, validator.ValidationError{
				Field:   "employment_type",
				Message: "employment_type must be one of: permanent, probation, contract, internship, freelance",
			})
		}
	}

	// Validate avatar file if provided
	if r.FileHeader != nil {
		filename := r.FileHeader.Filename
		ext := strings.ToLower(filename[strings.LastIndex(filename, "."):])
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			errs = append(errs, validator.ValidationError{
				Field:   "avatar",
				Message: "invalid file type: only jpg, jpeg, png allowed",
			})
		} else if r.FileHeader.Size > 5<<20 { // 5MB
			errs = append(errs, validator.ValidationError{
				Field:   "avatar",
				Message: "avatar size must not exceed 5MB",
			})
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

// UpdateEmployeeRequest for updating an existing employee
type UpdateEmployeeRequest struct {
	ID                    string           `json:"-"`
	PositionID            *string          `json:"position_id,omitempty"`
	GradeID               *string          `json:"grade_id,omitempty"`
	BranchID              *string          `json:"branch_id,omitempty"`
	EmployeeCode          *string          `json:"employee_code,omitempty"`
	FullName              *string          `json:"full_name,omitempty"`
	NIK                   *string          `json:"nik,omitempty"`
	Gender                *string          `json:"gender,omitempty"`
	PhoneNumber           *string          `json:"phone_number,omitempty"`
	Address               *string          `json:"address,omitempty"`
	PlaceOfBirth          *string          `json:"place_of_birth,omitempty"`
	DOB                   *string          `json:"dob,omitempty"`
	Education             *string          `json:"education,omitempty"`
	HireDate              *string          `json:"hire_date,omitempty"`
	ResignationDate       *string          `json:"resignation_date,omitempty"`
	EmploymentType        *string          `json:"employment_type,omitempty"`
	EmploymentStatus      *string          `json:"employment_status,omitempty"`
	WarningLetter         *string          `json:"warning_letter,omitempty"`
	BankName              *string          `json:"bank_name,omitempty"`
	BankAccountHolderName *string          `json:"bank_account_holder_name,omitempty"`
	BankAccountNumber     *string          `json:"bank_account_number,omitempty"`
	BaseSalary            *decimal.Decimal `json:"base_salary,omitempty"`
}

func (r *UpdateEmployeeRequest) Validate(role string) error {
	var errs validator.ValidationErrors

	// Field-level restrictions: employees can only update specific fields
	if role == "employee" {
		// Check if employee is trying to update restricted fields
		var restrictedFields []string

		if r.PositionID != nil {
			restrictedFields = append(restrictedFields, "position_id")
		}
		if r.GradeID != nil {
			restrictedFields = append(restrictedFields, "grade_id")
		}
		if r.BranchID != nil {
			restrictedFields = append(restrictedFields, "branch_id")
		}
		if r.EmployeeCode != nil {
			restrictedFields = append(restrictedFields, "employee_code")
		}
		if r.FullName != nil {
			restrictedFields = append(restrictedFields, "full_name")
		}
		if r.NIK != nil {
			restrictedFields = append(restrictedFields, "nik")
		}
		if r.Gender != nil {
			restrictedFields = append(restrictedFields, "gender")
		}
		if r.HireDate != nil {
			restrictedFields = append(restrictedFields, "hire_date")
		}
		if r.ResignationDate != nil {
			restrictedFields = append(restrictedFields, "resignation_date")
		}
		if r.EmploymentType != nil {
			restrictedFields = append(restrictedFields, "employment_type")
		}
		if r.EmploymentStatus != nil {
			restrictedFields = append(restrictedFields, "employment_status")
		}
		if r.WarningLetter != nil {
			restrictedFields = append(restrictedFields, "warning_letter")
		}
		if r.BaseSalary != nil {
			restrictedFields = append(restrictedFields, "base_salary")
		}

		if len(restrictedFields) > 0 {
			errs = append(errs, validator.ValidationError{
				Field:   "restricted_fields",
				Message: "employee cannot update the following fields: " + strings.Join(restrictedFields, ", ") + ". Only phone_number, address, place_of_birth, dob, education, and bank details can be updated",
			})
		}
	}

	if r.Gender != nil && *r.Gender != "" {
		if *r.Gender != "Male" && *r.Gender != "Female" {
			errs = append(errs, validator.ValidationError{
				Field:   "gender",
				Message: "gender must be Male or Female",
			})
		}
	}

	if r.PhoneNumber != nil && *r.PhoneNumber != "" {
		if len(*r.PhoneNumber) < 10 || len(*r.PhoneNumber) > 13 {
			errs = append(errs, validator.ValidationError{
				Field:   "phone_number",
				Message: "phone_number must be between 10 and 13 digits",
			})
		}
	}

	if r.NIK != nil && *r.NIK != "" && len(*r.NIK) != 16 {
		errs = append(errs, validator.ValidationError{
			Field:   "nik",
			Message: "NIK must be exactly 16 digits",
		})
	}

	if r.HireDate != nil && *r.HireDate != "" {
		if _, valid := validator.IsValidDate(*r.HireDate); !valid {
			errs = append(errs, validator.ValidationError{
				Field:   "hire_date",
				Message: "hire_date must be in YYYY-MM-DD format",
			})
		}
	}

	if r.ResignationDate != nil && *r.ResignationDate != "" {
		if _, valid := validator.IsValidDate(*r.ResignationDate); !valid {
			errs = append(errs, validator.ValidationError{
				Field:   "resignation_date",
				Message: "resignation_date must be in YYYY-MM-DD format",
			})
		}
	}

	if r.DOB != nil && *r.DOB != "" {
		if _, valid := validator.IsValidDate(*r.DOB); !valid {
			errs = append(errs, validator.ValidationError{
				Field:   "dob",
				Message: "dob must be in YYYY-MM-DD format",
			})
		}
	}

	if r.EmploymentType != nil && *r.EmploymentType != "" {
		validTypes := []string{"permanent", "probation", "contract", "internship", "freelance"}
		if !validator.IsInSlice(strings.ToLower(*r.EmploymentType), validTypes) {
			errs = append(errs, validator.ValidationError{
				Field:   "employment_type",
				Message: "employment_type must be one of: permanent, probation, contract, internship, freelance",
			})
		}
	}

	if r.EmploymentStatus != nil && *r.EmploymentStatus != "" {
		validStatuses := []string{"active", "resigned", "terminated"}
		if !validator.IsInSlice(strings.ToLower(*r.EmploymentStatus), validStatuses) {
			errs = append(errs, validator.ValidationError{
				Field:   "employment_status",
				Message: "employment_status must be one of: active, resigned, terminated",
			})
		}
	}

	if r.WarningLetter != nil && *r.WarningLetter != "" {
		validWarnings := []string{"light", "medium", "heavy"}
		if !validator.IsInSlice(strings.ToLower(*r.WarningLetter), validWarnings) {
			errs = append(errs, validator.ValidationError{
				Field:   "warning_letter",
				Message: "warning_letter must be one of: light, medium, heavy",
			})
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

// EmployeeResponse for returning employee data with joined names
type EmployeeResponse struct {
	ID                    string           `json:"id"`
	Email                 *string          `json:"email,omitempty"`
	UserID                *string          `json:"user_id,omitempty"`
	CompanyID             string           `json:"company_id"`
	WorkScheduleID        *string          `json:"work_schedule_id,omitempty"`
	WorkScheduleName      *string          `json:"work_schedule_name,omitempty"`
	PositionID            *string          `json:"position_id,omitempty"`
	PositionName          *string          `json:"position_name,omitempty"`
	GradeID               *string          `json:"grade_id,omitempty"`
	GradeName             *string          `json:"grade_name,omitempty"`
	BranchID              *string          `json:"branch_id,omitempty"`
	BranchName            *string          `json:"branch_name,omitempty"`
	EmployeeCode          string           `json:"employee_code"`
	FullName              string           `json:"full_name"`
	NIK                   *string          `json:"nik,omitempty"`
	Gender                string           `json:"gender"`
	PhoneNumber           string           `json:"phone_number"`
	Address               *string          `json:"address,omitempty"`
	PlaceOfBirth          *string          `json:"place_of_birth,omitempty"`
	DOB                   *string          `json:"dob,omitempty"`
	AvatarURL             *string          `json:"avatar_url,omitempty"`
	Education             *string          `json:"education,omitempty"`
	HireDate              string           `json:"hire_date"`
	ResignationDate       *string          `json:"resignation_date,omitempty"`
	EmploymentType        string           `json:"employment_type"`
	EmploymentStatus      string           `json:"employment_status"`
	WarningLetter         *string          `json:"warning_letter,omitempty"`
	BankName              *string          `json:"bank_name,omitempty"`
	BankAccountHolderName *string          `json:"bank_account_holder_name,omitempty"`
	BankAccountNumber     *string          `json:"bank_account_number,omitempty"`
	BaseSalary            *decimal.Decimal `json:"base_salary,omitempty"`
	CreatedAt             string           `json:"created_at"`
	UpdatedAt             string           `json:"updated_at"`
}

// EmployeeFilter for filtering employee list
type EmployeeFilter struct {
	// Search
	Search *string `json:"search,omitempty"` // Search by full_name, employee_code, nik

	// Filters
	WorkScheduleID   *string `json:"work_schedule_id,omitempty"`
	PositionID       *string `json:"position_id,omitempty"`
	GradeID          *string `json:"grade_id,omitempty"`
	BranchID         *string `json:"branch_id,omitempty"`
	EmploymentType   *string `json:"employment_type,omitempty"`
	EmploymentStatus *string `json:"employment_status,omitempty"`
	WarningLetter    *string `json:"warning_letter,omitempty"`

	// Date ranges
	HireDateFrom        *string `json:"hire_date_from,omitempty"`
	HireDateTo          *string `json:"hire_date_to,omitempty"`
	ResignationDateFrom *string `json:"resignation_date_from,omitempty"`
	ResignationDateTo   *string `json:"resignation_date_to,omitempty"`
	DOBFrom             *string `json:"dob_from,omitempty"`
	DOBTo               *string `json:"dob_to,omitempty"`

	// Pagination
	Page  int `json:"page"`
	Limit int `json:"limit"`

	// Sorting
	SortBy    string `json:"sort_by"`    // full_name, employee_code, hire_date, created_at
	SortOrder string `json:"sort_order"` // asc, desc
}

func (f *EmployeeFilter) Validate() error {
	var errs validator.ValidationErrors

	// Page validation
	if f.Page < 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "page",
			Message: "page must be a positive number",
		})
	}
	if f.Page == 0 {
		f.Page = 1 // Default page
	}

	// Limit validation
	if f.Limit < 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "limit",
			Message: "limit must be a positive number",
		})
	}
	if f.Limit == 0 {
		f.Limit = 20 // Default limit
	}
	if f.Limit > 100 {
		errs = append(errs, validator.ValidationError{
			Field:   "limit",
			Message: "limit must not exceed 100",
		})
	}

	// Employment type validation
	if f.EmploymentType != nil && *f.EmploymentType != "" {
		validTypes := []string{"permanent", "probation", "contract", "internship", "freelance"}
		if !validator.IsInSlice(strings.ToLower(*f.EmploymentType), validTypes) {
			errs = append(errs, validator.ValidationError{
				Field:   "employment_type",
				Message: "employment_type must be one of: permanent, probation, contract, internship, freelance",
			})
		}
	}

	// Employment status validation
	if f.EmploymentStatus != nil && *f.EmploymentStatus != "" {
		validStatuses := []string{"active", "resigned", "terminated"}
		if !validator.IsInSlice(strings.ToLower(*f.EmploymentStatus), validStatuses) {
			errs = append(errs, validator.ValidationError{
				Field:   "employment_status",
				Message: "employment_status must be one of: active, resigned, terminated",
			})
		}
	}

	// Warning letter validation
	if f.WarningLetter != nil && *f.WarningLetter != "" {
		validWarnings := []string{"light", "medium", "heavy"}
		if !validator.IsInSlice(strings.ToLower(*f.WarningLetter), validWarnings) {
			errs = append(errs, validator.ValidationError{
				Field:   "warning_letter",
				Message: "warning_letter must be one of: light, medium, heavy",
			})
		}
	}

	// Date validations
	if f.HireDateFrom != nil && *f.HireDateFrom != "" {
		if _, valid := validator.IsValidDate(*f.HireDateFrom); !valid {
			errs = append(errs, validator.ValidationError{
				Field:   "hire_date_from",
				Message: "hire_date_from must be in YYYY-MM-DD format",
			})
		}
	}
	if f.HireDateTo != nil && *f.HireDateTo != "" {
		if _, valid := validator.IsValidDate(*f.HireDateTo); !valid {
			errs = append(errs, validator.ValidationError{
				Field:   "hire_date_to",
				Message: "hire_date_to must be in YYYY-MM-DD format",
			})
		}
	}
	if f.ResignationDateFrom != nil && *f.ResignationDateFrom != "" {
		if _, valid := validator.IsValidDate(*f.ResignationDateFrom); !valid {
			errs = append(errs, validator.ValidationError{
				Field:   "resignation_date_from",
				Message: "resignation_date_from must be in YYYY-MM-DD format",
			})
		}
	}
	if f.ResignationDateTo != nil && *f.ResignationDateTo != "" {
		if _, valid := validator.IsValidDate(*f.ResignationDateTo); !valid {
			errs = append(errs, validator.ValidationError{
				Field:   "resignation_date_to",
				Message: "resignation_date_to must be in YYYY-MM-DD format",
			})
		}
	}
	if f.DOBFrom != nil && *f.DOBFrom != "" {
		if _, valid := validator.IsValidDate(*f.DOBFrom); !valid {
			errs = append(errs, validator.ValidationError{
				Field:   "dob_from",
				Message: "dob_from must be in YYYY-MM-DD format",
			})
		}
	}
	if f.DOBTo != nil && *f.DOBTo != "" {
		if _, valid := validator.IsValidDate(*f.DOBTo); !valid {
			errs = append(errs, validator.ValidationError{
				Field:   "dob_to",
				Message: "dob_to must be in YYYY-MM-DD format",
			})
		}
	}

	// Sort validation
	if f.SortBy != "" {
		validSortFields := []string{"full_name", "employee_code", "hire_date", "created_at", "employment_status"}
		if !validator.IsInSlice(f.SortBy, validSortFields) {
			errs = append(errs, validator.ValidationError{
				Field:   "sort_by",
				Message: "sort_by must be one of: full_name, employee_code, hire_date, created_at, employment_status",
			})
		}
	} else {
		f.SortBy = "created_at" // Default sort
	}

	if f.SortOrder != "" {
		validSortOrders := []string{"asc", "desc"}
		if !validator.IsInSlice(strings.ToLower(f.SortOrder), validSortOrders) {
			errs = append(errs, validator.ValidationError{
				Field:   "sort_order",
				Message: "sort_order must be one of: asc, desc",
			})
		}
	} else {
		f.SortOrder = "desc" // Default descending (newest first)
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

// ListEmployeeResponse for paginated list
type ListEmployeeResponse struct {
	TotalCount int64              `json:"total_count"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"total_pages"`
	Showing    string             `json:"showing"`
	Employees  []EmployeeResponse `json:"employees"`
}

// SearchEmployeeRequest for autocomplete search
type SearchEmployeeRequest struct {
	Query string `json:"query"` // Search query for full_name, employee_code
	Limit int    `json:"limit"` // Max results (default 10)
}

func (r *SearchEmployeeRequest) Validate() error {
	var errs validator.ValidationErrors

	if validator.IsEmpty(r.Query) {
		errs = append(errs, validator.ValidationError{
			Field:   "query",
			Message: "query is required",
		})
	}

	if r.Limit <= 0 {
		r.Limit = 10 // Default limit
	}
	if r.Limit > 50 {
		r.Limit = 50 // Max limit
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

// SearchEmployeeResponse for autocomplete results
type SearchEmployeeResponse struct {
	ID           string  `json:"id"`
	EmployeeCode string  `json:"employee_code"`
	FullName     string  `json:"full_name"`
	PositionName *string `json:"position_name,omitempty"`
	AvatarURL    *string `json:"avatar_url,omitempty"`
}

// InactivateEmployeeRequest for inactivating an employee
type InactivateEmployeeRequest struct {
	ID              string `json:"-"`
	ResignationDate string `json:"resignation_date"`
}

func (r *InactivateEmployeeRequest) Validate() error {
	var errs validator.ValidationErrors

	if validator.IsEmpty(r.ResignationDate) {
		errs = append(errs, validator.ValidationError{
			Field:   "resignation_date",
			Message: "resignation_date is required",
		})
	} else if _, valid := validator.IsValidDate(r.ResignationDate); !valid {
		errs = append(errs, validator.ValidationError{
			Field:   "resignation_date",
			Message: "resignation_date must be in YYYY-MM-DD format",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

// UploadAvatarRequest for avatar upload
type UploadAvatarRequest struct {
	EmployeeID string                `json:"-"`
	File       multipart.File        `json:"-"`
	FileHeader *multipart.FileHeader `json:"-"`
}

func (r *UploadAvatarRequest) Validate() error {
	var errs validator.ValidationErrors

	if r.FileHeader == nil {
		errs = append(errs, validator.ValidationError{
			Field:   "avatar",
			Message: "avatar file is required",
		})
	} else {
		filename := r.FileHeader.Filename
		ext := strings.ToLower(filename[strings.LastIndex(filename, "."):])
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			errs = append(errs, validator.ValidationError{
				Field:   "avatar",
				Message: "invalid file type: only jpg, jpeg, png allowed",
			})
		} else if r.FileHeader.Size > 5<<20 { // 5MB
			errs = append(errs, validator.ValidationError{
				Field:   "avatar",
				Message: "avatar size must not exceed 5MB",
			})
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
