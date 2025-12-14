# Notification System Implementation Plan

## Overview

Implementasi sistem notifikasi real-time untuk HRIS backend dengan fitur:
- Push notification via SSE (Server-Sent Events)
- Background job processing dengan Go channels
- Batch insert untuk efisiensi database
- Notification preferences per user
- Authentication & Security untuk SSE endpoint

---

## 1. Database Schema

### Table: notifications

```sql
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    recipient_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    sender_id UUID REFERENCES users(id) ON DELETE SET NULL,
    type VARCHAR(50) NOT NULL, -- 'attendance_clock_in', 'attendance_clock_out', 'leave_request', 'leave_approved', etc.
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    data JSONB, -- additional payload (employee_id, attendance_id, leave_id, etc.)
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT valid_notification_type CHECK (type IN (
        'attendance_clock_in',
        'attendance_clock_out', 
        'leave_request',
        'leave_approved',
        'leave_rejected',
        'payroll_generated',
        'schedule_updated',
        'invitation_sent',
        'employee_joined'
    ))
);

CREATE INDEX idx_notifications_recipient ON notifications(recipient_id, is_read, created_at DESC);
CREATE INDEX idx_notifications_company ON notifications(company_id, created_at DESC);
```

### Table: notification_preferences

```sql
CREATE TABLE notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    notification_type VARCHAR(50) NOT NULL,
    email_enabled BOOLEAN DEFAULT TRUE,
    push_enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(user_id, notification_type)
);

CREATE INDEX idx_notification_preferences_user ON notification_preferences(user_id);
```

---

## 2. Domain Layer

### 2.1 Entity (`internal/domain/notification/entity.go`)

```go
package notification

import (
    "time"
    "github.com/google/uuid"
)

type NotificationType string

const (
    TypeAttendanceClockIn  NotificationType = "attendance_clock_in"
    TypeAttendanceClockOut NotificationType = "attendance_clock_out"
    TypeLeaveRequest       NotificationType = "leave_request"
    TypeLeaveApproved      NotificationType = "leave_approved"
    TypeLeaveRejected      NotificationType = "leave_rejected"
    TypePayrollGenerated   NotificationType = "payroll_generated"
    TypeScheduleUpdated    NotificationType = "schedule_updated"
    TypeInvitationSent     NotificationType = "invitation_sent"
    TypeEmployeeJoined     NotificationType = "employee_joined"
)

type Notification struct {
    ID          uuid.UUID
    CompanyID   uuid.UUID
    RecipientID uuid.UUID
    SenderID    *uuid.UUID
    Type        NotificationType
    Title       string
    Message     string
    Data        map[string]interface{}
    IsRead      bool
    ReadAt      *time.Time
    CreatedAt   time.Time
}

type NotificationPreference struct {
    ID               uuid.UUID
    UserID           uuid.UUID
    NotificationType NotificationType
    EmailEnabled     bool
    PushEnabled      bool
    CreatedAt        time.Time
    UpdatedAt        time.Time
}
```

### 2.2 DTO (`internal/domain/notification/dto.go`)

```go
package notification

import (
    "time"
    "github.com/google/uuid"
)

// Request DTOs
type CreateNotificationRequest struct {
    CompanyID   uuid.UUID
    RecipientID uuid.UUID
    SenderID    *uuid.UUID
    Type        NotificationType
    Title       string
    Message     string
    Data        map[string]interface{}
}

type MarkAsReadRequest struct {
    NotificationIDs []uuid.UUID `json:"notification_ids" validate:"required,min=1"`
}

type UpdatePreferenceRequest struct {
    NotificationType NotificationType `json:"notification_type" validate:"required"`
    EmailEnabled     bool             `json:"email_enabled"`
    PushEnabled      bool             `json:"push_enabled"`
}

type ListNotificationsRequest struct {
    UserID   uuid.UUID
    Page     int
    PageSize int
    Unread   *bool // filter by unread only
}

// Response DTOs
type NotificationResponse struct {
    ID        uuid.UUID              `json:"id"`
    Type      NotificationType       `json:"type"`
    Title     string                 `json:"title"`
    Message   string                 `json:"message"`
    Data      map[string]interface{} `json:"data,omitempty"`
    IsRead    bool                   `json:"is_read"`
    ReadAt    *time.Time             `json:"read_at,omitempty"`
    CreatedAt time.Time              `json:"created_at"`
}

type NotificationListResponse struct {
    Notifications []NotificationResponse `json:"notifications"`
    Total         int                    `json:"total"`
    UnreadCount   int                    `json:"unread_count"`
    Page          int                    `json:"page"`
    PageSize      int                    `json:"page_size"`
}

type PreferenceResponse struct {
    NotificationType NotificationType `json:"notification_type"`
    EmailEnabled     bool             `json:"email_enabled"`
    PushEnabled      bool             `json:"push_enabled"`
}

// SSE Event
type SSEEvent struct {
    Event string               `json:"event"`
    Data  NotificationResponse `json:"data"`
}
```

### 2.3 Repository Interface (`internal/domain/notification/repository.go`)

```go
package notification

import (
    "context"
    "github.com/google/uuid"
)

type Repository interface {
    // Notifications
    Create(ctx context.Context, notification *Notification) error
    CreateBatch(ctx context.Context, notifications []*Notification) error
    GetByID(ctx context.Context, id uuid.UUID) (*Notification, error)
    GetByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int, unreadOnly bool) ([]*Notification, int, error)
    GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
    MarkAsRead(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) error
    MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
    Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
    
    // Preferences
    GetPreferences(ctx context.Context, userID uuid.UUID) ([]*NotificationPreference, error)
    GetPreference(ctx context.Context, userID uuid.UUID, notifType NotificationType) (*NotificationPreference, error)
    UpsertPreference(ctx context.Context, pref *NotificationPreference) error
    IsNotificationEnabled(ctx context.Context, userID uuid.UUID, notifType NotificationType) (bool, error)
}
```

### 2.4 Service Interface (`internal/domain/notification/service.go`)

```go
package notification

import (
    "context"
    "github.com/google/uuid"
)

type Service interface {
    // Queue notification (async processing)
    QueueNotification(ctx context.Context, req CreateNotificationRequest) error
    QueueBulkNotification(ctx context.Context, reqs []CreateNotificationRequest) error
    
    // Direct operations
    GetNotifications(ctx context.Context, userID uuid.UUID, page, pageSize int, unreadOnly bool) (*NotificationListResponse, error)
    GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
    MarkAsRead(ctx context.Context, userID uuid.UUID, req MarkAsReadRequest) error
    MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
    Delete(ctx context.Context, userID uuid.UUID, notificationID uuid.UUID) error
    
    // Preferences
    GetPreferences(ctx context.Context, userID uuid.UUID) ([]PreferenceResponse, error)
    UpdatePreference(ctx context.Context, userID uuid.UUID, req UpdatePreferenceRequest) error
    
    // SSE
    Subscribe(ctx context.Context, userID uuid.UUID) (<-chan SSEEvent, func())
}
```

---

## 3. SSE Hub (`internal/pkg/sse/hub.go`)

```go
package sse

import (
    "sync"
    "github.com/google/uuid"
)

type Event struct {
    UserID uuid.UUID
    Event  string
    Data   interface{}
}

type Hub struct {
    mu          sync.RWMutex
    subscribers map[uuid.UUID]map[chan Event]struct{}
}

func NewHub() *Hub {
    return &Hub{
        subscribers: make(map[uuid.UUID]map[chan Event]struct{}),
    }
}

// Subscribe registers a new subscriber for a user
func (h *Hub) Subscribe(userID uuid.UUID) (chan Event, func()) {
    h.mu.Lock()
    defer h.mu.Unlock()
    
    ch := make(chan Event, 10)
    
    if h.subscribers[userID] == nil {
        h.subscribers[userID] = make(map[chan Event]struct{})
    }
    h.subscribers[userID][ch] = struct{}{}
    
    // Return channel and cleanup function
    cleanup := func() {
        h.mu.Lock()
        defer h.mu.Unlock()
        delete(h.subscribers[userID], ch)
        close(ch)
        if len(h.subscribers[userID]) == 0 {
            delete(h.subscribers, userID)
        }
    }
    
    return ch, cleanup
}

// Publish sends an event to all subscribers of a user
func (h *Hub) Publish(userID uuid.UUID, event Event) {
    h.mu.RLock()
    defer h.mu.RUnlock()
    
    if subs, ok := h.subscribers[userID]; ok {
        for ch := range subs {
            select {
            case ch <- event:
            default:
                // Skip if channel is full (non-blocking)
            }
        }
    }
}

// PublishToMany sends an event to multiple users
func (h *Hub) PublishToMany(userIDs []uuid.UUID, event Event) {
    for _, userID := range userIDs {
        event.UserID = userID
        h.Publish(userID, event)
    }
}
```

---

## 4. Notification Service with Queue (`internal/service/notification/service.go`)

```go
package notification

import (
    "context"
    "log"
    "sync"
    "time"
    
    "github.com/google/uuid"
    notifDomain "hris-backend-go/internal/domain/notification"
    "hris-backend-go/internal/pkg/sse"
)

type Config struct {
    BatchSize     int           // default: 100
    FlushInterval time.Duration // default: 5 seconds
    WorkerCount   int           // default: 2
}

type service struct {
    repo       notifDomain.Repository
    hub        *sse.Hub
    config     Config
    
    queue      chan notifDomain.CreateNotificationRequest
    wg         sync.WaitGroup
    stopCh     chan struct{}
}

func NewService(repo notifDomain.Repository, hub *sse.Hub, cfg Config) notifDomain.Service {
    if cfg.BatchSize == 0 {
        cfg.BatchSize = 100
    }
    if cfg.FlushInterval == 0 {
        cfg.FlushInterval = 5 * time.Second
    }
    if cfg.WorkerCount == 0 {
        cfg.WorkerCount = 2
    }
    
    s := &service{
        repo:   repo,
        hub:    hub,
        config: cfg,
        queue:  make(chan notifDomain.CreateNotificationRequest, 1000),
        stopCh: make(chan struct{}),
    }
    
    // Start background workers
    for i := 0; i < cfg.WorkerCount; i++ {
        s.wg.Add(1)
        go s.worker(i)
    }
    
    return s
}

func (s *service) worker(id int) {
    defer s.wg.Done()
    
    batch := make([]notifDomain.CreateNotificationRequest, 0, s.config.BatchSize)
    ticker := time.NewTicker(s.config.FlushInterval)
    defer ticker.Stop()
    
    flush := func() {
        if len(batch) == 0 {
            return
        }
        
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        
        // Convert to entities
        notifications := make([]*notifDomain.Notification, len(batch))
        for i, req := range batch {
            notifications[i] = &notifDomain.Notification{
                ID:          uuid.New(),
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

func (s *service) QueueNotification(ctx context.Context, req notifDomain.CreateNotificationRequest) error {
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

func (s *service) QueueBulkNotification(ctx context.Context, reqs []notifDomain.CreateNotificationRequest) error {
    for _, req := range reqs {
        if err := s.QueueNotification(ctx, req); err != nil {
            log.Printf("Failed to queue notification: %v", err)
        }
    }
    return nil
}

func (s *service) directInsert(ctx context.Context, req notifDomain.CreateNotificationRequest) error {
    n := &notifDomain.Notification{
        ID:          uuid.New(),
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

func (s *service) toResponse(n *notifDomain.Notification) notifDomain.NotificationResponse {
    return notifDomain.NotificationResponse{
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

// ... implement remaining methods (GetNotifications, MarkAsRead, etc.)

func (s *service) Subscribe(ctx context.Context, userID uuid.UUID) (<-chan notifDomain.SSEEvent, func()) {
    ch, cleanup := s.hub.Subscribe(userID)
    
    out := make(chan notifDomain.SSEEvent, 10)
    
    go func() {
        defer close(out)
        for {
            select {
            case event, ok := <-ch:
                if !ok {
                    return
                }
                if resp, ok := event.Data.(notifDomain.NotificationResponse); ok {
                    out <- notifDomain.SSEEvent{
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

func (s *service) Stop() {
    close(s.stopCh)
    s.wg.Wait()
}
```

---

## 5. Authentication & Security

### 5.1 SSE Authentication Strategy

SSE tidak mendukung custom headers, sehingga perlu pendekatan alternatif:

#### Option A: Query Parameter Token (Recommended)

```go
// Generate short-lived SSE token
func (s *authService) GenerateSSEToken(userID uuid.UUID) (string, error) {
    claims := jwt.MapClaims{
        "sub":  userID.String(),
        "type": "sse",
        "exp":  time.Now().Add(5 * time.Minute).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(s.sseSecret)
}
```

#### SSE Endpoint dengan Token Validation

```go
// GET /api/v1/notifications/stream?token=<sse_token>
func (h *NotificationHandler) Stream(w http.ResponseWriter, r *http.Request) {
    // 1. Validate SSE token from query param
    tokenStr := r.URL.Query().Get("token")
    if tokenStr == "" {
        http.Error(w, "Missing token", http.StatusUnauthorized)
        return
    }
    
    userID, err := h.authService.ValidateSSEToken(tokenStr)
    if err != nil {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }
    
    // 2. Set SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("X-Accel-Buffering", "no")
    
    // 3. Subscribe to notifications
    events, cleanup := h.notifService.Subscribe(r.Context(), userID)
    defer cleanup()
    
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming not supported", http.StatusInternalServerError)
        return
    }
    
    // 4. Send initial ping
    fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"connected\"}\n\n")
    flusher.Flush()
    
    // 5. Stream events
    for {
        select {
        case event, ok := <-events:
            if !ok {
                return
            }
            data, _ := json.Marshal(event.Data)
            fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Event, data)
            flusher.Flush()
            
        case <-r.Context().Done():
            return
            
        case <-time.After(30 * time.Second):
            // Keepalive ping
            fmt.Fprintf(w, "event: ping\ndata: {}\n\n")
            flusher.Flush()
        }
    }
}
```

### 5.2 Security Measures

```go
// 1. Rate Limiting untuk SSE connections per user
type SSERateLimiter struct {
    connections map[uuid.UUID]int
    maxPerUser  int // default: 3
    mu          sync.RWMutex
}

// 2. Token Validation Middleware
func SSETokenMiddleware(authService auth.Service) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := r.URL.Query().Get("token")
            userID, err := authService.ValidateSSEToken(token)
            if err != nil {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            
            ctx := context.WithValue(r.Context(), "user_id", userID)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// 3. CORS Configuration
func CORSConfig() func(http.Handler) http.Handler {
    return cors.Handler(cors.Options{
        AllowedOrigins:   []string{"https://your-frontend.com"},
        AllowedMethods:   []string{"GET"},
        AllowCredentials: true,
    })
}
```

### 5.3 SSE Token Endpoint

```go
// GET /api/v1/notifications/token
// Requires: JWT Bearer token in header
// Returns: Short-lived SSE token
func (h *NotificationHandler) GetSSEToken(w http.ResponseWriter, r *http.Request) {
    userID := getUserIDFromContext(r.Context())
    
    token, err := h.authService.GenerateSSEToken(userID)
    if err != nil {
        respondError(w, http.StatusInternalServerError, "Failed to generate token")
        return
    }
    
    respondJSON(w, http.StatusOK, map[string]string{
        "token":      token,
        "expires_in": "300", // 5 minutes
    })
}
```

---

## 6. API Endpoints

### Routes (`internal/handler/http/router.go`)

```go
r.Route("/notifications", func(r chi.Router) {
    r.Use(jwtauth.Verifier(tokenAuth))
    r.Use(jwtauth.Authenticator(tokenAuth))
    
    // Get SSE token (protected by JWT)
    r.Get("/token", notificationHandler.GetSSEToken)
    
    // SSE Stream (protected by SSE token in query param)
    r.With(SSETokenMiddleware(authService)).Get("/stream", notificationHandler.Stream)
    
    // CRUD operations (protected by JWT)
    r.Get("/", notificationHandler.List)
    r.Get("/unread-count", notificationHandler.UnreadCount)
    r.Post("/mark-read", notificationHandler.MarkAsRead)
    r.Post("/mark-all-read", notificationHandler.MarkAllAsRead)
    r.Delete("/{id}", notificationHandler.Delete)
    
    // Preferences
    r.Get("/preferences", notificationHandler.GetPreferences)
    r.Put("/preferences", notificationHandler.UpdatePreference)
})
```

### Handler (`internal/handler/http/notification.go`)

```go
package http

import (
    "encoding/json"
    "net/http"
    
    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"
    notifDomain "hris-backend-go/internal/domain/notification"
)

type NotificationHandler struct {
    notifService notifDomain.Service
    authService  auth.Service
}

func NewNotificationHandler(ns notifDomain.Service, as auth.Service) *NotificationHandler {
    return &NotificationHandler{
        notifService: ns,
        authService:  as,
    }
}

func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
    userID := getUserIDFromContext(r.Context())
    
    page := getIntQuery(r, "page", 1)
    pageSize := getIntQuery(r, "page_size", 20)
    unreadOnly := getBoolQuery(r, "unread_only", false)
    
    result, err := h.notifService.GetNotifications(r.Context(), userID, page, pageSize, unreadOnly)
    if err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    respondJSON(w, http.StatusOK, result)
}

func (h *NotificationHandler) UnreadCount(w http.ResponseWriter, r *http.Request) {
    userID := getUserIDFromContext(r.Context())
    
    count, err := h.notifService.GetUnreadCount(r.Context(), userID)
    if err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    respondJSON(w, http.StatusOK, map[string]int{"unread_count": count})
}

func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
    userID := getUserIDFromContext(r.Context())
    
    var req notifDomain.MarkAsReadRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "Invalid request body")
        return
    }
    
    if err := h.notifService.MarkAsRead(r.Context(), userID, req); err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    respondJSON(w, http.StatusOK, map[string]string{"message": "Notifications marked as read"})
}

func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
    userID := getUserIDFromContext(r.Context())
    
    if err := h.notifService.MarkAllAsRead(r.Context(), userID); err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    respondJSON(w, http.StatusOK, map[string]string{"message": "All notifications marked as read"})
}

func (h *NotificationHandler) Delete(w http.ResponseWriter, r *http.Request) {
    userID := getUserIDFromContext(r.Context())
    notifID, err := uuid.Parse(chi.URLParam(r, "id"))
    if err != nil {
        respondError(w, http.StatusBadRequest, "Invalid notification ID")
        return
    }
    
    if err := h.notifService.Delete(r.Context(), userID, notifID); err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    respondJSON(w, http.StatusOK, map[string]string{"message": "Notification deleted"})
}

func (h *NotificationHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
    userID := getUserIDFromContext(r.Context())
    
    prefs, err := h.notifService.GetPreferences(r.Context(), userID)
    if err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    respondJSON(w, http.StatusOK, prefs)
}

func (h *NotificationHandler) UpdatePreference(w http.ResponseWriter, r *http.Request) {
    userID := getUserIDFromContext(r.Context())
    
    var req notifDomain.UpdatePreferenceRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "Invalid request body")
        return
    }
    
    if err := h.notifService.UpdatePreference(r.Context(), userID, req); err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    respondJSON(w, http.StatusOK, map[string]string{"message": "Preference updated"})
}

func (h *NotificationHandler) GetSSEToken(w http.ResponseWriter, r *http.Request) {
    userID := getUserIDFromContext(r.Context())
    
    token, err := h.authService.GenerateSSEToken(userID)
    if err != nil {
        respondError(w, http.StatusInternalServerError, "Failed to generate token")
        return
    }
    
    respondJSON(w, http.StatusOK, map[string]interface{}{
        "token":      token,
        "expires_in": 300,
    })
}

func (h *NotificationHandler) Stream(w http.ResponseWriter, r *http.Request) {
    // Implementation as shown in section 5.2
}
```

---

## 7. Integration with AttendanceService

### Trigger Notification on Clock In/Out

```go
// internal/service/attendance/service.go

func (s *service) ClockIn(ctx context.Context, req attendance.ClockInRequest) (*attendance.AttendanceResponse, error) {
    // ... existing clock in logic ...
    
    // Get employee info
    employee, _ := s.employeeRepo.GetByID(ctx, req.EmployeeID)
    
    // Get managers of the company
    managers, _ := s.employeeRepo.GetManagersByCompanyID(ctx, employee.CompanyID)
    
    // Queue notifications to all managers
    for _, manager := range managers {
        s.notifService.QueueNotification(ctx, notification.CreateNotificationRequest{
            CompanyID:   employee.CompanyID,
            RecipientID: manager.UserID,
            SenderID:    &employee.UserID,
            Type:        notification.TypeAttendanceClockIn,
            Title:       "Employee Clock In",
            Message:     fmt.Sprintf("%s has clocked in at %s", employee.FullName, time.Now().Format("15:04")),
            Data: map[string]interface{}{
                "employee_id":   employee.ID,
                "attendance_id": attendance.ID,
                "clock_in_time": attendance.ClockIn,
            },
        })
    }
    
    return response, nil
}
```

---

## 8. Implementation Steps

### Phase 1: Core Infrastructure
1. [ ] Create database migrations for `notifications` and `notification_preferences` tables
2. [ ] Create `internal/domain/notification/` - entity.go, dto.go, repository.go, service.go, errors.go
3. [ ] Create `internal/pkg/sse/hub.go` - SSE Hub implementation

### Phase 2: Repository & Service
4. [ ] Create `internal/repository/postgresql/notification.go` - Repository implementation
5. [ ] Create `internal/service/notification/service.go` - Service with queue worker

### Phase 3: Authentication
6. [ ] Add `GenerateSSEToken` and `ValidateSSEToken` to auth service
7. [ ] Create SSE token middleware

### Phase 4: Handler & Routes
8. [ ] Create `internal/handler/http/notification.go` - All handlers including SSE stream
9. [ ] Update `internal/handler/http/router.go` - Add notification routes

### Phase 5: Integration
10. [ ] Add `GetManagersByCompanyID` to employee repository
11. [ ] Integrate notification service into AttendanceService (ClockIn/ClockOut)
12. [ ] Wire all dependencies in `cmd/api/main.go`

### Phase 6: Testing
13. [ ] Unit tests for notification service queue
14. [ ] Integration tests for SSE endpoint
15. [ ] End-to-end test: Clock in → Manager receives notification

---

## 9. Frontend Integration Example

```javascript
// Get SSE token first
const tokenResponse = await fetch('/api/v1/notifications/token', {
    headers: { 'Authorization': `Bearer ${jwtToken}` }
});
const { token } = await tokenResponse.json();

// Connect to SSE stream
const eventSource = new EventSource(`/api/v1/notifications/stream?token=${token}`);

eventSource.addEventListener('connected', (e) => {
    console.log('SSE connected:', e.data);
});

eventSource.addEventListener('notification', (e) => {
    const notification = JSON.parse(e.data);
    console.log('New notification:', notification);
    // Update UI, show toast, etc.
});

eventSource.addEventListener('ping', () => {
    console.log('Keepalive ping received');
});

eventSource.onerror = (e) => {
    console.error('SSE error:', e);
    eventSource.close();
    // Implement reconnection logic with exponential backoff
};
```

---

## 10. Configuration

### Environment Variables

```env
# Notification Service
NOTIFICATION_BATCH_SIZE=100
NOTIFICATION_FLUSH_INTERVAL=5s
NOTIFICATION_WORKER_COUNT=2
NOTIFICATION_QUEUE_SIZE=1000

# SSE
SSE_SECRET_KEY=your-sse-secret-key
SSE_TOKEN_EXPIRY=5m
SSE_MAX_CONNECTIONS_PER_USER=3
SSE_KEEPALIVE_INTERVAL=30s
```

---

## Summary

| Component | File Path | Status |
|-----------|-----------|--------|
| Entity | `internal/domain/notification/entity.go` | ⏳ Pending |
| DTO | `internal/domain/notification/dto.go` | ⏳ Pending |
| Repository Interface | `internal/domain/notification/repository.go` | ⏳ Pending |
| Service Interface | `internal/domain/notification/service.go` | ⏳ Pending |
| SSE Hub | `internal/pkg/sse/hub.go` | ⏳ Pending |
| Repository Impl | `internal/repository/postgresql/notification.go` | ⏳ Pending |
| Service Impl | `internal/service/notification/service.go` | ⏳ Pending |
| Handler | `internal/handler/http/notification.go` | ⏳ Pending |
| Router Update | `internal/handler/http/router.go` | ⏳ Pending |
| Main Wire | `cmd/api/main.go` | ⏳ Pending |
