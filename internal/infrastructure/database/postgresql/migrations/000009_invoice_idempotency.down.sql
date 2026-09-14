-- Remove the one-pending-invoice-per-company guard
DROP INDEX IF EXISTS idx_invoices_one_pending_per_company;