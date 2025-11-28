-- =========================
-- Initial HRIS Database Schema
-- =========================

CREATE EXTENSION IF NOT EXISTS btree_gist;

-- Table: companies
-- Stores client companies.
CREATE TABLE companies (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(255),
    username VARCHAR(50) NOT NULL UNIQUE,
    address TEXT,
    logo_url TEXT, 
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_username_format CHECK (
        username ~ '^[A-Za-z0-9._-]{3,50}$'
    ),
    CONSTRAINT chk_updated_at_not_before_created_at CHECK (updated_at >= created_at)
);

-- Table: users
-- User accounts for login, linked to companies.
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    email VARCHAR(254) NOT NULL UNIQUE,
    password_hash VARCHAR(255),
    role VARCHAR(20), -- 'owner', 'manager', 'employee'
    oauth_provider VARCHAR(50) DEFAULT NULL CHECK (oauth_provider IS NULL OR oauth_provider = 'google'),
    oauth_provider_id VARCHAR(255),
    email_verified BOOLEAN NOT NULL DEFAULT false,
    email_verification_token VARCHAR(255),
    email_verification_sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_updated_at_not_before_created_at CHECK (updated_at >= created_at),
    -- UNIQUE(company_id, email), 1 Email can only join one Company
    UNIQUE(oauth_provider, oauth_provider_id),
    CONSTRAINT chk_password_hash_length CHECK (char_length(password_hash) >= 8),
    -- Basic email format validation
    CONSTRAINT chk_email_format CHECK (
            char_length(email) >= 6 AND
            char_length(email) <= 254 AND
            email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'
        )
);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_company_role ON users(company_id, role);

-- Store refresh tokens for revocation
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE, -- SHA256 hash of token
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,
    
    -- Track device/session
    user_agent TEXT,
    ip_address VARCHAR(45)
);

-- Table: positions
-- Master table for job positions.
CREATE TABLE positions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    company_id UUID NOT NULL REFERENCES companies(id),
    -- unit_id UUID NOT NULL REFERENCES organization_units(id), -- Uncomment if using organization_units
    name VARCHAR(100) NOT NULL,
    UNIQUE(company_id, name)
);

-- Table: grades
-- Master table for compensation grades/levels.
CREATE TABLE grades (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    company_id UUID NOT NULL REFERENCES companies(id),
    name VARCHAR(100) NOT NULL,
    -- level SMALLINT, -- Removed for simplicity
    UNIQUE(company_id, name)
);

-- Table: branches
-- Master table for company branches.
CREATE TABLE branches (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    company_id UUID NOT NULL REFERENCES companies(id),
    name VARCHAR(100) NOT NULL,
    address TEXT,
    timezone VARCHAR(50) NOT NULL DEFAULT 'Asia/Jakarta',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, name),
    CONSTRAINT chk_updated_at_not_before_created_at CHECK (updated_at >= created_at)
);

-- Enum: employment_type_enum
-- Types of employment contracts.
CREATE TYPE employment_type_enum AS ENUM (
    'permanent',
    'probation',
    'contract',
    'internship',
    'freelance'
);

-- Enum: employment_status_enum
-- Employee lifecycle status.
CREATE TYPE employment_status_enum AS ENUM (
    'active',
    'resigned',
    'terminated'
);

-- Table: work_schedules
-- Work schedule templates (WFO/WFA/Hybrid).
CREATE TABLE work_schedules (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    company_id UUID NOT NULL REFERENCES companies(id),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('WFO', 'WFA', 'Hybrid')),
    grace_period_minutes SMALLINT NOT NULL DEFAULT 15, 
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_updated_at_not_before_created_at CHECK (updated_at >= created_at),
    UNIQUE(company_id, name)
);

-- Table: employees
-- Main employee table, links to user, position, grade, branch, etc.
CREATE TABLE employees (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    company_id UUID NOT NULL REFERENCES companies(id),
    work_schedule_id UUID NOT NULL REFERENCES work_schedules(id), -- Default schedule
    position_id UUID NOT NULL REFERENCES positions(id),
    grade_id UUID NOT NULL REFERENCES grades(id),
    branch_id UUID NOT NULL REFERENCES branches(id),

    -- Personal & employment data
    employee_code VARCHAR(50) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    nik VARCHAR(16), -- Indonesian national ID
    gender VARCHAR(10) CHECK (gender IN ('Male', 'Female')),
    phone_number VARCHAR(13),
    address TEXT,
    place_of_birth VARCHAR(100),
    dob DATE,
    avatar_url TEXT,
    education VARCHAR(50),
    hire_date DATE NOT NULL,
    resignation_date DATE,
    employment_type employment_type_enum NOT NULL,
    employment_status employment_status_enum NOT NULL DEFAULT 'active',

    -- Warning Letter (Surat Peringatan)
    warning_letter VARCHAR(10) CHECK (warning_letter IS NULL OR warning_letter IN ('light', 'medium', 'heavy')),

    -- Bank info
    bank_name VARCHAR(50),
    bank_account_holder_name VARCHAR(255),
    bank_account_number VARCHAR(50),

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_updated_at_not_before_created_at CHECK (updated_at >= created_at),
    deleted_at TIMESTAMPTZ,
    UNIQUE(company_id, nik),
    UNIQUE(company_id, employee_code),
    CONSTRAINT chk_nik_length CHECK (nik IS NULL OR char_length(nik) = 16),
    CONSTRAINT chk_phone_number_length CHECK (phone_number IS NULL OR (char_length(phone_number) >= 10 AND char_length(phone_number) <= 13))
);

-- Table: work_schedule_times
-- Details of work hours for each schedule template.
CREATE TABLE work_schedule_times (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    work_schedule_id UUID NOT NULL REFERENCES work_schedules(id) ON DELETE CASCADE,
    day_of_week SMALLINT NOT NULL, -- 1=Monday, ..., 7=Sunday
    clock_in_time TIME NOT NULL,
    break_start_time TIME,
    break_end_time TIME,
    clock_out_time TIME NOT NULL,
    is_next_day_checkout BOOLEAN NOT NULL DEFAULT false,
    location_type VARCHAR(10) NOT NULL DEFAULT 'WFO',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_updated_at_not_before_created_at CHECK (updated_at >= created_at),
    CONSTRAINT chk_break_end_required_if_break_start CHECK (
        break_start_time IS NULL OR break_end_time IS NOT NULL
    ),
    CONSTRAINT chk_location_type CHECK (location_type IN ('WFO', 'WFA', 'Hybrid')),
    UNIQUE(work_schedule_id, day_of_week)
);

-- Table: work_schedule_locations
-- Locations for WFO schedules.
CREATE TABLE work_schedule_locations (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    work_schedule_id UUID NOT NULL REFERENCES work_schedules(id) ON DELETE CASCADE,
    location_name VARCHAR(255) NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    radius_meters INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_updated_at_not_before_created_at CHECK (updated_at >= created_at)
);

-- Table: employee_schedule_assignments
-- (Optional) For shift rotation assignments.
CREATE TABLE employee_schedule_assignments (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    employee_id UUID NOT NULL REFERENCES employees(id),
    work_schedule_id UUID NOT NULL REFERENCES work_schedules(id),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_end_date_after_start CHECK (end_date >= start_date),
    CONSTRAINT chk_updated_at_not_before_created_at CHECK (updated_at >= created_at),
    CONSTRAINT no_overlapping_schedules EXCLUDE USING GIST (
        employee_id WITH =,  -- Jika employee_id SAMA
        daterange(start_date, end_date, '[]') WITH && -- DAN tanggal BERIRISAN
    )
);

-- Table: leave_types
-- Master table for leave types.
CREATE TABLE leave_types (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    
    -- Basic Info
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50), -- 'ANNUAL', 'SICK', 'MATERNITY', 'UNPAID'
    description TEXT,
    color VARCHAR(7), -- #FF5733 for calendar display
    
    -- Policy Rules
    is_active BOOLEAN NOT NULL DEFAULT true,
    requires_approval BOOLEAN NOT NULL DEFAULT true,
    requires_attachment BOOLEAN NOT NULL DEFAULT true,
    attachment_required_after_days SMALLINT, -- e.g., sick leave > 2 days need doctor note
    
    -- Quota Rules
    has_quota BOOLEAN NOT NULL DEFAULT true, -- false for unpaid leave
    accrual_method VARCHAR(20) NOT NULL DEFAULT 'yearly', -- 'yearly', 'monthly', 'daily', 'none'
    
    -- Deduction Rules
    deduction_type VARCHAR(20) NOT NULL DEFAULT 'working_days', -- 'working_days', 'calendar_days'
    allow_half_day BOOLEAN NOT NULL DEFAULT false,
    
    -- Request Rules
    max_days_per_request SMALLINT, -- max 14 days for annual leave
    min_notice_days SMALLINT NOT NULL DEFAULT 0, -- must request 3 days before
    max_advance_days SMALLINT, -- can only request up to 90 days in advance
    allow_backdate BOOLEAN NOT NULL DEFAULT false,
    backdate_max_days SMALLINT, -- can backdate up to 7 days
    
    -- Rollover Rules
    allow_rollover BOOLEAN NOT NULL DEFAULT false,
    max_rollover_days SMALLINT,
    rollover_expiry_month SMALLINT, -- expires March = 3
    
    quota_calculation_type VARCHAR(20) NOT NULL DEFAULT 'fixed', -- 'fixed', 'tenure_based', 'position_based', 'employment_type_based', 'grade_based'
    quota_rules JSONB NOT NULL,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(company_id, code),
    UNIQUE(company_id, name)
);

-- Table: attendances
-- Daily attendance records.
CREATE TABLE attendances (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    employee_id UUID NOT NULL REFERENCES employees(id),
    date DATE NOT NULL,
    work_schedule_time_id UUID REFERENCES work_schedule_times(id),
    actual_location_type VARCHAR(10),
    clock_in TIMESTAMPTZ,
    clock_out TIMESTAMPTZ,
    work_hours_in_minutes SMALLINT,
    clock_in_latitude DOUBLE PRECISION,
    clock_in_longitude DOUBLE PRECISION,
    clock_in_proof_url TEXT,
    clock_out_latitude DOUBLE PRECISION,
    clock_out_longitude DOUBLE PRECISION,
    clock_out_proof_url TEXT,
    status VARCHAR(50) NOT NULL,
    company_id UUID NOT NULL REFERENCES companies(id),
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    rejection_reason TEXT,
    leave_type_id UUID REFERENCES leave_types(id), -- Reference if status is 'leave'
    late_minutes SMALLINT, 
    early_leave_minutes SMALLINT,
    overtime_minutes SMALLINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_updated_at_not_before_created_at CHECK (updated_at >= created_at),
    CONSTRAINT chk_clock_out_after_clock_in CHECK (clock_out IS NULL OR clock_in IS NULL OR clock_out >= clock_in),
    CONSTRAINT chk_work_hours_non_negative CHECK (work_hours_in_minutes IS NULL OR work_hours_in_minutes >= 0),
    UNIQUE(employee_id, date),
    CONSTRAINT chk_actual_location_type CHECK (actual_location_type IN ('WFO', 'WFA', 'Hybrid'))
);

-- Table: leave_quotas
-- Leave quota per employee per year/type.
CREATE TABLE leave_quotas (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    leave_type_id UUID NOT NULL REFERENCES leave_types(id) ON DELETE CASCADE,
    year SMALLINT NOT NULL,
    
    -- Quota Breakdown
    opening_balance SMALLINT DEFAULT 0, -- quota di awal tahun
    earned_quota SMALLINT DEFAULT 0, -- accrued during year
    rollover_quota SMALLINT DEFAULT 0, -- carried forward from previous year
    adjustment_quota SMALLINT DEFAULT 0, -- manual adjustment by HR
    
    -- Usage
    used_quota DECIMAL(4,1) DEFAULT 0, -- support half-day: 0.5, 1.5, etc
    pending_quota DECIMAL(4,1) DEFAULT 0, -- pending approval requests
    
    -- Calculated Fields
    available_quota DECIMAL(4,1) GENERATED ALWAYS AS (
        opening_balance + earned_quota + rollover_quota + adjustment_quota - used_quota - pending_quota
    ) STORED,
    
    -- Expiry
    rollover_expiry_date DATE,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(employee_id, leave_type_id, year),
    CONSTRAINT chk_quotas_non_negative CHECK (
        opening_balance >= 0 AND 
        earned_quota >= 0 AND 
        used_quota >= 0
    )
);

-- Enum: leave_request_status_enum
-- Status for leave requests.
CREATE TYPE leave_request_status_enum AS ENUM ('cancelled', 'waiting_approval', 'approved', 'rejected');
CREATE TYPE leave_duration_enum AS ENUM ('full_day', 'half_day_morning', 'half_day_afternoon');

-- Table: leave_requests
-- Leave request transactions.
CREATE TABLE leave_requests (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    leave_type_id UUID NOT NULL REFERENCES leave_types(id) ON DELETE CASCADE,
    
    -- Date Range
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    
    -- Duration Details
    duration_type leave_duration_enum DEFAULT 'full_day',
    total_days DECIMAL(4,1) NOT NULL, -- calculated: 1.5 days for half-day
    working_days DECIMAL(4,1) NOT NULL, -- excluding weekends/holidays
    
    -- Request Details
    reason TEXT NOT NULL,
    attachment_url TEXT,
    emergency_leave BOOLEAN DEFAULT false,
    is_backdate BOOLEAN DEFAULT false,
    
    -- Approval Workflow
    status leave_request_status_enum NOT NULL DEFAULT 'waiting_approval',
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    rejection_reason TEXT,
    
    -- Cancellation
    cancelled_by UUID REFERENCES users(id),
    cancelled_at TIMESTAMPTZ,
    cancellation_reason TEXT,
    
    -- Metadata
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_end_date_not_before_start_date CHECK (end_date >= start_date),
    CONSTRAINT chk_total_days_positive CHECK (total_days > 0)
);

CREATE INDEX idx_leave_requests_employee_status ON leave_requests(employee_id, status);
CREATE INDEX idx_leave_requests_date_range ON leave_requests(start_date, end_date);

CREATE TABLE public_holidays (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    
    name VARCHAR(255) NOT NULL, -- "Hari Kemerdekaan RI"
    date DATE NOT NULL,
    is_recurring BOOLEAN DEFAULT false, -- true for annual holidays
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(company_id, date)
);
CREATE INDEX idx_public_holidays_date ON public_holidays(company_id, date);

-- Table: document_types
-- Master table for document types.
CREATE TABLE document_types (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    company_id UUID NOT NULL REFERENCES companies(id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    UNIQUE(company_id, name)
);

-- Table: document_templates
-- (Optional) Document templates for generation.
CREATE TABLE document_templates (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    company_id UUID NOT NULL REFERENCES companies(id),
    document_type_id UUID NOT NULL REFERENCES document_types(id),
    name VARCHAR(255) NOT NULL,
    content TEXT NOT NULL -- Template content (e.g., HTML)
);

-- Table: employee_documents
-- Stores employee documents (uploaded/generated).
CREATE TABLE employee_documents (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    document_type_id UUID NOT NULL REFERENCES document_types(id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    file_url TEXT,
    issue_date DATE,
    expiry_date DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_updated_at_not_before_created_at CHECK (updated_at >= created_at)
);

-- Table: employee_job_history
-- Tracks employee job changes (promotion, transfer, etc.).
CREATE TABLE employee_job_history (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    position_id UUID NOT NULL REFERENCES positions(id),
    grade_id UUID NOT NULL REFERENCES grades(id),
    branch_id UUID REFERENCES branches(id),
    work_schedule_id UUID REFERENCES work_schedules(id),
    employment_type employment_type_enum,
    employment_status employment_status_enum,
    warning_letter VARCHAR(10),
    start_date DATE NOT NULL,
    end_date DATE,
    change_reason TEXT, -- e.g., "Annual Promotion", "Branch Transfer"
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Enum: audit_action
-- Actions for audit trail.
CREATE TYPE audit_action AS ENUM ('CREATE', 'UPDATE', 'DELETE', 'APPROVE', 'REJECT', 'LOGIN_SUCCESS', 'LOGIN_FAIL');

-- Table: audit_trails
-- Logs all important actions for auditing.
CREATE TABLE audit_trails (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID REFERENCES users(id), -- Actor (can be NULL for system actions)
    action audit_action NOT NULL, -- Action performed
    
    table_name VARCHAR(255), -- Affected table name
    record_id UUID, -- Affected record ID
    
    -- Change details in JSONB for flexibility
    old_value JSONB, -- Data before change
    new_value JSONB, -- Data after change
    
    description TEXT, -- Additional info, e.g., "User approved attendance"
    ip_address VARCHAR(45), -- Source IP address
    user_agent TEXT, -- Browser/device info
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- =========================
-- Indexes for Performance
-- Based on Repository Query Patterns Analysis
-- =========================

-- Enable pg_trgm for text search optimization
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- =====================
-- USERS TABLE
-- =====================
CREATE INDEX idx_users_company_id ON users(company_id);

-- =====================
-- EMPLOYEES TABLE
-- =====================
CREATE INDEX idx_employees_company_id ON employees(company_id);
CREATE INDEX idx_employees_user_id ON employees(user_id);
CREATE INDEX idx_employees_position_id ON employees(position_id);
CREATE INDEX idx_employees_grade_id ON employees(grade_id);
CREATE INDEX idx_employees_branch_id ON employees(branch_id);
-- Query: GetActiveByCompanyID (WHERE company_id = ? AND employment_status = 'active')
CREATE INDEX idx_employees_company_status ON employees(company_id, employment_status);
-- Query: GetByCompanyID with name search (ILIKE on full_name)
CREATE INDEX idx_employees_full_name_trgm ON employees USING gin(full_name gin_trgm_ops);

-- =====================
-- LEAVE_TYPES TABLE
-- =====================
-- Query: GetByCompanyID, GetActiveByCompanyID (WHERE company_id = ? AND is_active = true)
CREATE INDEX idx_leave_types_company_id ON leave_types(company_id);
CREATE INDEX idx_leave_types_company_active ON leave_types(company_id, is_active) WHERE is_active = true;

-- =====================
-- LEAVE_QUOTAS TABLE
-- =====================
-- Query: GetByEmployee (WHERE employee_id = ?)
CREATE INDEX idx_leave_quotas_employee_id ON leave_quotas(employee_id);
-- Query: GetByEmployeeYear (WHERE employee_id = ? AND year = ?)
CREATE INDEX idx_leave_quotas_employee_year ON leave_quotas(employee_id, year);
-- Optimize leave_quotas JOIN employees JOIN leave_types  
CREATE INDEX idx_leave_quotas_composite ON leave_quotas(employee_id, leave_type_id, year DESC);

-- =====================
-- LEAVE_REQUESTS TABLE  
-- =====================
CREATE INDEX idx_leave_requests_employee_id ON leave_requests(employee_id);
CREATE INDEX idx_leave_requests_status ON leave_requests(status);
-- Query: GetByEmployeeID with filters (WHERE employee_id = ? AND status = ? AND leave_type_id = ?)
CREATE INDEX idx_leave_requests_employee_leave_type ON leave_requests(employee_id, leave_type_id);
-- Query: CheckOverlapping (WHERE employee_id = ? AND status IN ('waiting_approval', 'approved') AND date range)
CREATE INDEX idx_leave_requests_employee_date_overlap ON leave_requests(employee_id, status, start_date, end_date);
-- Query: Sorting by submitted_at (ORDER BY submitted_at DESC)
CREATE INDEX idx_leave_requests_submitted_at ON leave_requests(submitted_at DESC);
-- Optimize leave_requests JOIN employees JOIN leave_types
CREATE INDEX idx_leave_requests_composite ON leave_requests(employee_id, leave_type_id, status, submitted_at DESC);

-- =====================
-- ATTENDANCES TABLE
-- =====================
CREATE INDEX idx_attendances_employee_id ON attendances(employee_id);
CREATE INDEX idx_attendances_date ON attendances(date);
CREATE INDEX idx_attendances_employee_date ON attendances(employee_id, date);
CREATE INDEX idx_attendances_company_id ON attendances(company_id);
-- Query: GetOpenSession (WHERE employee_id = ? AND clock_out IS NULL ORDER BY clock_in DESC)
CREATE INDEX idx_attendances_employee_open_session ON attendances(employee_id, clock_in DESC) WHERE clock_out IS NULL;
-- Query: Update (WHERE id = ? AND company_id = ?)
CREATE INDEX idx_attendances_id_company ON attendances(id, company_id);

-- =====================
-- WORK_SCHEDULES TABLE
-- =====================
CREATE INDEX idx_work_schedules_company_id ON work_schedules (company_id);
CREATE INDEX idx_work_schedules_type ON work_schedules (type);
CREATE INDEX idx_work_schedules_deleted_at ON work_schedules (deleted_at);
CREATE UNIQUE INDEX idx_unique_schedule_name ON work_schedules (company_id, name) WHERE deleted_at IS NULL;
-- Query: GetByCompanyID with filters (WHERE company_id = ? AND deleted_at IS NULL AND name ILIKE ?)
CREATE INDEX idx_work_schedules_name_trgm ON work_schedules USING gin(name gin_trgm_ops);
-- Query: GetByID (WHERE id = ? AND company_id = ? AND deleted_at IS NULL)
CREATE INDEX idx_work_schedules_id_company_active ON work_schedules(id, company_id) WHERE deleted_at IS NULL;

-- =====================
-- WORK_SCHEDULE_TIMES TABLE
-- =====================
-- Query: GetByWorkScheduleID (WHERE work_schedule_id = ? ORDER BY day_of_week)
CREATE INDEX idx_work_schedule_times_schedule_id ON work_schedule_times(work_schedule_id);

-- =====================
-- WORK_SCHEDULE_LOCATIONS TABLE
-- =====================
-- Query: GetByWorkScheduleID (WHERE work_schedule_id = ?)
CREATE INDEX idx_work_schedule_locations_schedule_id ON work_schedule_locations(work_schedule_id);

-- =====================
-- EMPLOYEE_SCHEDULE_ASSIGNMENTS TABLE
-- =====================
-- Query: GetByEmployeeID (WHERE employee_id = ? ORDER BY start_date DESC)
CREATE INDEX idx_employee_schedule_assignments_employee_id ON employee_schedule_assignments(employee_id);
-- Query: GetActiveSchedule (WHERE employee_id = ? AND date BETWEEN start_date AND end_date)
CREATE INDEX idx_employee_schedule_assignments_employee_dates ON employee_schedule_assignments(employee_id, start_date, end_date);
-- Query: GetScheduleAssignments with date range overlap
CREATE INDEX idx_employee_schedule_assignments_date_range ON employee_schedule_assignments(employee_id, start_date DESC);

-- =====================
-- REFRESH_TOKENS TABLE
-- =====================
-- Query: Token lookup by user_id and expiry
CREATE INDEX idx_refresh_tokens_user_expires ON refresh_tokens(user_id, expires_at) WHERE revoked_at IS NULL;

-- =====================
-- EMPLOYEE_DOCUMENTS TABLE
-- =====================
CREATE INDEX idx_employee_documents_employee_id ON employee_documents(employee_id);

-- =====================
-- EMPLOYEE_JOB_HISTORY TABLE
-- =====================
CREATE INDEX idx_employee_job_history_employee_id ON employee_job_history(employee_id);
CREATE INDEX idx_employee_job_history_position_id ON employee_job_history(position_id);
CREATE INDEX idx_employee_job_history_grade_id ON employee_job_history(grade_id);
CREATE INDEX idx_employee_job_history_branch_id ON employee_job_history(branch_id);
CREATE INDEX idx_employee_job_history_work_schedule_id ON employee_job_history(work_schedule_id);
-- Query: Historical lookups by employee and date
CREATE INDEX idx_employee_job_history_employee_dates ON employee_job_history(employee_id, start_date DESC);

-- =====================
-- AUDIT_TRAILS TABLE
-- =====================
CREATE INDEX idx_audit_trails_record ON audit_trails (table_name, record_id);
CREATE INDEX idx_audit_trails_user ON audit_trails (user_id);
-- Query: GetByUserId with date range
CREATE INDEX idx_audit_trails_user_created ON audit_trails(user_id, created_at DESC);

-- =====================
-- PUBLIC_HOLIDAYS TABLE
-- =====================
-- Query: GetByCompanyID and date range (for leave calculation)
CREATE INDEX idx_public_holidays_company_date_range ON public_holidays(company_id, date);

-- End of schema