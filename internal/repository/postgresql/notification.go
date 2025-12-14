package postgresql

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/notification"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type notificationRepository struct {
	db *database.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *database.DB) notification.Repository {
	return &notificationRepository{db: db}
}

// Create creates a new notification
func (r *notificationRepository) Create(ctx context.Context, n *notification.Notification) error {
	q := GetQuerier(ctx, r.db)

	if n.ID == "" {
		n.ID = uuid.New().String()
	}

	dataJSON, err := json.Marshal(n.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal notification data: %w", err)
	}

	query := `
		INSERT INTO notifications (id, company_id, recipient_id, sender_id, type, title, message, data, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = q.Exec(ctx, query,
		n.ID,
		n.CompanyID,
		n.RecipientID,
		n.SenderID,
		string(n.Type),
		n.Title,
		n.Message,
		dataJSON,
		n.IsRead,
		n.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

// CreateBatch creates multiple notifications in a single transaction
func (r *notificationRepository) CreateBatch(ctx context.Context, notifications []*notification.Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	q := GetQuerier(ctx, r.db)

	// Build batch insert query
	valueStrings := make([]string, 0, len(notifications))
	valueArgs := make([]interface{}, 0, len(notifications)*10)

	for i, n := range notifications {
		if n.ID == "" {
			n.ID = uuid.New().String()
		}

		dataJSON, err := json.Marshal(n.Data)
		if err != nil {
			return fmt.Errorf("failed to marshal notification data: %w", err)
		}

		base := i * 10
		valueStrings = append(valueStrings, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9, base+10,
		))
		valueArgs = append(valueArgs,
			n.ID,
			n.CompanyID,
			n.RecipientID,
			n.SenderID,
			string(n.Type),
			n.Title,
			n.Message,
			dataJSON,
			n.IsRead,
			n.CreatedAt,
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO notifications (id, company_id, recipient_id, sender_id, type, title, message, data, is_read, created_at)
		VALUES %s
	`, strings.Join(valueStrings, ", "))

	_, err := q.Exec(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to batch create notifications: %w", err)
	}

	return nil
}

// GetByID retrieves a notification by ID
func (r *notificationRepository) GetByID(ctx context.Context, id string) (*notification.Notification, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, company_id, recipient_id, sender_id, type, title, message, data, is_read, read_at, created_at
		FROM notifications
		WHERE id = $1
	`

	var n notification.Notification
	var dataJSON []byte
	var notifType string

	err := q.QueryRow(ctx, query, id).Scan(
		&n.ID,
		&n.CompanyID,
		&n.RecipientID,
		&n.SenderID,
		&notifType,
		&n.Title,
		&n.Message,
		&dataJSON,
		&n.IsRead,
		&n.ReadAt,
		&n.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, notification.ErrNotificationNotFound
		}
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	n.Type = notification.NotificationType(notifType)
	if dataJSON != nil {
		if err := json.Unmarshal(dataJSON, &n.Data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal notification data: %w", err)
		}
	}

	return &n, nil
}

// GetByUserID retrieves notifications for a user with pagination
func (r *notificationRepository) GetByUserID(ctx context.Context, userID string, page, pageSize int, unreadOnly bool) ([]*notification.Notification, int, error) {
	q := GetQuerier(ctx, r.db)

	offset := (page - 1) * pageSize

	// Build query with optional unread filter
	whereClause := "recipient_id = $1"
	args := []interface{}{userID}
	argIndex := 2

	if unreadOnly {
		whereClause += " AND is_read = false"
	}

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM notifications WHERE %s", whereClause)
	var total int
	if err := q.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	// Data query
	query := fmt.Sprintf(`
		SELECT id, company_id, recipient_id, sender_id, type, title, message, data, is_read, read_at, created_at
		FROM notifications
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, pageSize, offset)

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*notification.Notification
	for rows.Next() {
		var n notification.Notification
		var dataJSON []byte
		var notifType string

		if err := rows.Scan(
			&n.ID,
			&n.CompanyID,
			&n.RecipientID,
			&n.SenderID,
			&notifType,
			&n.Title,
			&n.Message,
			&dataJSON,
			&n.IsRead,
			&n.ReadAt,
			&n.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan notification: %w", err)
		}

		n.Type = notification.NotificationType(notifType)
		if dataJSON != nil {
			if err := json.Unmarshal(dataJSON, &n.Data); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal notification data: %w", err)
			}
		}

		notifications = append(notifications, &n)
	}

	return notifications, total, nil
}

// GetUnreadCount returns the count of unread notifications for a user
func (r *notificationRepository) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	q := GetQuerier(ctx, r.db)

	query := `SELECT COUNT(*) FROM notifications WHERE recipient_id = $1 AND is_read = false`
	var count int
	if err := q.QueryRow(ctx, query, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	return count, nil
}

// MarkAsRead marks specific notifications as read
func (r *notificationRepository) MarkAsRead(ctx context.Context, ids []string, userID string) error {
	if len(ids) == 0 {
		return nil
	}

	q := GetQuerier(ctx, r.db)

	// Build placeholders
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids)+2)
	args[0] = time.Now()
	args[1] = userID

	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+3)
		args[i+2] = id
	}

	query := fmt.Sprintf(`
		UPDATE notifications
		SET is_read = true, read_at = $1
		WHERE recipient_id = $2 AND id IN (%s)
	`, strings.Join(placeholders, ", "))

	_, err := q.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to mark notifications as read: %w", err)
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID string) error {
	q := GetQuerier(ctx, r.db)

	query := `
		UPDATE notifications
		SET is_read = true, read_at = $1
		WHERE recipient_id = $2 AND is_read = false
	`

	_, err := q.Exec(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

// Delete deletes a notification
func (r *notificationRepository) Delete(ctx context.Context, id string, userID string) error {
	q := GetQuerier(ctx, r.db)

	query := `DELETE FROM notifications WHERE id = $1 AND recipient_id = $2`
	result, err := q.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	if result.RowsAffected() == 0 {
		return notification.ErrNotificationNotFound
	}

	return nil
}

// ============= Preferences =============

// GetPreferences retrieves all notification preferences for a user
func (r *notificationRepository) GetPreferences(ctx context.Context, userID string) ([]*notification.NotificationPreference, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, user_id, notification_type, email_enabled, push_enabled, created_at, updated_at
		FROM notification_preferences
		WHERE user_id = $1
	`

	rows, err := q.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query preferences: %w", err)
	}
	defer rows.Close()

	var prefs []*notification.NotificationPreference
	for rows.Next() {
		var p notification.NotificationPreference
		var notifType string

		if err := rows.Scan(
			&p.ID,
			&p.UserID,
			&notifType,
			&p.EmailEnabled,
			&p.PushEnabled,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan preference: %w", err)
		}

		p.NotificationType = notification.NotificationType(notifType)
		prefs = append(prefs, &p)
	}

	return prefs, nil
}

// GetPreference retrieves a specific notification preference
func (r *notificationRepository) GetPreference(ctx context.Context, userID string, notifType notification.NotificationType) (*notification.NotificationPreference, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT id, user_id, notification_type, email_enabled, push_enabled, created_at, updated_at
		FROM notification_preferences
		WHERE user_id = $1 AND notification_type = $2
	`

	var p notification.NotificationPreference
	var nt string

	err := q.QueryRow(ctx, query, userID, string(notifType)).Scan(
		&p.ID,
		&p.UserID,
		&nt,
		&p.EmailEnabled,
		&p.PushEnabled,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, notification.ErrPreferenceNotFound
		}
		return nil, fmt.Errorf("failed to get preference: %w", err)
	}

	p.NotificationType = notification.NotificationType(nt)
	return &p, nil
}

// UpsertPreference creates or updates a notification preference
func (r *notificationRepository) UpsertPreference(ctx context.Context, pref *notification.NotificationPreference) error {
	q := GetQuerier(ctx, r.db)

	if pref.ID == "" {
		pref.ID = uuid.New().String()
	}

	query := `
		INSERT INTO notification_preferences (id, user_id, notification_type, email_enabled, push_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, notification_type)
		DO UPDATE SET email_enabled = $4, push_enabled = $5, updated_at = $7
	`

	now := time.Now()
	_, err := q.Exec(ctx, query,
		pref.ID,
		pref.UserID,
		string(pref.NotificationType),
		pref.EmailEnabled,
		pref.PushEnabled,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert preference: %w", err)
	}

	return nil
}

// IsNotificationEnabled checks if push notifications are enabled for a user and type
func (r *notificationRepository) IsNotificationEnabled(ctx context.Context, userID string, notifType notification.NotificationType) (bool, error) {
	q := GetQuerier(ctx, r.db)

	query := `
		SELECT push_enabled
		FROM notification_preferences
		WHERE user_id = $1 AND notification_type = $2
	`

	var enabled bool
	err := q.QueryRow(ctx, query, userID, string(notifType)).Scan(&enabled)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Default to enabled if no preference exists
			return true, nil
		}
		return false, fmt.Errorf("failed to check notification enabled: %w", err)
	}

	return enabled, nil
}
