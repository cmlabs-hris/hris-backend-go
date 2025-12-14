package notification

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/notification"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/sse"
	"github.com/google/uuid"
)

// Config holds notification service configuration
type Config struct {
	BatchSize     int           // default: 100
	FlushInterval time.Duration // default: 5 seconds
	WorkerCount   int           // default: 2
	QueueSize     int           // default: 1000
}

type service struct {
	repo   notification.Repository
	hub    *sse.Hub
	config Config

	queue  chan notification.CreateNotificationRequest
	wg     sync.WaitGroup
	stopCh chan struct{}
}

// NewNotificationService creates a new notification service with background workers
func NewNotificationService(repo notification.Repository, hub *sse.Hub, cfg Config) notification.Service {
	// Set defaults
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 100
	}
	if cfg.FlushInterval == 0 {
		cfg.FlushInterval = 5 * time.Second
	}
	if cfg.WorkerCount == 0 {
		cfg.WorkerCount = 2
	}
	if cfg.QueueSize == 0 {
		cfg.QueueSize = 1000
	}

	s := &service{
		repo:   repo,
		hub:    hub,
		config: cfg,
		queue:  make(chan notification.CreateNotificationRequest, cfg.QueueSize),
		stopCh: make(chan struct{}),
	}

	// Start background workers
	for i := 0; i < cfg.WorkerCount; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}

	log.Printf("[NotificationService] Started with %d workers, batch size %d, flush interval %v",
		cfg.WorkerCount, cfg.BatchSize, cfg.FlushInterval)

	return s
}

// worker is the background worker that processes notification queue
func (s *service) worker(id int) {
	defer s.wg.Done()

	batch := make([]notification.CreateNotificationRequest, 0, s.config.BatchSize)
	ticker := time.NewTicker(s.config.FlushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Convert to entities
		notifications := make([]*notification.Notification, len(batch))
		for i, req := range batch {
			notifications[i] = &notification.Notification{
				ID:          uuid.New().String(),
				CompanyID:   req.CompanyID,
				RecipientID: req.RecipientID,
				SenderID:    req.SenderID,
				Type:        req.Type,
				Title:       req.Title,
				Message:     req.Message,
				Data:        req.Data,
				IsRead:      false,
				CreatedAt:   time.Now(),
			}
		}

		// Batch insert
		if err := s.repo.CreateBatch(ctx, notifications); err != nil {
			log.Printf("[NotificationWorker-%d] Failed to batch insert: %v", id, err)
		} else {
			log.Printf("[NotificationWorker-%d] Inserted %d notifications", id, len(notifications))

			// Push to SSE subscribers
			for _, n := range notifications {
				s.hub.Publish(n.RecipientID, sse.Event{
					UserID: n.RecipientID,
					Event:  "notification",
					Data:   s.toResponse(n),
				})
			}
		}

		batch = batch[:0]
	}

	for {
		select {
		case req := <-s.queue:
			batch = append(batch, req)
			if len(batch) >= s.config.BatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-s.stopCh:
			flush()
			return
		}
	}
}

// QueueNotification queues a notification for async processing
func (s *service) QueueNotification(ctx context.Context, req notification.CreateNotificationRequest) error {
	// Check if push notification is enabled for this user/type
	enabled, err := s.repo.IsNotificationEnabled(ctx, req.RecipientID, req.Type)
	if err != nil {
		return err
	}
	if !enabled {
		return nil // Skip if disabled
	}

	select {
	case s.queue <- req:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Queue full, try direct insert
		return s.directInsert(ctx, req)
	}
}

// QueueBulkNotification queues multiple notifications for async processing
func (s *service) QueueBulkNotification(ctx context.Context, reqs []notification.CreateNotificationRequest) error {
	for _, req := range reqs {
		if err := s.QueueNotification(ctx, req); err != nil {
			log.Printf("[NotificationService] Failed to queue notification: %v", err)
		}
	}
	return nil
}

// directInsert inserts a notification directly when queue is full
func (s *service) directInsert(ctx context.Context, req notification.CreateNotificationRequest) error {
	n := &notification.Notification{
		ID:          uuid.New().String(),
		CompanyID:   req.CompanyID,
		RecipientID: req.RecipientID,
		SenderID:    req.SenderID,
		Type:        req.Type,
		Title:       req.Title,
		Message:     req.Message,
		Data:        req.Data,
		IsRead:      false,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, n); err != nil {
		return err
	}

	s.hub.Publish(n.RecipientID, sse.Event{
		UserID: n.RecipientID,
		Event:  "notification",
		Data:   s.toResponse(n),
	})

	return nil
}

// toResponse converts a Notification entity to NotificationResponse
func (s *service) toResponse(n *notification.Notification) notification.NotificationResponse {
	return notification.NotificationResponse{
		ID:        n.ID,
		Type:      n.Type,
		Title:     n.Title,
		Message:   n.Message,
		Data:      n.Data,
		IsRead:    n.IsRead,
		ReadAt:    n.ReadAt,
		CreatedAt: n.CreatedAt,
	}
}

// GetNotifications retrieves paginated notifications for a user
func (s *service) GetNotifications(ctx context.Context, userID string, page, pageSize int, unreadOnly bool) (*notification.NotificationListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	notifications, total, err := s.repo.GetByUserID(ctx, userID, page, pageSize, unreadOnly)
	if err != nil {
		return nil, err
	}

	unreadCount, err := s.repo.GetUnreadCount(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]notification.NotificationResponse, len(notifications))
	for i, n := range notifications {
		responses[i] = s.toResponse(n)
	}

	return &notification.NotificationListResponse{
		Notifications: responses,
		Total:         total,
		UnreadCount:   unreadCount,
		Page:          page,
		PageSize:      pageSize,
	}, nil
}

// GetUnreadCount returns the count of unread notifications
func (s *service) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	return s.repo.GetUnreadCount(ctx, userID)
}

// MarkAsRead marks specified notifications as read
func (s *service) MarkAsRead(ctx context.Context, userID string, req notification.MarkAsReadRequest) error {
	return s.repo.MarkAsRead(ctx, req.NotificationIDs, userID)
}

// MarkAllAsRead marks all notifications as read for a user
func (s *service) MarkAllAsRead(ctx context.Context, userID string) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}

// Delete removes a notification
func (s *service) Delete(ctx context.Context, userID string, notificationID string) error {
	return s.repo.Delete(ctx, notificationID, userID)
}

// GetPreferences retrieves all notification preferences for a user
func (s *service) GetPreferences(ctx context.Context, userID string) ([]notification.PreferenceResponse, error) {
	prefs, err := s.repo.GetPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Create a map of existing preferences
	prefMap := make(map[notification.NotificationType]*notification.NotificationPreference)
	for _, p := range prefs {
		prefMap[p.NotificationType] = p
	}

	// Return all notification types with their preferences (defaults if not set)
	allTypes := notification.AllNotificationTypes()
	responses := make([]notification.PreferenceResponse, len(allTypes))

	for i, t := range allTypes {
		if p, ok := prefMap[t]; ok {
			responses[i] = notification.PreferenceResponse{
				NotificationType: t,
				EmailEnabled:     p.EmailEnabled,
				PushEnabled:      p.PushEnabled,
			}
		} else {
			// Default preferences
			responses[i] = notification.PreferenceResponse{
				NotificationType: t,
				EmailEnabled:     true,
				PushEnabled:      true,
			}
		}
	}

	return responses, nil
}

// UpdatePreference updates a notification preference
func (s *service) UpdatePreference(ctx context.Context, userID string, req notification.UpdatePreferenceRequest) error {
	pref := &notification.NotificationPreference{
		UserID:           userID,
		NotificationType: req.NotificationType,
		EmailEnabled:     req.EmailEnabled,
		PushEnabled:      req.PushEnabled,
		UpdatedAt:        time.Now(),
	}

	return s.repo.UpsertPreference(ctx, pref)
}

// Subscribe creates an SSE subscription for a user
func (s *service) Subscribe(ctx context.Context, userID string) (<-chan notification.SSEEvent, func()) {
	ch, cleanup := s.hub.Subscribe(userID)

	out := make(chan notification.SSEEvent, 10)

	go func() {
		defer close(out)
		for {
			select {
			case event, ok := <-ch:
				if !ok {
					return
				}
				if resp, ok := event.Data.(notification.NotificationResponse); ok {
					out <- notification.SSEEvent{
						Event: event.Event,
						Data:  resp,
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, cleanup
}

// Stop gracefully stops the notification service
func (s *service) Stop() {
	close(s.stopCh)
	s.wg.Wait()
	log.Println("[NotificationService] Stopped")
}
