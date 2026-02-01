-- =========================
-- Rollback Seat Management Enhancement
-- =========================

-- Drop indexes
DROP INDEX IF EXISTS idx_invoices_is_prorated;
DROP INDEX IF EXISTS idx_subscriptions_pending_max_seats;

-- Drop columns
ALTER TABLE invoices DROP COLUMN IF EXISTS is_prorated;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS pending_max_seats;
