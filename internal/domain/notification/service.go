package notification

import (
	"context"
)

// Service defines the notification service interface
type Service interface {
	// Queue notification (async processing via background workers)
	QueueNotification(ctx context.Context, req CreateNotificationRequest) error
	QueueBulkNotification(ctx context.Context, reqs []CreateNotificationRequest) error

	// Direct operations
	GetNotifications(ctx context.Context, userID string, page, pageSize int, unreadOnly bool) (*NotificationListResponse, error)
	GetUnreadCount(ctx context.Context, userID string) (int, error)
	MarkAsRead(ctx context.Context, userID string, req MarkAsReadRequest) error
	MarkAllAsRead(ctx context.Context, userID string) error
	Delete(ctx context.Context, userID string, notificationID string) error

	// Preferences
	GetPreferences(ctx context.Context, userID string) ([]PreferenceResponse, error)
	UpdatePreference(ctx context.Context, userID string, req UpdatePreferenceRequest) error

	// SSE subscription
	Subscribe(ctx context.Context, userID string) (<-chan SSEEvent, func())

	// Lifecycle
	Stop()
}
