-- ============================================
-- Invoice Idempotency & Concurrency Safeguards
-- ============================================
-- Ensures a company can only have ONE pending invoice at a time.
-- This is the database-level backstop (not just application check-then-act)
-- that prevents double payment on the same billing cycle when two checkout
-- requests race concurrently (double-click / client retry).
--
-- NOTE: If existing data already contains duplicate pending invoices for a
-- company, this migration will fail. Resolve duplicate pending invoices
-- (expire/cancel extras) before applying.
CREATE UNIQUE INDEX idx_invoices_one_pending_per_company
ON invoices(company_id) WHERE status = 'pending';