package subscription

import "errors"

var (
	// Subscription errors
	ErrSubscriptionNotFound     = errors.New("subscription not found")
	ErrSubscriptionExpired      = errors.New("subscription has expired")
	ErrSubscriptionCancelled    = errors.New("subscription has been cancelled")
	ErrAlreadySubscribed        = errors.New("company already has an active subscription")
	ErrInvalidSubscriptionState = errors.New("invalid subscription state for this operation")

	// Plan errors
	ErrPlanNotFound         = errors.New("plan not found")
	ErrPlanNotActive        = errors.New("plan is not active")
	ErrInvalidPlanDowngrade = errors.New("cannot downgrade to a higher tier plan")
	ErrInvalidPlanUpgrade   = errors.New("cannot upgrade to a lower tier plan")
	ErrSamePlan             = errors.New("already subscribed to this plan")
	ErrNotAnUpgrade         = errors.New("target plan is not an upgrade from current plan")
	ErrNotADowngrade        = errors.New("target plan is not a downgrade from current plan")

	// Seat errors
	ErrInsufficientSeats   = errors.New("seat count must be greater than or equal to active employees")
	ErrMaxSeatsReached     = errors.New("maximum seats limit reached")
	ErrExceedsPlanMaxSeats = errors.New("requested seats exceed plan maximum")
	ErrSeatLimitExceeded   = errors.New("seat limit exceeded for current subscription")
	ErrSeatsBelowActive    = errors.New("seat count cannot be less than active employees")

	// Feature errors
	ErrFeatureNotFound     = errors.New("feature not found")
	ErrFeatureNotAllowed   = errors.New("feature not available in current plan")
	ErrFeatureNotAvailable = errors.New("feature not available in current subscription")

	// Invoice errors
	ErrInvoiceNotFound      = errors.New("invoice not found")
	ErrInvoiceAlreadyPaid   = errors.New("invoice has already been paid")
	ErrInvoiceExpired       = errors.New("invoice has expired")
	ErrPendingInvoiceExists = errors.New("pending invoice already exists")

	// Webhook errors
	ErrInvalidWebhookSignature = errors.New("invalid webhook signature")
	ErrWebhookProcessingFailed = errors.New("failed to process webhook")
)
