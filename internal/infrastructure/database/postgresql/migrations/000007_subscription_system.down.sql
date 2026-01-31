-- =========================
-- Rollback Subscription & Payment System
-- =========================

-- Drop tables in reverse order (respect foreign keys)
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS plan_features;
DROP TABLE IF EXISTS subscription_plans;
DROP TABLE IF EXISTS features;

-- Drop enum types
DROP TYPE IF EXISTS billing_cycle_enum;
DROP TYPE IF EXISTS invoice_status;
DROP TYPE IF EXISTS subscription_status;
