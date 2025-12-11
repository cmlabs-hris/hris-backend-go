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
    )
);

-- 5. Table: payroll_components
-- Master table for salary components (allowances/deductions)
CREATE TABLE payroll_components (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    type payroll_component_type NOT NULL,
    description TEXT,
    is_taxable BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT uk_payroll_component_name UNIQUE (company_id, name)
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
    
    CONSTRAINT chk_amount_non_negative CHECK (amount >= 0),
    CONSTRAINT chk_date_range CHECK (end_date IS NULL OR end_date >= effective_date)
);

-- 7. Table: payroll_records
-- Generated payroll per employee per period
CREATE TABLE payroll_records (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    period_month SMALLINT NOT NULL,
    period_year SMALLINT NOT NULL,
    
    -- Base
    base_salary DECIMAL(15,2) NOT NULL,
    
    -- Components
    total_allowances DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_deductions DECIMAL(15,2) NOT NULL DEFAULT 0,
    allowances_detail JSONB,
    deductions_detail JSONB,
    
    -- Attendance
    total_work_days INT NOT NULL DEFAULT 0,
    total_late_minutes INT NOT NULL DEFAULT 0,
    late_deduction_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_early_leave_minutes INT NOT NULL DEFAULT 0,
    early_leave_deduction_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_overtime_minutes INT NOT NULL DEFAULT 0,
    overtime_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    
    -- Calculated
    gross_salary DECIMAL(15,2) NOT NULL,
    net_salary DECIMAL(15,2) NOT NULL,
    
    -- Status
    status payroll_status NOT NULL DEFAULT 'draft',
    paid_at TIMESTAMPTZ,
    paid_by UUID REFERENCES users(id),
    notes TEXT,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT uk_employee_period UNIQUE (employee_id, period_month, period_year),
    CONSTRAINT chk_period_month CHECK (period_month BETWEEN 1 AND 12),
    CONSTRAINT chk_period_year CHECK (period_year >= 2020)
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
