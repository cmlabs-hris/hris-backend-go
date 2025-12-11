# Full Planning: Implementasi Fitur Payroll MVP

## Overview

Fitur payroll akan menggunakan data attendance yang sudah tersedia (`late_minutes`, `overtime_minutes`, `early_leave_minutes`) untuk menghitung gaji dengan pendekatan **minute-based rate**.

---

## Dependencies

Tambahkan dependency untuk monetary precision:

```bash
go get github.com/shopspring/decimal
```

**Catatan Penting:**
- Semua field monetary (gaji, potongan, tunjangan) menggunakan `decimal.Decimal` dari `github.com/shopspring/decimal`
- Database menggunakan `DECIMAL(15,2)` untuk presisi yang sama
- Hindari `float64` untuk kalkulasi uang karena masalah floating point precision

---

## STEP 1: Database Migration

### File: `internal/infrastructure/database/postgresql/migrations/000004_payroll.up.sql`

```sql
-- =========================
-- Payroll Feature Schema
-- =========================

-- 1. Add base_salary to employees table
ALTER TABLE employees ADD COLUMN base_salary DECIMAL(15,2);

-- 2. Enum for payroll component type
CREATE TYPE payroll_component_type AS ENUM ('allowance', 'deduction');

-- 3. Enum for payroll status
CREATE TYPE payroll_status AS ENUM ('draft', 'paid');

-- 4. Table: payroll_settings
-- Company-wide payroll configuration
CREATE TABLE payroll_settings (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE UNIQUE,
    
    -- Late Deduction Settings
    late_deduction_enabled BOOLEAN NOT NULL DEFAULT true,
    late_deduction_per_minute DECIMAL(10,2) NOT NULL DEFAULT 0,
    
    -- Overtime Settings  
    overtime_enabled BOOLEAN NOT NULL DEFAULT true,
    overtime_pay_per_minute DECIMAL(10,2) NOT NULL DEFAULT 0,
    
    -- Early Leave Settings
    early_leave_deduction_enabled BOOLEAN NOT NULL DEFAULT false,
    early_leave_deduction_per_minute DECIMAL(10,2) NOT NULL DEFAULT 0,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_rates_non_negative CHECK (
        late_deduction_per_minute >= 0 AND 
        overtime_pay_per_minute >= 0 AND
        early_leave_deduction_per_minute >= 0
    ),
    CONSTRAINT chk_updated_at_not_before_created_at CHECK (updated_at >= created_at)
);

-- 5. Table: payroll_components
-- Master table for salary components (allowances/deductions)
CREATE TABLE payroll_components (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    
    name VARCHAR(100) NOT NULL,
    type payroll_component_type NOT NULL,
    description TEXT,
    is_taxable BOOLEAN NOT NULL DEFAULT true,
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(company_id, name),
    CONSTRAINT chk_updated_at_not_before_created_at CHECK (updated_at >= created_at)
);

-- 6. Table: employee_payroll_components
-- Assignment of components to employees
CREATE TABLE employee_payroll_components (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    payroll_component_id UUID NOT NULL REFERENCES payroll_components(id) ON DELETE CASCADE,
    
    amount DECIMAL(15,2) NOT NULL,
    
    effective_date DATE NOT NULL DEFAULT CURRENT_DATE,
    end_date DATE,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_amount_positive CHECK (amount >= 0),
    CONSTRAINT chk_end_date_after_effective CHECK (end_date IS NULL OR end_date >= effective_date),
    CONSTRAINT chk_updated_at_not_before_created_at CHECK (updated_at >= created_at)
);

-- 7. Table: payroll_records
-- Generated payroll per employee per period
CREATE TABLE payroll_records (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    
    -- Period
    period_month SMALLINT NOT NULL CHECK (period_month BETWEEN 1 AND 12),
    period_year SMALLINT NOT NULL CHECK (period_year >= 2020),
    
    -- Base Components (snapshot at generation time)
    base_salary DECIMAL(15,2) NOT NULL,
    
    -- Allowances & Deductions from Components
    total_allowances DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_deductions DECIMAL(15,2) NOT NULL DEFAULT 0,
    allowances_detail JSONB,
    deductions_detail JSONB,
    
    -- Attendance-based Calculations
    total_work_days INT NOT NULL DEFAULT 0,
    total_late_minutes INT NOT NULL DEFAULT 0,
    late_deduction_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_early_leave_minutes INT NOT NULL DEFAULT 0,
    early_leave_deduction_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_overtime_minutes INT NOT NULL DEFAULT 0,
    overtime_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    
    -- Final Calculation
    gross_salary DECIMAL(15,2) NOT NULL,
    net_salary DECIMAL(15,2) NOT NULL,
    
    -- Status & Metadata
    status payroll_status NOT NULL DEFAULT 'draft',
    paid_at TIMESTAMPTZ,
    paid_by UUID REFERENCES users(id),
    notes TEXT,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(employee_id, period_month, period_year),
    CONSTRAINT chk_salaries_non_negative CHECK (
        base_salary >= 0 AND 
        gross_salary >= 0
    ),
    CONSTRAINT chk_updated_at_not_before_created_at CHECK (updated_at >= created_at)
);

-- =========================
-- Indexes for Payroll Tables
-- =========================

-- payroll_settings
CREATE INDEX idx_payroll_settings_company ON payroll_settings(company_id);

-- payroll_components
CREATE INDEX idx_payroll_components_company ON payroll_components(company_id);
CREATE INDEX idx_payroll_components_type ON payroll_components(company_id, type);
CREATE INDEX idx_payroll_components_active ON payroll_components(company_id, is_active) WHERE is_active = true;

-- employee_payroll_components
CREATE INDEX idx_emp_payroll_comp_employee ON employee_payroll_components(employee_id);
CREATE INDEX idx_emp_payroll_comp_component ON employee_payroll_components(payroll_component_id);
CREATE INDEX idx_emp_payroll_comp_active ON employee_payroll_components(employee_id, effective_date, end_date);

-- payroll_records
CREATE INDEX idx_payroll_records_company ON payroll_records(company_id);
CREATE INDEX idx_payroll_records_employee ON payroll_records(employee_id);
CREATE INDEX idx_payroll_records_period ON payroll_records(company_id, period_year DESC, period_month DESC);
CREATE INDEX idx_payroll_records_status ON payroll_records(company_id, status);
CREATE INDEX idx_payroll_records_employee_period ON payroll_records(employee_id, period_year DESC, period_month DESC);
```

### File: `internal/infrastructure/database/postgresql/migrations/000004_payroll.down.sql`

```sql
-- Rollback payroll schema
DROP INDEX IF EXISTS idx_payroll_records_employee_period;
DROP INDEX IF EXISTS idx_payroll_records_status;
DROP INDEX IF EXISTS idx_payroll_records_period;
DROP INDEX IF EXISTS idx_payroll_records_employee;
DROP INDEX IF EXISTS idx_payroll_records_company;
DROP INDEX IF EXISTS idx_emp_payroll_comp_active;
DROP INDEX IF EXISTS idx_emp_payroll_comp_component;
DROP INDEX IF EXISTS idx_emp_payroll_comp_employee;
DROP INDEX IF EXISTS idx_payroll_components_active;
DROP INDEX IF EXISTS idx_payroll_components_type;
DROP INDEX IF EXISTS idx_payroll_components_company;
DROP INDEX IF EXISTS idx_payroll_settings_company;

DROP TABLE IF EXISTS payroll_records;
DROP TABLE IF EXISTS employee_payroll_components;
DROP TABLE IF EXISTS payroll_components;
DROP TABLE IF EXISTS payroll_settings;

DROP TYPE IF EXISTS payroll_status;
DROP TYPE IF EXISTS payroll_component_type;

ALTER TABLE employees DROP COLUMN IF EXISTS base_salary;
```

---

## STEP 2: Domain Layer

### File: `internal/domain/payroll/entity.go`

```go
package payroll

import (
    "time"
    
    "github.com/shopspring/decimal"
)

// PayrollSettings - Company payroll configuration
type PayrollSettings struct {
    ID                            string
    CompanyID                     string
    LateDeductionEnabled          bool
    LateDeductionPerMinute        decimal.Decimal
    OvertimeEnabled               bool
    OvertimePayPerMinute          decimal.Decimal
    EarlyLeaveDeductionEnabled    bool
    EarlyLeaveDeductionPerMinute  decimal.Decimal
    CreatedAt                     time.Time
    UpdatedAt                     time.Time
}

// ComponentType enum
type ComponentType string

const (
    ComponentTypeAllowance  ComponentType = "allowance"
    ComponentTypeDeduction  ComponentType = "deduction"
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
    ID                       string
    EmployeeID               string
    CompanyID                string
    PeriodMonth              int
    PeriodYear               int
    BaseSalary               decimal.Decimal
    TotalAllowances          decimal.Decimal
    TotalDeductions          decimal.Decimal
    AllowancesDetail         map[string]decimal.Decimal // {"Tunjangan Transport": 500000}
    DeductionsDetail         map[string]decimal.Decimal // {"BPJS": 100000}
    TotalWorkDays            int
    TotalLateMinutes         int
    LateDeductionAmount      decimal.Decimal
    TotalEarlyLeaveMinutes   int
    EarlyLeaveDeductionAmount decimal.Decimal
    TotalOvertimeMinutes     int
    OvertimeAmount           decimal.Decimal
    GrossSalary              decimal.Decimal
    NetSalary                decimal.Decimal
    Status                   PayrollStatus
    PaidAt                   *time.Time
    PaidBy                   *string
    Notes                    *string
    CreatedAt                time.Time
    UpdatedAt                time.Time
    
    // Joined fields
    EmployeeName     *string
    EmployeeCode     *string
    PositionName     *string
    BranchName       *string
}

// AttendanceSummary - Aggregate from attendances table
type AttendanceSummary struct {
    EmployeeID           string
    TotalWorkDays        int
    TotalLateMinutes     int
    TotalEarlyLeaveMinutes int
    TotalOvertimeMinutes int
}
```

### File: `internal/domain/payroll/dto.go`

```go
package payroll

import (
    "github.com/cmlabs-hris/hris-backend-go/internal/pkg/validator"
    "github.com/shopspring/decimal"
)

// ========== SETTINGS DTOs ==========

type PayrollSettingsResponse struct {
    ID                            string          `json:"id"`
    CompanyID                     string          `json:"company_id"`
    LateDeductionEnabled          bool            `json:"late_deduction_enabled"`
    LateDeductionPerMinute        decimal.Decimal `json:"late_deduction_per_minute"`
    OvertimeEnabled               bool            `json:"overtime_enabled"`
    OvertimePayPerMinute          decimal.Decimal `json:"overtime_pay_per_minute"`
    EarlyLeaveDeductionEnabled    bool            `json:"early_leave_deduction_enabled"`
    EarlyLeaveDeductionPerMinute  decimal.Decimal `json:"early_leave_deduction_per_minute"`
}

type UpdatePayrollSettingsRequest struct {
    LateDeductionEnabled          *bool            `json:"late_deduction_enabled,omitempty"`
    LateDeductionPerMinute        *decimal.Decimal `json:"late_deduction_per_minute,omitempty"`
    OvertimeEnabled               *bool            `json:"overtime_enabled,omitempty"`
    OvertimePayPerMinute          *decimal.Decimal `json:"overtime_pay_per_minute,omitempty"`
    EarlyLeaveDeductionEnabled    *bool            `json:"early_leave_deduction_enabled,omitempty"`
    EarlyLeaveDeductionPerMinute  *decimal.Decimal `json:"early_leave_deduction_per_minute,omitempty"`
}

func (r *UpdatePayrollSettingsRequest) Validate() error {
    if r.LateDeductionPerMinute != nil && r.LateDeductionPerMinute.IsNegative() {
        return validator.NewValidationError("late_deduction_per_minute", "must be non-negative")
    }
    if r.OvertimePayPerMinute != nil && r.OvertimePayPerMinute.IsNegative() {
        return validator.NewValidationError("overtime_pay_per_minute", "must be non-negative")
    }
    if r.EarlyLeaveDeductionPerMinute != nil && r.EarlyLeaveDeductionPerMinute.IsNegative() {
        return validator.NewValidationError("early_leave_deduction_per_minute", "must be non-negative")
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
    if r.Name == "" {
        return validator.NewValidationError("name", "is required")
    }
    if r.Type != "allowance" && r.Type != "deduction" {
        return validator.NewValidationError("type", "must be 'allowance' or 'deduction'")
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
    EmployeeID         string           `json:"-"`
    PayrollComponentID string           `json:"payroll_component_id"`
    Amount             decimal.Decimal  `json:"amount"`
    EffectiveDate      *string          `json:"effective_date,omitempty"`
    EndDate            *string          `json:"end_date,omitempty"`
}

func (r *AssignComponentRequest) Validate() error {
    if r.PayrollComponentID == "" {
        return validator.NewValidationError("payroll_component_id", "is required")
    }
    if r.Amount.IsNegative() {
        return validator.NewValidationError("amount", "must be non-negative")
    }
    return nil
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
    if r.PeriodMonth < 1 || r.PeriodMonth > 12 {
        return validator.NewValidationError("period_month", "must be between 1 and 12")
    }
    if r.PeriodYear < 2020 {
        return validator.NewValidationError("period_year", "must be 2020 or later")
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
    if len(r.RecordIDs) == 0 {
        return validator.NewValidationError("record_ids", "at least one record is required")
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
    PeriodMonth       int             `json:"period_month"`
    PeriodYear        int             `json:"period_year"`
    TotalEmployees    int             `json:"total_employees"`
    TotalBaseSalary   decimal.Decimal `json:"total_base_salary"`
    TotalAllowances   decimal.Decimal `json:"total_allowances"`
    TotalDeductions   decimal.Decimal `json:"total_deductions"`
    TotalLateDeduction decimal.Decimal `json:"total_late_deduction"`
    TotalOvertime     decimal.Decimal `json:"total_overtime"`
    TotalGrossSalary  decimal.Decimal `json:"total_gross_salary"`
    TotalNetSalary    decimal.Decimal `json:"total_net_salary"`
    DraftCount        int             `json:"draft_count"`
    PaidCount         int             `json:"paid_count"`
}
```

### File: `internal/domain/payroll/repository.go`

```go
package payroll

import "context"

type PayrollRepository interface {
    // Settings
    GetSettings(ctx context.Context, companyID string) (PayrollSettings, error)
    UpsertSettings(ctx context.Context, settings PayrollSettings) (PayrollSettings, error)
    
    // Components
    CreateComponent(ctx context.Context, component PayrollComponent) (PayrollComponent, error)
    GetComponentByID(ctx context.Context, id string, companyID string) (PayrollComponent, error)
    GetComponentsByCompanyID(ctx context.Context, companyID string, activeOnly bool) ([]PayrollComponent, error)
    UpdateComponent(ctx context.Context, req UpdatePayrollComponentRequest) error
    DeleteComponent(ctx context.Context, id string, companyID string) error
    
    // Employee Components
    AssignComponentToEmployee(ctx context.Context, assignment EmployeePayrollComponent) (EmployeePayrollComponent, error)
    GetEmployeeComponents(ctx context.Context, employeeID string, activeOnly bool) ([]EmployeePayrollComponent, error)
    RemoveEmployeeComponent(ctx context.Context, id string) error
    
    // Payroll Records
    CreatePayrollRecord(ctx context.Context, record PayrollRecord) (PayrollRecord, error)
    GetPayrollRecordByID(ctx context.Context, id string, companyID string) (PayrollRecord, error)
    GetPayrollRecordByEmployeePeriod(ctx context.Context, employeeID string, month, year int) (PayrollRecord, error)
    ListPayrollRecords(ctx context.Context, companyID string, filter PayrollFilter) ([]PayrollRecord, int64, error)
    UpdatePayrollRecord(ctx context.Context, req UpdatePayrollRecordRequest) error
    FinalizePayrollRecords(ctx context.Context, ids []string, paidBy string) error
    DeletePayrollRecord(ctx context.Context, id string, companyID string) error
    
    // Aggregations
    GetAttendanceSummary(ctx context.Context, companyID string, month, year int, employeeIDs []string) ([]AttendanceSummary, error)
    GetPayrollSummary(ctx context.Context, companyID string, month, year int) (PayrollSummaryResponse, error)
}
```

### File: `internal/domain/payroll/service.go`

```go
package payroll

import "context"

type PayrollService interface {
    // Settings
    GetSettings(ctx context.Context) (PayrollSettingsResponse, error)
    UpdateSettings(ctx context.Context, req UpdatePayrollSettingsRequest) (PayrollSettingsResponse, error)
    
    // Components
    CreateComponent(ctx context.Context, req CreatePayrollComponentRequest) (PayrollComponentResponse, error)
    GetComponent(ctx context.Context, id string) (PayrollComponentResponse, error)
    ListComponents(ctx context.Context, activeOnly bool) ([]PayrollComponentResponse, error)
    UpdateComponent(ctx context.Context, req UpdatePayrollComponentRequest) error
    DeleteComponent(ctx context.Context, id string) error
    
    // Employee Components
    AssignComponentToEmployee(ctx context.Context, req AssignComponentRequest) (EmployeeComponentResponse, error)
    GetEmployeeComponents(ctx context.Context, employeeID string) ([]EmployeeComponentResponse, error)
    RemoveEmployeeComponent(ctx context.Context, id string) error
    
    // Payroll Generation & Management
    GeneratePayroll(ctx context.Context, req GeneratePayrollRequest) ([]PayrollRecordResponse, error)
    GetPayrollRecord(ctx context.Context, id string) (PayrollRecordResponse, error)
    ListPayrollRecords(ctx context.Context, filter PayrollFilter) (ListPayrollRecordResponse, error)
    UpdatePayrollRecord(ctx context.Context, req UpdatePayrollRecordRequest) (PayrollRecordResponse, error)
    FinalizePayroll(ctx context.Context, req FinalizePayrollRequest) error
    DeletePayrollRecord(ctx context.Context, id string) error
    
    // Summary
    GetPayrollSummary(ctx context.Context, month, year int) (PayrollSummaryResponse, error)
}
```

### File: `internal/domain/payroll/errors.go`

```go
package payroll

import "errors"

var (
    ErrPayrollSettingsNotFound    = errors.New("payroll settings not found")
    ErrPayrollComponentNotFound   = errors.New("payroll component not found")
    ErrPayrollComponentNameExists = errors.New("payroll component name already exists")
    ErrPayrollRecordNotFound      = errors.New("payroll record not found")
    ErrPayrollRecordAlreadyExists = errors.New("payroll record already exists for this period")
    ErrPayrollRecordAlreadyPaid   = errors.New("payroll record already paid, cannot modify")
    ErrInvalidPeriod              = errors.New("invalid payroll period")
    ErrEmployeeHasNoBaseSalary    = errors.New("employee has no base salary configured")
    ErrCannotDeletePaidRecord     = errors.New("cannot delete paid payroll record")
    ErrEmployeeComponentNotFound  = errors.New("employee component assignment not found")
)
```

---

## STEP 3: Repository Layer

### File: `internal/repository/postgresql/payroll.go`

Implementasi repository dengan methods:

1. **Settings**: Get/Upsert payroll settings per company
2. **Components**: CRUD master components
3. **Employee Components**: Assign/Get/Remove components
4. **Payroll Records**: Create/Get/List/Update/Finalize/Delete
5. **GetAttendanceSummary**: Aggregate query dari attendances

**Key Query - Attendance Summary**:
```sql
-- Note: Status can be 'on_time', 'late', or leave type names (dynamic).
-- Exclude: 'rejected', 'approved', 'absent', 'waiting_approval'
SELECT 
    employee_id,
    COUNT(*) as total_work_days,
    COALESCE(SUM(late_minutes), 0) as total_late_minutes,
    COALESCE(SUM(early_leave_minutes), 0) as total_early_leave_minutes,
    COALESCE(SUM(overtime_minutes), 0) as total_overtime_minutes
FROM attendances
WHERE company_id = $1 
    AND EXTRACT(MONTH FROM date) = $2
    AND EXTRACT(YEAR FROM date) = $3
    AND status NOT IN ('rejected', 'approved', 'absent', 'waiting_approval')
    AND ($4::uuid[] IS NULL OR employee_id = ANY($4))
GROUP BY employee_id
```

---

## STEP 4: Service Layer

### File: `internal/service/payroll/service.go`

**Core Logic - GeneratePayroll**:

```go
func (s *PayrollServiceImpl) GeneratePayroll(ctx context.Context, req GeneratePayrollRequest) ([]PayrollRecordResponse, error) {
    // 1. Validate request
    if err := req.Validate(); err != nil {
        return nil, err
    }
    
    // 2. Get company ID from JWT
    companyID := getCompanyIDFromContext(ctx)
    
    // 3. Get payroll settings
    settings, err := s.payrollRepo.GetSettings(ctx, companyID)
    if err != nil {
        // Use default settings if not found
    }
    
    // 4. Get employees (all active or selected)
    var employees []employee.Employee
    if len(req.EmployeeIDs) > 0 {
        employees = s.employeeRepo.GetByIDs(ctx, req.EmployeeIDs, companyID)
    } else {
        employees = s.employeeRepo.GetActiveByCompanyID(ctx, companyID)
    }
    
    // 5. Get attendance summary
    attendanceSummaries, _ := s.payrollRepo.GetAttendanceSummary(ctx, companyID, req.PeriodMonth, req.PeriodYear, nil)
    attendanceMap := make(map[string]AttendanceSummary)
    for _, a := range attendanceSummaries {
        attendanceMap[a.EmployeeID] = a
    }
    
    // 6. Generate payroll for each employee
    var records []PayrollRecord
    for _, emp := range employees {
        if emp.BaseSalary == nil || emp.BaseSalary.IsZero() {
            continue // Skip employees without base salary
        }
        
        // Get employee components
        components, _ := s.payrollRepo.GetEmployeeComponents(ctx, emp.ID, true)
        
        totalAllowances := decimal.Zero
        totalDeductions := decimal.Zero
        allowancesDetail := make(map[string]decimal.Decimal)
        deductionsDetail := make(map[string]decimal.Decimal)
        
        for _, comp := range components {
            if comp.ComponentType != nil {
                if *comp.ComponentType == ComponentTypeAllowance {
                    totalAllowances = totalAllowances.Add(comp.Amount)
                    allowancesDetail[*comp.ComponentName] = comp.Amount
                } else {
                    totalDeductions = totalDeductions.Add(comp.Amount)
                    deductionsDetail[*comp.ComponentName] = comp.Amount
                }
            }
        }
        
        // Get attendance data
        att := attendanceMap[emp.ID]
        
        // Calculate late/overtime deductions using decimal
        lateDeduction := decimal.Zero
        earlyLeaveDeduction := decimal.Zero
        overtimeAmount := decimal.Zero
        
        if settings.LateDeductionEnabled {
            lateDeduction = decimal.NewFromInt(int64(att.TotalLateMinutes)).Mul(settings.LateDeductionPerMinute)
        }
        if settings.EarlyLeaveDeductionEnabled {
            earlyLeaveDeduction = decimal.NewFromInt(int64(att.TotalEarlyLeaveMinutes)).Mul(settings.EarlyLeaveDeductionPerMinute)
        }
        if settings.OvertimeEnabled {
            overtimeAmount = decimal.NewFromInt(int64(att.TotalOvertimeMinutes)).Mul(settings.OvertimePayPerMinute)
        }
        
        // Calculate final salary using decimal arithmetic
        grossSalary := emp.BaseSalary.Add(totalAllowances).Add(overtimeAmount)
        netSalary := grossSalary.Sub(totalDeductions).Sub(lateDeduction).Sub(earlyLeaveDeduction)
        
        record := PayrollRecord{
            EmployeeID:               emp.ID,
            CompanyID:                companyID,
            PeriodMonth:              req.PeriodMonth,
            PeriodYear:               req.PeriodYear,
            BaseSalary:               *emp.BaseSalary,
            TotalAllowances:          totalAllowances,
            TotalDeductions:          totalDeductions,
            AllowancesDetail:         allowancesDetail,
            DeductionsDetail:         deductionsDetail,
            TotalWorkDays:            att.TotalWorkDays,
            TotalLateMinutes:         att.TotalLateMinutes,
            LateDeductionAmount:      lateDeduction,
            TotalEarlyLeaveMinutes:   att.TotalEarlyLeaveMinutes,
            EarlyLeaveDeductionAmount: earlyLeaveDeduction,
            TotalOvertimeMinutes:     att.TotalOvertimeMinutes,
            OvertimeAmount:           overtimeAmount,
            GrossSalary:              grossSalary,
            NetSalary:                netSalary,
            Status:                   PayrollStatusDraft,
        }
        
        // Insert or update record
        created, _ := s.payrollRepo.CreatePayrollRecord(ctx, record)
        records = append(records, created)
    }
    
    return mapToRecordResponses(records), nil
}
```

---

## STEP 5: Handler Layer

### File: `internal/handler/http/payroll.go`

```go
package http

import (
    "encoding/json"
    "net/http"
    "strconv"
    
    "github.com/cmlabs-hris/hris-backend-go/internal/domain/payroll"
    "github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
    "github.com/go-chi/chi/v5"
)

type PayrollHandler interface {
    // Settings
    GetSettings(w http.ResponseWriter, r *http.Request)
    UpdateSettings(w http.ResponseWriter, r *http.Request)
    
    // Components
    CreateComponent(w http.ResponseWriter, r *http.Request)
    GetComponent(w http.ResponseWriter, r *http.Request)
    ListComponents(w http.ResponseWriter, r *http.Request)
    UpdateComponent(w http.ResponseWriter, r *http.Request)
    DeleteComponent(w http.ResponseWriter, r *http.Request)
    
    // Employee Components
    AssignComponent(w http.ResponseWriter, r *http.Request)
    GetEmployeeComponents(w http.ResponseWriter, r *http.Request)
    RemoveEmployeeComponent(w http.ResponseWriter, r *http.Request)
    
    // Payroll Records
    GeneratePayroll(w http.ResponseWriter, r *http.Request)
    GetPayrollRecord(w http.ResponseWriter, r *http.Request)
    ListPayrollRecords(w http.ResponseWriter, r *http.Request)
    UpdatePayrollRecord(w http.ResponseWriter, r *http.Request)
    FinalizePayroll(w http.ResponseWriter, r *http.Request)
    DeletePayrollRecord(w http.ResponseWriter, r *http.Request)
    
    // Summary
    GetPayrollSummary(w http.ResponseWriter, r *http.Request)
}

type payrollHandlerImpl struct {
    payrollService payroll.PayrollService
}

func NewPayrollHandler(payrollService payroll.PayrollService) PayrollHandler {
    return &payrollHandlerImpl{payrollService: payrollService}
}

// Implementation of each handler method...
```

---

## STEP 6: Router & Main Wiring

### Update: `internal/handler/http/router.go`

Tambah parameter `payrollHandler PayrollHandler` dan routes:

```go
// Add payrollHandler parameter
func NewRouter(..., payrollHandler PayrollHandler, ...) *chi.Mux {
    // ...existing code...
    
    // Payroll routes (owner/manager only)
    r.Route("/payroll", func(r chi.Router) {
        r.Use(jwtauth.Verifier(tokenAuth), jwtauth.Authenticator(tokenAuth))
        r.Use(middleware.RequireCompany)
        r.Use(middleware.RequireRole("owner", "manager"))
        
        // Settings
        r.Get("/settings", payrollHandler.GetSettings)
        r.Put("/settings", payrollHandler.UpdateSettings)
        
        // Components
        r.Post("/components", payrollHandler.CreateComponent)
        r.Get("/components", payrollHandler.ListComponents)
        r.Get("/components/{id}", payrollHandler.GetComponent)
        r.Put("/components/{id}", payrollHandler.UpdateComponent)
        r.Delete("/components/{id}", payrollHandler.DeleteComponent)
        
        // Employee Components
        r.Post("/employees/{employeeId}/components", payrollHandler.AssignComponent)
        r.Get("/employees/{employeeId}/components", payrollHandler.GetEmployeeComponents)
        r.Delete("/employee-components/{id}", payrollHandler.RemoveEmployeeComponent)
        
        // Payroll Records
        r.Post("/generate", payrollHandler.GeneratePayroll)
        r.Get("/records", payrollHandler.ListPayrollRecords)
        r.Get("/records/{id}", payrollHandler.GetPayrollRecord)
        r.Put("/records/{id}", payrollHandler.UpdatePayrollRecord)
        r.Delete("/records/{id}", payrollHandler.DeletePayrollRecord)
        r.Post("/finalize", payrollHandler.FinalizePayroll)
        
        // Summary
        r.Get("/summary", payrollHandler.GetPayrollSummary)
    })
}
```

### Update: `cmd/api/main.go`

```go
// Add import
import payrollService "github.com/cmlabs-hris/hris-backend-go/internal/service/payroll"

// Add repository
payrollRepo := postgresql.NewPayrollRepository(db)

// Add service
payrollService := payrollService.NewPayrollService(db, payrollRepo, employeeRepo)

// Add handler
payrollHandler := appHTTP.NewPayrollHandler(payrollService)

// Update router call
router := appHTTP.NewRouter(
    JWTService,
    authHandler,
    companyHandler,
    leaveHandler,
    masterHandler,
    scheduleHandler,
    attendanceHandler,
    employeeHandler,
    invitationHandler,
    payrollHandler,  // Add this
    cfg.Storage.BasePath,
)
```

---

## STEP 7: Update Employee Domain

### Update: `internal/domain/employee/entity.go`

Add `BaseSalary` field:

```go
import "github.com/shopspring/decimal"

type Employee struct {
    // ...existing fields...
    BaseSalary *decimal.Decimal  // Add this
}
```

### Update: `internal/domain/employee/dto.go`

Add to DTOs:

```go
import "github.com/shopspring/decimal"

type CreateEmployeeRequest struct {
    // ...existing fields...
    BaseSalary *decimal.Decimal `json:"base_salary,omitempty"`
}

type UpdateEmployeeRequest struct {
    // ...existing fields...
    BaseSalary *decimal.Decimal `json:"base_salary,omitempty"`
}

type EmployeeResponse struct {
    // ...existing fields...
    BaseSalary *decimal.Decimal `json:"base_salary,omitempty"`
}
```

### Update: `internal/repository/postgresql/employee.go`

Add `base_salary` to Create/Update/Get queries.

---

## API Endpoints Summary

| Endpoint | Method | Description | Role |
|----------|--------|-------------|------|
| `/payroll/settings` | GET | Get company settings | owner, manager |
| `/payroll/settings` | PUT | Update settings | owner |
| `/payroll/components` | POST | Create component | owner, manager |
| `/payroll/components` | GET | List components | owner, manager |
| `/payroll/components/{id}` | GET | Get component | owner, manager |
| `/payroll/components/{id}` | PUT | Update component | owner, manager |
| `/payroll/components/{id}` | DELETE | Delete component | owner |
| `/payroll/employees/{id}/components` | POST | Assign to employee | owner, manager |
| `/payroll/employees/{id}/components` | GET | Get employee components | owner, manager |
| `/payroll/employee-components/{id}` | DELETE | Remove assignment | owner, manager |
| `/payroll/generate` | POST | Generate payroll | owner, manager |
| `/payroll/records` | GET | List records | owner, manager |
| `/payroll/records/{id}` | GET | Get record | owner, manager |
| `/payroll/records/{id}` | PUT | Update record | owner, manager |
| `/payroll/records/{id}` | DELETE | Delete record | owner |
| `/payroll/finalize` | POST | Mark as paid | owner |
| `/payroll/summary` | GET | Get period summary | owner, manager |

---

## Formula Perhitungan

```
Gross Salary = Base Salary + Total Allowances + Overtime Amount
Net Salary = Gross Salary - Total Deductions - Late Deduction - Early Leave Deduction

Where:
- Late Deduction = total_late_minutes × late_deduction_per_minute
- Overtime Amount = total_overtime_minutes × overtime_pay_per_minute
- Early Leave Deduction = total_early_leave_minutes × early_leave_deduction_per_minute
```

---

## Implementation Order

1. ☐ Step 1: Create migration files (up & down)
2. ☐ Step 2: Create domain layer (entity, dto, repository interface, service interface, errors)
3. ☐ Step 3: Create repository implementation
4. ☐ Step 4: Create service implementation
5. ☐ Step 5: Create handler
6. ☐ Step 6: Update router & main.go
7. ☐ Step 7: Update employee entity/dto/repository for base_salary
