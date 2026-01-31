package subscription

import "context"

// SubscriptionService handles subscription business logic
type SubscriptionService interface {
	// ==================== Plan Operations ====================

	// GetPlans retrieves all active plans with their features
	GetPlans(ctx context.Context) ([]PlanResponse, error)

	// GetPlanByID retrieves a specific plan by ID
	GetPlanByID(ctx context.Context, id string) (PlanResponse, error)

	// GetFeatures retrieves all features
	GetFeatures(ctx context.Context) ([]FeatureResponse, error)

	// ==================== Subscription Operations ====================

	// GetMySubscription retrieves the subscription for a specific company
	GetMySubscription(ctx context.Context, companyID string) (SubscriptionResponse, error)

	// CreateTrialSubscription creates a trial subscription for a new company
	// Called during company registration
	CreateTrialSubscription(ctx context.Context, companyID string) (Subscription, error)

	// ==================== Checkout & Payment ====================

	// Checkout creates a new invoice for subscription purchase/renewal
	// Validates: seat_count >= active employees
	Checkout(ctx context.Context, companyID string, req CheckoutRequest) (InvoiceResponse, error)

	// HandleWebhook processes Xendit webhook callback
	HandleWebhook(ctx context.Context, payload XenditWebhookPayload) error

	// ==================== Plan Changes ====================

	// UpgradePlan upgrades to a higher tier plan (immediate, full price new period)
	UpgradePlan(ctx context.Context, companyID string, req UpgradeRequest) (InvoiceResponse, error)

	// DowngradePlan downgrades to a lower tier plan (effective next period)
	DowngradePlan(ctx context.Context, companyID string, req DowngradeRequest) error

	// CancelSubscription cancels the subscription (access until period end)
	CancelSubscription(ctx context.Context, companyID string, req CancelRequest) error

	// ==================== Seat Management ====================

	// ChangeSeats changes the number of seats
	ChangeSeats(ctx context.Context, companyID string, req ChangeSeatRequest) (InvoiceResponse, error)

	// CanAddEmployee checks if more employees can be added to the subscription
	CanAddEmployee(ctx context.Context, companyID string) (bool, error)

	// ==================== Invoice Operations ====================

	// GetInvoices retrieves all invoices for the specified company
	GetInvoices(ctx context.Context, companyID string) ([]InvoiceResponse, error)

	// GetInvoiceByID retrieves a specific invoice
	GetInvoiceByID(ctx context.Context, companyID string, invoiceID string) (InvoiceResponse, error)

	// ==================== Cron Job Operations ====================

	// UpdateExpiredSubscriptions updates subscription statuses based on period end
	// Called by cron job: trial/active -> past_due, past_due -> expired (after grace period)
	UpdateExpiredSubscriptions(ctx context.Context) error

	// CleanupStaleInvoices marks old pending invoices as expired
	// Called by cron job
	CleanupStaleInvoices(ctx context.Context) error

	// ApplyPendingDowngrades applies pending plan downgrades for expired subscriptions
	// Called by cron job
	ApplyPendingDowngrades(ctx context.Context) error

	// ==================== Feature Check ====================

	// HasFeature checks if the company's subscription includes a specific feature
	HasFeature(ctx context.Context, companyID string, featureCode string) (bool, error)

	// GetSubscriptionFeatures retrieves all features for a company's subscription
	GetSubscriptionFeatures(ctx context.Context, companyID string) ([]string, error)
}
