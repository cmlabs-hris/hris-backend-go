package leave

import "errors"

var (
	ErrLeaveRequestNotFound         = errors.New("Leave request not found")
	ErrInsufficientQuota            = errors.New("Insufficient leave quota")
	ErrLeaveRequestAlreadyProcessed = errors.New("Leave request already processed")
)
