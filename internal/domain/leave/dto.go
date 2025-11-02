package leave

import (
	"encoding/json"
	"mime/multipart"
	"strings"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/validator"
)

// ========================================
// LEAVE TYPE DTOs
// ========================================

type CreateLeaveTypeRequest struct {
	Name        string  `json:"name"`
	Code        *string `json:"code,omitempty"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty"`

	IsActive                    *bool `json:"is_active,omitempty"`
	RequiresApproval            *bool `json:"requires_approval,omitempty"`
	RequiresAttachment          *bool `json:"requires_attachment,omitempty"`
	AttachmentRequiredAfterDays *int  `json:"attachment_required_after_days,omitempty"`

	HasQuota      *bool   `json:"has_quota,omitempty"`
	AccrualMethod *string `json:"accrual_method,omitempty"`

	DeductionType *string `json:"deduction_type,omitempty"`
	AllowHalfDay  *bool   `json:"allow_half_day,omitempty"`

	MaxDaysPerRequest *int  `json:"max_days_per_request,omitempty"`
	MinNoticeDays     *int  `json:"min_notice_days,omitempty"`
	MaxAdvanceDays    *int  `json:"max_advance_days,omitempty"`
	AllowBackdate     *bool `json:"allow_backdate,omitempty"`
	BackdateMaxDays   *int  `json:"backdate_max_days,omitempty"`

	AllowRollover       *bool `json:"allow_rollover,omitempty"`
	MaxRolloverDays     *int  `json:"max_rollover_days,omitempty"`
	RolloverExpiryMonth *int  `json:"rollover_expiry_month,omitempty"`

	QuotaCalculationType string                 `json:"quota_calculation_type"`
	QuotaRules           map[string]interface{} `json:"quota_rules"`
}

func (r *CreateLeaveTypeRequest) Validate() error {
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

	// Code
	if r.Code != nil && len(*r.Code) > 50 {
		errs = append(errs, validator.ValidationError{
			Field:   "code",
			Message: "code must not exceed 50 characters",
		})
	}

	// Description
	if r.Description != nil && len(*r.Description) > 1000 {
		errs = append(errs, validator.ValidationError{
			Field:   "description",
			Message: "description must not exceed 1000 characters",
		})
	}

	// Color
	if r.Color != nil {
		if len(*r.Color) > 7 {
			errs = append(errs, validator.ValidationError{
				Field:   "color",
				Message: "color must not exceed 7 characters",
			})
		}
		if len(*r.Color) > 0 && (*r.Color)[0] != '#' {
			errs = append(errs, validator.ValidationError{
				Field:   "color",
				Message: "color must start with #",
			})
		}
	}

	// AttachmentRequiredAfterDays
	if r.AttachmentRequiredAfterDays != nil && *r.AttachmentRequiredAfterDays < 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "attachment_required_after_days",
			Message: "attachment_required_after_days must not be negative",
		})
	}

	// AccrualMethod
	if r.AccrualMethod != nil {
		validAccrual := []string{"yearly", "monthly", "daily", "none"}
		if !validator.IsInSlice(*r.AccrualMethod, validAccrual) {
			errs = append(errs, validator.ValidationError{
				Field:   "accrual_method",
				Message: "accrual_method must be one of: yearly, monthly, daily, none",
			})
		}
	}

	// DeductionType
	if r.DeductionType != nil {
		validDeduction := []string{"working_days", "calendar_days"}
		if !validator.IsInSlice(*r.DeductionType, validDeduction) {
			errs = append(errs, validator.ValidationError{
				Field:   "deduction_type",
				Message: "deduction_type must be one of: working_days, calendar_days",
			})
		}
	}

	// MaxDaysPerRequest
	if r.MaxDaysPerRequest != nil && *r.MaxDaysPerRequest < 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "max_days_per_request",
			Message: "max_days_per_request must not be negative",
		})
	}

	// MinNoticeDays
	if r.MinNoticeDays != nil && *r.MinNoticeDays < 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "min_notice_days",
			Message: "min_notice_days must not be negative",
		})
	}

	// MaxAdvanceDays
	if r.MaxAdvanceDays != nil && *r.MaxAdvanceDays < 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "max_advance_days",
			Message: "max_advance_days must not be negative",
		})
	}

	// BackdateMaxDays
	if r.BackdateMaxDays != nil && *r.BackdateMaxDays < 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "backdate_max_days",
			Message: "backdate_max_days must not be negative",
		})
	}

	// MaxRolloverDays
	if r.MaxRolloverDays != nil && *r.MaxRolloverDays < 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "max_rollover_days",
			Message: "max_rollover_days must not be negative",
		})
	}

	// RolloverExpiryMonth
	if r.RolloverExpiryMonth != nil {
		if *r.RolloverExpiryMonth < 1 || *r.RolloverExpiryMonth > 12 {
			errs = append(errs, validator.ValidationError{
				Field:   "rollover_expiry_month",
				Message: "rollover_expiry_month must be between 1 and 12",
			})
		}
	}

	// QuotaCalculationType
	if validator.IsEmpty(r.QuotaCalculationType) {
		errs = append(errs, validator.ValidationError{
			Field:   "quota_calculation_type",
			Message: "quota_calculation_type is required",
		})
	} else {
		validQuotaCalc := []string{"fixed", "tenure_based", "position_based", "employment_type_based", "grade_based"}
		if !validator.IsInSlice(r.QuotaCalculationType, validQuotaCalc) {
			errs = append(errs, validator.ValidationError{
				Field:   "quota_calculation_type",
				Message: "quota_calculation_type must be one of: fixed, tenure_based, position_based, employment_type_based, grade_based",
			})
		}
	}

	// QuotaRules
	if r.QuotaRules == nil {
		errs = append(errs, validator.ValidationError{
			Field:   "quota_rules",
			Message: "quota_rules is required",
		})
	} else {
		// Validate QuotaRules structure
		// Try to marshal/unmarshal to QuotaRules struct
		qrBytes, err := json.Marshal(r.QuotaRules)
		if err != nil {
			errs = append(errs, validator.ValidationError{
				Field:   "quota_rules",
				Message: "quota_rules must be a valid object",
			})
		} else {
			var qr QuotaRules
			if err := json.Unmarshal(qrBytes, &qr); err != nil {
				errs = append(errs, validator.ValidationError{
					Field:   "quota_rules",
					Message: "quota_rules structure is invalid",
				})
			} else {
				// Validate QuotaRules.Type
				validTypes := []string{"fixed", "tenure", "position", "grade", "employment_type", "combined"}
				if validator.IsEmpty(qr.Type) {
					errs = append(errs, validator.ValidationError{
						Field:   "quota_rules.type",
						Message: "quota_rules.type is required",
					})
				} else if !validator.IsInSlice(qr.Type, validTypes) {
					errs = append(errs, validator.ValidationError{
						Field:   "quota_rules.type",
						Message: "quota_rules.type must be one of: fixed, tenure, position, grade, employment_type, combined",
					})
				}
				// Validate each rule
				for i, rule := range qr.Rules {
					if rule.Quota < 0 {
						errs = append(errs, validator.ValidationError{
							Field:   "quota_rules.rules[" + validator.Itoa(i) + "].quota",
							Message: "quota must not be negative",
						})
					}
					// For tenure-based, min/max months must be >= 0
					if rule.MinMonths != nil && *rule.MinMonths < 0 {
						errs = append(errs, validator.ValidationError{
							Field:   "quota_rules.rules[" + validator.Itoa(i) + "].min_months",
							Message: "min_months must not be negative",
						})
					}
					if rule.MaxMonths != nil && *rule.MaxMonths < 0 {
						errs = append(errs, validator.ValidationError{
							Field:   "quota_rules.rules[" + validator.Itoa(i) + "].max_months",
							Message: "max_months must not be negative",
						})
					}
					// For combined rules, validate conditions
					if rule.Conditions != nil {
						if rule.Conditions.MinTenureMonths != nil && *rule.Conditions.MinTenureMonths < 0 {
							errs = append(errs, validator.ValidationError{
								Field:   "quota_rules.rules[" + validator.Itoa(i) + "].conditions.min_tenure_months",
								Message: "min_tenure_months must not be negative",
							})
						}
						if rule.Conditions.MaxTenureMonths != nil && *rule.Conditions.MaxTenureMonths < 0 {
							errs = append(errs, validator.ValidationError{
								Field:   "quota_rules.rules[" + validator.Itoa(i) + "].conditions.max_tenure_months",
								Message: "max_tenure_months must not be negative",
							})
						}
					}
				}
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type LeaveTypeResponse struct {
	ID                   string     `json:"id"`
	CompanyID            string     `json:"company_id"`
	Name                 string     `json:"name"`
	Code                 *string    `json:"code,omitempty"`
	Description          *string    `json:"description,omitempty"`
	Color                *string    `json:"color,omitempty"`
	IsActive             *bool      `json:"is_active,omitempty"`
	RequiresApproval     *bool      `json:"requires_approval,omitempty"`
	HasQuota             *bool      `json:"has_quota,omitempty"`
	AccrualMethod        *string    `json:"accrual_method,omitempty"`
	QuotaCalculationType string     `json:"quota_calculation_type"`
	QuotaRules           QuotaRules `json:"quota_rules"`
}

type UpdateLeaveTypeRequest struct {
	ID                          string                 `json:"id"`
	Name                        *string                `json:"name,omitempty"`
	Code                        *string                `json:"code,omitempty"`
	Description                 *string                `json:"description,omitempty"`
	Color                       *string                `json:"color,omitempty"`
	IsActive                    *bool                  `json:"is_active,omitempty"`
	RequiresApproval            *bool                  `json:"requires_approval,omitempty"`
	RequiresAttachment          *bool                  `json:"requires_attachment,omitempty"`
	AttachmentRequiredAfterDays *int                   `json:"attachment_required_after_days,omitempty"`
	HasQuota                    *bool                  `json:"has_quota,omitempty"`
	AccrualMethod               *string                `json:"accrual_method,omitempty"`
	DeductionType               *string                `json:"deduction_type,omitempty"`
	AllowHalfDay                *bool                  `json:"allow_half_day,omitempty"`
	MaxDaysPerRequest           *int                   `json:"max_days_per_request,omitempty"`
	MinNoticeDays               *int                   `json:"min_notice_days,omitempty"`
	MaxAdvanceDays              *int                   `json:"max_advance_days,omitempty"`
	AllowBackdate               *bool                  `json:"allow_backdate,omitempty"`
	BackdateMaxDays             *int                   `json:"backdate_max_days,omitempty"`
	AllowRollover               *bool                  `json:"allow_rollover,omitempty"`
	MaxRolloverDays             *int                   `json:"max_rollover_days,omitempty"`
	RolloverExpiryMonth         *int                   `json:"rollover_expiry_month,omitempty"`
	QuotaCalculationType        *string                `json:"quota_calculation_type,omitempty"`
	QuotaRules                  map[string]interface{} `json:"quota_rules,omitempty"`
}

func (r *UpdateLeaveTypeRequest) Validate() error {
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

	// Code
	if r.Code != nil {
		if len(*r.Code) > 50 {
			errs = append(errs, validator.ValidationError{
				Field:   "code",
				Message: "code must not exceed 50 characters",
			})
		}
	}

	// Description
	if r.Description != nil {
		if len(*r.Description) > 1000 {
			errs = append(errs, validator.ValidationError{
				Field:   "description",
				Message: "description must not exceed 1000 characters",
			})
		}
	}

	// Color
	if r.Color != nil {
		if len(*r.Color) > 7 {
			errs = append(errs, validator.ValidationError{
				Field:   "color",
				Message: "color must not exceed 7 characters",
			})
		}
		if len(*r.Color) > 0 && (*r.Color)[0] != '#' {
			errs = append(errs, validator.ValidationError{
				Field:   "color",
				Message: "color must start with #",
			})
		}
	}

	// IsActive
	if r.IsActive != nil {
		// BOOLEAN NOT NULL, no further validation needed
	}

	// RequiresApproval
	if r.RequiresApproval != nil {
		// BOOLEAN NOT NULL, no further validation needed
	}

	// RequiresAttachment
	if r.RequiresAttachment != nil {
		// BOOLEAN NOT NULL, no further validation needed
	}

	// AttachmentRequiredAfterDays
	if r.AttachmentRequiredAfterDays != nil {
		if *r.AttachmentRequiredAfterDays < 0 || *r.AttachmentRequiredAfterDays > 32767 {
			errs = append(errs, validator.ValidationError{
				Field:   "attachment_required_after_days",
				Message: "attachment_required_after_days must be between 0 and 32767",
			})
		}
	}

	// HasQuota
	if r.HasQuota != nil {
		// BOOLEAN NOT NULL, no further validation needed
	}

	// AccrualMethod
	if r.AccrualMethod != nil {
		validAccrual := []string{"yearly", "monthly", "daily", "none"}
		if !validator.IsInSlice(*r.AccrualMethod, validAccrual) {
			errs = append(errs, validator.ValidationError{
				Field:   "accrual_method",
				Message: "accrual_method must be one of: yearly, monthly, daily, none",
			})
		}
		if len(*r.AccrualMethod) > 20 {
			errs = append(errs, validator.ValidationError{
				Field:   "accrual_method",
				Message: "accrual_method must not exceed 20 characters",
			})
		}
	}

	// DeductionType
	if r.DeductionType != nil {
		validDeduction := []string{"working_days", "calendar_days"}
		if !validator.IsInSlice(*r.DeductionType, validDeduction) {
			errs = append(errs, validator.ValidationError{
				Field:   "deduction_type",
				Message: "deduction_type must be one of: working_days, calendar_days",
			})
		}
		if len(*r.DeductionType) > 20 {
			errs = append(errs, validator.ValidationError{
				Field:   "deduction_type",
				Message: "deduction_type must not exceed 20 characters",
			})
		}
	}

	// AllowHalfDay
	if r.AllowHalfDay != nil {
		// BOOLEAN NOT NULL, no further validation needed
	}

	// MaxDaysPerRequest
	if r.MaxDaysPerRequest != nil {
		if *r.MaxDaysPerRequest < 0 || *r.MaxDaysPerRequest > 32767 {
			errs = append(errs, validator.ValidationError{
				Field:   "max_days_per_request",
				Message: "max_days_per_request must be between 0 and 32767",
			})
		}
	}

	// MinNoticeDays
	if r.MinNoticeDays != nil {
		if *r.MinNoticeDays < 0 || *r.MinNoticeDays > 32767 {
			errs = append(errs, validator.ValidationError{
				Field:   "min_notice_days",
				Message: "min_notice_days must be between 0 and 32767",
			})
		}
	}

	// MaxAdvanceDays
	if r.MaxAdvanceDays != nil {
		if *r.MaxAdvanceDays < 0 || *r.MaxAdvanceDays > 32767 {
			errs = append(errs, validator.ValidationError{
				Field:   "max_advance_days",
				Message: "max_advance_days must be between 0 and 32767",
			})
		}
	}

	// AllowBackdate
	if r.AllowBackdate != nil {
		// BOOLEAN NOT NULL, no further validation needed
	}

	// BackdateMaxDays
	if r.BackdateMaxDays != nil {
		if *r.BackdateMaxDays < 0 || *r.BackdateMaxDays > 32767 {
			errs = append(errs, validator.ValidationError{
				Field:   "backdate_max_days",
				Message: "backdate_max_days must be between 0 and 32767",
			})
		}
	}

	// AllowRollover
	if r.AllowRollover != nil {
		// BOOLEAN NOT NULL, no further validation needed
	}

	// MaxRolloverDays
	if r.MaxRolloverDays != nil {
		if *r.MaxRolloverDays < 0 || *r.MaxRolloverDays > 32767 {
			errs = append(errs, validator.ValidationError{
				Field:   "max_rollover_days",
				Message: "max_rollover_days must be between 0 and 32767",
			})
		}
	}

	// RolloverExpiryMonth
	if r.RolloverExpiryMonth != nil {
		if *r.RolloverExpiryMonth < 1 || *r.RolloverExpiryMonth > 12 {
			errs = append(errs, validator.ValidationError{
				Field:   "rollover_expiry_month",
				Message: "rollover_expiry_month must be between 1 and 12",
			})
		}
	}

	// QuotaCalculationType
	if r.QuotaCalculationType != nil {
		validQuotaCalc := []string{"fixed", "tenure_based", "position_based", "employment_type_based", "grade_based"}
		if !validator.IsInSlice(*r.QuotaCalculationType, validQuotaCalc) {
			errs = append(errs, validator.ValidationError{
				Field:   "quota_calculation_type",
				Message: "quota_calculation_type must be one of: fixed, tenure_based, position_based, employment_type_based, grade_based",
			})
		}
		if len(*r.QuotaCalculationType) > 20 {
			errs = append(errs, validator.ValidationError{
				Field:   "quota_calculation_type",
				Message: "quota_calculation_type must not exceed 20 characters",
			})
		}
	}

	// QuotaRules
	if r.QuotaRules != nil {
		qrBytes, err := json.Marshal(r.QuotaRules)
		if err != nil {
			errs = append(errs, validator.ValidationError{
				Field:   "quota_rules",
				Message: "quota_rules must be a valid object",
			})
		} else {
			var qr QuotaRules
			if err := json.Unmarshal(qrBytes, &qr); err != nil {
				errs = append(errs, validator.ValidationError{
					Field:   "quota_rules",
					Message: "quota_rules structure is invalid",
				})
			} else {
				validTypes := []string{"fixed", "tenure", "position", "grade", "employment_type", "combined"}
				if validator.IsEmpty(qr.Type) {
					errs = append(errs, validator.ValidationError{
						Field:   "quota_rules.type",
						Message: "quota_rules.type is required",
					})
				} else if !validator.IsInSlice(qr.Type, validTypes) {
					errs = append(errs, validator.ValidationError{
						Field:   "quota_rules.type",
						Message: "quota_rules.type must be one of: fixed, tenure, position, grade, employment_type, combined",
					})
				}
				for i, rule := range qr.Rules {
					if rule.Quota < 0 {
						errs = append(errs, validator.ValidationError{
							Field:   "quota_rules.rules[" + validator.Itoa(i) + "].quota",
							Message: "quota must not be negative",
						})
					}
					if rule.MinMonths != nil && *rule.MinMonths < 0 {
						errs = append(errs, validator.ValidationError{
							Field:   "quota_rules.rules[" + validator.Itoa(i) + "].min_months",
							Message: "min_months must not be negative",
						})
					}
					if rule.MaxMonths != nil && *rule.MaxMonths < 0 {
						errs = append(errs, validator.ValidationError{
							Field:   "quota_rules.rules[" + validator.Itoa(i) + "].max_months",
							Message: "max_months must not be negative",
						})
					}
					if rule.Conditions != nil {
						if rule.Conditions.MinTenureMonths != nil && *rule.Conditions.MinTenureMonths < 0 {
							errs = append(errs, validator.ValidationError{
								Field:   "quota_rules.rules[" + validator.Itoa(i) + "].conditions.min_tenure_months",
								Message: "min_tenure_months must not be negative",
							})
						}
						if rule.Conditions.MaxTenureMonths != nil && *rule.Conditions.MaxTenureMonths < 0 {
							errs = append(errs, validator.ValidationError{
								Field:   "quota_rules.rules[" + validator.Itoa(i) + "].conditions.max_tenure_months",
								Message: "max_tenure_months must not be negative",
							})
						}
					}
				}
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type LeaveRequestFilter struct {
	// Search & Filter
	EmployeeID   *string `json:"employee_id,omitempty"`
	EmployeeName *string `json:"employee_name,omitempty"` // ← NEW: Search by name
	LeaveTypeID  *string `json:"leave_type_id,omitempty"`
	Status       *string `json:"status,omitempty"`
	StartDate    *string `json:"start_date,omitempty"`
	EndDate      *string `json:"end_date,omitempty"`

	// Pagination
	Page  int `json:"page"`
	Limit int `json:"limit"`

	// Sorting
	SortBy    string `json:"sort_by"`    // ← NEW: submitted_at, employee_name, start_date
	SortOrder string `json:"sort_order"` // ← NEW: asc, desc
}

// MyLeaveRequestFilter - Filter for /leave/requests/my (no employee filters)
type MyLeaveRequestFilter struct {
	// Search & Filter (no employee_id/employee_name)
	LeaveTypeID *string `json:"leave_type_id,omitempty"`
	Status      *string `json:"status,omitempty"`
	StartDate   *string `json:"start_date,omitempty"`
	EndDate     *string `json:"end_date,omitempty"`

	// Pagination
	Page  int `json:"page"`
	Limit int `json:"limit"`

	// Sorting
	SortBy    string `json:"sort_by"`    // submitted_at, start_date, end_date, status
	SortOrder string `json:"sort_order"` // asc, desc
}

func (f *LeaveRequestFilter) Validate() error {
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

	// Status validation
	if f.Status != nil {
		validStatuses := []string{"waiting_approval", "approved", "rejected", "cancelled"}
		if !validator.IsInSlice(*f.Status, validStatuses) {
			errs = append(errs, validator.ValidationError{
				Field:   "status",
				Message: "status must be one of: waiting_approval, approved, rejected, cancelled",
			})
		}
	}

	// Date validation
	if f.StartDate != nil && *f.StartDate != "" {
		if _, valid := validator.IsValidDate(*f.StartDate); !valid {
			errs = append(errs, validator.ValidationError{
				Field:   "start_date",
				Message: "start_date must be in YYYY-MM-DD format",
			})
		}
	}

	if f.EndDate != nil && *f.EndDate != "" {
		if _, valid := validator.IsValidDate(*f.EndDate); !valid {
			errs = append(errs, validator.ValidationError{
				Field:   "end_date",
				Message: "end_date must be in YYYY-MM-DD format",
			})
		}
	}

	// Sort validation
	if f.SortBy != "" {
		validSortFields := []string{"submitted_at", "employee_name", "start_date", "end_date", "status"}
		if !validator.IsInSlice(f.SortBy, validSortFields) {
			errs = append(errs, validator.ValidationError{
				Field:   "sort_by",
				Message: "sort_by must be one of: submitted_at, employee_name, start_date, end_date, status",
			})
		}
	} else {
		f.SortBy = "submitted_at" // Default sort
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

func (f *MyLeaveRequestFilter) Validate() error {
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

	// Status validation
	if f.Status != nil {
		validStatuses := []string{"waiting_approval", "approved", "rejected", "cancelled"}
		if !validator.IsInSlice(*f.Status, validStatuses) {
			errs = append(errs, validator.ValidationError{
				Field:   "status",
				Message: "status must be one of: waiting_approval, approved, rejected, cancelled",
			})
		}
	}

	// Date validation
	if f.StartDate != nil && *f.StartDate != "" {
		if _, valid := validator.IsValidDate(*f.StartDate); !valid {
			errs = append(errs, validator.ValidationError{
				Field:   "start_date",
				Message: "start_date must be in YYYY-MM-DD format",
			})
		}
	}

	if f.EndDate != nil && *f.EndDate != "" {
		if _, valid := validator.IsValidDate(*f.EndDate); !valid {
			errs = append(errs, validator.ValidationError{
				Field:   "end_date",
				Message: "end_date must be in YYYY-MM-DD format",
			})
		}
	}

	// Sort validation (no employee_name for my requests)
	if f.SortBy != "" {
		validSortFields := []string{"submitted_at", "start_date", "end_date", "status"}
		if !validator.IsInSlice(f.SortBy, validSortFields) {
			errs = append(errs, validator.ValidationError{
				Field:   "sort_by",
				Message: "sort_by must be one of: submitted_at, start_date, end_date, status",
			})
		}
	} else {
		f.SortBy = "submitted_at" // Default sort
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

// LeaveRequestResponse - Enhanced with employee details
type LeaveRequestResponse struct {
	ID              string     `json:"id"`
	EmployeeID      string     `json:"employee_id"`
	EmployeeName    string     `json:"employee_name"`
	LeaveTypeID     string     `json:"leave_type_id"`
	LeaveTypeName   string     `json:"leave_type_name"`
	StartDate       time.Time  `json:"start_date"`
	EndDate         time.Time  `json:"end_date"`
	DurationType    string     `json:"duration_type"`
	TotalDays       float64    `json:"total_days"`
	WorkingDays     float64    `json:"working_days"`
	Reason          string     `json:"reason"`
	AttachmentURL   *string    `json:"attachment_url,omitempty"`
	Status          string     `json:"status"`
	SubmittedAt     time.Time  `json:"submitted_at"`
	ApprovedBy      *string    `json:"approved_by,omitempty"`
	ApprovedAt      *time.Time `json:"approved_at,omitempty"`
	RejectionReason *string    `json:"rejection_reason,omitempty"`
}

// ListLeaveRequestResponse - Enhanced with pagination metadata
type ListLeaveRequestResponse struct {
	TotalCount int64                  `json:"total_count"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
	TotalPages int                    `json:"total_pages"` // ← NEW
	Showing    string                 `json:"showing"`     // ← NEW: "21-40 of 150 results"
	Requests   []LeaveRequestResponse `json:"requests"`
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
	ID                 string   `json:"id"`
	EmployeeID         *string  `json:"employee_id,omitempty"`
	LeaveTypeID        *string  `json:"leave_type_id,omitempty"`
	Year               *int     `json:"year,omitempty"`
	OpeningBalance     *int     `json:"opening_balance,omitempty"`
	EarnedQuota        *int     `json:"earned_quota,omitempty"`
	RolloverQuota      *int     `json:"rollover_quota,omitempty"`
	AdjustmentQuota    *int     `json:"adjustment_quota,omitempty"`
	UsedQuota          *float64 `json:"used_quota,omitempty"`
	PendingQuota       *float64 `json:"pending_quota,omitempty"`
	RolloverExpiryDate *string  `json:"rollover_expiry_date,omitempty"` // ISO8601 date string
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
	if r.EmployeeID != nil && validator.IsEmpty(*r.EmployeeID) {
		errs = append(errs, validator.ValidationError{
			Field:   "employee_id",
			Message: "employee_id must not be empty",
		})
	}

	// LeaveTypeID
	if r.LeaveTypeID != nil && validator.IsEmpty(*r.LeaveTypeID) {
		errs = append(errs, validator.ValidationError{
			Field:   "leave_type_id",
			Message: "leave_type_id must not be empty",
		})
	}

	// Year
	if r.Year != nil && *r.Year <= 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "year",
			Message: "year must be a positive integer",
		})
	}

	// OpeningBalance
	if r.OpeningBalance != nil && *r.OpeningBalance < 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "opening_balance",
			Message: "opening_balance must not be negative",
		})
	}

	// EarnedQuota
	if r.EarnedQuota != nil && *r.EarnedQuota < 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "earned_quota",
			Message: "earned_quota must not be negative",
		})
	}

	// RolloverQuota
	if r.RolloverQuota != nil && *r.RolloverQuota < 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "rollover_quota",
			Message: "rollover_quota must not be negative",
		})
	}

	// AdjustmentQuota
	if r.AdjustmentQuota != nil && *r.AdjustmentQuota < 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "adjustment_quota",
			Message: "adjustment_quota must not be negative",
		})
	}

	// UsedQuota
	if r.UsedQuota != nil && *r.UsedQuota < 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "used_quota",
			Message: "used_quota must not be negative",
		})
	}

	// PendingQuota
	if r.PendingQuota != nil && *r.PendingQuota < 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "pending_quota",
			Message: "pending_quota must not be negative",
		})
	}

	// RolloverExpiryDate
	if r.RolloverExpiryDate != nil && *r.RolloverExpiryDate != "" {
		if _, valid := validator.IsValidDate(*r.RolloverExpiryDate); !valid {
			errs = append(errs, validator.ValidationError{
				Field:   "rollover_expiry_date",
				Message: "rollover_expiry_date must be a valid date (YYYY-MM-DD)",
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

// ========================================
// LEAVE REQUEST DTOs
// ========================================

type CreateLeaveRequestRequest struct {
	EmployeeID    string                `json:"employee_id"`
	LeaveTypeID   string                `json:"leave_type_id"`
	StartDate     string                `json:"start_date"` // "2024-01-15"
	EndDate       string                `json:"end_date"`
	DurationType  string                `json:"duration_type"` // 'full_day', 'half_day_morning', 'half_day_afternoon'
	Reason        string                `json:"reason"`
	AttachmentURL *string               `json:"-"`
	File          multipart.File        `json:"-"`
	FileHeader    *multipart.FileHeader `json:"-"`
}

func (r *CreateLeaveRequestRequest) Validate() error {
	var errs validator.ValidationErrors

	if validator.IsEmpty(r.EmployeeID) {
		errs = append(errs, validator.ValidationError{
			Field:   "employee_id",
			Message: "employee ID is required",
		})
	}
	if validator.IsEmpty(r.LeaveTypeID) {
		errs = append(errs, validator.ValidationError{
			Field:   "leave_type_id",
			Message: "leave type ID is required",
		})
	} else if !validator.IsValidUUID(r.LeaveTypeID) {
		errs = append(errs, validator.ValidationError{
			Field:   "leave_type_id",
			Message: "leave type ID must be a valid UUID",
		})
	}

	if validator.IsEmpty(r.StartDate) {
		errs = append(errs, validator.ValidationError{
			Field:   "start_date",
			Message: "start date is required",
		})
	} else {
		if _, valid := validator.IsValidDate(r.StartDate); !valid {
			errs = append(errs, validator.ValidationError{
				Field:   "start_date",
				Message: "start date format is invalid (use YYYY-MM-DD)",
			})
		}
	}

	if validator.IsEmpty(r.EndDate) {
		errs = append(errs, validator.ValidationError{
			Field:   "end_date",
			Message: "end date is required",
		})
	} else {
		if _, valid := validator.IsValidDate(r.EndDate); !valid {
			errs = append(errs, validator.ValidationError{
				Field:   "end_date",
				Message: "end date format is invalid (use YYYY-MM-DD)",
			})
		}
	}

	validDurationTypes := []string{"full_day", "half_day_morning", "half_day_afternoon"}
	if r.DurationType == "" {
		r.DurationType = "full_day"
	} else if !validator.IsInSlice(r.DurationType, validDurationTypes) {
		errs = append(errs, validator.ValidationError{
			Field:   "duration_type",
			Message: "duration type must be one of: full_day, half_day_morning, half_day_afternoon",
		})
	}

	if validator.IsEmpty(r.Reason) {
		errs = append(errs, validator.ValidationError{
			Field:   "reason",
			Message: "reason is required",
		})
	} else if len(r.Reason) < 10 {
		errs = append(errs, validator.ValidationError{
			Field:   "reason",
			Message: "reason must be at least 10 characters",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type UpdateLeaveRequestRequest struct {
	ID                 string     `json:"id"`
	EmployeeID         *string    `json:"employee_id,omitempty"`
	LeaveTypeID        *string    `json:"leave_type_id,omitempty"`
	StartDate          *string    `json:"start_date,omitempty"`
	EndDate            *string    `json:"end_date,omitempty"`
	DurationType       *string    `json:"duration_type,omitempty"`
	TotalDays          *float64   `json:"total_days,omitempty"`
	WorkingDays        *float64   `json:"working_days,omitempty"`
	Reason             *string    `json:"reason,omitempty"`
	AttachmentURL      *string    `json:"attachment_url,omitempty"`
	EmergencyLeave     *bool      `json:"emergency_leave,omitempty"`
	IsBackdate         *bool      `json:"is_backdate,omitempty"`
	Status             *string    `json:"status,omitempty"`
	ApprovedBy         *string    `json:"approved_by,omitempty"`
	ApprovedAt         *time.Time `json:"approved_at,omitempty"`
	RejectionReason    *string    `json:"rejection_reason,omitempty"`
	CancelledBy        *string    `json:"cancelled_by,omitempty"`
	CancelledAt        *time.Time `json:"cancelled_at,omitempty"`
	CancellationReason *string    `json:"cancellation_reason,omitempty"`
}

func (r *UpdateLeaveRequestRequest) Validate() error {
	var errs validator.ValidationErrors

	// ID
	if validator.IsEmpty(r.ID) {
		errs = append(errs, validator.ValidationError{
			Field:   "id",
			Message: "id is required",
		})
	}

	// EmployeeID
	if r.EmployeeID != nil && validator.IsEmpty(*r.EmployeeID) {
		errs = append(errs, validator.ValidationError{
			Field:   "employee_id",
			Message: "employee_id must not be empty",
		})
	}

	// LeaveTypeID
	if r.LeaveTypeID != nil && validator.IsEmpty(*r.LeaveTypeID) {
		errs = append(errs, validator.ValidationError{
			Field:   "leave_type_id",
			Message: "leave_type_id must not be empty",
		})
	}

	// StartDate and EndDate validation
	var start, end *validator.Date

	if r.StartDate != nil {
		if validator.IsEmpty(*r.StartDate) {
			errs = append(errs, validator.ValidationError{
				Field:   "start_date",
				Message: "start_date must not be empty",
			})
		} else {
			s, err := validator.ParseDate(*r.StartDate)
			if err != nil {
				errs = append(errs, validator.ValidationError{
					Field:   "start_date",
					Message: "start_date must be a valid date (YYYY-MM-DD)",
				})
			} else {
				start = &s
			}
		}
	}

	if r.EndDate != nil {
		if validator.IsEmpty(*r.EndDate) {
			errs = append(errs, validator.ValidationError{
				Field:   "end_date",
				Message: "end_date must not be empty",
			})
		} else {
			e, err := validator.ParseDate(*r.EndDate)
			if err != nil {
				errs = append(errs, validator.ValidationError{
					Field:   "end_date",
					Message: "end_date must be a valid date (YYYY-MM-DD)",
				})
			} else {
				end = &e
			}
		}
	}

	if start != nil && end != nil && end.Before(*start) {
		errs = append(errs, validator.ValidationError{
			Field:   "end_date",
			Message: "end_date must not be before start_date",
		})
	}

	// DurationType
	if r.DurationType != nil {
		validDurationTypes := []string{"full_day", "half_day_morning", "half_day_afternoon"}
		if !validator.IsInSlice(*r.DurationType, validDurationTypes) {
			errs = append(errs, validator.ValidationError{
				Field:   "duration_type",
				Message: "duration_type must be one of: full_day, half_day_morning, half_day_afternoon",
			})
		}
	}

	// TotalDays
	if r.TotalDays != nil && *r.TotalDays <= 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "total_days",
			Message: "total_days must be a positive number",
		})
	}

	// WorkingDays
	if r.WorkingDays != nil && *r.WorkingDays < 0 {
		errs = append(errs, validator.ValidationError{
			Field:   "working_days",
			Message: "working_days must not be negative",
		})
	}

	// Reason
	if r.Reason != nil && len(*r.Reason) < 10 {
		errs = append(errs, validator.ValidationError{
			Field:   "reason",
			Message: "reason must be at least 10 characters",
		})
	}

	// AttachmentURL
	if r.AttachmentURL != nil && len(*r.AttachmentURL) > 2048 {
		errs = append(errs, validator.ValidationError{
			Field:   "attachment_url",
			Message: "attachment_url must not exceed 2048 characters",
		})
	}

	// Status
	if r.Status != nil {
		validStatuses := []string{"waiting_approval", "approved", "rejected", "cancelled"}
		if !validator.IsInSlice(*r.Status, validStatuses) {
			errs = append(errs, validator.ValidationError{
				Field:   "status",
				Message: "status must be one of: waiting_approval, approved, rejected, cancelled",
			})
		}
	}

	// ApprovedBy
	if r.ApprovedBy != nil && !validator.IsEmpty(*r.ApprovedBy) {
		if !validator.IsValidUUID(*r.ApprovedBy) {
			errs = append(errs, validator.ValidationError{
				Field:   "approved_by",
				Message: "approved_by must be a valid UUID",
			})
		}
	}

	// CancelledBy
	if r.CancelledBy != nil && !validator.IsEmpty(*r.CancelledBy) {
		if !validator.IsValidUUID(*r.CancelledBy) {
			errs = append(errs, validator.ValidationError{
				Field:   "cancelled_by",
				Message: "cancelled_by must be a valid UUID",
			})
		}
	}

	// RejectionReason
	if r.RejectionReason != nil && len(*r.RejectionReason) < 10 {
		errs = append(errs, validator.ValidationError{
			Field:   "rejection_reason",
			Message: "rejection_reason must be at least 10 characters",
		})
	}

	// CancellationReason
	if r.CancellationReason != nil && len(*r.CancellationReason) < 10 {
		errs = append(errs, validator.ValidationError{
			Field:   "cancellation_reason",
			Message: "cancellation_reason must be at least 10 characters",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

// ========================================
// LEAVE QUOTA DTOs
// ========================================

type LeaveQuotaResponse struct {
	ID              string  `json:"id"`
	EmployeeID      string  `json:"employee_id"`
	LeaveTypeID     string  `json:"leave_type_id"`
	LeaveTypeName   string  `json:"leave_type_name"`
	Year            int     `json:"year"`
	OpeningBalance  int     `json:"opening_balance"`
	EarnedQuota     int     `json:"earned_quota"`
	RolloverQuota   int     `json:"rollover_quota"`
	AdjustmentQuota int     `json:"adjustment_quota"`
	UsedQuota       float64 `json:"used_quota"`
	PendingQuota    float64 `json:"pending_quota"`
	AvailableQuota  float64 `json:"available_quota"`
}

type AdjustQuotaRequest struct {
	EmployeeID  string `json:"employee_id"`
	LeaveTypeID string `json:"leave_type_id"`
	Year        int    `json:"year"`
	Adjustment  int    `json:"adjustment"`
	Reason      string `json:"reason"`
}

func (r *AdjustQuotaRequest) Validate() error {
	var errs validator.ValidationErrors

	if validator.IsEmpty(r.EmployeeID) {
		errs = append(errs, validator.ValidationError{
			Field:   "employee_id",
			Message: "employee ID is required",
		})
	}

	if validator.IsEmpty(r.LeaveTypeID) {
		errs = append(errs, validator.ValidationError{
			Field:   "leave_type_id",
			Message: "leave type ID is required",
		})
	}

	if r.Year == 0 {
		r.Year = time.Now().Year()
	}

	if validator.IsEmpty(r.Reason) {
		errs = append(errs, validator.ValidationError{
			Field:   "reason",
			Message: "reason is required",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type ApproveRequestRequest struct {
	RequestID string  `json:"request_id"`
	Reason    *string `json:"reason,omitempty"`
}

func (r *ApproveRequestRequest) Validate() error {
	var errs validator.ValidationErrors

	// RequestID
	if validator.IsEmpty(r.RequestID) {
		errs = append(errs, validator.ValidationError{
			Field:   "request_id",
			Message: "request_id is required",
		})
	}

	// Reason
	if r.Reason != nil && len(*r.Reason) < 10 {
		errs = append(errs, validator.ValidationError{
			Field:   "reason",
			Message: "reason must be at least 10 characters",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type RejectRequestRequest struct {
	RequestID string  `json:"request_id"`
	Reason    *string `json:"reason,omitempty"`
}

func (r *RejectRequestRequest) Validate() error {
	var errs validator.ValidationErrors

	// RequestID
	if validator.IsEmpty(r.RequestID) {
		errs = append(errs, validator.ValidationError{
			Field:   "request_id",
			Message: "request_id is required",
		})
	}

	// Reason
	if r.Reason != nil && len(*r.Reason) < 10 {
		errs = append(errs, validator.ValidationError{
			Field:   "reason",
			Message: "reason must be at least 10 characters",
		})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
