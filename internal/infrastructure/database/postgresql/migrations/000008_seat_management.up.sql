-- =========================
-- Seat Management Enhancement
-- =========================
-- Purpose: Support prorated upsell and scheduled downsell
-- Features:
-- - pending_max_seats: Scheduled seat changes (apply at renewal)
-- - is_prorated: Distinguish mid-cycle upsells from regular renewals

-- =========================
-- SUBSCRIPTIONS TABLE
-- =========================

-- Add pending_max_seats column for scheduled downsells
ALTER TABLE subscriptions 
ADD COLUMN pending_max_seats INT NULL CHECK (pending_max_seats > 0);

COMMENT ON COLUMN subscriptions.pending_max_seats IS 'Scheduled seat count to apply at next renewal (downsell only)';

-- Add index for cron job to find pending seat changes
CREATE INDEX idx_subscriptions_pending_max_seats ON subscriptions(pending_max_seats) WHERE pending_max_seats IS NOT NULL;

-- =========================
-- INVOICES TABLE
-- =========================

-- Add is_prorated flag to distinguish invoice types
ALTER TABLE invoices 
ADD COLUMN is_prorated BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN invoices.is_prorated IS 'TRUE = mid-cycle seat increase (prorated), FALSE = regular renewal';

-- Add index for webhook processing optimization
CREATE INDEX idx_invoices_is_prorated ON invoices(is_prorated);
