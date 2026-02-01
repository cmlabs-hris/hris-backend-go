package subscription

import (
	"fmt"

	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/validator"
	"github.com/shopspring/decimal"
)

// ==================== Request DTOs ====================

// CheckoutRequest represents a request to create a new subscription invoice
type CheckoutRequest struct {
	PlanID       string       `json:"plan_id"`
	SeatCount    int          `json:"seat_count"`
	BillingCycle BillingCycle `json:"billing_cycle"`
	PayerEmail   string       `json:"payer_email"`
}

func (r *CheckoutRequest) Validate() error {
	var errs validator.ValidationErrors

	if r.PlanID == "" {
		errs = append(errs, validator.ValidationError{Field: "plan_id", Message: "plan_id is required"})
	}
	if r.SeatCount < 1 {
		errs = append(errs, validator.ValidationError{Field: "seat_count", Message: "seat_count must be at least 1"})
	}
	if r.BillingCycle != BillingCycleMonthly && r.BillingCycle != BillingCycleYearly {
		errs = append(errs, validator.ValidationError{Field: "billing_cycle", Message: "billing_cycle must be 'monthly' or 'yearly'"})
	}
	if r.PayerEmail == "" {
		errs = append(errs, validator.ValidationError{Field: "payer_email", Message: "payer_email is required"})
	} else if !validator.IsValidEmail(r.PayerEmail) {
		errs = append(errs, validator.ValidationError{
			Field:   "email",
			Message: "email must be a valid email address (letters, numbers, ., _, %, +, - allowed before @; must contain @; domain must contain letters, numbers, ., -; must end with a dot and at least 2 letters, e.g. user@example.com)",
		})
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// UpgradeRequest represents a request to upgrade subscription plan
type UpgradeRequest struct {
	PlanID     string `json:"plan_id"`
	SeatCount  int    `json:"seat_count"` // New seat count (must be >= current employees)
	PayerEmail string `json:"payer_email"`
}

func (r *UpgradeRequest) Validate() error {
	var errs validator.ValidationErrors

	if r.PlanID == "" {
		errs = append(errs, validator.ValidationError{Field: "plan_id", Message: "plan_id is required"})
	}
	if r.SeatCount < 1 {
		errs = append(errs, validator.ValidationError{Field: "seat_count", Message: "seat_count must be at least 1"})
	}
	if r.PayerEmail == "" {
		errs = append(errs, validator.ValidationError{Field: "payer_email", Message: "payer_email is required"})
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// DowngradeRequest represents a request to downgrade subscription plan
type DowngradeRequest struct {
	PlanID string `json:"plan_id"`
}

func (r *DowngradeRequest) Validate() error {
	var errs validator.ValidationErrors

	if r.PlanID == "" {
		errs = append(errs, validator.ValidationError{Field: "plan_id", Message: "plan_id is required"})
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// CancelRequest represents a request to cancel subscription
type CancelRequest struct {
	Reason string `json:"reason,omitempty"`
}

// ChangeSeatRequest represents a request to change seat count
type ChangeSeatRequest struct {
	SeatCount int `json:"seat_count"`
}

func (r *ChangeSeatRequest) Validate() error {
	var errs validator.ValidationErrors

	if r.SeatCount < 1 {
		errs = append(errs, validator.ValidationError{Field: "seat_count", Message: "seat_count must be at least 1"})
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// ==================== Response DTOs ====================

// PlanResponse represents a plan in API responses
type PlanResponse struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	PricePerSeat decimal.Decimal   `json:"price_per_seat"`
	TierLevel    int               `json:"tier_level"`
	MaxSeats     *int              `json:"max_seats,omitempty"`
	Features     []FeatureResponse `json:"features"`
}

// FeatureResponse represents a feature in API responses
type FeatureResponse struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// SubscriptionResponse represents a subscription in API responses
type SubscriptionResponse struct {
	ID                 string             `json:"id"`
	Status             SubscriptionStatus `json:"status"`
	Plan               PlanResponse       `json:"plan"`
	MaxSeats           int                `json:"max_seats"`
	PendingMaxSeats    *int               `json:"pending_max_seats,omitempty"`
	UsedSeats          int                `json:"used_seats"`
	CurrentPeriodStart string             `json:"current_period_start"`
	CurrentPeriodEnd   string             `json:"current_period_end"`
	TrialEndsAt        *string            `json:"trial_ends_at,omitempty"`
	BillingCycle       BillingCycle       `json:"billing_cycle"`
	AutoRenew          bool               `json:"auto_renew"`
	PendingPlan        *PlanResponse      `json:"pending_plan,omitempty"`
	Features           []string           `json:"features"` // List of feature codes
}

// InvoiceResponse represents an invoice in API responses
type InvoiceResponse struct {
	ID             string          `json:"id"`
	Amount         decimal.Decimal `json:"amount"`
	Status         InvoiceStatus   `json:"status"`
	IsProrated     bool            `json:"is_prorated"`
	PlanName       string          `json:"plan_name"`
	SeatCount      int             `json:"seat_count"`
	PricePerSeat   decimal.Decimal `json:"price_per_seat"`
	BillingCycle   BillingCycle    `json:"billing_cycle"`
	PeriodStart    string          `json:"period_start"`
	PeriodEnd      string          `json:"period_end"`
	IssueDate      string          `json:"issue_date"`
	PaymentURL     *string         `json:"payment_url,omitempty"`
	ExpiryDate     *string         `json:"expiry_date,omitempty"`
	PaidAt         *string         `json:"paid_at,omitempty"`
	PaymentMethod  *string         `json:"payment_method,omitempty"`
	PaymentChannel *string         `json:"payment_channel,omitempty"`
}

// CheckoutResponse represents the response after creating a checkout invoice
type CheckoutResponse struct {
	Invoice    InvoiceResponse `json:"invoice"`
	PaymentURL string          `json:"payment_url"`
	ExpiresAt  string          `json:"expires_at"`
}

// ChangeSeatResponse represents the response after changing seat count
type ChangeSeatResponse struct {
	Invoice         *InvoiceResponse `json:"invoice,omitempty"`
	Message         string           `json:"message"`
	IsPending       bool             `json:"is_pending"`
	PendingMaxSeats *int             `json:"pending_max_seats,omitempty"`
}

// ==================== Webhook DTOs ====================

// XenditWebhookPayload represents the webhook payload from Xendit
type XenditWebhookPayload struct {
	ID                 string  `json:"id"`
	ExternalID         string  `json:"external_id"`
	UserID             string  `json:"user_id"`
	Status             string  `json:"status"` // PAID, EXPIRED, PENDING
	MerchantName       string  `json:"merchant_name"`
	Amount             float64 `json:"amount"`
	PaidAmount         float64 `json:"paid_amount"`
	BankCode           string  `json:"bank_code"`
	PaidAt             string  `json:"paid_at"`
	PayerEmail         string  `json:"payer_email"`
	Description        string  `json:"description"`
	Currency           string  `json:"currency"`
	PaymentMethod      string  `json:"payment_method"`
	PaymentChannel     string  `json:"payment_channel"`
	PaymentDestination string  `json:"payment_destination"`
}

// ==================== Helper Functions ====================

// ToResponse converts a Plan entity to PlanResponse
func (p *Plan) ToResponse() PlanResponse {
	features := make([]FeatureResponse, len(p.Features))
	for i, f := range p.Features {
		features[i] = FeatureResponse{
			Code:        f.Code,
			Name:        f.Name,
			Description: f.Description,
		}
	}

	return PlanResponse{
		ID:           p.ID,
		Name:         p.Name,
		PricePerSeat: p.PricePerSeat,
		TierLevel:    p.TierLevel,
		MaxSeats:     p.MaxSeats,
		Features:     features,
	}
}

// ToResponse converts a Subscription entity to SubscriptionResponse
func (s *Subscription) ToResponse(usedSeats int, pendingPlan *Plan) SubscriptionResponse {
	var planResp PlanResponse
	if s.Plan != nil {
		planResp = s.Plan.ToResponse()
	}

	featureCodes := make([]string, len(s.Features))
	for i, f := range s.Features {
		featureCodes[i] = f.Code
	}

	resp := SubscriptionResponse{
		ID:                 s.ID,
		Status:             s.Status,
		Plan:               planResp,
		MaxSeats:           s.MaxSeats,
		PendingMaxSeats:    s.PendingMaxSeats,
		UsedSeats:          usedSeats,
		CurrentPeriodStart: s.CurrentPeriodStart.Format("2006-01-02T15:04:05Z07:00"),
		CurrentPeriodEnd:   s.CurrentPeriodEnd.Format("2006-01-02T15:04:05Z07:00"),
		BillingCycle:       s.BillingCycle,
		AutoRenew:          s.AutoRenew,
		Features:           featureCodes,
	}

	if s.TrialEndsAt != nil {
		t := s.TrialEndsAt.Format("2006-01-02T15:04:05Z07:00")
		resp.TrialEndsAt = &t
	}

	if pendingPlan != nil {
		pr := pendingPlan.ToResponse()
		resp.PendingPlan = &pr
	}

	return resp
}

// ToResponse converts an Invoice entity to InvoiceResponse
func (i *Invoice) ToResponse() InvoiceResponse {
	resp := InvoiceResponse{
		ID:           i.ID,
		Amount:       i.Amount,
		Status:       i.Status,
		IsProrated:   i.IsProrated,
		PlanName:     i.PlanSnapshotName,
		SeatCount:    i.SeatCountSnapshot,
		PricePerSeat: i.PricePerSeatSnapshot,
		BillingCycle: i.BillingCycleSnapshot,
		PeriodStart:  i.PeriodStart.Format("2006-01-02"),
		PeriodEnd:    i.PeriodEnd.Format("2006-01-02"),
		IssueDate:    i.IssueDate.Format("2006-01-02T15:04:05Z07:00"),
	}

	if i.XenditInvoiceURL != nil {
		resp.PaymentURL = i.XenditInvoiceURL
	}
	if i.XenditExpiryDate != nil {
		t := i.XenditExpiryDate.Format("2006-01-02T15:04:05Z07:00")
		resp.ExpiryDate = &t
	}
	if i.PaidAt != nil {
		t := i.PaidAt.Format("2006-01-02T15:04:05Z07:00")
		resp.PaidAt = &t
	}
	if i.PaymentMethod != nil {
		resp.PaymentMethod = i.PaymentMethod
	}
	if i.PaymentChannel != nil {
		resp.PaymentChannel = i.PaymentChannel
	}

	return resp
}

// ValidationError for seat count check
func NewInsufficientSeatsError(required int) error {
	return fmt.Errorf("%w: minimum %d seats required", ErrInsufficientSeats, required)
}

// ValidationError for max seats check
func NewExceedsPlanMaxSeatsError(planMax int) error {
	return fmt.Errorf("%w: plan allows maximum %d seats", ErrExceedsPlanMaxSeats, planMax)
}
