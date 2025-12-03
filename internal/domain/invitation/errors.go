package invitation

import "errors"

var (
	ErrInvitationNotFound    = errors.New("invitation not found")
	ErrInvitationExpired     = errors.New("invitation has expired")
	ErrInvitationAlreadyUsed = errors.New("invitation has already been used")
	ErrInvitationRevoked     = errors.New("invitation has been revoked")
	ErrEmailMismatch         = errors.New("your email does not match the invitation email")
	ErrEmailAlreadyInvited   = errors.New("email already has a pending invitation in this company")
	ErrEmployeeAlreadyLinked = errors.New("employee already linked to a user")
	ErrUserAlreadyHasCompany = errors.New("user already belongs to a company")
	ErrCannotRevokeAccepted  = errors.New("cannot revoke an accepted invitation")
	ErrNoPendingInvitation   = errors.New("no pending invitation found for this employee")
)
