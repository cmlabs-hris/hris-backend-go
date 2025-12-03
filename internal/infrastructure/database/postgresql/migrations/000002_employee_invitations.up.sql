-- =========================
-- Employee Invitations Schema
-- =========================

-- Table: employee_invitations
-- Stores employee invitation links for onboarding new employees
CREATE TABLE employee_invitations (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    invited_by_employee_id UUID NOT NULL REFERENCES employees(id),
    email VARCHAR(254) NOT NULL,
    token UUID UNIQUE NOT NULL DEFAULT uuidv7(),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_invitation_status CHECK (status IN ('pending', 'accepted', 'revoked')),
    CONSTRAINT chk_updated_at_not_before_created_at CHECK (updated_at >= created_at)
);

-- Indexes for performance
CREATE INDEX idx_invitations_token ON employee_invitations(token);
CREATE INDEX idx_invitations_email_company ON employee_invitations(email, company_id);
CREATE INDEX idx_invitations_employee_id ON employee_invitations(employee_id);
CREATE INDEX idx_invitations_company_id ON employee_invitations(company_id);
CREATE INDEX idx_invitations_status_expires ON employee_invitations(status, expires_at) WHERE status = 'pending';

COMMENT ON TABLE employee_invitations IS 'Stores employee invitation links for secure onboarding';
COMMENT ON COLUMN employee_invitations.token IS 'UUIDv7 token used in invitation link';
COMMENT ON COLUMN employee_invitations.status IS 'pending: waiting for user, accepted: user joined, revoked: cancelled by admin';
