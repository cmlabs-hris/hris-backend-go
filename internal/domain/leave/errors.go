package leave

import "errors"

var (
	// General
	ErrFileSizeExceeds              = errors.New("File size exceeds 5MB")
	ErrFileTypeNotAllowed           = errors.New("File type not allowed. Allowed: pdf, jpg, jpeg, png")
	ErrUnauthorizedAccess           = errors.New("unauthorized access to leave request")
	ErrUnauthorizedAccessQuota      = errors.New("unauthorized access to leave quota")
	ErrLeaveRequestAlreadyProcessed = errors.New("Leave request already processed")

	// Leave Type errors
	ErrLeaveTypeNotFound   = errors.New("leave type not found")
	ErrLeaveTypesNotFound  = errors.New("leave types not found")
	ErrLeaveTypeCodeExists = errors.New("leave type code already exists")
	ErrLeaveTypeNameExists = errors.New("leave type name already exists")
	ErrLeaveTypeInactive   = errors.New("leave type is not active")

	// Leave Quota errors
	ErrQuotaNotFound     = errors.New("leave quota not found")
	ErrInsufficientQuota = errors.New("insufficient leave quota")

	// Leave Request errors
	ErrLeaveRequestNotFound  = errors.New("leave request not found")
	ErrOverlappingLeave      = errors.New("leave dates overlap with existing request")
	ErrLeaveAlreadyProcessed = errors.New("leave request already processed")
	ErrBackdateNotAllowed    = errors.New("backdate leave is not allowed")
	ErrBackdateTooOld        = errors.New("backdate exceeds maximum allowed days")
	ErrInsufficientNotice    = errors.New("insufficient notice period")
	ErrTooFarAdvance         = errors.New("leave date is too far in advance")
	ErrExceedsMaxDays        = errors.New("leave duration exceeds maximum days per request")
	ErrAttachmentRequired    = errors.New("attachment is required for this leave type")

	// Eligibility errors
	ErrNotEligible                = errors.New("employee is not eligible for this leave type")
	ErrInsufficientTenure         = errors.New("insufficient tenure for this leave type")
	ErrProbationNotEligible       = errors.New("probation employees are not eligible")
	ErrQuotaNotAvailable          = errors.New("no quota available for this leave type")
	ErrPositionNotEligible        = errors.New("employee position is not eligible for this leave type")
	ErrGradeNotEligible           = errors.New("employee grade is not eligible for this leave type")
	ErrEmploymentTypeNotEligible  = errors.New("employee employment type is not eligible for this leave type")
	ErrCombinedRequirementsNotMet = errors.New("employee does not meet combined eligibility requirements")
	ErrMinimumTenureNotMet        = errors.New("employee does not meet minimum tenure requirement")

	// Leave Adjustment errors
	ErrNegativeQuota = errors.New("adjustment would result in negative available quota")
)
