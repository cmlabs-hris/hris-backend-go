package xendit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// WebhookVerifier handles webhook signature verification
type WebhookVerifier struct {
	webhookToken string
}

// NewWebhookVerifier creates a new webhook verifier
func NewWebhookVerifier(webhookToken string) *WebhookVerifier {
	return &WebhookVerifier{
		webhookToken: webhookToken,
	}
}

// VerifySignature verifies the webhook signature from Xendit
// Xendit sends the signature in the x-callback-token header
func (v *WebhookVerifier) VerifySignature(callbackToken string) bool {
	// Xendit uses a simple token comparison for callback verification
	// The x-callback-token header should match the webhook verification token
	return strings.TrimSpace(callbackToken) == strings.TrimSpace(v.webhookToken)
}

// VerifyHMACSignature verifies HMAC-SHA256 signature (alternative method)
// Some Xendit webhooks may use HMAC signing
func (v *WebhookVerifier) VerifyHMACSignature(payload []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(v.webhookToken))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expectedMAC), []byte(signature))
}

// WebhookEvent represents the type of webhook event
type WebhookEvent string

const (
	WebhookEventInvoicePaid    WebhookEvent = "invoices.paid"
	WebhookEventInvoiceExpired WebhookEvent = "invoices.expired"
)

// InvoiceWebhookPayload represents the webhook payload for invoice events
type InvoiceWebhookPayload struct {
	ID                     string  `json:"id"`
	ExternalID             string  `json:"external_id"`
	UserID                 string  `json:"user_id"`
	IsHigh                 bool    `json:"is_high"`
	Status                 string  `json:"status"`
	MerchantName           string  `json:"merchant_name"`
	Amount                 float64 `json:"amount"`
	PaidAmount             float64 `json:"paid_amount"`
	BankCode               string  `json:"bank_code"`
	PaidAt                 string  `json:"paid_at"`
	PayerEmail             string  `json:"payer_email"`
	Description            string  `json:"description"`
	AdjustedReceivedAmount float64 `json:"adjusted_received_amount"`
	FeesPaidAmount         float64 `json:"fees_paid_amount"`
	Updated                string  `json:"updated"`
	Created                string  `json:"created"`
	Currency               string  `json:"currency"`
	PaymentMethod          string  `json:"payment_method"`
	PaymentChannel         string  `json:"payment_channel"`
	PaymentDestination     string  `json:"payment_destination"`
}
