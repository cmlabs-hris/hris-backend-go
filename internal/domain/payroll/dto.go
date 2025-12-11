package payroll

import (
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/validator"
	"github.com/shopspring/decimal"
)

// ========== SETTINGS DTOs ==========

type PayrollSettingsResponse struct {
	ID                           string          `json:"id"`
	CompanyID                    string          `json:"company_id"`
	LateDeductionEnabled         bool            `json:"late_deduction_enabled"`
	LateDeductionPerMinute       decimal.Decimal `json:"late_deduction_per_minute"`
	OvertimeEnabled              bool            `json:"overtime_enabled"`
	OvertimePayPerMinute         decimal.Decimal `json:"overtime_pay_per_minute"`
	EarlyLeaveDeductionEnabled   bool            `json:"early_leave_deduction_enabled"`
	EarlyLeaveDeductionPerMinute decimal.Decimal `json:"early_leave_deduction_per_minute"`
}

type UpdatePayrollSettingsRequest struct {
	LateDeductionEnabled         *bool            `json:"late_deduction_enabled,omitempty"`
	LateDeductionPerMinute       *decimal.Decimal `json:"late_deduction_per_minute,omitempty"`
	OvertimeEnabled              *bool            `json:"overtime_enabled,omitempty"`
	OvertimePayPerMinute         *decimal.Decimal `json:"overtime_pay_per_minute,omitempty"`
	EarlyLeaveDeductionEnabled   *bool            `json:"early_leave_deduction_enabled,omitempty"`
	EarlyLeaveDeductionPerMinute *decimal.Decimal `json:"early_leave_deduction_per_minute,omitempty"`
}

func (r *UpdatePayrollSettingsRequest) Validate() error {
	var errs validator.ValidationErrors

	if r.LateDeductionPerMinute != nil && r.LateDeductionPerMinute.IsNegative() {
		errs = append(errs, validator.ValidationError{Field: "late_deduction_per_minute", Message: "must be non-negative"})
	}
	if r.OvertimePayPerMinute != nil && r.OvertimePayPerMinute.IsNegative() {
		errs = append(errs, validator.ValidationError{Field: "overtime_pay_per_minute", Message: "must be non-negative"})
	}
	if r.EarlyLeaveDeductionPerMinute != nil && r.EarlyLeaveDeductionPerMinute.IsNegative() {
		errs = append(errs, validator.ValidationError{Field: "early_leave_deduction_per_minute", Message: "must be non-negative"})
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// ========== COMPONENT DTOs ==========

type CreatePayrollComponentRequest struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"` // "allowance" or "deduction"
	Description *string `json:"description,omitempty"`
	IsTaxable   *bool   `json:"is_taxable,omitempty"`
}

func (r *CreatePayrollComponentRequest) Validate() error {
	var errs validator.ValidationErrors

	if r.Name == "" {
		errs = append(errs, validator.ValidationError{Field: "name", Message: "is required"})
	}
	if r.Type != "allowance" && r.Type != "deduction" {
		errs = append(errs, validator.ValidationError{Field: "type", Message: "must be 'allowance' or 'deduction'"})
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type UpdatePayrollComponentRequest struct {
	ID          string
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	IsTaxable   *bool   `json:"is_taxable,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

type PayrollComponentResponse struct {
	ID          string  `json:"id"`
	CompanyID   string  `json:"company_id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Description *string `json:"description,omitempty"`
	IsTaxable   bool    `json:"is_taxable"`
	IsActive    bool    `json:"is_active"`
}

// ========== EMPLOYEE COMPONENT DTOs ==========

type AssignComponentRequest struct {
	EmployeeID         string          `json:"-"`
	PayrollComponentID string          `json:"payroll_component_id"`
	Amount             decimal.Decimal `json:"amount"`
	EffectiveDate      *string         `json:"effective_date,omitempty"`
	EndDate            *string         `json:"end_date,omitempty"`
}

func (r *AssignComponentRequest) Validate() error {
	var errs validator.ValidationErrors

	if r.PayrollComponentID == "" {
		errs = append(errs, validator.ValidationError{Field: "payroll_component_id", Message: "is required"})
	}
	if r.Amount.IsNegative() {
		errs = append(errs, validator.ValidationError{Field: "amount", Message: "must be non-negative"})
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type UpdateEmployeeComponentRequest struct {
	ID            string
	Amount        *decimal.Decimal `json:"amount,omitempty"`
	EffectiveDate *string          `json:"effective_date,omitempty"`
	EndDate       *string          `json:"end_date,omitempty"`
}

type EmployeeComponentResponse struct {
	ID                 string          `json:"id"`
	EmployeeID         string          `json:"employee_id"`
	PayrollComponentID string          `json:"payroll_component_id"`
	ComponentName      string          `json:"component_name"`
	ComponentType      string          `json:"component_type"`
	Amount             decimal.Decimal `json:"amount"`
	EffectiveDate      string          `json:"effective_date"`
	EndDate            *string         `json:"end_date,omitempty"`
}

// ========== PAYROLL RECORD DTOs ==========

type GeneratePayrollRequest struct {
	PeriodMonth int      `json:"period_month"`
	PeriodYear  int      `json:"period_year"`
	EmployeeIDs []string `json:"employee_ids,omitempty"` // Empty = all active employees
}

func (r *GeneratePayrollRequest) Validate() error {
	var errs validator.ValidationErrors

	if r.PeriodMonth < 1 || r.PeriodMonth > 12 {
		errs = append(errs, validator.ValidationError{Field: "period_month", Message: "must be between 1 and 12"})
	}
	if r.PeriodYear < 2020 {
		errs = append(errs, validator.ValidationError{Field: "period_year", Message: "must be 2020 or later"})
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type UpdatePayrollRecordRequest struct {
	ID                  string
	BaseSalary          *decimal.Decimal `json:"base_salary,omitempty"`
	TotalAllowances     *decimal.Decimal `json:"total_allowances,omitempty"`
	TotalDeductions     *decimal.Decimal `json:"total_deductions,omitempty"`
	LateDeductionAmount *decimal.Decimal `json:"late_deduction_amount,omitempty"`
	OvertimeAmount      *decimal.Decimal `json:"overtime_amount,omitempty"`
	Notes               *string          `json:"notes,omitempty"`
}

type FinalizePayrollRequest struct {
	RecordIDs []string `json:"record_ids"`
}

func (r *FinalizePayrollRequest) Validate() error {
	var errs validator.ValidationErrors

	if len(r.RecordIDs) == 0 {
		errs = append(errs, validator.ValidationError{Field: "record_ids", Message: "at least one record is required"})
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type PayrollRecordResponse struct {
	ID                        string                     `json:"id"`
	EmployeeID                string                     `json:"employee_id"`
	EmployeeName              string                     `json:"employee_name"`
	EmployeeCode              string                     `json:"employee_code"`
	PositionName              *string                    `json:"position_name,omitempty"`
	BranchName                *string                    `json:"branch_name,omitempty"`
	PeriodMonth               int                        `json:"period_month"`
	PeriodYear                int                        `json:"period_year"`
	BaseSalary                decimal.Decimal            `json:"base_salary"`
	TotalAllowances           decimal.Decimal            `json:"total_allowances"`
	TotalDeductions           decimal.Decimal            `json:"total_deductions"`
	AllowancesDetail          map[string]decimal.Decimal `json:"allowances_detail,omitempty"`
	DeductionsDetail          map[string]decimal.Decimal `json:"deductions_detail,omitempty"`
	TotalWorkDays             int                        `json:"total_work_days"`
	TotalLateMinutes          int                        `json:"total_late_minutes"`
	LateDeductionAmount       decimal.Decimal            `json:"late_deduction_amount"`
	TotalEarlyLeaveMinutes    int                        `json:"total_early_leave_minutes"`
	EarlyLeaveDeductionAmount decimal.Decimal            `json:"early_leave_deduction_amount"`
	TotalOvertimeMinutes      int                        `json:"total_overtime_minutes"`
	OvertimeAmount            decimal.Decimal            `json:"overtime_amount"`
	GrossSalary               decimal.Decimal            `json:"gross_salary"`
	NetSalary                 decimal.Decimal            `json:"net_salary"`
	Status                    string                     `json:"status"`
	PaidAt                    *string                    `json:"paid_at,omitempty"`
	Notes                     *string                    `json:"notes,omitempty"`
}

type PayrollFilter struct {
	PeriodMonth *int    `json:"period_month,omitempty"`
	PeriodYear  *int    `json:"period_year,omitempty"`
	Status      *string `json:"status,omitempty"`
	EmployeeID  *string `json:"employee_id,omitempty"`
	Page        int     `json:"page"`
	Limit       int     `json:"limit"`
	SortBy      string  `json:"sort_by"`
	SortOrder   string  `json:"sort_order"`
}

type ListPayrollRecordResponse struct {
	Data       []PayrollRecordResponse `json:"data"`
	TotalCount int64                   `json:"total_count"`
	Page       int                     `json:"page"`
	Limit      int                     `json:"limit"`
}

type PayrollSummaryResponse struct {
	PeriodMonth        int             `json:"period_month"`
	PeriodYear         int             `json:"period_year"`
	TotalEmployees     int             `json:"total_employees"`
	TotalBaseSalary    decimal.Decimal `json:"total_base_salary"`
	TotalAllowances    decimal.Decimal `json:"total_allowances"`
	TotalDeductions    decimal.Decimal `json:"total_deductions"`
	TotalLateDeduction decimal.Decimal `json:"total_late_deduction"`
	TotalOvertime      decimal.Decimal `json:"total_overtime"`
	TotalGrossSalary   decimal.Decimal `json:"total_gross_salary"`
	TotalNetSalary     decimal.Decimal `json:"total_net_salary"`
	DraftCount         int             `json:"draft_count"`
	PaidCount          int             `json:"paid_count"`
}
