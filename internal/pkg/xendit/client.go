package xendit

import (
	"fmt"

	"github.com/cmlabs-hris/hris-backend-go/internal/config"
	xenditSDK "github.com/xendit/xendit-go/v7"
	"github.com/xendit/xendit-go/v7/invoice"
)

// Client wraps the official Xendit SDK
type Client struct {
	sdk         *xenditSDK.APIClient
	invoiceAPI  invoice.InvoiceApi
	environment string
}

// NewClient creates a new Xendit client using the official SDK
func NewClient(cfg config.XenditConfig) *Client {
	sdk := xenditSDK.NewClient(cfg.APIKey)

	return &Client{
		sdk:         sdk,
		invoiceAPI:  sdk.InvoiceApi,
		environment: cfg.Environment,
	}
}

// APIError represents a Xendit API error
type APIError struct {
	StatusCode int
	ErrorCode  string
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("xendit API error [%d] %s: %s", e.StatusCode, e.ErrorCode, e.Message)
}

// IsSandbox returns true if running in sandbox mode
func (c *Client) IsSandbox() bool {
	return c.environment == "sandbox"
}
