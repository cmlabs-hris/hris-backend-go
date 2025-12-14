package notification

import "errors"

// Notification domain errors
var (
	ErrNotificationNotFound    = errors.New("notification not found")
	ErrUnauthorized            = errors.New("unauthorized to access this notification")
	ErrInvalidNotificationType = errors.New("invalid notification type")
	ErrPreferenceNotFound      = errors.New("notification preference not found")
	ErrQueueFull               = errors.New("notification queue is full")
)
