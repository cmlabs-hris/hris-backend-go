-- =========================
-- Subscription & Payment System
-- =========================
-- Business Rules:
-- - Grace Period: 7 days (past_due before expired)
-- - Downgrade: Effective at next period
-- - Seat Limit: Block employee creation when max_seats reached
-- - No proration: Mid-cycle upgrade = full price for new period
-- - Checkout validation: seat_count must be >= active employees

-- =========================
-- ENUM TYPES
-- =========================

-- Subscription Status
-- 'trial': User baru (14 hari)
-- 'active': User berbayar aktif
-- 'past_due': Telat bayar (Grace Period 7 hari)
-- 'cancelled': Stop langganan (masih bisa akses sampai period_end)
-- 'expired': Akses mati total
CREATE TYPE subscription_status AS ENUM ('trial', 'active', 'past_due', 'cancelled', 'expired');

-- Invoice Status (Mapping Xendit)
CREATE TYPE invoice_status AS ENUM ('pending', 'paid', 'expired', 'failed');

-- Billing Cycle
CREATE TYPE billing_cycle_enum AS ENUM ('monthly', 'yearly');

-- =========================
-- MASTER DATA TABLES
-- =========================

-- Table: features
-- Kamus Fitur - Referensi untuk middleware
CREATE TABLE features (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    code VARCHAR(50) NOT NULL UNIQUE,  -- Key untuk coding: 'payroll', 'attendance'
    name VARCHAR(100) NOT NULL,        -- Label UI: 'Payroll System'
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: subscription_plans
-- Paket Langganan (termasuk Free Trial dengan price = 0)
CREATE TABLE subscription_plans (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(50) NOT NULL UNIQUE,       -- 'Free Trial', 'Standard', 'Premium'
    price_per_seat DECIMAL(15,2) NOT NULL,  -- Harga per karyawan (Rupiah)
    tier_level SMALLINT NOT NULL DEFAULT 0, -- 0=Trial, 1=Standard, 2=Premium, 3=Ultra
    max_seats INT,                          -- NULL = unlimited
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_price_non_negative CHECK (price_per_seat >= 0),
    CONSTRAINT chk_tier_level_range CHECK (tier_level >= 0 AND tier_level <= 10)
);

-- Table: plan_features
-- Pivot table - Menentukan paket mana dapat fitur apa
CREATE TABLE plan_features (
    plan_id UUID NOT NULL REFERENCES subscription_plans(id) ON DELETE CASCADE,
    feature_id UUID NOT NULL REFERENCES features(id) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    PRIMARY KEY (plan_id, feature_id)
);

CREATE INDEX idx_plan_features_plan_id ON plan_features(plan_id);
CREATE INDEX idx_plan_features_feature_id ON plan_features(feature_id);

-- =========================
-- TRANSACTION TABLES
-- =========================

-- Table: subscriptions
-- Status Langganan Perusahaan (Jantung Sistem)
-- 1:1 dengan company
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES subscription_plans(id),
    
    status subscription_status NOT NULL DEFAULT 'trial',
    
    -- Limit Karyawan (Seat-Based)
    max_seats INT NOT NULL DEFAULT 5,
    
    -- Period Fields (untuk middleware check)
    -- Block jika NOW() > current_period_end
    current_period_start TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    current_period_end TIMESTAMPTZ NOT NULL,
    trial_ends_at TIMESTAMPTZ, -- Penanda visual kapan trial habis
    
    -- Pending downgrade (efektif di periode berikutnya)
    pending_plan_id UUID REFERENCES subscription_plans(id),
    
    billing_cycle billing_cycle_enum NOT NULL DEFAULT 'monthly',
    
    -- Auto-renewal
    auto_renew BOOLEAN NOT NULL DEFAULT true,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- 1 company = 1 subscription
    CONSTRAINT uk_company_subscription UNIQUE (company_id),
    CONSTRAINT chk_period_valid CHECK (current_period_end > current_period_start),
    CONSTRAINT chk_max_seats_positive CHECK (max_seats > 0)
);

CREATE INDEX idx_subscriptions_company_id ON subscriptions(company_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_period_end ON subscriptions(current_period_end);

-- Table: invoices
-- Log Invoice & Integrasi Xendit dengan Snapshot Data
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    company_id UUID NOT NULL REFERENCES companies(id),
    subscription_id UUID NOT NULL REFERENCES subscriptions(id),
    
    -- Kolom Integrasi Xendit
    xendit_invoice_id VARCHAR(255),      -- ID untuk Webhook matching
    xendit_invoice_url TEXT,             -- Link bayar user
    xendit_expiry_date TIMESTAMPTZ,      -- Kapan link mati
    
    -- SNAPSHOT DATA (Immutable - nilai saat transaksi)
    amount DECIMAL(15,2) NOT NULL,                -- Total Rupiah
    plan_snapshot_name VARCHAR(50) NOT NULL,      -- Nama paket saat beli
    price_per_seat_snapshot DECIMAL(15,2) NOT NULL, -- Harga per seat saat beli
    seat_count_snapshot INT NOT NULL,             -- Jumlah seat saat beli
    billing_cycle_snapshot billing_cycle_enum NOT NULL, -- Cycle saat beli
    
    -- Period yang dibeli
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    
    status invoice_status NOT NULL DEFAULT 'pending',
    
    -- Data Pelunasan (Diisi oleh Webhook)
    issue_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    paid_at TIMESTAMPTZ,
    payment_method VARCHAR(50),   -- 'BANK_TRANSFER', 'EWALLET', 'CREDIT_CARD'
    payment_channel VARCHAR(50),  -- 'BCA', 'MANDIRI', 'OVO', 'GOPAY'
    
    -- Metadata
    description TEXT,
    notes TEXT,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_amount_positive CHECK (amount > 0),
    CONSTRAINT chk_seat_count_positive CHECK (seat_count_snapshot > 0)
);

CREATE INDEX idx_invoices_company_id ON invoices(company_id);
CREATE INDEX idx_invoices_subscription_id ON invoices(subscription_id);
CREATE INDEX idx_invoices_xendit_id ON invoices(xendit_invoice_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_issue_date ON invoices(issue_date);

-- =========================
-- SEED DATA
-- =========================

-- 1. Insert Features (Kamus Fitur)
INSERT INTO features (code, name, description) VALUES
('attendance', 'Attendance System', 'Clock in/out, attendance tracking, GPS-based attendance'),
('leave', 'Leave Management', 'Leave requests, approvals, quota management'),
('payroll', 'Payroll System', 'Salary calculation, payslips, payroll reports'),
('invitation', 'Employee Invitation', 'Invite employees via email with role assignment'),
('schedule', 'Work Schedule', 'Work schedule management and assignment'),
('report', 'Advanced Reports', 'Detailed reports and analytics');

-- 2. Insert Subscription Plans
INSERT INTO subscription_plans (name, price_per_seat, tier_level, max_seats) VALUES
('Free Trial', 0, 0, 5),           -- Trial: max 5 employees, 14 hari
('Standard', 12000, 1, 50),        -- Standard: max 50 employees
('Premium', 15000, 2, 200),        -- Premium: max 200 employees
('Ultra', 20000, 3, NULL);         -- Ultra: unlimited employees

-- 3. Insert Plan Features Mapping

-- Free Trial: Only attendance + leave (limited)
INSERT INTO plan_features (plan_id, feature_id)
SELECT p.id, f.id 
FROM subscription_plans p, features f
WHERE p.name = 'Free Trial' AND f.code IN ('attendance', 'leave');

-- Standard: attendance, leave, invitation, schedule
INSERT INTO plan_features (plan_id, feature_id)
SELECT p.id, f.id 
FROM subscription_plans p, features f
WHERE p.name = 'Standard' AND f.code IN ('attendance', 'leave', 'invitation', 'schedule');

-- Premium: All features
INSERT INTO plan_features (plan_id, feature_id)
SELECT p.id, f.id 
FROM subscription_plans p, features f
WHERE p.name = 'Premium' AND f.code IN ('attendance', 'leave', 'invitation', 'schedule', 'payroll', 'report');

-- Ultra: All features
INSERT INTO plan_features (plan_id, feature_id)
SELECT p.id, f.id 
FROM subscription_plans p, features f
WHERE p.name = 'Ultra' AND f.code IN ('attendance', 'leave', 'invitation', 'schedule', 'payroll', 'report');
