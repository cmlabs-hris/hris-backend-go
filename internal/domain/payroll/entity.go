package payroll

import (
	"time"

	"github.com/shopspring/decimal"
)

// PayrollSettings - Company payroll configuration
type PayrollSettings struct {
	ID                           string
	CompanyID                    string
	LateDeductionEnabled         bool
	LateDeductionPerMinute       decimal.Decimal
	OvertimeEnabled              bool
	OvertimePayPerMinute         decimal.Decimal
	EarlyLeaveDeductionEnabled   bool
	EarlyLeaveDeductionPerMinute decimal.Decimal
	CreatedAt                    time.Time
	UpdatedAt                    time.Time
}

// ComponentType enum
type ComponentType string

const (
	ComponentTypeAllowance ComponentType = "allowance"
	ComponentTypeDeduction ComponentType = "deduction"
)

// PayrollComponent - Master payroll component
type PayrollComponent struct {
	ID          string
	CompanyID   string
	Name        string
	Type        ComponentType
	Description *string
	IsTaxable   bool
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// EmployeePayrollComponent - Component assignment to employee
type EmployeePayrollComponent struct {
	ID                 string
	EmployeeID         string
	PayrollComponentID string
	Amount             decimal.Decimal
	EffectiveDate      time.Time
	EndDate            *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time

	// Joined fields
	ComponentName *string
	ComponentType *ComponentType
}

// PayrollStatus enum
type PayrollStatus string

const (
	PayrollStatusDraft PayrollStatus = "draft"
	PayrollStatusPaid  PayrollStatus = "paid"
)

// PayrollRecord - Generated payroll result
type PayrollRecord struct {
	ID                        string
	EmployeeID                string
	CompanyID                 string
	PeriodMonth               int
	PeriodYear                int
	BaseSalary                decimal.Decimal
	TotalAllowances           decimal.Decimal
	TotalDeductions           decimal.Decimal
	AllowancesDetail          map[string]decimal.Decimal // {"Tunjangan Transport": 500000}
	DeductionsDetail          map[string]decimal.Decimal // {"BPJS": 100000}
	TotalWorkDays             int
	TotalLateMinutes          int
	LateDeductionAmount       decimal.Decimal
	TotalEarlyLeaveMinutes    int
	EarlyLeaveDeductionAmount decimal.Decimal
	TotalOvertimeMinutes      int
	OvertimeAmount            decimal.Decimal
	GrossSalary               decimal.Decimal
	NetSalary                 decimal.Decimal
	Status                    PayrollStatus
	PaidAt                    *time.Time
	PaidBy                    *string
	Notes                     *string
	CreatedAt                 time.Time
	UpdatedAt                 time.Time

	// Joined fields
	EmployeeName *string
	EmployeeCode *string
	PositionName *string
	BranchName   *string
}

// AttendanceSummary - Aggregate from attendances table
type AttendanceSummary struct {
	EmployeeID             string
	TotalWorkDays          int
	TotalLateMinutes       int
	TotalEarlyLeaveMinutes int
	TotalOvertimeMinutes   int
}
