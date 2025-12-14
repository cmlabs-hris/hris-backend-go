package notification

import (
	"context"
)

// Repository defines the notification repository interface
type Repository interface {
	// Notifications CRUD
	Create(ctx context.Context, notification *Notification) error
	CreateBatch(ctx context.Context, notifications []*Notification) error
	GetByID(ctx context.Context, id string) (*Notification, error)
	GetByUserID(ctx context.Context, userID string, page, pageSize int, unreadOnly bool) ([]*Notification, int, error)
	GetUnreadCount(ctx context.Context, userID string) (int, error)
	MarkAsRead(ctx context.Context, ids []string, userID string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	Delete(ctx context.Context, id string, userID string) error

	// Preferences
	GetPreferences(ctx context.Context, userID string) ([]*NotificationPreference, error)
	GetPreference(ctx context.Context, userID string, notifType NotificationType) (*NotificationPreference, error)
	UpsertPreference(ctx context.Context, pref *NotificationPreference) error
	IsNotificationEnabled(ctx context.Context, userID string, notifType NotificationType) (bool, error)
}
