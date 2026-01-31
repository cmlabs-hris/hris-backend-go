package response

import (
	"errors"
	"log"
	"net/http"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/attendance"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/auth"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/company"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/employee"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/invitation"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/leave"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/master/branch"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/master/grade"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/master/position"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/schedule"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/subscription"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/user"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/validator"
)

// HandleError maps domain errors to HTTP responses
func HandleError(w http.ResponseWriter, err error) {
	// Check if it's a validation error
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		ValidationError(w, validationErrs.ToMap())
		return
	}

	switch {
	// Security: generic message for registration conflicts
	// case errors.Is(err, user.ErrUserEmailExists), errors.Is(err, company.ErrCompanyUsernameExists):
	// 	Conflict(w, "Registration failed: one or more fields already exist")

	// Auth domain errors
	case errors.Is(err, auth.ErrEmailAlreadyExists):
		Conflict(w, "Account with this email already exists")
	case errors.Is(err, auth.ErrInvalidCredentials):
		Unauthorized(w, err.Error())
	case errors.Is(err, auth.ErrInvalidEmployeeCodeCredentials):
		Unauthorized(w, err.Error())
	case errors.Is(err, auth.ErrTokenExpired):
		Unauthorized(w, "Token expired")
	case errors.Is(err, auth.ErrRefreshTokenRevoked):
		Unauthorized(w, "Refresh token revoked")
	case errors.Is(err, auth.ErrEmailNotVerified):
		Forbidden(w, "Email not verified")
	case errors.Is(err, auth.ErrUserNotFound):
		NotFound(w, "User not found")
	case errors.Is(err, auth.ErrCompanyNotFound):
		NotFound(w, "Company not found")
	case errors.Is(err, auth.ErrAccountLocked):
		Forbidden(w, "Account is locked")
	case errors.Is(err, auth.ErrInvalidToken):
		Unauthorized(w, "Invalid or expired token")
	case errors.Is(err, auth.ErrStateCookieNotFound):
		Unauthorized(w, "State cookie not found")
	case errors.Is(err, auth.ErrStateMismatch):
		Unauthorized(w, "State mismatch: value from cookie does not match value from URL parameter")
	case errors.Is(err, auth.ErrStateParamEmpty):
		Unauthorized(w, "State param is empty")
	case errors.Is(err, auth.ErrStateCookieEmpty):
		Unauthorized(w, "State cookie is empty")
	case errors.Is(err, auth.ErrCodeValueEmpty):
		BadRequest(w, "Code value is empty", nil)
	case errors.Is(err, auth.ErrGoogleAccessDeniedByUser):
		Unauthorized(w, "Google access denied by user")
	case errors.Is(err, auth.ErrRefreshTokenCookieNotFound):
		Unauthorized(w, "Refresh token cookie not found")
	case errors.Is(err, auth.ErrRefreshTokenCookieEmpty):
		Unauthorized(w, "Refresh token cookie is empty")

	// Employee domain errors
	case errors.Is(err, employee.ErrEmployeeNotFound):
		NotFound(w, "Employee not found")
	case errors.Is(err, employee.ErrEmployeeCodeExists):
		Conflict(w, "Employee code already exists")
	case errors.Is(err, employee.ErrNIKExists):
		Conflict(w, "NIK already registered")
	case errors.Is(err, employee.ErrEmailExists):
		Conflict(w, "Email already registered in this company")
	case errors.Is(err, employee.ErrInvalidEmployeeCode):
		BadRequest(w, "Invalid employee code format", nil)
	case errors.Is(err, employee.ErrInvalidNIK):
		BadRequest(w, "NIK must be exactly 16 digits", nil)
	case errors.Is(err, employee.ErrInvalidPhoneNumber):
		BadRequest(w, "Phone number must be 10-13 digits", nil)
	case errors.Is(err, employee.ErrInvalidGender):
		BadRequest(w, "Gender must be Male or Female", nil)
	case errors.Is(err, employee.ErrMinimumAge):
		BadRequest(w, "Employee must be at least 17 years old", nil)
	case errors.Is(err, employee.ErrFutureDateNotAllowed):
		BadRequest(w, "Date cannot be in the future", nil)

	// Leave domain errors
	case errors.Is(err, leave.ErrLeaveRequestNotFound):
		NotFound(w, "Leave request not found")
	case errors.Is(err, leave.ErrInsufficientQuota):
		BadRequest(w, "Insufficient leave quota", nil)
	case errors.Is(err, leave.ErrLeaveRequestAlreadyProcessed):
		Conflict(w, "Leave request already processed")
	case errors.Is(err, leave.ErrLeaveTypeNotFound):
		NotFound(w, "Leave type not found")
	case errors.Is(err, leave.ErrLeaveTypesNotFound):
		NotFound(w, "Leave types not found")
	case errors.Is(err, leave.ErrLeaveTypeCodeExists):
		Conflict(w, "Leave type code already exists")
	case errors.Is(err, leave.ErrLeaveTypeNameExists):
		Conflict(w, "Leave type name already exists")
	case errors.Is(err, leave.ErrLeaveTypeInactive):
		BadRequest(w, "Leave type is not active", nil)
	case errors.Is(err, leave.ErrQuotaNotFound):
		NotFound(w, "Leave quota not found")
	case errors.Is(err, leave.ErrOverlappingLeave):
		Conflict(w, "Leave dates overlap with existing request")
	case errors.Is(err, leave.ErrLeaveAlreadyProcessed):
		Conflict(w, "leave request is not in waiting approval status")
	case errors.Is(err, leave.ErrBackdateNotAllowed):
		BadRequest(w, "Backdate leave is not allowed", nil)
	case errors.Is(err, leave.ErrBackdateTooOld):
		BadRequest(w, "Backdate exceeds maximum allowed days", nil)
	case errors.Is(err, leave.ErrInsufficientNotice):
		BadRequest(w, "Insufficient notice period", nil)
	case errors.Is(err, leave.ErrTooFarAdvance):
		BadRequest(w, "Leave date is too far in advance", nil)
	case errors.Is(err, leave.ErrExceedsMaxDays):
		BadRequest(w, "Leave duration exceeds maximum days per request", nil)
	case errors.Is(err, leave.ErrAttachmentRequired):
		BadRequest(w, "Attachment is required for this leave type", nil)
	case errors.Is(err, leave.ErrNotEligible):
		Forbidden(w, "Employee is not eligible for this leave type")
	case errors.Is(err, leave.ErrInsufficientTenure):
		Forbidden(w, "Insufficient tenure for this leave type")
	case errors.Is(err, leave.ErrProbationNotEligible):
		Forbidden(w, "Probation employees are not eligible")
	case errors.Is(err, leave.ErrQuotaNotAvailable):
		BadRequest(w, "No quota available for this leave type", nil)
	case errors.Is(err, leave.ErrPositionNotEligible):
		Forbidden(w, "Employee position is not eligible for this leave type")
	case errors.Is(err, leave.ErrGradeNotEligible):
		Forbidden(w, "Employee grade is not eligible for this leave type")
	case errors.Is(err, leave.ErrEmploymentTypeNotEligible):
		Forbidden(w, "Employee employment type is not eligible for this leave type")
	case errors.Is(err, leave.ErrCombinedRequirementsNotMet):
		Forbidden(w, "Employee does not meet combined eligibility requirements")
	case errors.Is(err, leave.ErrMinimumTenureNotMet):
		Forbidden(w, "Employee does not meet minimum tenure requirement")
	case errors.Is(err, leave.ErrFileSizeExceeds):
		BadRequest(w, "File size exceeds 5MB", nil)
	case errors.Is(err, leave.ErrFileTypeNotAllowed):
		BadRequest(w, "File type not allowed. Allowed: pdf, jpg, jpeg, png", nil)
	case errors.Is(err, leave.ErrUnauthorizedAccess):
		Forbidden(w, "Unauthorized access to leave request")
	case errors.Is(err, leave.ErrUnauthorizedAccessQuota):
		Forbidden(w, "Unauthorized access to leave quota")
	case errors.Is(err, leave.ErrNegativeQuota):
		BadRequest(w, "Adjustment would result in negative available quota", nil)

	// User domain errors
	case errors.Is(err, user.ErrUserNotFound):
		NotFound(w, "User not found")
	case errors.Is(err, user.ErrInvalidEmailFormat):
		BadRequest(w, "Invalid email format", nil)
	case errors.Is(err, user.ErrInvalidPasswordLength):
		BadRequest(w, "Password must be at least 8 characters", nil)
	case errors.Is(err, user.ErrInvalidOAuthProvider):
		BadRequest(w, "Invalid oauth provider", nil)
	case errors.Is(err, user.ErrOAuthProviderIDExists):
		Conflict(w, "OAuth provider id already registered")
	case errors.Is(err, user.ErrEmailNotVerified):
		Forbidden(w, "Email not verified")
	case errors.Is(err, user.ErrEmailVerificationTokenEmpty):
		BadRequest(w, "Email verification token is empty", nil)
	case errors.Is(err, user.ErrAdminPrivilegeRequired):
		Forbidden(w, "Admin privilege required")
	case errors.Is(err, user.ErrOwnerAccessRequired):
		Forbidden(w, "Owner access required")
	case errors.Is(err, user.ErrPendingRoleRequired):
		Forbidden(w, "Pending role required")
	case errors.Is(err, user.ErrManagerAccessRequired):
		Forbidden(w, "Manager access required")
	case errors.Is(err, user.ErrPendingRoleAccessRequired):
		Forbidden(w, "Pending role access required")
	case errors.Is(err, user.ErrInsufficientPermissions):
		Forbidden(w, "Insufficient permissions")
	case errors.Is(err, user.ErrCompanyIDRequired):
		Forbidden(w, "Create a company or join a company to access")
	case errors.Is(err, user.ErrUpdatedAtBeforeCreatedAt):
		BadRequest(w, "updated_at cannot be before created_at", nil)

	// Company domain errors
	case errors.Is(err, company.ErrCompanyNotFound):
		NotFound(w, "Company not found")
	case errors.Is(err, company.ErrInvalidCompanyUsernameFormat):
		BadRequest(w, "Invalid company username format", nil)
	case errors.Is(err, company.ErrInvalidCompanyName):
		BadRequest(w, "Company name cannot be empty", nil)
	case errors.Is(err, company.ErrUpdatedAtBeforeCreatedAt):
		BadRequest(w, "updated_at cannot be before created_at", nil)
	case errors.Is(err, company.ErrCompanyUsernameExists):
		Conflict(w, "Company username already exists")
	case errors.Is(err, company.ErrFileSizeExceeds):
		BadRequest(w, "File size exceeds 5MB", nil)

	// Master data - Branch domain errors
	case errors.Is(err, branch.ErrBranchNotFound):
		NotFound(w, "Branch not found")
	case errors.Is(err, branch.ErrBranchNameExists):
		Conflict(w, "Branch with this name already exists")
	case errors.Is(err, branch.ErrBranchesNotFound):
		NotFound(w, "Branches not found")
	case errors.Is(err, branch.ErrUnauthorizedAccess):
		Forbidden(w, "Unauthorized access to branch")
	case errors.Is(err, branch.ErrInvalidTimezone):
		BadRequest(w, "Invalid timezone", nil)

	// Master data - Grade domain errors
	case errors.Is(err, grade.ErrGradeNotFound):
		NotFound(w, "Grade not found")
	case errors.Is(err, grade.ErrGradeNameExists):
		Conflict(w, "Grade with this name already exists")
	case errors.Is(err, grade.ErrGradesNotFound):
		NotFound(w, "Grades not found")
	case errors.Is(err, grade.ErrUnauthorizedAccess):
		Forbidden(w, "Unauthorized access to grade")

	// Master data - Position domain errors
	case errors.Is(err, position.ErrPositionNotFound):
		NotFound(w, "Position not found")
	case errors.Is(err, position.ErrPositionNameExists):
		Conflict(w, "Position with this name already exists")
	case errors.Is(err, position.ErrPositionsNotFound):
		NotFound(w, "Positions not found")
	case errors.Is(err, position.ErrUnauthorizedAccess):
		Forbidden(w, "Unauthorized access to position")

	// Schedule errors
	// Schedule domain errors
	case errors.Is(err, schedule.ErrWorkScheduleNotFound):
		NotFound(w, "Work schedule not found")
	case errors.Is(err, schedule.ErrWorkScheduleAlreadyDeleted):
		NotFound(w, "Work schedule not found or already deleted")
	case errors.Is(err, schedule.ErrWorkScheduleNameExists):
		Conflict(w, "Work schedule with this name already exists")
	case errors.Is(err, schedule.ErrWorkScheduleTimeNotFound):
		NotFound(w, "Work schedule time not found")
	case errors.Is(err, schedule.ErrWorkScheduleLocationNotFound):
		NotFound(w, "Work schedule location not found")
	case errors.Is(err, schedule.ErrEmployeeScheduleAssignmentNotFound):
		NotFound(w, "Employee schedule assignment not found")
	case errors.Is(err, schedule.ErrOverlappingScheduleAssignment):
		Conflict(w, "Overlapping schedule assignment detected")
	case errors.Is(err, schedule.ErrEmployeeIDRequired):
		BadRequest(w, "Employee ID is required", nil)
	case errors.Is(err, schedule.ErrInvalidDateFormat):
		BadRequest(w, "Invalid date format. Use YYYY-MM-DD", nil)
	case errors.Is(err, schedule.ErrWorkScheduleTimeExists):
		Conflict(w, "Work schedule time already exists")
	case errors.Is(err, schedule.ErrInvalidWorkScheduleType):
		BadRequest(w, "Work schedule type must be 'WFO' or 'Hybrid'", nil)
	case errors.Is(err, schedule.ErrEmployeeScheduleTimelineNotFound):
		NotFound(w, "Employee schedule timeline not found")
	case errors.Is(err, schedule.ErrMismatchedLocationType):
		BadRequest(w, "Mismatched location type for work schedule", nil)

	// Attendance domain errors
	case errors.Is(err, attendance.ErrAlreadyCheckedIn):
		Conflict(w, "You have already checked in today")
	case errors.Is(err, attendance.ErrNoScheduleFound):
		NotFound(w, "No schedule found for today")
	case errors.Is(err, attendance.ErrOutsideAllowedRadius):
		Forbidden(w, "You are outside the allowed radius")
	case errors.Is(err, attendance.ErrTooEarlyToCheckIn):
		BadRequest(w, "Too early to check in", nil)
	case errors.Is(err, attendance.ErrNotCheckedIn):
		BadRequest(w, "You have not checked in yet", nil)
	case errors.Is(err, attendance.ErrAlreadyCheckedOut):
		Conflict(w, "You have already checked out")
	case errors.Is(err, attendance.ErrAttendanceNotFound):
		NotFound(w, "Attendance record not found")
	case errors.Is(err, attendance.ErrUnauthorized):
		Forbidden(w, "Unauthorized to access this attendance record")

	// Invitation domain errors
	case errors.Is(err, invitation.ErrInvitationNotFound):
		NotFound(w, "Invitation not found")
	case errors.Is(err, invitation.ErrInvitationExpired):
		BadRequest(w, "Invitation has expired", nil)
	case errors.Is(err, invitation.ErrInvitationAlreadyUsed):
		Conflict(w, "Invitation has already been used")
	case errors.Is(err, invitation.ErrInvitationRevoked):
		BadRequest(w, "Invitation has been revoked", nil)
	case errors.Is(err, invitation.ErrEmailAlreadyInvited):
		Conflict(w, "This email already has a pending invitation")
	case errors.Is(err, invitation.ErrEmailMismatch):
		Forbidden(w, "Your email does not match the invitation")
	case errors.Is(err, invitation.ErrNoPendingInvitation):
		NotFound(w, "No pending invitation found for this employee")
	case errors.Is(err, invitation.ErrEmployeeAlreadyLinked):
		Conflict(w, "Employee is already linked to a user")
	case errors.Is(err, invitation.ErrUserAlreadyHasCompany):
		Conflict(w, "User already belongs to a company")
	case errors.Is(err, invitation.ErrCannotRevokeAccepted):
		BadRequest(w, "Cannot revoke an accepted invitation", nil)
	case errors.Is(err, employee.ErrCannotDeleteSelf):
		Forbidden(w, "You cannot delete your own employee record")

	// Subscription domain errors
	case errors.Is(err, subscription.ErrSubscriptionNotFound):
		NotFound(w, "Subscription not found")
	case errors.Is(err, subscription.ErrSubscriptionExpired):
		Forbidden(w, "Subscription has expired")
	case errors.Is(err, subscription.ErrSubscriptionCancelled):
		Forbidden(w, "Subscription has been cancelled")
	case errors.Is(err, subscription.ErrAlreadySubscribed):
		Conflict(w, "Company already has an active subscription")
	case errors.Is(err, subscription.ErrInvalidSubscriptionState):
		BadRequest(w, "Invalid subscription state for this operation", nil)

	// Plan errors
	case errors.Is(err, subscription.ErrPlanNotFound):
		NotFound(w, "Subscription plan not found")
	case errors.Is(err, subscription.ErrPlanNotActive):
		BadRequest(w, "Subscription plan is not active", nil)
	case errors.Is(err, subscription.ErrInvalidPlanDowngrade):
		BadRequest(w, "Cannot downgrade to a higher tier plan", nil)
	case errors.Is(err, subscription.ErrInvalidPlanUpgrade):
		BadRequest(w, "Cannot upgrade to a lower tier plan", nil)
	case errors.Is(err, subscription.ErrSamePlan):
		Conflict(w, "Already subscribed to this plan")
	case errors.Is(err, subscription.ErrNotAnUpgrade):
		BadRequest(w, "Target plan is not an upgrade from current plan", nil)
	case errors.Is(err, subscription.ErrNotADowngrade):
		BadRequest(w, "Target plan is not a downgrade from current plan", nil)

	// Seat errors
	case errors.Is(err, subscription.ErrInsufficientSeats):
		BadRequest(w, "Seat count must be greater than or equal to active employees", nil)
	case errors.Is(err, subscription.ErrMaxSeatsReached):
		Forbidden(w, "Maximum seats limit reached")
	case errors.Is(err, subscription.ErrExceedsPlanMaxSeats):
		BadRequest(w, "Requested seats exceed plan maximum", nil)
	case errors.Is(err, subscription.ErrSeatLimitExceeded):
		Forbidden(w, "Seat limit exceeded for current subscription")
	case errors.Is(err, subscription.ErrSeatsBelowActive):
		BadRequest(w, "Seat count cannot be less than active employees", nil)

	// Feature errors
	case errors.Is(err, subscription.ErrFeatureNotFound):
		NotFound(w, "Feature not found")
	case errors.Is(err, subscription.ErrFeatureNotAllowed):
		Forbidden(w, "Feature not available in current plan")
	case errors.Is(err, subscription.ErrFeatureNotAvailable):
		Forbidden(w, "Feature not available in current subscription")

	// Invoice errors
	case errors.Is(err, subscription.ErrInvoiceNotFound):
		NotFound(w, "Invoice not found")
	case errors.Is(err, subscription.ErrInvoiceAlreadyPaid):
		Conflict(w, "Invoice has already been paid")
	case errors.Is(err, subscription.ErrInvoiceExpired):
		BadRequest(w, "Invoice has expired", nil)
	case errors.Is(err, subscription.ErrPendingInvoiceExists):
		Conflict(w, "Pending invoice already exists")

	// Webhook errors
	case errors.Is(err, subscription.ErrInvalidWebhookSignature):
		Forbidden(w, "Invalid webhook signature")
	case errors.Is(err, subscription.ErrWebhookProcessingFailed):
		InternalServerError(w, "Failed to process webhook")

	// Default
	default:
		// Log the error for debugging purposes
		log.Printf("Unhandled error: %v", err)
		InternalServerError(w, "An unexpected error occurred")
	}
}
