-- =========================
-- Remove Role from Employee Invitations
-- =========================

ALTER TABLE employee_invitations DROP CONSTRAINT IF EXISTS chk_invitation_role;
ALTER TABLE employee_invitations DROP COLUMN IF EXISTS role;
