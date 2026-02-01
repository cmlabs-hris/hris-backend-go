package http

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/subscription"
	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/xendit"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

// SubscriptionHandler handles subscription-related HTTP requests
type SubscriptionHandler interface {
	// Public endpoints
	GetPlans(w http.ResponseWriter, r *http.Request)
	HandleWebhook(w http.ResponseWriter, r *http.Request)

	// Authenticated endpoints
	GetMySubscription(w http.ResponseWriter, r *http.Request)
	GetInvoices(w http.ResponseWriter, r *http.Request)
	GetInvoiceByID(w http.ResponseWriter, r *http.Request)

	// Owner-only endpoints
	Checkout(w http.ResponseWriter, r *http.Request)
	UpgradePlan(w http.ResponseWriter, r *http.Request)
	DowngradePlan(w http.ResponseWriter, r *http.Request)
	CancelSubscription(w http.ResponseWriter, r *http.Request)
	ChangeSeats(w http.ResponseWriter, r *http.Request)
	CancelPendingInvoice(w http.ResponseWriter, r *http.Request)
}

type subscriptionHandlerImpl struct {
	subscriptionService subscription.SubscriptionService
	webhookVerifier     *xendit.WebhookVerifier
}

// NewSubscriptionHandler creates a new subscription handler
func NewSubscriptionHandler(
	subscriptionService subscription.SubscriptionService,
	webhookVerifier *xendit.WebhookVerifier,
) SubscriptionHandler {
	return &subscriptionHandlerImpl{
		subscriptionService: subscriptionService,
		webhookVerifier:     webhookVerifier,
	}
}

// GetPlans retrieves all available subscription plans
// GET /api/v1/plans - Public
func (h *subscriptionHandlerImpl) GetPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.subscriptionService.GetPlans(r.Context())
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, plans)
}

// GetMySubscription retrieves the current company's subscription
// GET /api/v1/subscription/my - Authenticated
func (h *subscriptionHandlerImpl) GetMySubscription(w http.ResponseWriter, r *http.Request) {
	companyID, ok := getCompanyIDFromContext(r)
	if !ok {
		response.Forbidden(w, "no company associated with this user")
		return
	}

	sub, err := h.subscriptionService.GetMySubscription(r.Context(), companyID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, sub)
}

// GetInvoices retrieves all invoices for the current company
// GET /api/v1/subscription/invoices - Authenticated
func (h *subscriptionHandlerImpl) GetInvoices(w http.ResponseWriter, r *http.Request) {
	companyID, ok := getCompanyIDFromContext(r)
	if !ok {
		response.Forbidden(w, "no company associated with this user")
		return
	}

	invoices, err := h.subscriptionService.GetInvoices(r.Context(), companyID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, invoices)
}

// GetInvoiceByID retrieves a specific invoice
// GET /api/v1/subscription/invoices/{id} - Authenticated
func (h *subscriptionHandlerImpl) GetInvoiceByID(w http.ResponseWriter, r *http.Request) {
	companyID, ok := getCompanyIDFromContext(r)
	if !ok {
		response.Forbidden(w, "no company associated with this user")
		return
	}

	invoiceID := chi.URLParam(r, "id")
	if invoiceID == "" {
		response.BadRequest(w, "invoice ID is required", nil)
		return
	}

	invoice, err := h.subscriptionService.GetInvoiceByID(r.Context(), companyID, invoiceID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, invoice)
}

// Checkout creates a new subscription invoice
// POST /api/v1/subscription/checkout - Owner only
func (h *subscriptionHandlerImpl) Checkout(w http.ResponseWriter, r *http.Request) {
	companyID, ok := getCompanyIDFromContext(r)
	if !ok {
		response.Forbidden(w, "no company associated with this user")
		return
	}

	var req subscription.CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body", nil)
		return
	}

	if err := req.Validate(); err != nil {
		response.HandleError(w, err)
		return
	}

	result, err := h.subscriptionService.Checkout(r.Context(), companyID, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Created(w, "invoice created", result)
}

// UpgradePlan upgrades to a higher tier plan
// POST /api/v1/subscription/upgrade - Owner only
func (h *subscriptionHandlerImpl) UpgradePlan(w http.ResponseWriter, r *http.Request) {
	companyID, ok := getCompanyIDFromContext(r)
	if !ok {
		response.Forbidden(w, "no company associated with this user")
		return
	}

	var req subscription.UpgradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body", nil)
		return
	}

	if err := req.Validate(); err != nil {
		response.HandleError(w, err)
		return
	}

	result, err := h.subscriptionService.UpgradePlan(r.Context(), companyID, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

// DowngradePlan downgrades to a lower tier plan (effective next period)
// POST /api/v1/subscription/downgrade - Owner only
func (h *subscriptionHandlerImpl) DowngradePlan(w http.ResponseWriter, r *http.Request) {
	companyID, ok := getCompanyIDFromContext(r)
	if !ok {
		response.Forbidden(w, "no company associated with this user")
		return
	}

	var req subscription.DowngradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body", nil)
		return
	}

	if err := req.Validate(); err != nil {
		response.HandleError(w, err)
		return
	}

	err := h.subscriptionService.DowngradePlan(r.Context(), companyID, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, map[string]string{
		"message": "Downgrade scheduled for next billing period",
	})
}

// CancelSubscription cancels the subscription
// POST /api/v1/subscription/cancel - Owner only
func (h *subscriptionHandlerImpl) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	companyID, ok := getCompanyIDFromContext(r)
	if !ok {
		response.Forbidden(w, "no company associated with this user")
		return
	}

	var req subscription.CancelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body", nil)
		return
	}

	err := h.subscriptionService.CancelSubscription(r.Context(), companyID, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, map[string]string{
		"message": "Subscription cancelled. Access will continue until the end of the current period.",
	})
}

// ChangeSeats changes the number of seats
// POST /api/v1/subscription/seats - Owner only
func (h *subscriptionHandlerImpl) ChangeSeats(w http.ResponseWriter, r *http.Request) {
	companyID, ok := getCompanyIDFromContext(r)
	if !ok {
		response.Forbidden(w, "no company associated with this user")
		return
	}

	var req subscription.ChangeSeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body", nil)
		return
	}

	if err := req.Validate(); err != nil {
		response.HandleError(w, err)
		return
	}

	result, err := h.subscriptionService.ChangeSeats(r.Context(), companyID, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

// CancelPendingInvoice cancels a pending invoice
// DELETE /api/v1/subscription/invoices/{id} - Requires owner role
func (h *subscriptionHandlerImpl) CancelPendingInvoice(w http.ResponseWriter, r *http.Request) {
	companyID, ok := getCompanyIDFromContext(r)
	if !ok {
		response.Forbidden(w, "no company associated with this user")
		return
	}

	invoiceID := chi.URLParam(r, "id")
	if invoiceID == "" {
		response.BadRequest(w, "invoice_id is required", nil)
		return
	}

	err := h.subscriptionService.CancelPendingInvoice(r.Context(), companyID, invoiceID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, map[string]string{
		"message": "invoice cancelled successfully",
	})
}

// HandleWebhook processes Xendit webhook callbacks
// POST /api/v1/webhook/xendit - Public (signature verified)
func (h *subscriptionHandlerImpl) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Read the raw body for signature verification
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.BadRequest(w, "failed to read request body", nil)
		return
	}

	// Get the callback token from header
	callbackToken := r.Header.Get("X-Callback-Token")
	if callbackToken == "" {
		response.Unauthorized(w, "missing callback token")
		return
	}

	// Verify the webhook signature
	if !h.webhookVerifier.VerifySignature(callbackToken) {
		response.Unauthorized(w, "invalid callback token")
		return
	}

	// Parse the webhook payload
	var payload subscription.XenditWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		response.BadRequest(w, "invalid webhook payload", nil)
		return
	}

	// Process the webhook
	if err := h.subscriptionService.HandleWebhook(r.Context(), payload); err != nil {
		response.HandleError(w, err)
		return
	}

	// Return 200 OK to acknowledge receipt
	response.Success(w, map[string]string{
		"status": "received",
	})
}

// Helper function to get company ID from JWT claims
func getCompanyIDFromContext(r *http.Request) (string, bool) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return "", false
	}

	companyID, ok := claims["company_id"].(string)
	return companyID, ok && companyID != ""
}
