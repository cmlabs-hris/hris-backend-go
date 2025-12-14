package notification

import (
	"time"
)

// ============= Request DTOs =============

// CreateNotificationRequest represents a request to create a notification
type CreateNotificationRequest struct {
	CompanyID   string
	RecipientID string
	SenderID    *string
	Type        NotificationType
	Title       string
	Message     string
	Data        map[string]interface{}
}

// MarkAsReadRequest represents a request to mark notifications as read
type MarkAsReadRequest struct {
	NotificationIDs []string `json:"notification_ids" validate:"required,min=1"`
}

// UpdatePreferenceRequest represents a request to update notification preference
type UpdatePreferenceRequest struct {
	NotificationType NotificationType `json:"notification_type" validate:"required"`
	EmailEnabled     bool             `json:"email_enabled"`
	PushEnabled      bool             `json:"push_enabled"`
}

// ListNotificationsRequest represents a request to list notifications
type ListNotificationsRequest struct {
	UserID   string
	Page     int
	PageSize int
	Unread   *bool
}

// ============= Response DTOs =============

// NotificationResponse represents a notification in API responses
type NotificationResponse struct {
	ID        string                 `json:"id"`
	Type      NotificationType       `json:"type"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	IsRead    bool                   `json:"is_read"`
	ReadAt    *time.Time             `json:"read_at,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// NotificationListResponse represents a paginated list of notifications
type NotificationListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	Total         int                    `json:"total"`
	UnreadCount   int                    `json:"unread_count"`
	Page          int                    `json:"page"`
	PageSize      int                    `json:"page_size"`
}

// PreferenceResponse represents a notification preference in API responses
type PreferenceResponse struct {
	NotificationType NotificationType `json:"notification_type"`
	EmailEnabled     bool             `json:"email_enabled"`
	PushEnabled      bool             `json:"push_enabled"`
}

// UnreadCountResponse represents unread count response
type UnreadCountResponse struct {
	UnreadCount int `json:"unread_count"`
}

// SSETokenResponse represents the SSE token response
type SSETokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
}

// ============= SSE Event =============

// SSEEvent represents a Server-Sent Event
type SSEEvent struct {
	Event string               `json:"event"`
	Data  NotificationResponse `json:"data"`
}
