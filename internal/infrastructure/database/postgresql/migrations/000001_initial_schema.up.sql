-- =========================
-- Initial HRIS Database Schema
-- =========================

-- Table: companies
-- Stores client companies.
CREATE TABLE companies (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(255) NOT NULL,
    username VARCHAR(50) NOT NULL UNIQUE,
    address TEXT,
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
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    email VARCHAR(254) NOT NULL UNIQUE,
    password_hash VARCHAR(255),
    is_admin BOOLEAN NOT NULL DEFAULT false,
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
    address TEXT, -- Removed for simplicity
    UNIQUE(company_id, name)
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
    UNIQUE(company_id, name)
);

-- Table: employees
-- Main employee table, links to user, position, grade, branch, etc.
CREATE TABLE employees (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    company_id UUID NOT NULL REFERENCES companies(id),
    work_schedule_id UUID REFERENCES work_schedules(id), -- Default schedule
    position_id UUID NOT NULL REFERENCES positions(id),
    grade_id UUID REFERENCES grades(id),
    branch_id UUID REFERENCES branches(id),

    -- Personal & employment data
    employee_code VARCHAR(50),
    full_name VARCHAR(255) NOT NULL,
    nik VARCHAR(16) NOT NULL, -- Indonesian national ID
    gender VARCHAR(10) CHECK (gender IN ('Male', 'Female')),
    phone_number VARCHAR(13) NOT NULL,
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
    bank_name VARCHAR(50) NOT NULL,
    bank_account_holder_name VARCHAR(255),
    bank_account_number VARCHAR(50) NOT NULL,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_updated_at_not_before_created_at CHECK (updated_at >= created_at),
    deleted_at TIMESTAMPTZ,
    UNIQUE(company_id, nik),
    UNIQUE(company_id, employee_code),
    CONSTRAINT chk_nik_length CHECK (char_length(nik) = 16 AND nik <> ''),
    CONSTRAINT chk_phone_number_length CHECK (char_length(phone_number) >= 10 AND char_length(phone_number) <= 13)
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
    location_type VARCHAR(10) NOT NULL DEFAULT 'WFO',
    CONSTRAINT chk_clock_out_after_clock_in CHECK (clock_out_time > clock_in_time),
    CONSTRAINT chk_break_start_after_clock_in CHECK (break_start_time IS NULL OR break_start_time >= clock_in_time),
    CONSTRAINT chk_break_end_required_if_break_start CHECK (
        break_start_time IS NULL OR break_end_time IS NOT NULL
    ),
    CONSTRAINT chk_location_type CHECK (location_type IN ('WFO', 'WFA', 'Hybrid'))
);

-- Table: work_schedule_locations
-- Locations for WFO schedules.
CREATE TABLE work_schedule_locations (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    work_schedule_id UUID NOT NULL REFERENCES work_schedules(id) ON DELETE CASCADE,
    location_name VARCHAR(255) NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    radius_meters INT NOT NULL
);

-- Table: employee_schedule_assignments
-- (Optional) For shift rotation assignments.
CREATE TABLE employee_schedule_assignments (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    employee_id UUID NOT NULL REFERENCES employees(id),
    work_schedule_id UUID NOT NULL REFERENCES work_schedules(id),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL
);

-- Table: leave_types
-- Master table for leave types.
CREATE TABLE leave_types (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    company_id UUID NOT NULL REFERENCES companies(id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    UNIQUE(company_id, name)
);

-- Table: attendances
-- Daily attendance records.
CREATE TABLE attendances (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    employee_id UUID NOT NULL REFERENCES employees(id),
    date DATE NOT NULL,
    work_schedule_time_id UUID REFERENCES work_schedule_times(id),
    actual_location_type VARCHAR(10) NOT NULL,
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
    employee_id UUID NOT NULL REFERENCES employees(id),
    leave_type_id UUID NOT NULL REFERENCES leave_types(id),
    year SMALLINT NOT NULL,
    total_quota SMALLINT NOT NULL,
    taken_quota SMALLINT NOT NULL DEFAULT 0,
    UNIQUE(employee_id, leave_type_id, year)
);

-- Enum: leave_request_status_enum
-- Status for leave requests.
CREATE TYPE leave_request_status_enum AS ENUM ('waiting_approval', 'approved', 'rejected');

-- Table: leave_requests
-- Leave request transactions.
CREATE TABLE leave_requests (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    employee_id UUID NOT NULL REFERENCES employees(id),
    leave_type_id UUID NOT NULL REFERENCES leave_types(id),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    reason TEXT,
    status leave_request_status_enum NOT NULL DEFAULT 'waiting_approval',
    attachment_url TEXT,
    approved_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_end_date_not_before_start_date CHECK (end_date >= start_date)
);

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
-- =========================

-- Frequently used columns for filtering/searching
CREATE INDEX idx_users_company_id ON users(company_id);
CREATE INDEX idx_employees_company_id ON employees(company_id);
CREATE INDEX idx_employees_user_id ON employees(user_id);
CREATE INDEX idx_employees_position_id ON employees(position_id);
CREATE INDEX idx_employees_grade_id ON employees(grade_id);
CREATE INDEX idx_employees_branch_id ON employees(branch_id);
CREATE INDEX idx_attendances_employee_id ON attendances(employee_id);
CREATE INDEX idx_attendances_date ON attendances(date);
CREATE INDEX idx_attendances_employee_date ON attendances(employee_id, date);
CREATE INDEX idx_leave_requests_employee_id ON leave_requests(employee_id);
CREATE INDEX idx_leave_requests_status ON leave_requests(status);
CREATE INDEX idx_employee_documents_employee_id ON employee_documents(employee_id);
CREATE INDEX idx_employee_job_history_employee_id ON employee_job_history(employee_id);
CREATE INDEX idx_employee_job_history_position_id ON employee_job_history(position_id);
CREATE INDEX idx_employee_job_history_grade_id ON employee_job_history(grade_id);
CREATE INDEX idx_employee_job_history_branch_id ON employee_job_history(branch_id);
CREATE INDEX idx_employee_job_history_work_schedule_id ON employee_job_history(work_schedule_id);

-- Indexes for audit trail queries
CREATE INDEX idx_audit_trails_record ON audit_trails (table_name, record_id);
CREATE INDEX idx_audit_trails_user ON audit_trails (user_id);

-- End of schema