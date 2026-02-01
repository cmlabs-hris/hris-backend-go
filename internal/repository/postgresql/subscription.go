package postgresql

import (
	"context"
	"fmt"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/subscription"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// ==================== Feature Repository ====================

type featureRepository struct {
	db *database.DB
}

func NewFeatureRepository(db *database.DB) subscription.FeatureRepository {
	return &featureRepository{db: db}
}

func (r *featureRepository) GetByCode(ctx context.Context, code string) (subscription.Feature, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, code, name, description, created_at
		FROM features
		WHERE code = $1
	`

	var f subscription.Feature
	var desc *string
	err := q.QueryRow(ctx, query, code).Scan(
		&f.ID, &f.Code, &f.Name, &desc, &f.CreatedAt,
	)
	if err != nil {
		return subscription.Feature{}, err
	}
	f.Description = desc
	return f, nil
}

func (r *featureRepository) List(ctx context.Context) ([]subscription.Feature, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, code, name, description, created_at
		FROM features
		ORDER BY name
	`

	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var features []subscription.Feature
	for rows.Next() {
		var f subscription.Feature
		var desc *string
		if err := rows.Scan(&f.ID, &f.Code, &f.Name, &desc, &f.CreatedAt); err != nil {
			return nil, err
		}
		f.Description = desc
		features = append(features, f)
	}
	return features, nil
}

// ==================== Plan Repository ====================

type planRepository struct {
	db *database.DB
}

func NewPlanRepository(db *database.DB) subscription.PlanRepository {
	return &planRepository{db: db}
}

func (r *planRepository) GetByID(ctx context.Context, id string) (subscription.Plan, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, name, price_per_seat, tier_level, max_seats, is_active, created_at, updated_at
		FROM subscription_plans
		WHERE id = $1
	`

	var p subscription.Plan
	err := q.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.PricePerSeat, &p.TierLevel, &p.MaxSeats, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return subscription.Plan{}, err
	}

	// Get features for this plan
	features, err := r.GetFeaturesByPlanID(ctx, id)
	if err != nil {
		return subscription.Plan{}, err
	}
	p.Features = features

	return p, nil
}

func (r *planRepository) GetByName(ctx context.Context, name string) (subscription.Plan, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, name, price_per_seat, tier_level, max_seats, is_active, created_at, updated_at
		FROM subscription_plans
		WHERE name = $1
	`

	var p subscription.Plan
	err := q.QueryRow(ctx, query, name).Scan(
		&p.ID, &p.Name, &p.PricePerSeat, &p.TierLevel, &p.MaxSeats, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return subscription.Plan{}, err
	}

	// Get features for this plan
	features, err := r.GetFeaturesByPlanID(ctx, p.ID)
	if err != nil {
		return subscription.Plan{}, err
	}
	p.Features = features

	return p, nil
}

func (r *planRepository) ListActive(ctx context.Context) ([]subscription.Plan, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, name, price_per_seat, tier_level, max_seats, is_active, created_at, updated_at
		FROM subscription_plans
		WHERE is_active = true
		ORDER BY tier_level
	`

	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []subscription.Plan
	for rows.Next() {
		var p subscription.Plan
		if err := rows.Scan(
			&p.ID, &p.Name, &p.PricePerSeat, &p.TierLevel, &p.MaxSeats, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}

		// Get features for each plan
		features, err := r.GetFeaturesByPlanID(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		p.Features = features

		plans = append(plans, p)
	}
	return plans, nil
}

func (r *planRepository) GetFeaturesByPlanID(ctx context.Context, planID string) ([]subscription.Feature, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT f.id, f.code, f.name, f.description, f.created_at
		FROM features f
		JOIN plan_features pf ON f.id = pf.feature_id
		WHERE pf.plan_id = $1 AND pf.is_active = true
		ORDER BY f.name
	`

	rows, err := q.Query(ctx, query, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var features []subscription.Feature
	for rows.Next() {
		var f subscription.Feature
		var desc *string
		if err := rows.Scan(&f.ID, &f.Code, &f.Name, &desc, &f.CreatedAt); err != nil {
			return nil, err
		}
		f.Description = desc
		features = append(features, f)
	}
	return features, nil
}

// ==================== Subscription Repository ====================

type subscriptionRepository struct {
	db *database.DB
}

func NewSubscriptionRepository(db *database.DB) subscription.SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

func (r *subscriptionRepository) GetByID(ctx context.Context, id string) (subscription.Subscription, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, company_id, plan_id, status, max_seats, pending_max_seats, current_period_start, current_period_end,
			   trial_ends_at, pending_plan_id, billing_cycle, auto_renew, created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`

	var s subscription.Subscription
	err := q.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.CompanyID, &s.PlanID, &s.Status, &s.MaxSeats, &s.PendingMaxSeats,
		&s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.TrialEndsAt,
		&s.PendingPlanID, &s.BillingCycle, &s.AutoRenew, &s.CreatedAt, &s.UpdatedAt,
	)
	return s, err
}

func (r *subscriptionRepository) GetByCompanyID(ctx context.Context, companyID string) (subscription.Subscription, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, company_id, plan_id, status, max_seats, pending_max_seats, current_period_start, current_period_end,
			   trial_ends_at, pending_plan_id, billing_cycle, auto_renew, created_at, updated_at
		FROM subscriptions
		WHERE company_id = $1
	`

	var s subscription.Subscription
	err := q.QueryRow(ctx, query, companyID).Scan(
		&s.ID, &s.CompanyID, &s.PlanID, &s.Status, &s.MaxSeats, &s.PendingMaxSeats,
		&s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.TrialEndsAt,
		&s.PendingPlanID, &s.BillingCycle, &s.AutoRenew, &s.CreatedAt, &s.UpdatedAt,
	)
	return s, err
}

func (r *subscriptionRepository) GetByCompanyIDWithFeatures(ctx context.Context, companyID string) (subscription.Subscription, error) {
	// Get subscription
	s, err := r.GetByCompanyID(ctx, companyID)
	if err != nil {
		return subscription.Subscription{}, err
	}

	q := GetQuerier(ctx, r.db)

	// Get plan
	planQuery := `
		SELECT id, name, price_per_seat, tier_level, max_seats, is_active, created_at, updated_at
		FROM subscription_plans
		WHERE id = $1
	`
	var plan subscription.Plan
	err = q.QueryRow(ctx, planQuery, s.PlanID).Scan(
		&plan.ID, &plan.Name, &plan.PricePerSeat, &plan.TierLevel, &plan.MaxSeats, &plan.IsActive, &plan.CreatedAt, &plan.UpdatedAt,
	)
	if err != nil {
		return subscription.Subscription{}, err
	}
	s.Plan = &plan

	// Get plan features from plan_features table
	featuresQuery := `
		SELECT f.id, f.code, f.name, f.description, f.created_at
		FROM features f
		JOIN plan_features pf ON f.id = pf.feature_id
		WHERE pf.plan_id = $1 AND pf.is_active = true
		ORDER BY f.name
	`
	rows, err := q.Query(ctx, featuresQuery, s.PlanID)
	if err != nil {
		return subscription.Subscription{}, err
	}
	defer rows.Close()

	var features []subscription.Feature
	for rows.Next() {
		var f subscription.Feature
		var desc *string
		if err := rows.Scan(&f.ID, &f.Code, &f.Name, &desc, &f.CreatedAt); err != nil {
			return subscription.Subscription{}, err
		}
		f.Description = desc
		features = append(features, f)
	}

	// Assign features to both plan and subscription
	// Plan features show what the plan includes
	// Subscription features show what's currently active for this subscription
	plan.Features = features
	s.Features = features

	return s, nil
}

func (r *subscriptionRepository) Create(ctx context.Context, s subscription.Subscription) (subscription.Subscription, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		INSERT INTO subscriptions (company_id, plan_id, status, max_seats, current_period_start, current_period_end,
								   trial_ends_at, pending_plan_id, billing_cycle, auto_renew)
		VALUES ($1, $2, $3::subscription_status, $4, $5, $6, $7, $8, $9::billing_cycle_enum, $10)
		RETURNING id, created_at, updated_at
	`

	err := q.QueryRow(ctx, query,
		s.CompanyID, s.PlanID, string(s.Status), s.MaxSeats, s.CurrentPeriodStart, s.CurrentPeriodEnd,
		s.TrialEndsAt, s.PendingPlanID, string(s.BillingCycle), s.AutoRenew,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)

	return s, err
}

func (r *subscriptionRepository) Update(ctx context.Context, s subscription.Subscription) error {
	q := GetQuerier(ctx, r.db)

	query := `
		UPDATE subscriptions
		SET plan_id = $2, status = $3::subscription_status, max_seats = $4, current_period_start = $5, current_period_end = $6,
			trial_ends_at = $7, pending_plan_id = $8, billing_cycle = $9::billing_cycle_enum, auto_renew = $10, updated_at = NOW()
		WHERE id = $1
	`

	_, err := q.Exec(ctx, query,
		s.ID, s.PlanID, string(s.Status), s.MaxSeats, s.CurrentPeriodStart, s.CurrentPeriodEnd,
		s.TrialEndsAt, s.PendingPlanID, string(s.BillingCycle), s.AutoRenew,
	)
	return err
}

func (r *subscriptionRepository) UpdateStatus(ctx context.Context, id string, status subscription.SubscriptionStatus) error {
	q := GetQuerier(ctx, r.db)

	query := `UPDATE subscriptions SET status = $2::subscription_status, updated_at = NOW() WHERE id = $1`
	_, err := q.Exec(ctx, query, id, string(status))
	return err
}

func (r *subscriptionRepository) UpdatePlan(ctx context.Context, id string, planID string, maxSeats int, periodEnd interface{}) error {
	q := GetQuerier(ctx, r.db)

	query := `
		UPDATE subscriptions
		SET plan_id = $2, max_seats = $3, current_period_end = $4, status = 'active', pending_plan_id = NULL, updated_at = NOW()
		WHERE id = $1
	`
	_, err := q.Exec(ctx, query, id, planID, maxSeats, periodEnd)
	return err
}

func (r *subscriptionRepository) SetPendingPlan(ctx context.Context, id string, pendingPlanID *string) error {
	q := GetQuerier(ctx, r.db)

	query := `UPDATE subscriptions SET pending_plan_id = $2, updated_at = NOW() WHERE id = $1`
	_, err := q.Exec(ctx, query, id, pendingPlanID)
	return err
}

func (r *subscriptionRepository) ListExpiring(ctx context.Context, before interface{}) ([]subscription.Subscription, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, company_id, plan_id, status, max_seats, pending_max_seats, current_period_start, current_period_end,
			   trial_ends_at, pending_plan_id, billing_cycle, auto_renew, created_at, updated_at
		FROM subscriptions
		WHERE current_period_end < $1 AND status IN ('trial', 'active')
		ORDER BY current_period_end
	`

	rows, err := q.Query(ctx, query, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []subscription.Subscription
	for rows.Next() {
		var s subscription.Subscription
		if err := rows.Scan(
			&s.ID, &s.CompanyID, &s.PlanID, &s.Status, &s.MaxSeats, &s.PendingMaxSeats,
			&s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.TrialEndsAt,
			&s.PendingPlanID, &s.BillingCycle, &s.AutoRenew, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, nil
}

func (r *subscriptionRepository) ListByStatus(ctx context.Context, status subscription.SubscriptionStatus) ([]subscription.Subscription, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, company_id, plan_id, status, max_seats, pending_max_seats, current_period_start, current_period_end,
			   trial_ends_at, pending_plan_id, billing_cycle, auto_renew, created_at, updated_at
		FROM subscriptions
		WHERE status = $1
		ORDER BY current_period_end
	`

	rows, err := q.Query(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []subscription.Subscription
	for rows.Next() {
		var s subscription.Subscription
		if err := rows.Scan(
			&s.ID, &s.CompanyID, &s.PlanID, &s.Status, &s.MaxSeats, &s.PendingMaxSeats,
			&s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.TrialEndsAt,
			&s.PendingPlanID, &s.BillingCycle, &s.AutoRenew, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, nil
}

func (r *subscriptionRepository) UpdateExpiredToStatus(ctx context.Context, cutoffTime interface{}, fromStatuses []subscription.SubscriptionStatus, toStatus subscription.SubscriptionStatus) (int64, error) {
	q := GetQuerier(ctx, r.db)

	// Convert []SubscriptionStatus to []string for pgx encoding
	fromStatusStrings := make([]string, len(fromStatuses))
	for i, s := range fromStatuses {
		fromStatusStrings[i] = string(s)
	}

	query := `
		UPDATE subscriptions
		SET status = $1, updated_at = NOW()
		WHERE current_period_end < $2 AND status = ANY($3::subscription_status[])
	`

	tag, err := q.Exec(ctx, query, string(toStatus), cutoffTime, fromStatusStrings)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *subscriptionRepository) UpdateMaxSeats(ctx context.Context, id string, maxSeats int) error {
	q := GetQuerier(ctx, r.db)

	query := `UPDATE subscriptions SET max_seats = $2, updated_at = NOW() WHERE id = $1`
	_, err := q.Exec(ctx, query, id, maxSeats)
	return err
}

func (r *subscriptionRepository) SetPendingMaxSeats(ctx context.Context, id string, pendingMaxSeats *int) error {
	q := GetQuerier(ctx, r.db)

	query := `UPDATE subscriptions SET pending_max_seats = $2, updated_at = NOW() WHERE id = $1`
	_, err := q.Exec(ctx, query, id, pendingMaxSeats)
	return err
}

func (r *subscriptionRepository) ApplyPendingMaxSeats(ctx context.Context, id string) error {
	q := GetQuerier(ctx, r.db)

	// Apply the pending seat count and clear pending_max_seats
	query := `
		UPDATE subscriptions 
		SET max_seats = pending_max_seats, pending_max_seats = NULL, updated_at = NOW() 
		WHERE id = $1 AND pending_max_seats IS NOT NULL
	`
	_, err := q.Exec(ctx, query, id)
	return err
}

func (r *subscriptionRepository) ListSubscriptionsWithPendingSeats(ctx context.Context) ([]subscription.Subscription, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, company_id, plan_id, status, max_seats, pending_max_seats, current_period_start, current_period_end,
			   trial_ends_at, pending_plan_id, billing_cycle, auto_renew, created_at, updated_at
		FROM subscriptions
		WHERE pending_max_seats IS NOT NULL 
		  AND current_period_end <= NOW()
		  AND status IN ('active', 'cancelled')
		ORDER BY current_period_end
	`

	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []subscription.Subscription
	for rows.Next() {
		var s subscription.Subscription
		if err := rows.Scan(
			&s.ID, &s.CompanyID, &s.PlanID, &s.Status, &s.MaxSeats, &s.PendingMaxSeats,
			&s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.TrialEndsAt,
			&s.PendingPlanID, &s.BillingCycle, &s.AutoRenew, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, nil
}

func (r *subscriptionRepository) ListWithPendingDowngrade(ctx context.Context) ([]subscription.Subscription, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, company_id, plan_id, status, max_seats, pending_max_seats, current_period_start, current_period_end,
			   trial_ends_at, pending_plan_id, billing_cycle, auto_renew, created_at, updated_at
		FROM subscriptions
		WHERE pending_plan_id IS NOT NULL
		ORDER BY current_period_end
	`

	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []subscription.Subscription
	for rows.Next() {
		var s subscription.Subscription
		if err := rows.Scan(
			&s.ID, &s.CompanyID, &s.PlanID, &s.Status, &s.MaxSeats, &s.PendingMaxSeats,
			&s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.TrialEndsAt,
			&s.PendingPlanID, &s.BillingCycle, &s.AutoRenew, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, nil
}

func (r *subscriptionRepository) ApplyPendingPlan(ctx context.Context, id string) error {
	q := GetQuerier(ctx, r.db)

	// Apply the pending plan and clear the pending_plan_id
	query := `
		UPDATE subscriptions 
		SET plan_id = pending_plan_id, pending_plan_id = NULL, updated_at = NOW() 
		WHERE id = $1 AND pending_plan_id IS NOT NULL
	`
	_, err := q.Exec(ctx, query, id)
	return err
}

// ==================== Invoice Repository ====================

type invoiceRepository struct {
	db *database.DB
}

func NewInvoiceRepository(db *database.DB) subscription.InvoiceRepository {
	return &invoiceRepository{db: db}
}

func (r *invoiceRepository) GetByID(ctx context.Context, id string) (subscription.Invoice, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, company_id, subscription_id, xendit_invoice_id, xendit_invoice_url, xendit_expiry_date,
			   amount, is_prorated, plan_snapshot_name, price_per_seat_snapshot, seat_count_snapshot, billing_cycle_snapshot,
			   period_start, period_end, status, issue_date, paid_at, payment_method, payment_channel,
			   description, notes, created_at, updated_at
		FROM invoices
		WHERE id = $1
	`

	var inv subscription.Invoice
	err := q.QueryRow(ctx, query, id).Scan(
		&inv.ID, &inv.CompanyID, &inv.SubscriptionID, &inv.XenditInvoiceID, &inv.XenditInvoiceURL, &inv.XenditExpiryDate,
		&inv.Amount, &inv.IsProrated, &inv.PlanSnapshotName, &inv.PricePerSeatSnapshot, &inv.SeatCountSnapshot, &inv.BillingCycleSnapshot,
		&inv.PeriodStart, &inv.PeriodEnd, &inv.Status, &inv.IssueDate, &inv.PaidAt, &inv.PaymentMethod, &inv.PaymentChannel,
		&inv.Description, &inv.Notes, &inv.CreatedAt, &inv.UpdatedAt,
	)
	return inv, err
}

func (r *invoiceRepository) GetByXenditID(ctx context.Context, xenditID string) (subscription.Invoice, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, company_id, subscription_id, xendit_invoice_id, xendit_invoice_url, xendit_expiry_date,
			   amount, is_prorated, plan_snapshot_name, price_per_seat_snapshot, seat_count_snapshot, billing_cycle_snapshot,
			   period_start, period_end, status, issue_date, paid_at, payment_method, payment_channel,
			   description, notes, created_at, updated_at
		FROM invoices
		WHERE xendit_invoice_id = $1
	`

	var inv subscription.Invoice
	err := q.QueryRow(ctx, query, xenditID).Scan(
		&inv.ID, &inv.CompanyID, &inv.SubscriptionID, &inv.XenditInvoiceID, &inv.XenditInvoiceURL, &inv.XenditExpiryDate,
		&inv.Amount, &inv.IsProrated, &inv.PlanSnapshotName, &inv.PricePerSeatSnapshot, &inv.SeatCountSnapshot, &inv.BillingCycleSnapshot,
		&inv.PeriodStart, &inv.PeriodEnd, &inv.Status, &inv.IssueDate, &inv.PaidAt, &inv.PaymentMethod, &inv.PaymentChannel,
		&inv.Description, &inv.Notes, &inv.CreatedAt, &inv.UpdatedAt,
	)
	return inv, err
}

func (r *invoiceRepository) Create(ctx context.Context, inv subscription.Invoice) (subscription.Invoice, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		INSERT INTO invoices (company_id, subscription_id, xendit_invoice_id, xendit_invoice_url, xendit_expiry_date,
						  amount, is_prorated, plan_snapshot_name, price_per_seat_snapshot, seat_count_snapshot, billing_cycle_snapshot,
						  period_start, period_end, status, description, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::billing_cycle_enum, $12, $13, $14::invoice_status, $15, $16)
		RETURNING id, issue_date, created_at, updated_at
	`

	err := q.QueryRow(ctx, query,
		inv.CompanyID, inv.SubscriptionID, inv.XenditInvoiceID, inv.XenditInvoiceURL, inv.XenditExpiryDate,
		inv.Amount, inv.IsProrated, inv.PlanSnapshotName, inv.PricePerSeatSnapshot, inv.SeatCountSnapshot, string(inv.BillingCycleSnapshot),
		inv.PeriodStart, inv.PeriodEnd, string(inv.Status), inv.Description, inv.Notes,
	).Scan(&inv.ID, &inv.IssueDate, &inv.CreatedAt, &inv.UpdatedAt)

	return inv, err
}

func (r *invoiceRepository) UpdateStatus(ctx context.Context, id string, status subscription.InvoiceStatus) error {
	q := GetQuerier(ctx, r.db)

	query := `UPDATE invoices SET status = $2::invoice_status, updated_at = NOW() WHERE id = $1`
	_, err := q.Exec(ctx, query, id, string(status))
	return err
}

func (r *invoiceRepository) UpdatePayment(ctx context.Context, id string, status subscription.InvoiceStatus, paidAt interface{}, method, channel string) error {
	q := GetQuerier(ctx, r.db)

	query := `
		UPDATE invoices
		SET status = $2::invoice_status, paid_at = $3, payment_method = $4, payment_channel = $5, updated_at = NOW()
		WHERE id = $1
	`
	_, err := q.Exec(ctx, query, id, string(status), paidAt, method, channel)
	return err
}

func (r *invoiceRepository) ListByCompanyID(ctx context.Context, companyID string) ([]subscription.Invoice, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, company_id, subscription_id, xendit_invoice_id, xendit_invoice_url, xendit_expiry_date,
			   amount, is_prorated, plan_snapshot_name, price_per_seat_snapshot, seat_count_snapshot, billing_cycle_snapshot,
			   period_start, period_end, status, issue_date, paid_at, payment_method, payment_channel,
			   description, notes, created_at, updated_at
		FROM invoices
		WHERE company_id = $1
		ORDER BY issue_date DESC
	`

	return r.scanInvoices(ctx, q, query, companyID)
}

func (r *invoiceRepository) ListBySubscriptionID(ctx context.Context, subscriptionID string) ([]subscription.Invoice, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, company_id, subscription_id, xendit_invoice_id, xendit_invoice_url, xendit_expiry_date,
			   amount, is_prorated, plan_snapshot_name, price_per_seat_snapshot, seat_count_snapshot, billing_cycle_snapshot,
			   period_start, period_end, status, issue_date, paid_at, payment_method, payment_channel,
			   description, notes, created_at, updated_at
		FROM invoices
		WHERE subscription_id = $1
		ORDER BY issue_date DESC
	`

	return r.scanInvoices(ctx, q, query, subscriptionID)
}

func (r *invoiceRepository) ListPending(ctx context.Context) ([]subscription.Invoice, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, company_id, subscription_id, xendit_invoice_id, xendit_invoice_url, xendit_expiry_date,
			   amount, is_prorated, plan_snapshot_name, price_per_seat_snapshot, seat_count_snapshot, billing_cycle_snapshot,
			   period_start, period_end, status, issue_date, paid_at, payment_method, payment_channel,
			   description, notes, created_at, updated_at
		FROM invoices
		WHERE status = 'pending'
		ORDER BY issue_date DESC
	`

	return r.scanInvoicesNoArg(ctx, q, query)
}

func (r *invoiceRepository) ListPendingOlderThan(ctx context.Context, olderThan interface{}) ([]subscription.Invoice, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, company_id, subscription_id, xendit_invoice_id, xendit_invoice_url, xendit_expiry_date,
			   amount, is_prorated, plan_snapshot_name, price_per_seat_snapshot, seat_count_snapshot, billing_cycle_snapshot,
			   period_start, period_end, status, issue_date, paid_at, payment_method, payment_channel,
			   description, notes, created_at, updated_at
		FROM invoices
		WHERE status = 'pending' AND created_at < $1
		ORDER BY issue_date DESC
	`

	return r.scanInvoices(ctx, q, query, olderThan)
}

func (r *invoiceRepository) HasPendingInvoice(ctx context.Context, companyID string) (bool, error) {
	q := GetQuerier(ctx, r.db)

	query := `SELECT EXISTS(SELECT 1 FROM invoices WHERE company_id = $1 AND status = 'pending')`
	var exists bool
	err := q.QueryRow(ctx, query, companyID).Scan(&exists)
	return exists, err
}

func (r *invoiceRepository) CountPendingInvoicesBySubscription(ctx context.Context, subscriptionID string) (int, error) {
	q := GetQuerier(ctx, r.db)

	query := `SELECT COUNT(*) FROM invoices WHERE subscription_id = $1 AND status = 'pending'`
	var count int
	err := q.QueryRow(ctx, query, subscriptionID).Scan(&count)
	return count, err
}

func (r *invoiceRepository) ExpireStaleInvoices(ctx context.Context, olderThan interface{}) (int64, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		UPDATE invoices
		SET status = 'expired', updated_at = NOW()
		WHERE status = 'pending' AND created_at < $1
	`

	tag, err := q.Exec(ctx, query, olderThan)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// Helper to scan invoice rows
func (r *invoiceRepository) scanInvoices(ctx context.Context, q database.Querier, query string, arg interface{}) ([]subscription.Invoice, error) {
	rows, err := q.Query(ctx, query, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.parseInvoiceRows(rows)
}

func (r *invoiceRepository) scanInvoicesNoArg(ctx context.Context, q database.Querier, query string) ([]subscription.Invoice, error) {
	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.parseInvoiceRows(rows)
}

func (r *invoiceRepository) parseInvoiceRows(rows pgx.Rows) ([]subscription.Invoice, error) {
	var invoices []subscription.Invoice
	for rows.Next() {
		var inv subscription.Invoice
		if err := rows.Scan(
			&inv.ID, &inv.CompanyID, &inv.SubscriptionID, &inv.XenditInvoiceID, &inv.XenditInvoiceURL, &inv.XenditExpiryDate,
			&inv.Amount, &inv.IsProrated, &inv.PlanSnapshotName, &inv.PricePerSeatSnapshot, &inv.SeatCountSnapshot, &inv.BillingCycleSnapshot,
			&inv.PeriodStart, &inv.PeriodEnd, &inv.Status, &inv.IssueDate, &inv.PaidAt, &inv.PaymentMethod, &inv.PaymentChannel,
			&inv.Description, &inv.Notes, &inv.CreatedAt, &inv.UpdatedAt,
		); err != nil {
			return nil, err
		}
		invoices = append(invoices, inv)
	}
	return invoices, nil
}

// ==================== Employee Counter (for seat validation) ====================

type employeeCounter struct {
	db *database.DB
}

func NewEmployeeCounter(db *database.DB) subscription.EmployeeCounter {
	return &employeeCounter{db: db}
}

func (r *employeeCounter) CountActiveByCompanyID(ctx context.Context, companyID string) (int, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT COUNT(*)
		FROM employees
		WHERE company_id = $1 AND status = 'active' AND deleted_at IS NULL
	`

	var count int
	err := q.QueryRow(ctx, query, companyID).Scan(&count)
	return count, err
}

// ==================== Helper Functions ====================

// CalculatePeriodEnd calculates the end date based on billing cycle
func CalculatePeriodEnd(start time.Time, cycle subscription.BillingCycle) time.Time {
	switch cycle {
	case subscription.BillingCycleYearly:
		return start.AddDate(1, 0, 0)
	default: // monthly
		return start.AddDate(0, 1, 0)
	}
}

// CalculateAmount calculates total amount based on price per seat and seat count
func CalculateAmount(pricePerSeat decimal.Decimal, seatCount int, cycle subscription.BillingCycle) decimal.Decimal {
	amount := pricePerSeat.Mul(decimal.NewFromInt(int64(seatCount)))
	if cycle == subscription.BillingCycleYearly {
		// 12 months with discount (e.g., 10 months = 2 months free)
		amount = amount.Mul(decimal.NewFromInt(10))
	}
	return amount
}

// FormatInvoiceDescription creates a description for invoice
func FormatInvoiceDescription(planName string, seatCount int, cycle subscription.BillingCycle) string {
	cycleStr := "Monthly"
	if cycle == subscription.BillingCycleYearly {
		cycleStr = "Yearly"
	}
	return fmt.Sprintf("HRIS %s Plan - %d seats (%s)", planName, seatCount, cycleStr)
}
