package subscription

import "context"

// FeatureRepository handles feature data operations
type FeatureRepository interface {
	// GetByCode retrieves a feature by its code
	GetByCode(ctx context.Context, code string) (Feature, error)

	// List retrieves all features
	List(ctx context.Context) ([]Feature, error)
}

// PlanRepository handles subscription plan data operations
type PlanRepository interface {
	// GetByID retrieves a plan by its ID
	GetByID(ctx context.Context, id string) (Plan, error)

	// GetByName retrieves a plan by its name
	GetByName(ctx context.Context, name string) (Plan, error)

	// ListActive retrieves all active plans with their features
	ListActive(ctx context.Context) ([]Plan, error)

	// GetFeaturesByPlanID retrieves all features for a plan
	GetFeaturesByPlanID(ctx context.Context, planID string) ([]Feature, error)
}

// SubscriptionRepository handles subscription data operations
type SubscriptionRepository interface {
	// GetByID retrieves a subscription by its ID
	GetByID(ctx context.Context, id string) (Subscription, error)

	// GetByCompanyID retrieves a subscription by company ID
	GetByCompanyID(ctx context.Context, companyID string) (Subscription, error)

	// GetByCompanyIDWithFeatures retrieves subscription with plan and features
	GetByCompanyIDWithFeatures(ctx context.Context, companyID string) (Subscription, error)

	// Create creates a new subscription
	Create(ctx context.Context, subscription Subscription) (Subscription, error)

	// Update updates an existing subscription
	Update(ctx context.Context, subscription Subscription) error

	// UpdateStatus updates subscription status
	UpdateStatus(ctx context.Context, id string, status SubscriptionStatus) error

	// UpdateMaxSeats updates subscription max seats
	UpdateMaxSeats(ctx context.Context, id string, maxSeats int) error

	// SetPendingMaxSeats sets or clears pending seat count
	SetPendingMaxSeats(ctx context.Context, id string, pendingMaxSeats *int) error

	// ApplyPendingMaxSeats applies pending seat count to max_seats
	ApplyPendingMaxSeats(ctx context.Context, id string) error

	// UpdatePlan updates subscription plan (for upgrade)
	UpdatePlan(ctx context.Context, id string, planID string, maxSeats int, periodEnd interface{}) error

	// SetPendingPlan sets pending plan for downgrade
	SetPendingPlan(ctx context.Context, id string, pendingPlanID *string) error

	// ListExpiring retrieves subscriptions expiring before a given time
	ListExpiring(ctx context.Context, before interface{}) ([]Subscription, error)

	// ListByStatus retrieves subscriptions by status
	ListByStatus(ctx context.Context, status SubscriptionStatus) ([]Subscription, error)

	// ListWithPendingDowngrade retrieves subscriptions with pending downgrades
	ListWithPendingDowngrade(ctx context.Context) ([]Subscription, error)

	// ListSubscriptionsWithPendingSeats retrieves subscriptions with pending seat changes that are ready to apply
	ListSubscriptionsWithPendingSeats(ctx context.Context) ([]Subscription, error)

	// ApplyPendingPlan applies the pending plan to the subscription
	ApplyPendingPlan(ctx context.Context, id string) error

	// UpdateExpiredToStatus bulk updates subscriptions that have passed their period end
	UpdateExpiredToStatus(ctx context.Context, cutoffTime interface{}, fromStatuses []SubscriptionStatus, toStatus SubscriptionStatus) (int64, error)
}

// InvoiceRepository handles invoice data operations
type InvoiceRepository interface {
	// GetByID retrieves an invoice by its ID
	GetByID(ctx context.Context, id string) (Invoice, error)

	// GetByXenditID retrieves an invoice by Xendit invoice ID
	GetByXenditID(ctx context.Context, xenditID string) (Invoice, error)

	// Create creates a new invoice
	Create(ctx context.Context, invoice Invoice) (Invoice, error)

	// UpdateStatus updates invoice status
	UpdateStatus(ctx context.Context, id string, status InvoiceStatus) error

	// UpdatePayment updates invoice with payment details
	UpdatePayment(ctx context.Context, id string, status InvoiceStatus, paidAt interface{}, method, channel string) error

	// ListByCompanyID retrieves all invoices for a company
	ListByCompanyID(ctx context.Context, companyID string) ([]Invoice, error)

	// ListBySubscriptionID retrieves all invoices for a subscription
	ListBySubscriptionID(ctx context.Context, subscriptionID string) ([]Invoice, error)

	// ListPending retrieves all pending invoices
	ListPending(ctx context.Context) ([]Invoice, error)

	// ListPendingOlderThan retrieves pending invoices older than a given time
	ListPendingOlderThan(ctx context.Context, olderThan interface{}) ([]Invoice, error)

	// HasPendingInvoice checks if company has a pending invoice
	HasPendingInvoice(ctx context.Context, companyID string) (bool, error)

	// CountPendingInvoicesBySubscription counts pending invoices for a subscription
	CountPendingInvoicesBySubscription(ctx context.Context, subscriptionID string) (int, error)

	// ExpireStaleInvoices marks old pending invoices as expired
	ExpireStaleInvoices(ctx context.Context, olderThan interface{}) (int64, error)
}

// EmployeeCounter provides method to count active employees
// This is implemented by employee repository
type EmployeeCounter interface {
	// CountActiveByCompanyID counts active employees for a company
	CountActiveByCompanyID(ctx context.Context, companyID string) (int, error)
}
