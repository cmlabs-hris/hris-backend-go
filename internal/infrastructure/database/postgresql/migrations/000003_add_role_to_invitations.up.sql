-- =========================
-- Add Role to Employee Invitations
-- =========================

-- Add role column to store the intended role when invitation is accepted
ALTER TABLE employee_invitations 
ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'employee';

-- Add check constraint to ensure valid roles (only employee or manager allowed)
ALTER TABLE employee_invitations 
ADD CONSTRAINT chk_invitation_role CHECK (role IN ('employee', 'manager'));

COMMENT ON COLUMN employee_invitations.role IS 'The role to assign when invitation is accepted (employee or manager)';
