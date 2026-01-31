package middleware

import (
	"net/http"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/subscription"
	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
	"github.com/go-chi/jwtauth/v5"
)

// SubscriptionMiddleware provides middleware functions for subscription checks
type SubscriptionMiddleware struct {
	subscriptionService subscription.SubscriptionService
}

// NewSubscriptionMiddleware creates a new subscription middleware
func NewSubscriptionMiddleware(subscriptionService subscription.SubscriptionService) *SubscriptionMiddleware {
	return &SubscriptionMiddleware{
		subscriptionService: subscriptionService,
	}
}

// RequireActiveSubscription checks if the company has an active subscription
// This performs a double check: JWT claims + database verification
// The JWT check provides early rejection for expired subscriptions (stale JWT protection)
func (m *SubscriptionMiddleware) RequireActiveSubscription(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _, err := jwtauth.FromContext(r.Context())
		if err != nil || token == nil {
			response.Unauthorized(w, "unauthorized")
			return
		}

		claims, err := token.AsMap(r.Context())
		if err != nil {
			response.Unauthorized(w, "invalid token claims")
			return
		}

		// Get company_id from JWT claims
		companyID, ok := claims["company_id"].(string)
		if !ok || companyID == "" {
			response.Forbidden(w, "no company associated with this user")
			return
		}

		// JWT Check 1: subscription_expires_at - Early rejection for stale JWTs
		// This protects against scenarios where cron hasn't updated status yet
		if expiresAtUnix, ok := claims["subscription_expires_at"].(float64); ok {
			expiresAt := time.Unix(int64(expiresAtUnix), 0)
			if time.Now().After(expiresAt) {
				// Subscription period has ended - reject even if DB says active
				// User needs to refresh token after renewing subscription
				response.HandleError(w, subscription.ErrSubscriptionExpired)
				return
			}
		}

		// DB verification: Check subscription status in database
		sub, err := m.subscriptionService.GetMySubscription(r.Context(), companyID)
		if err != nil {
			response.HandleError(w, subscription.ErrSubscriptionNotFound)
			return
		}

		// Check if subscription is active, trial, or in grace period (past_due)
		status := subscription.SubscriptionStatus(sub.Status)
		if !isActiveStatus(status) {
			response.HandleError(w, subscription.ErrSubscriptionExpired)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireFeature checks if the company's subscription has access to a specific feature
// This performs a double check: JWT claims + database verification
func (m *SubscriptionMiddleware) RequireFeature(featureCode string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, _, err := jwtauth.FromContext(r.Context())
			if err != nil || token == nil {
				response.Unauthorized(w, "unauthorized")
				return
			}

			claims, err := token.AsMap(r.Context())
			if err != nil {
				response.Unauthorized(w, "invalid token claims")
				return
			}

			// Get company_id from JWT claims
			companyID, ok := claims["company_id"].(string)
			if !ok || companyID == "" {
				response.Forbidden(w, "no company associated with this user")
				return
			}

			// Quick check from JWT claims (if features are included)
			// This is a fast-path optimization
			if features, ok := claims["features"].([]interface{}); ok {
				hasFeature := false
				for _, f := range features {
					if code, ok := f.(string); ok && code == featureCode {
						hasFeature = true
						break
					}
				}
				if hasFeature {
					next.ServeHTTP(w, r)
					return
				}
			}

			// DB verification: Check feature access in database
			hasFeature, err := m.subscriptionService.HasFeature(r.Context(), companyID, featureCode)
			if err != nil {
				response.InternalServerError(w, "failed to check feature access")
				return
			}

			if !hasFeature {
				response.HandleError(w, subscription.ErrFeatureNotAvailable)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireCanAddEmployee checks if the company can add more employees based on seat limit
func (m *SubscriptionMiddleware) RequireCanAddEmployee(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _, err := jwtauth.FromContext(r.Context())
		if err != nil || token == nil {
			response.Unauthorized(w, "unauthorized")
			return
		}

		claims, err := token.AsMap(r.Context())
		if err != nil {
			response.Unauthorized(w, "invalid token claims")
			return
		}

		// Get company_id from JWT claims
		companyID, ok := claims["company_id"].(string)
		if !ok || companyID == "" {
			response.Forbidden(w, "no company associated with this user")
			return
		}

		// Check if company can add more employees
		canAdd, err := m.subscriptionService.CanAddEmployee(r.Context(), companyID)
		if err != nil {
			response.InternalServerError(w, "failed to check seat limit")
			return
		}

		if !canAdd {
			response.HandleError(w, subscription.ErrSeatLimitExceeded)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isActiveStatus checks if subscription status allows access
// Cancelled status is allowed because time-based check enforces period_end
func isActiveStatus(status subscription.SubscriptionStatus) bool {
	switch status {
	case subscription.StatusActive, subscription.StatusTrial, subscription.StatusPastDue, subscription.StatusCancelled:
		return true
	default:
		return false
	}
}

// Feature codes for easy reference - Must match database feature codes
const (
	FeatureAttendance = "attendance" // Clock in/out, attendance tracking
	FeatureLeave      = "leave"      // Leave requests, approvals, quota management
	FeaturePayroll    = "payroll"    // Salary calculation, payslips
	FeatureInvitation = "invitation" // Invite employees via email
	FeatureSchedule   = "schedule"   // Work schedule management
	FeatureReport     = "report"     // Advanced reports and analytics
)
