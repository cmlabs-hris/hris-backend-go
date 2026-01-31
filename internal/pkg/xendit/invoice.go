package xendit

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/xendit/xendit-go/v7/invoice"
)

// CreateInvoiceRequest represents the request to create an invoice
// Maintains backward compatibility with existing service code
type CreateInvoiceRequest struct {
	ExternalID         string            `json:"external_id"`
	Amount             decimal.Decimal   `json:"amount"`
	Description        string            `json:"description"`
	PayerEmail         string            `json:"payer_email"`
	CustomerName       string            `json:"customer_name,omitempty"`
	Currency           string            `json:"currency,omitempty"`         // Default: IDR
	InvoiceDuration    int               `json:"invoice_duration,omitempty"` // In seconds
	SuccessRedirectURL string            `json:"success_redirect_url,omitempty"`
	FailureRedirectURL string            `json:"failure_redirect_url,omitempty"`
	Items              []InvoiceItem     `json:"items,omitempty"`
	Fees               []InvoiceFee      `json:"fees,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
}

// InvoiceItem represents an item in the invoice
type InvoiceItem struct {
	Name     string          `json:"name"`
	Quantity int             `json:"quantity"`
	Price    decimal.Decimal `json:"price"`
}

// InvoiceFee represents additional fees
type InvoiceFee struct {
	Type  string          `json:"type"`
	Value decimal.Decimal `json:"value"`
}

// InvoiceResponse represents the response from creating/getting an invoice
type InvoiceResponse struct {
	ID                 string    `json:"id"`
	ExternalID         string    `json:"external_id"`
	UserID             string    `json:"user_id"`
	Status             string    `json:"status"` // PENDING, PAID, SETTLED, EXPIRED
	MerchantName       string    `json:"merchant_name"`
	MerchantProfilePic string    `json:"merchant_profile_picture_url"`
	Amount             float64   `json:"amount"`
	PayerEmail         string    `json:"payer_email"`
	Description        string    `json:"description"`
	InvoiceURL         string    `json:"invoice_url"`
	ExpiryDate         time.Time `json:"expiry_date"`
	Currency           string    `json:"currency"`
	Created            time.Time `json:"created"`
	Updated            time.Time `json:"updated"`
	PaidAt             *string   `json:"paid_at,omitempty"`
	PaymentMethod      string    `json:"payment_method,omitempty"`
	PaymentChannel     string    `json:"payment_channel,omitempty"`
	PaymentDestination string    `json:"payment_destination,omitempty"`
}

// CreateInvoice creates a new invoice using the official Xendit SDK
func (c *Client) CreateInvoice(req CreateInvoiceRequest) (*InvoiceResponse, error) {
	ctx := context.Background()

	currency := req.Currency
	if currency == "" {
		currency = "IDR"
	}

	// Convert decimal to float64 for SDK
	amount, _ := req.Amount.Float64()

	// Build SDK request
	sdkReq := *invoice.NewCreateInvoiceRequest(req.ExternalID, amount)

	if req.PayerEmail != "" {
		sdkReq.SetPayerEmail(req.PayerEmail)
	}
	if req.Description != "" {
		sdkReq.SetDescription(req.Description)
	}
	if req.InvoiceDuration > 0 {
		sdkReq.SetInvoiceDuration(float32(req.InvoiceDuration))
	}
	if req.SuccessRedirectURL != "" {
		sdkReq.SetSuccessRedirectUrl(req.SuccessRedirectURL)
	}
	if req.FailureRedirectURL != "" {
		sdkReq.SetFailureRedirectUrl(req.FailureRedirectURL)
	}
	sdkReq.SetCurrency(currency)

	// Convert items if present
	if len(req.Items) > 0 {
		items := make([]invoice.InvoiceItem, len(req.Items))
		for i, item := range req.Items {
			price, _ := item.Price.Float64()
			items[i] = *invoice.NewInvoiceItem(item.Name, float32(price), float32(item.Quantity))
		}
		sdkReq.SetItems(items)
	}

	// Convert fees if present
	if len(req.Fees) > 0 {
		fees := make([]invoice.InvoiceFee, len(req.Fees))
		for i, fee := range req.Fees {
			value, _ := fee.Value.Float64()
			fees[i] = *invoice.NewInvoiceFee(fee.Type, float32(value))
		}
		sdkReq.SetFees(fees)
	}

	// Convert metadata if present
	if len(req.Metadata) > 0 {
		metadata := make(map[string]interface{})
		for k, v := range req.Metadata {
			metadata[k] = v
		}
		sdkReq.SetMetadata(metadata)
	}

	// Execute API call
	resp, _, err := c.invoiceAPI.CreateInvoice(ctx).
		CreateInvoiceRequest(sdkReq).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	return toInvoiceResponse(resp), nil
}

// GetInvoice retrieves an invoice by ID using the official Xendit SDK
func (c *Client) GetInvoice(invoiceID string) (*InvoiceResponse, error) {
	ctx := context.Background()

	resp, _, err := c.invoiceAPI.GetInvoiceById(ctx, invoiceID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	return toInvoiceResponse(resp), nil
}

// ExpireInvoice expires an invoice using the official Xendit SDK
func (c *Client) ExpireInvoice(invoiceID string) (*InvoiceResponse, error) {
	ctx := context.Background()

	resp, _, err := c.invoiceAPI.ExpireInvoice(ctx, invoiceID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to expire invoice: %w", err)
	}

	return toInvoiceResponse(resp), nil
}

// toInvoiceResponse converts SDK Invoice to our InvoiceResponse
func toInvoiceResponse(inv *invoice.Invoice) *InvoiceResponse {
	if inv == nil {
		return nil
	}

	resp := &InvoiceResponse{
		ID:         inv.GetId(),
		ExternalID: inv.GetExternalId(),
		UserID:     inv.GetUserId(),
		Status:     string(inv.GetStatus()),
		Amount:     inv.GetAmount(),
		InvoiceURL: inv.GetInvoiceUrl(),
		ExpiryDate: inv.GetExpiryDate(),
		Currency:   string(inv.GetCurrency()),
		Created:    inv.GetCreated(),
		Updated:    inv.GetUpdated(),
	}

	// Set merchant info if available
	resp.MerchantName = inv.GetMerchantName()
	resp.MerchantProfilePic = inv.GetMerchantProfilePictureUrl()

	// Set optional fields
	if inv.HasPayerEmail() {
		resp.PayerEmail = inv.GetPayerEmail()
	}
	if inv.HasDescription() {
		resp.Description = inv.GetDescription()
	}
	// Note: PaidAt, PaymentMethod, PaymentChannel, PaymentDestination
	// are only available in webhook callback (InvoiceCallback), not in Invoice response
	// The service layer will get this info from webhook payload

	return resp
}

// InvoiceStatus constants
const (
	InvoiceStatusPending = "PENDING"
	InvoiceStatusPaid    = "PAID"
	InvoiceStatusSettled = "SETTLED"
	InvoiceStatusExpired = "EXPIRED"
)
