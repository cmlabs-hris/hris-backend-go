package cron

import (
	"context"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/subscription"
)

// SubscriptionJobs contains subscription-related cron jobs
type SubscriptionJobs struct {
	subscriptionService subscription.SubscriptionService
}

// NewSubscriptionJobs creates subscription cron jobs
func NewSubscriptionJobs(subscriptionService subscription.SubscriptionService) *SubscriptionJobs {
	return &SubscriptionJobs{
		subscriptionService: subscriptionService,
	}
}

// RegisterJobs registers all subscription-related cron jobs
func (j *SubscriptionJobs) RegisterJobs(scheduler *Scheduler) {
	// Update expired subscriptions every hour
	scheduler.AddJob(
		"update_expired_subscriptions",
		1*time.Hour,
		j.UpdateExpiredSubscriptions,
	)

	// Cleanup stale invoices every 6 hours
	scheduler.AddJob(
		"cleanup_stale_invoices",
		6*time.Hour,
		j.CleanupStaleInvoices,
	)

	// Apply pending downgrades every day at midnight (check every hour)
	scheduler.AddJob(
		"apply_pending_downgrades",
		1*time.Hour,
		j.ApplyPendingDowngrades,
	)
}

// UpdateExpiredSubscriptions updates subscription statuses
// trial/active -> past_due (when period ended)
// past_due -> expired (after 7 days grace period)
func (j *SubscriptionJobs) UpdateExpiredSubscriptions(ctx context.Context) error {
	return j.subscriptionService.UpdateExpiredSubscriptions(ctx)
}

// CleanupStaleInvoices marks old pending invoices as expired
func (j *SubscriptionJobs) CleanupStaleInvoices(ctx context.Context) error {
	return j.subscriptionService.CleanupStaleInvoices(ctx)
}

// ApplyPendingDowngrades applies scheduled downgrades at period end
func (j *SubscriptionJobs) ApplyPendingDowngrades(ctx context.Context) error {
	return j.subscriptionService.ApplyPendingDowngrades(ctx)
}
