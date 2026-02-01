package subscription

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/config"
	"github.com/cmlabs-hris/hris-backend-go/internal/domain/subscription"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/xendit"
	"github.com/cmlabs-hris/hris-backend-go/internal/repository/postgresql"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// Constants
const (
	TrialPlanName       = "Free Trial"
	TrialDurationDays   = 14
	GracePeriodDays     = 7
	YearlyMonthsCharged = 10 // 12 months for price of 10 (2 months free)
)

// exceedsPlanMaxSeats checks if seat count exceeds the plan's max seats limit
// Returns false if plan has no max limit (nil MaxSeats)
func exceedsPlanMaxSeats(plan subscription.Plan, seatCount int) bool {
	if plan.MaxSeats == nil {
		return false // No limit
	}
	return seatCount > *plan.MaxSeats
}

// planMaxSeatsValue returns the plan's max seats or a default value if nil
func planMaxSeatsValue(plan subscription.Plan, defaultVal int) int {
	if plan.MaxSeats == nil {
		return defaultVal
	}
	return *plan.MaxSeats
}

type subscriptionService struct {
	featureRepo      subscription.FeatureRepository
	planRepo         subscription.PlanRepository
	subscriptionRepo subscription.SubscriptionRepository
	invoiceRepo      subscription.InvoiceRepository
	employeeCounter  subscription.EmployeeCounter
	xenditClient     *xendit.Client
	db               *database.DB
	cfg              *config.Config
}

func NewSubscriptionService(
	featureRepo subscription.FeatureRepository,
	planRepo subscription.PlanRepository,
	subscriptionRepo subscription.SubscriptionRepository,
	invoiceRepo subscription.InvoiceRepository,
	employeeCounter subscription.EmployeeCounter,
	xenditClient *xendit.Client,
	db *database.DB,
	cfg *config.Config,
) subscription.SubscriptionService {
	return &subscriptionService{
		featureRepo:      featureRepo,
		planRepo:         planRepo,
		subscriptionRepo: subscriptionRepo,
		invoiceRepo:      invoiceRepo,
		employeeCounter:  employeeCounter,
		xenditClient:     xenditClient,
		db:               db,
		cfg:              cfg,
	}
}

// ==================== Plan & Feature Operations ====================

func (s *subscriptionService) GetPlans(ctx context.Context) ([]subscription.PlanResponse, error) {
	plans, err := s.planRepo.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("list plans: %w", err)
	}

	responses := make([]subscription.PlanResponse, len(plans))
	for i, plan := range plans {
		responses[i] = toPlanResponse(plan)
	}
	return responses, nil
}

func (s *subscriptionService) GetPlanByID(ctx context.Context, id string) (subscription.PlanResponse, error) {
	plan, err := s.planRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return subscription.PlanResponse{}, subscription.ErrPlanNotFound
		}
		return subscription.PlanResponse{}, fmt.Errorf("get plan: %w", err)
	}
	return toPlanResponse(plan), nil
}

func (s *subscriptionService) GetFeatures(ctx context.Context) ([]subscription.FeatureResponse, error) {
	features, err := s.featureRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list features: %w", err)
	}

	responses := make([]subscription.FeatureResponse, len(features))
	for i, f := range features {
		responses[i] = toFeatureResponse(f)
	}
	return responses, nil
}

// ==================== Subscription Operations ====================

func (s *subscriptionService) GetMySubscription(ctx context.Context, companyID string) (subscription.SubscriptionResponse, error) {
	sub, err := s.subscriptionRepo.GetByCompanyIDWithFeatures(ctx, companyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return subscription.SubscriptionResponse{}, subscription.ErrSubscriptionNotFound
		}
		return subscription.SubscriptionResponse{}, fmt.Errorf("get subscription: %w", err)
	}

	// Count active employees for used_seats
	usedSeats, err := s.employeeCounter.CountActiveByCompanyID(ctx, companyID)
	if err != nil {
		return subscription.SubscriptionResponse{}, fmt.Errorf("count active employees: %w", err)
	}

	return toSubscriptionResponse(sub, usedSeats), nil
}

func (s *subscriptionService) CreateTrialSubscription(ctx context.Context, companyID string) (subscription.Subscription, error) {
	// Get the Free Trial plan
	trialPlan, err := s.planRepo.GetByName(ctx, TrialPlanName)
	if err != nil {
		return subscription.Subscription{}, fmt.Errorf("get trial plan: %w", err)
	}

	now := time.Now()
	trialEnd := now.AddDate(0, 0, TrialDurationDays)

	// Default max seats for trial
	maxSeats := 5
	if trialPlan.MaxSeats != nil {
		maxSeats = *trialPlan.MaxSeats
	}

	sub := subscription.Subscription{
		CompanyID:          companyID,
		PlanID:             trialPlan.ID,
		Status:             subscription.StatusTrial,
		MaxSeats:           maxSeats,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   trialEnd,
		TrialEndsAt:        &trialEnd,
		BillingCycle:       subscription.BillingCycleMonthly,
		AutoRenew:          false,
	}

	created, err := s.subscriptionRepo.Create(ctx, sub)
	if err != nil {
		return subscription.Subscription{}, fmt.Errorf("create trial subscription: %w", err)
	}

	return created, nil
}

// ==================== Checkout & Payment ====================

func (s *subscriptionService) Checkout(ctx context.Context, companyID string, req subscription.CheckoutRequest) (subscription.InvoiceResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return subscription.InvoiceResponse{}, err
	}

	// Check if company already has a pending invoice
	hasPending, err := s.invoiceRepo.HasPendingInvoice(ctx, companyID)
	if err != nil {
		return subscription.InvoiceResponse{}, fmt.Errorf("check pending invoice: %w", err)
	}
	if hasPending {
		return subscription.InvoiceResponse{}, subscription.ErrPendingInvoiceExists
	}

	// Get the plan
	plan, err := s.planRepo.GetByID(ctx, req.PlanID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return subscription.InvoiceResponse{}, subscription.ErrPlanNotFound
		}
		return subscription.InvoiceResponse{}, fmt.Errorf("get plan: %w", err)
	}

	// Validate seats (if plan has max limit)
	if plan.MaxSeats != nil && req.SeatCount > *plan.MaxSeats {
		return subscription.InvoiceResponse{}, subscription.ErrSeatLimitExceeded
	}

	// Validate seat count >= active employees
	activeEmployees, err := s.employeeCounter.CountActiveByCompanyID(ctx, companyID)
	if err != nil {
		return subscription.InvoiceResponse{}, fmt.Errorf("count employees: %w", err)
	}
	if req.SeatCount < activeEmployees {
		return subscription.InvoiceResponse{}, subscription.ErrSeatsBelowActive
	}

	// Get or create subscription
	sub, err := s.subscriptionRepo.GetByCompanyID(ctx, companyID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return subscription.InvoiceResponse{}, fmt.Errorf("get subscription: %w", err)
	}

	// Calculate period
	billingCycle := subscription.BillingCycle(req.BillingCycle)
	now := time.Now()
	periodStart := now
	periodEnd := postgresql.CalculatePeriodEnd(periodStart, billingCycle)

	// Calculate amount (no proration - full price for new period)
	amount := postgresql.CalculateAmount(plan.PricePerSeat, req.SeatCount, billingCycle)
	description := postgresql.FormatInvoiceDescription(plan.Name, req.SeatCount, billingCycle)

	// Invoice expiry in seconds (InvoiceExpiry is in hours)
	invoiceExpirySecs := s.cfg.Xendit.InvoiceExpiry * 3600

	// Create Xendit invoice
	xenditReq := xendit.CreateInvoiceRequest{
		ExternalID:         fmt.Sprintf("sub-%s-%d", companyID, now.Unix()),
		Amount:             amount,
		PayerEmail:         req.PayerEmail,
		Description:        description,
		Currency:           "IDR",
		InvoiceDuration:    invoiceExpirySecs,
		SuccessRedirectURL: s.cfg.Xendit.SuccessRedirect,
		FailureRedirectURL: s.cfg.Xendit.FailureRedirect,
	}

	xenditResp, err := s.xenditClient.CreateInvoice(xenditReq)
	if err != nil {
		return subscription.InvoiceResponse{}, fmt.Errorf("create xendit invoice: %w", err)
	}

	// Create invoice record with snapshot data
	xenditID := xenditResp.ID
	xenditURL := xenditResp.InvoiceURL
	invoice := subscription.Invoice{
		CompanyID:            companyID,
		SubscriptionID:       sub.ID,
		XenditInvoiceID:      &xenditID,
		XenditInvoiceURL:     &xenditURL,
		XenditExpiryDate:     &xenditResp.ExpiryDate,
		Amount:               amount,
		PlanSnapshotName:     plan.Name,
		PricePerSeatSnapshot: plan.PricePerSeat,
		SeatCountSnapshot:    req.SeatCount,
		BillingCycleSnapshot: billingCycle,
		PeriodStart:          periodStart,
		PeriodEnd:            periodEnd,
		Status:               subscription.InvoiceStatusPending,
		Description:          &description,
	}

	created, err := s.invoiceRepo.Create(ctx, invoice)
	if err != nil {
		// Try to expire the Xendit invoice if we failed to save
		_, _ = s.xenditClient.ExpireInvoice(xenditResp.ID)
		return subscription.InvoiceResponse{}, fmt.Errorf("create invoice: %w", err)
	}

	return toInvoiceResponse(created), nil
}

func (s *subscriptionService) HandleWebhook(ctx context.Context, payload subscription.XenditWebhookPayload) error {
	// Get invoice by Xendit ID
	invoice, err := s.invoiceRepo.GetByXenditID(ctx, payload.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Unknown invoice - log and ignore
			log.Printf("Webhook: Unknown invoice ID %s", payload.ID)
			return nil
		}
		return fmt.Errorf("get invoice: %w", err)
	}

	// Handle based on status
	switch payload.Status {
	case "PAID", "SETTLED":
		return s.handlePaymentSuccess(ctx, invoice, payload)
	case "EXPIRED":
		return s.handlePaymentExpired(ctx, invoice)
	case "FAILED":
		return s.handlePaymentFailed(ctx, invoice)
	default:
		log.Printf("Webhook: Unhandled status %s for invoice %s", payload.Status, invoice.ID)
	}

	return nil
}

func (s *subscriptionService) handlePaymentSuccess(ctx context.Context, invoice subscription.Invoice, payload subscription.XenditWebhookPayload) error {
	// Update invoice
	paidAt := time.Now()
	if err := s.invoiceRepo.UpdatePayment(
		ctx,
		invoice.ID,
		subscription.InvoiceStatusPaid,
		paidAt,
		payload.PaymentMethod,
		payload.PaymentChannel,
	); err != nil {
		return fmt.Errorf("update invoice payment: %w", err)
	}

	// Get subscription
	sub, err := s.subscriptionRepo.GetByID(ctx, invoice.SubscriptionID)
	if err != nil {
		return fmt.Errorf("get subscription: %w", err)
	}

	// Get plan to get the plan ID for the snapshot name
	plan, err := s.planRepo.GetByName(ctx, invoice.PlanSnapshotName)
	if err != nil {
		return fmt.Errorf("get plan by name: %w", err)
	}

	// Update subscription based on invoice type
	if invoice.IsProrated {
		// Prorated invoice (mid-cycle seat increase) - update seats only, don't extend period
		sub.MaxSeats = invoice.SeatCountSnapshot
		sub.PendingMaxSeats = nil // Clear any pending downsell
		log.Printf("Prorated payment success: Company %s, Seats %d → %d (period unchanged)",
			invoice.CompanyID, sub.MaxSeats, invoice.SeatCountSnapshot)
	} else {
		// Regular renewal or upgrade - update seats AND extend period
		sub.PlanID = plan.ID
		sub.MaxSeats = invoice.SeatCountSnapshot
		sub.CurrentPeriodStart = invoice.PeriodStart
		sub.CurrentPeriodEnd = invoice.PeriodEnd
		sub.BillingCycle = subscription.BillingCycle(invoice.BillingCycleSnapshot)
		sub.PendingPlanID = nil
		sub.PendingMaxSeats = nil
		sub.TrialEndsAt = nil
		log.Printf("Payment success: Company %s, Plan %s, Seats %d",
			invoice.CompanyID, invoice.PlanSnapshotName, invoice.SeatCountSnapshot)
	}

	sub.Status = subscription.StatusActive

	if err := s.subscriptionRepo.Update(ctx, sub); err != nil {
		return fmt.Errorf("update subscription: %w", err)
	}

	return nil
}

func (s *subscriptionService) handlePaymentExpired(ctx context.Context, invoice subscription.Invoice) error {
	if err := s.invoiceRepo.UpdateStatus(ctx, invoice.ID, subscription.InvoiceStatusExpired); err != nil {
		return fmt.Errorf("update invoice status: %w", err)
	}
	log.Printf("Invoice expired: %s for company %s", invoice.ID, invoice.CompanyID)
	return nil
}

func (s *subscriptionService) handlePaymentFailed(ctx context.Context, invoice subscription.Invoice) error {
	if err := s.invoiceRepo.UpdateStatus(ctx, invoice.ID, subscription.InvoiceStatusFailed); err != nil {
		return fmt.Errorf("update invoice status: %w", err)
	}
	log.Printf("Payment failed: %s for company %s", invoice.ID, invoice.CompanyID)
	return nil
}

// ==================== Upgrade & Downgrade ====================

func (s *subscriptionService) UpgradePlan(ctx context.Context, companyID string, req subscription.UpgradeRequest) (subscription.InvoiceResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return subscription.InvoiceResponse{}, err
	}

	// Get current subscription
	sub, err := s.subscriptionRepo.GetByCompanyID(ctx, companyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return subscription.InvoiceResponse{}, subscription.ErrSubscriptionNotFound
		}
		return subscription.InvoiceResponse{}, fmt.Errorf("get subscription: %w", err)
	}

	// Check subscription is active
	if sub.Status != subscription.StatusActive && sub.Status != subscription.StatusTrial {
		return subscription.InvoiceResponse{}, subscription.ErrInvalidSubscriptionState
	}

	// Get current and new plans
	currentPlan, err := s.planRepo.GetByID(ctx, sub.PlanID)
	if err != nil {
		return subscription.InvoiceResponse{}, fmt.Errorf("get current plan: %w", err)
	}

	newPlan, err := s.planRepo.GetByID(ctx, req.PlanID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return subscription.InvoiceResponse{}, subscription.ErrPlanNotFound
		}
		return subscription.InvoiceResponse{}, fmt.Errorf("get new plan: %w", err)
	}

	// Validate it's an upgrade (higher tier)
	if newPlan.TierLevel <= currentPlan.TierLevel {
		return subscription.InvoiceResponse{}, subscription.ErrNotAnUpgrade
	}

	// Validate seats
	seatCount := req.SeatCount
	if seatCount == 0 {
		seatCount = sub.MaxSeats // Keep current seat count
	}

	if exceedsPlanMaxSeats(newPlan, seatCount) {
		return subscription.InvoiceResponse{}, subscription.ErrSeatLimitExceeded
	}

	// Validate seat count >= active employees
	activeEmployees, err := s.employeeCounter.CountActiveByCompanyID(ctx, companyID)
	if err != nil {
		return subscription.InvoiceResponse{}, fmt.Errorf("count employees: %w", err)
	}
	if seatCount < activeEmployees {
		return subscription.InvoiceResponse{}, subscription.ErrSeatsBelowActive
	}

	// Create checkout for upgrade (no proration - full price new period)
	return s.Checkout(ctx, companyID, subscription.CheckoutRequest{
		PlanID:       req.PlanID,
		SeatCount:    seatCount,
		BillingCycle: sub.BillingCycle,
		PayerEmail:   req.PayerEmail,
	})
}

func (s *subscriptionService) DowngradePlan(ctx context.Context, companyID string, req subscription.DowngradeRequest) error {
	// Validate request
	if err := req.Validate(); err != nil {
		return err
	}

	// Get current subscription
	sub, err := s.subscriptionRepo.GetByCompanyID(ctx, companyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return subscription.ErrSubscriptionNotFound
		}
		return fmt.Errorf("get subscription: %w", err)
	}

	// Check subscription is active
	if sub.Status != subscription.StatusActive {
		return subscription.ErrInvalidSubscriptionState
	}

	// Get current and new plans
	currentPlan, err := s.planRepo.GetByID(ctx, sub.PlanID)
	if err != nil {
		return fmt.Errorf("get current plan: %w", err)
	}

	newPlan, err := s.planRepo.GetByID(ctx, req.PlanID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return subscription.ErrPlanNotFound
		}
		return fmt.Errorf("get new plan: %w", err)
	}

	// Validate it's a downgrade (lower tier)
	if newPlan.TierLevel >= currentPlan.TierLevel {
		return subscription.ErrNotADowngrade
	}

	// Validate seat count will accommodate current employees
	activeEmployees, err := s.employeeCounter.CountActiveByCompanyID(ctx, companyID)
	if err != nil {
		return fmt.Errorf("count employees: %w", err)
	}
	// If plan has a max seats limit, check if it can accommodate active employees
	if newPlan.MaxSeats != nil && *newPlan.MaxSeats < activeEmployees {
		return subscription.ErrSeatLimitExceeded
	}

	// Set pending plan - will be applied at next renewal
	if err := s.subscriptionRepo.SetPendingPlan(ctx, sub.ID, &req.PlanID); err != nil {
		return fmt.Errorf("set pending plan: %w", err)
	}

	log.Printf("Downgrade scheduled: Company %s, from %s to %s (effective at period end)", companyID, currentPlan.Name, newPlan.Name)
	return nil
}

func (s *subscriptionService) CancelDowngrade(ctx context.Context, companyID string) error {
	// Get subscription
	sub, err := s.subscriptionRepo.GetByCompanyID(ctx, companyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return subscription.ErrSubscriptionNotFound
		}
		return fmt.Errorf("get subscription: %w", err)
	}

	// Clear pending plan
	if err := s.subscriptionRepo.SetPendingPlan(ctx, sub.ID, nil); err != nil {
		return fmt.Errorf("clear pending plan: %w", err)
	}

	return nil
}

// ==================== Invoice Operations ====================

func (s *subscriptionService) GetInvoices(ctx context.Context, companyID string) ([]subscription.InvoiceResponse, error) {
	invoices, err := s.invoiceRepo.ListByCompanyID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("list invoices: %w", err)
	}

	responses := make([]subscription.InvoiceResponse, len(invoices))
	for i, inv := range invoices {
		responses[i] = toInvoiceResponse(inv)
	}
	return responses, nil
}

func (s *subscriptionService) GetInvoiceByID(ctx context.Context, companyID, invoiceID string) (subscription.InvoiceResponse, error) {
	invoice, err := s.invoiceRepo.GetByID(ctx, invoiceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return subscription.InvoiceResponse{}, subscription.ErrInvoiceNotFound
		}
		return subscription.InvoiceResponse{}, fmt.Errorf("get invoice: %w", err)
	}

	// Verify company owns this invoice
	if invoice.CompanyID != companyID {
		return subscription.InvoiceResponse{}, subscription.ErrInvoiceNotFound
	}

	return toInvoiceResponse(invoice), nil
}

// ==================== Cron Jobs ====================

func (s *subscriptionService) ProcessExpiredTrials(ctx context.Context) error {
	now := time.Now()

	// Find trial subscriptions past their end date
	count, err := s.subscriptionRepo.UpdateExpiredToStatus(
		ctx,
		now,
		[]subscription.SubscriptionStatus{subscription.StatusTrial},
		subscription.StatusExpired,
	)
	if err != nil {
		return fmt.Errorf("update expired trials: %w", err)
	}

	if count > 0 {
		log.Printf("Cron: Expired %d trial subscriptions", count)
	}

	// Find cancelled subscriptions past their period end
	cancelledCount, err := s.subscriptionRepo.UpdateExpiredToStatus(
		ctx,
		now,
		[]subscription.SubscriptionStatus{subscription.StatusCancelled},
		subscription.StatusExpired,
	)
	if err != nil {
		return fmt.Errorf("update expired cancelled subscriptions: %w", err)
	}

	if cancelledCount > 0 {
		log.Printf("Cron: Expired %d cancelled subscriptions", cancelledCount)
	}

	return nil
}

func (s *subscriptionService) ProcessPastDueSubscriptions(ctx context.Context) error {
	now := time.Now()
	graceCutoff := now.AddDate(0, 0, -GracePeriodDays)

	// Find active subscriptions past their period end (enter grace period)
	subs, err := s.subscriptionRepo.ListExpiring(ctx, now)
	if err != nil {
		return fmt.Errorf("list expiring subscriptions: %w", err)
	}

	for _, sub := range subs {
		if sub.Status == subscription.StatusActive && sub.CurrentPeriodEnd.Before(now) {
			// Move to past_due (grace period)
			if err := s.subscriptionRepo.UpdateStatus(ctx, sub.ID, subscription.StatusPastDue); err != nil {
				log.Printf("Cron: Failed to set past_due for subscription %s: %v", sub.ID, err)
				continue
			}
			log.Printf("Cron: Subscription %s entered grace period", sub.ID)
		}
	}

	// Find past_due subscriptions past grace period -> expired
	count, err := s.subscriptionRepo.UpdateExpiredToStatus(
		ctx,
		graceCutoff,
		[]subscription.SubscriptionStatus{subscription.StatusPastDue},
		subscription.StatusExpired,
	)
	if err != nil {
		return fmt.Errorf("update expired past_due: %w", err)
	}

	if count > 0 {
		log.Printf("Cron: Expired %d past_due subscriptions after grace period", count)
	}

	return nil
}

func (s *subscriptionService) ExpireStaleInvoices(ctx context.Context) error {
	// Expire invoices older than configured expiry time (InvoiceExpiry is in hours)
	cutoff := time.Now().Add(-time.Duration(s.cfg.Xendit.InvoiceExpiry) * time.Hour)

	count, err := s.invoiceRepo.ExpireStaleInvoices(ctx, cutoff)
	if err != nil {
		return fmt.Errorf("expire stale invoices: %w", err)
	}

	if count > 0 {
		log.Printf("Cron: Expired %d stale invoices", count)
	}
	return nil
}

// ==================== Feature Checks ====================

func (s *subscriptionService) HasFeature(ctx context.Context, companyID string, featureCode string) (bool, error) {
	sub, err := s.subscriptionRepo.GetByCompanyIDWithFeatures(ctx, companyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("get subscription: %w", err)
	}

	return sub.HasFeature(featureCode), nil
}

func (s *subscriptionService) CanAddEmployee(ctx context.Context, companyID string) (bool, error) {
	sub, err := s.subscriptionRepo.GetByCompanyID(ctx, companyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("get subscription: %w", err)
	}

	// Check subscription is active
	if !sub.IsActive() {
		return false, nil
	}

	// Count current employees
	count, err := s.employeeCounter.CountActiveByCompanyID(ctx, companyID)
	if err != nil {
		return false, fmt.Errorf("count employees: %w", err)
	}

	return sub.CanAddEmployee(count), nil
}

// CancelSubscription cancels the subscription (access until period end)
// Voids all pending invoices and prevents future billing
func (s *subscriptionService) CancelSubscription(ctx context.Context, companyID string, req subscription.CancelRequest) error {
	// Get subscription
	sub, err := s.subscriptionRepo.GetByCompanyID(ctx, companyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return subscription.ErrSubscriptionNotFound
		}
		return fmt.Errorf("get subscription: %w", err)
	}

	// Can only cancel active or trial subscriptions
	if sub.Status != subscription.StatusActive && sub.Status != subscription.StatusTrial {
		return subscription.ErrInvalidSubscriptionState
	}

	// Use database transaction to ensure consistency
	var expiredCount int
	err = postgresql.WithTransaction(ctx, s.db, func(tx pgx.Tx) error {
		txCtx := context.WithValue(ctx, "tx", tx)

		// 1. Get all invoices for this subscription and filter pending ones
		allInvoices, err := s.invoiceRepo.ListBySubscriptionID(txCtx, sub.ID)
		if err != nil {
			return fmt.Errorf("list invoices: %w", err)
		}

		// Filter to get only pending invoices
		var pendingInvoices []subscription.Invoice
		for _, inv := range allInvoices {
			if inv.Status == subscription.InvoiceStatusPending {
				pendingInvoices = append(pendingInvoices, inv)
			}
		}

		// 2. Void pending invoices in Xendit and update DB status
		for _, inv := range pendingInvoices {
			// Expire invoice in Xendit if xendit_invoice_id exists
			if inv.XenditInvoiceID != nil && *inv.XenditInvoiceID != "" {
				_, err := s.xenditClient.ExpireInvoice(*inv.XenditInvoiceID)
				if err != nil {
					// Log error but don't fail the cancellation
					// The invoice will still be marked expired in our DB
					log.Printf("Warning: Failed to expire Xendit invoice %s: %v", *inv.XenditInvoiceID, err)
				} else {
					log.Printf("Expired Xendit invoice %s for cancelled subscription", *inv.XenditInvoiceID)
				}
			}

			// Update invoice status to expired in DB
			if err := s.invoiceRepo.UpdateStatus(txCtx, inv.ID, subscription.InvoiceStatusExpired); err != nil {
				return fmt.Errorf("expire invoice %s: %w", inv.ID, err)
			}
			expiredCount++
		}

		// 3. Update subscription status to cancelled
		if err := s.subscriptionRepo.UpdateStatus(txCtx, sub.ID, subscription.StatusCancelled); err != nil {
			return fmt.Errorf("update subscription status: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("cancel subscription: %w", err)
	}

	log.Printf("Subscription cancelled: Company %s, expired %d pending invoices, access until %v",
		companyID, expiredCount, sub.CurrentPeriodEnd)
	return nil
}

// CancelPendingInvoice cancels a pending invoice
func (s *subscriptionService) CancelPendingInvoice(ctx context.Context, companyID string, invoiceID string) error {
	// Get subscription to verify ownership
	sub, err := s.subscriptionRepo.GetByCompanyID(ctx, companyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return subscription.ErrSubscriptionNotFound
		}
		return fmt.Errorf("get subscription: %w", err)
	}

	// Get invoice
	invoice, err := s.invoiceRepo.GetByID(ctx, invoiceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return subscription.ErrInvoiceNotFound
		}
		return fmt.Errorf("get invoice: %w", err)
	}

	// Validate invoice belongs to this company's subscription
	if invoice.SubscriptionID != sub.ID {
		return subscription.ErrInvoiceNotFound // Don't expose existence of other invoices
	}

	// Validate invoice is pending
	if invoice.Status != subscription.InvoiceStatusPending {
		return subscription.ErrInvoiceNotPending
	}

	// Use transaction for consistency
	err = postgresql.WithTransaction(ctx, s.db, func(tx pgx.Tx) error {
		txCtx := context.WithValue(ctx, "tx", tx)

		// Void invoice in Xendit if xendit_invoice_id exists
		if invoice.XenditInvoiceID != nil && *invoice.XenditInvoiceID != "" {
			_, err := s.xenditClient.ExpireInvoice(*invoice.XenditInvoiceID)
			if err != nil {
				// Log warning but continue with DB update
				log.Printf("Warning: Failed to expire Xendit invoice %s: %v", *invoice.XenditInvoiceID, err)
			} else {
				log.Printf("Expired Xendit invoice %s upon user cancellation", *invoice.XenditInvoiceID)
			}
		}

		// Update invoice status to expired
		if err := s.invoiceRepo.UpdateStatus(txCtx, invoice.ID, subscription.InvoiceStatusExpired); err != nil {
			return fmt.Errorf("update invoice status: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("cancel invoice: %w", err)
	}

	log.Printf("Cancelled pending invoice %s for company %s", invoiceID, companyID)
	return nil
}

// ChangeSeats changes the number of seats
func (s *subscriptionService) ChangeSeats(ctx context.Context, companyID string, req subscription.ChangeSeatRequest) (subscription.ChangeSeatResponse, error) {
	if err := req.Validate(); err != nil {
		return subscription.ChangeSeatResponse{}, err
	}

	// Get subscription with features
	sub, err := s.subscriptionRepo.GetByCompanyIDWithFeatures(ctx, companyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return subscription.ChangeSeatResponse{}, subscription.ErrSubscriptionNotFound
		}
		return subscription.ChangeSeatResponse{}, fmt.Errorf("get subscription: %w", err)
	}

	// Validate not same as current seats
	if req.SeatCount == sub.MaxSeats {
		return subscription.ChangeSeatResponse{}, subscription.ErrSameAsCurrentSeats
	}

	// Validate seat count > 0
	if req.SeatCount < 1 {
		return subscription.ChangeSeatResponse{}, subscription.ErrInvalidSeatCount
	}

	// Check for pending invoices - block any seat changes if pending payment exists
	pendingCount, err := s.invoiceRepo.CountPendingInvoicesBySubscription(ctx, sub.ID)
	if err != nil {
		return subscription.ChangeSeatResponse{}, fmt.Errorf("count pending invoices: %w", err)
	}
	if pendingCount > 0 {
		return subscription.ChangeSeatResponse{}, subscription.ErrPendingInvoiceExists
	}

	// Get current plan
	plan, err := s.planRepo.GetByID(ctx, sub.PlanID)
	if err != nil {
		return subscription.ChangeSeatResponse{}, fmt.Errorf("get plan: %w", err)
	}

	// Check plan limits
	if exceedsPlanMaxSeats(plan, req.SeatCount) {
		return subscription.ChangeSeatResponse{}, subscription.ErrSeatLimitExceeded
	}

	now := time.Now()

	// UPSELL: Adding seats (prorated, immediate after payment)
	if req.SeatCount > sub.MaxSeats {
		// Block upsell during grace period (past_due status)
		if sub.Status == subscription.StatusPastDue {
			return subscription.ChangeSeatResponse{}, subscription.ErrCannotUpgradeDuringGracePeriod
		}

		// Calculate prorated amount
		daysRemaining := sub.CurrentPeriodEnd.Sub(now).Hours() / 24
		var totalDays float64
		if sub.BillingCycle == subscription.BillingCycleYearly {
			totalDays = 365
		} else {
			totalDays = 30
		}

		seatDifference := req.SeatCount - sub.MaxSeats
		proratedAmount := plan.PricePerSeat.
			Mul(decimal.NewFromInt(int64(seatDifference))).
			Mul(decimal.NewFromFloat(daysRemaining / totalDays))

		// Calculate invoice expiry matching subscription remaining days (minimum 24 hours)
		expiryHours := daysRemaining * 24
		if expiryHours < 24 {
			expiryHours = 24
		}
		invoiceExpirySecs := int(expiryHours * 3600)

		description := fmt.Sprintf("Additional Seats (Prorated) - %s Plan: %d → %d seats", plan.Name, sub.MaxSeats, req.SeatCount)

		// Create invoice with is_prorated=true
		invoice := subscription.Invoice{
			CompanyID:            companyID,
			SubscriptionID:       sub.ID,
			Amount:               proratedAmount,
			IsProrated:           true,
			PlanSnapshotName:     plan.Name,
			PricePerSeatSnapshot: plan.PricePerSeat,
			SeatCountSnapshot:    req.SeatCount, // New seat count
			BillingCycleSnapshot: sub.BillingCycle,
			PeriodStart:          sub.CurrentPeriodStart,
			PeriodEnd:            sub.CurrentPeriodEnd,
			Status:               subscription.InvoiceStatusPending,
			Description:          &description,
			IssueDate:            now,
		}

		var createdInvoice subscription.Invoice
		var xenditResp *xendit.InvoiceResponse

		// Use transaction for multi-step operation
		err = postgresql.WithTransaction(ctx, s.db, func(tx pgx.Tx) error {
			txCtx := context.WithValue(ctx, "tx", tx)

			// Get payer email from company
			// Note: You may need to add this to the company domain if not already present
			payerEmail := fmt.Sprintf("billing+%s@company.com", companyID) // Placeholder

			// Create Xendit invoice
			xenditReq := xendit.CreateInvoiceRequest{
				ExternalID:         fmt.Sprintf("seat-up-%s-%d", sub.ID, now.Unix()),
				Amount:             proratedAmount,
				PayerEmail:         payerEmail,
				Description:        description,
				Currency:           "IDR",
				InvoiceDuration:    invoiceExpirySecs,
				SuccessRedirectURL: s.cfg.Xendit.SuccessRedirect,
				FailureRedirectURL: s.cfg.Xendit.FailureRedirect,
			}

			xResp, err := s.xenditClient.CreateInvoice(xenditReq)
			if err != nil {
				return fmt.Errorf("create xendit invoice: %w", err)
			}
			xenditResp = xResp

			// Set Xendit details
			invoice.XenditInvoiceID = &xenditResp.ID
			invoice.XenditInvoiceURL = &xenditResp.InvoiceURL
			invoice.XenditExpiryDate = &xenditResp.ExpiryDate

			// Save invoice to database
			created, err := s.invoiceRepo.Create(txCtx, invoice)
			if err != nil {
				// Try to expire the Xendit invoice on failure
				_, _ = s.xenditClient.ExpireInvoice(*invoice.XenditInvoiceID)
				return fmt.Errorf("create invoice: %w", err)
			}
			createdInvoice = created

			// Clear pending_max_seats if exists (immediate action wins)
			if sub.PendingMaxSeats != nil {
				if err := s.subscriptionRepo.SetPendingMaxSeats(txCtx, sub.ID, nil); err != nil {
					return fmt.Errorf("clear pending seats: %w", err)
				}
			}

			return nil
		})

		if err != nil {
			return subscription.ChangeSeatResponse{}, fmt.Errorf("upsell transaction: %w", err)
		}

		// Return response with invoice for payment
		invoiceResp := createdInvoice.ToResponse()
		return subscription.ChangeSeatResponse{
			Invoice:   &invoiceResp,
			Message:   fmt.Sprintf("Payment required to add %d seats. Seats will be applied after payment.", seatDifference),
			IsPending: false,
		}, nil
	}

	// DOWNSELL: Reducing seats (scheduled for next period, no payment)
	if req.SeatCount < sub.MaxSeats {
		// Validate: new seat count must accommodate active employees
		activeEmployees, err := s.employeeCounter.CountActiveByCompanyID(ctx, companyID)
		if err != nil {
			return subscription.ChangeSeatResponse{}, fmt.Errorf("count employees: %w", err)
		}
		if req.SeatCount < activeEmployees {
			return subscription.ChangeSeatResponse{}, subscription.ErrSeatsBelowActiveEmployees
		}

		// Set pending_max_seats for application at next renewal
		pendingSeats := req.SeatCount
		if err := s.subscriptionRepo.SetPendingMaxSeats(ctx, sub.ID, &pendingSeats); err != nil {
			return subscription.ChangeSeatResponse{}, fmt.Errorf("set pending seats: %w", err)
		}

		periodEndStr := sub.CurrentPeriodEnd.Format("2006-01-02")
		return subscription.ChangeSeatResponse{
			Invoice:         nil,
			Message:         fmt.Sprintf("Seat count will be reduced from %d to %d at next renewal on %s. No charge.", sub.MaxSeats, req.SeatCount, periodEndStr),
			IsPending:       true,
			PendingMaxSeats: &pendingSeats,
		}, nil
	}

	// Should never reach here
	return subscription.ChangeSeatResponse{}, fmt.Errorf("unexpected state")
}

// UpdateExpiredSubscriptions updates subscription statuses based on period end
func (s *subscriptionService) UpdateExpiredSubscriptions(ctx context.Context) error {
	// Process expired trials
	if err := s.ProcessExpiredTrials(ctx); err != nil {
		log.Printf("Error processing expired trials: %v", err)
	}
	// Process past due subscriptions
	if err := s.ProcessPastDueSubscriptions(ctx); err != nil {
		log.Printf("Error processing past due subscriptions: %v", err)
	}
	return nil
}

// CleanupStaleInvoices marks old pending invoices as expired
func (s *subscriptionService) CleanupStaleInvoices(ctx context.Context) error {
	return s.ExpireStaleInvoices(ctx)
}

// ApplyPendingDowngrades applies pending plan downgrades for subscriptions at period end
func (s *subscriptionService) ApplyPendingDowngrades(ctx context.Context) error {
	// Handle pending seat downgrades
	seatSubs, err := s.subscriptionRepo.ListSubscriptionsWithPendingSeats(ctx)
	if err != nil {
		log.Printf("Cron: Failed to list pending seat changes: %v", err)
	} else if len(seatSubs) > 0 {
		applied := 0
		for _, sub := range seatSubs {
			oldSeats := sub.MaxSeats
			newSeats := *sub.PendingMaxSeats

			if err := s.subscriptionRepo.ApplyPendingMaxSeats(ctx, sub.ID); err != nil {
				log.Printf("Cron: Failed to apply pending seat downgrade for subscription %s: %v", sub.ID, err)
				continue
			}

			log.Printf("Cron: Applied pending seat downgrade for subscription %s from %d to %d seats",
				sub.ID, oldSeats, newSeats)
			applied++
		}
		if applied > 0 {
			log.Printf("Cron: Applied %d pending seat downgrades", applied)
		}
	}

	// Handle pending plan downgrades
	subs, err := s.subscriptionRepo.ListWithPendingDowngrade(ctx)
	if err != nil {
		return fmt.Errorf("list pending downgrades: %w", err)
	}

	for _, sub := range subs {
		// Only apply if period has ended
		if time.Now().Before(sub.CurrentPeriodEnd) {
			continue
		}

		if sub.PendingPlanID == nil {
			continue
		}

		// Apply the pending plan
		if err := s.subscriptionRepo.ApplyPendingPlan(ctx, sub.ID); err != nil {
			log.Printf("Failed to apply pending downgrade for subscription %s: %v", sub.ID, err)
			continue
		}

		log.Printf("Applied pending downgrade for subscription %s to plan %s", sub.ID, *sub.PendingPlanID)
	}

	return nil
}

// GetSubscriptionFeatures retrieves all features for a company's subscription
func (s *subscriptionService) GetSubscriptionFeatures(ctx context.Context, companyID string) ([]string, error) {
	sub, err := s.subscriptionRepo.GetByCompanyIDWithFeatures(ctx, companyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, subscription.ErrSubscriptionNotFound
		}
		return nil, fmt.Errorf("get subscription: %w", err)
	}

	features := make([]string, len(sub.Features))
	for i, f := range sub.Features {
		features[i] = f.Code
	}

	return features, nil
}

// ==================== Response Mappers ====================

func toPlanResponse(plan subscription.Plan) subscription.PlanResponse {
	features := make([]subscription.FeatureResponse, len(plan.Features))
	for i, f := range plan.Features {
		features[i] = toFeatureResponse(f)
	}

	return subscription.PlanResponse{
		ID:           plan.ID,
		Name:         plan.Name,
		PricePerSeat: plan.PricePerSeat,
		TierLevel:    plan.TierLevel,
		MaxSeats:     plan.MaxSeats,
		Features:     features,
	}
}

func toFeatureResponse(feature subscription.Feature) subscription.FeatureResponse {
	return subscription.FeatureResponse{
		Code:        feature.Code,
		Name:        feature.Name,
		Description: feature.Description,
	}
}

func toSubscriptionResponse(sub subscription.Subscription, usedSeats int) subscription.SubscriptionResponse {
	// Convert dates to RFC3339 strings
	periodStart := sub.CurrentPeriodStart.Format(time.RFC3339)
	periodEnd := sub.CurrentPeriodEnd.Format(time.RFC3339)

	var trialEndsAt *string
	if sub.TrialEndsAt != nil {
		s := sub.TrialEndsAt.Format(time.RFC3339)
		trialEndsAt = &s
	}

	// Extract feature codes
	featureCodes := make([]string, len(sub.Features))
	for i, f := range sub.Features {
		featureCodes[i] = f.Code
	}

	resp := subscription.SubscriptionResponse{
		ID:                 sub.ID,
		Status:             sub.Status,
		MaxSeats:           sub.MaxSeats,
		UsedSeats:          usedSeats,
		CurrentPeriodStart: periodStart,
		CurrentPeriodEnd:   periodEnd,
		TrialEndsAt:        trialEndsAt,
		BillingCycle:       sub.BillingCycle,
		AutoRenew:          sub.AutoRenew,
		Features:           featureCodes,
	}

	if sub.Plan != nil {
		resp.Plan = toPlanResponse(*sub.Plan)
	}

	return resp
}

func toInvoiceResponse(inv subscription.Invoice) subscription.InvoiceResponse {
	// Convert dates to RFC3339 strings
	periodStart := inv.PeriodStart.Format(time.RFC3339)
	periodEnd := inv.PeriodEnd.Format(time.RFC3339)
	issueDate := inv.IssueDate.Format(time.RFC3339)

	var paidAt *string
	if inv.PaidAt != nil {
		s := inv.PaidAt.Format(time.RFC3339)
		paidAt = &s
	}

	var expiryDate *string
	if inv.XenditExpiryDate != nil {
		s := inv.XenditExpiryDate.Format(time.RFC3339)
		expiryDate = &s
	}

	return subscription.InvoiceResponse{
		ID:             inv.ID,
		Amount:         inv.Amount,
		Status:         inv.Status,
		PlanName:       inv.PlanSnapshotName,
		SeatCount:      inv.SeatCountSnapshot,
		PricePerSeat:   inv.PricePerSeatSnapshot,
		BillingCycle:   inv.BillingCycleSnapshot,
		PeriodStart:    periodStart,
		PeriodEnd:      periodEnd,
		IssueDate:      issueDate,
		PaymentURL:     inv.XenditInvoiceURL,
		ExpiryDate:     expiryDate,
		PaidAt:         paidAt,
		PaymentMethod:  inv.PaymentMethod,
		PaymentChannel: inv.PaymentChannel,
	}
}
