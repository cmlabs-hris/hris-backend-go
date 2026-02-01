package subscription

import (
	"time"

	"github.com/shopspring/decimal"
)

// SubscriptionStatus represents the status of a subscription
type SubscriptionStatus string

const (
	StatusTrial     SubscriptionStatus = "trial"
	StatusActive    SubscriptionStatus = "active"
	StatusPastDue   SubscriptionStatus = "past_due"
	StatusCancelled SubscriptionStatus = "cancelled"
	StatusExpired   SubscriptionStatus = "expired"
)

// InvoiceStatus represents the status of an invoice
type InvoiceStatus string

const (
	InvoiceStatusPending InvoiceStatus = "pending"
	InvoiceStatusPaid    InvoiceStatus = "paid"
	InvoiceStatusExpired InvoiceStatus = "expired"
	InvoiceStatusFailed  InvoiceStatus = "failed"
)

// BillingCycle represents the billing cycle
type BillingCycle string

const (
	BillingCycleMonthly BillingCycle = "monthly"
	BillingCycleYearly  BillingCycle = "yearly"
)

// Feature represents a system feature that can be enabled/disabled per plan
type Feature struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// Plan represents a subscription plan
type Plan struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	PricePerSeat decimal.Decimal `json:"price_per_seat"`
	TierLevel    int             `json:"tier_level"`
	MaxSeats     *int            `json:"max_seats,omitempty"` // nil = unlimited
	IsActive     bool            `json:"is_active"`
	Features     []Feature       `json:"features,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// PlanFeature represents the many-to-many relationship between plans and features
type PlanFeature struct {
	PlanID    string `json:"plan_id"`
	FeatureID string `json:"feature_id"`
	IsActive  bool   `json:"is_active"`
}

// Subscription represents a company's subscription
type Subscription struct {
	ID                 string             `json:"id"`
	CompanyID          string             `json:"company_id"`
	PlanID             string             `json:"plan_id"`
	Status             SubscriptionStatus `json:"status"`
	MaxSeats           int                `json:"max_seats"`
	PendingMaxSeats    *int               `json:"pending_max_seats,omitempty"`
	CurrentPeriodStart time.Time          `json:"current_period_start"`
	CurrentPeriodEnd   time.Time          `json:"current_period_end"`
	TrialEndsAt        *time.Time         `json:"trial_ends_at,omitempty"`
	PendingPlanID      *string            `json:"pending_plan_id,omitempty"`
	BillingCycle       BillingCycle       `json:"billing_cycle"`
	AutoRenew          bool               `json:"auto_renew"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`

	// Joined data
	Plan     *Plan     `json:"plan,omitempty"`
	Features []Feature `json:"features,omitempty"`
}

// Invoice represents a payment invoice with snapshot data
type Invoice struct {
	ID               string          `json:"id"`
	CompanyID        string          `json:"company_id"`
	SubscriptionID   string          `json:"subscription_id"`
	XenditInvoiceID  *string         `json:"xendit_invoice_id,omitempty"`
	XenditInvoiceURL *string         `json:"xendit_invoice_url,omitempty"`
	XenditExpiryDate *time.Time      `json:"xendit_expiry_date,omitempty"`
	Amount           decimal.Decimal `json:"amount"`
	IsProrated       bool            `json:"is_prorated"`
	// Snapshot data (immutable - values at transaction time)
	PlanSnapshotName     string          `json:"plan_snapshot_name"`
	PricePerSeatSnapshot decimal.Decimal `json:"price_per_seat_snapshot"`
	SeatCountSnapshot    int             `json:"seat_count_snapshot"`
	BillingCycleSnapshot BillingCycle    `json:"billing_cycle_snapshot"`
	// Period being purchased
	PeriodStart    time.Time     `json:"period_start"`
	PeriodEnd      time.Time     `json:"period_end"`
	Status         InvoiceStatus `json:"status"`
	IssueDate      time.Time     `json:"issue_date"`
	PaidAt         *time.Time    `json:"paid_at,omitempty"`
	PaymentMethod  *string       `json:"payment_method,omitempty"`
	PaymentChannel *string       `json:"payment_channel,omitempty"`
	Description    *string       `json:"description,omitempty"`
	Notes          *string       `json:"notes,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

// IsActive checks if subscription is in an active state (active or trial)
func (s *Subscription) IsActive() bool {
	return s.Status == StatusActive || s.Status == StatusTrial || s.Status == StatusPastDue
}

// IsExpired checks if the subscription period has ended
func (s *Subscription) IsExpired() bool {
	return time.Now().After(s.CurrentPeriodEnd)
}

// IsInGracePeriod checks if subscription is in grace period (7 days after period end)
func (s *Subscription) IsInGracePeriod() bool {
	gracePeriodEnd := s.CurrentPeriodEnd.Add(7 * 24 * time.Hour)
	now := time.Now()
	return now.After(s.CurrentPeriodEnd) && now.Before(gracePeriodEnd)
}

// HasFeature checks if the subscription includes a specific feature
func (s *Subscription) HasFeature(featureCode string) bool {
	for _, f := range s.Features {
		if f.Code == featureCode {
			return true
		}
	}
	return false
}

// CanAddEmployee checks if more employees can be added
func (s *Subscription) CanAddEmployee(currentCount int) bool {
	return currentCount < s.MaxSeats
}
