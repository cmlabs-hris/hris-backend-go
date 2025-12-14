package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/notification"
	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/jwt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

// NotificationHandler defines the notification handler interface
type NotificationHandler interface {
	// Notifications
	List(w http.ResponseWriter, r *http.Request)
	UnreadCount(w http.ResponseWriter, r *http.Request)
	MarkAsRead(w http.ResponseWriter, r *http.Request)
	MarkAllAsRead(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)

	// Preferences
	GetPreferences(w http.ResponseWriter, r *http.Request)
	UpdatePreference(w http.ResponseWriter, r *http.Request)

	// SSE
	GetSSEToken(w http.ResponseWriter, r *http.Request)
	Stream(w http.ResponseWriter, r *http.Request)
}

type notificationHandlerImpl struct {
	notifService notification.Service
	jwtService   jwt.Service
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(notifService notification.Service, jwtService jwt.Service) NotificationHandler {
	return &notificationHandlerImpl{
		notifService: notifService,
		jwtService:   jwtService,
	}
}

// getUserIDFromContext extracts user_id from JWT context
func getUserIDFromContext(r *http.Request) string {
	_, claims, _ := jwtauth.FromContext(r.Context())
	if userID, ok := claims["user_id"].(string); ok {
		return userID
	}
	return ""
}

// getIntQueryParam gets an int query parameter with a default value
func getIntQueryParam(r *http.Request, key string, defaultVal int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultVal
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return intVal
}

// getBoolQueryParam gets a bool query parameter with a default value
func getBoolQueryParam(r *http.Request, key string, defaultVal bool) bool {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultVal
	}
	return val == "true" || val == "1"
}

// List returns paginated notifications for the authenticated user
func (h *notificationHandlerImpl) List(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == "" {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	page := getIntQueryParam(r, "page", 1)
	pageSize := getIntQueryParam(r, "page_size", 20)
	unreadOnly := getBoolQueryParam(r, "unread_only", false)

	result, err := h.notifService.GetNotifications(r.Context(), userID, page, pageSize, unreadOnly)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

// UnreadCount returns the count of unread notifications
func (h *notificationHandlerImpl) UnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == "" {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	count, err := h.notifService.GetUnreadCount(r.Context(), userID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, notification.UnreadCountResponse{UnreadCount: count})
}

// MarkAsRead marks specified notifications as read
func (h *notificationHandlerImpl) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == "" {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	var req notification.MarkAsReadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", nil)
		return
	}

	if len(req.NotificationIDs) == 0 {
		response.BadRequest(w, "notification_ids is required", nil)
		return
	}

	if err := h.notifService.MarkAsRead(r.Context(), userID, req); err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Notifications marked as read", nil)
}

// MarkAllAsRead marks all notifications as read
func (h *notificationHandlerImpl) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == "" {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	if err := h.notifService.MarkAllAsRead(r.Context(), userID); err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "All notifications marked as read", nil)
}

// Delete removes a notification
func (h *notificationHandlerImpl) Delete(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == "" {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	notifID := chi.URLParam(r, "id")
	if notifID == "" {
		response.BadRequest(w, "Notification ID is required", nil)
		return
	}

	if err := h.notifService.Delete(r.Context(), userID, notifID); err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Notification deleted", nil)
}

// GetPreferences retrieves notification preferences
func (h *notificationHandlerImpl) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == "" {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	prefs, err := h.notifService.GetPreferences(r.Context(), userID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, prefs)
}

// UpdatePreference updates a notification preference
func (h *notificationHandlerImpl) UpdatePreference(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == "" {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	var req notification.UpdatePreferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", nil)
		return
	}

	if req.NotificationType == "" {
		response.BadRequest(w, "notification_type is required", nil)
		return
	}

	if err := h.notifService.UpdatePreference(r.Context(), userID, req); err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Preference updated", nil)
}

// GetSSEToken generates a short-lived token for SSE connections
func (h *notificationHandlerImpl) GetSSEToken(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == "" {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	token, expiresIn, err := h.jwtService.GenerateSSEToken(userID)
	if err != nil {
		response.InternalServerError(w, "Failed to generate SSE token")
		return
	}

	response.Success(w, notification.SSETokenResponse{
		Token:     token,
		ExpiresIn: expiresIn,
	})
}

// Stream handles SSE connection for real-time notifications
func (h *notificationHandlerImpl) Stream(w http.ResponseWriter, r *http.Request) {
	// Get token from query parameter (SSE doesn't support custom headers)
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	// Validate SSE token
	userID, err := h.jwtService.ValidateSSEToken(tokenStr)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Check if streaming is supported
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Subscribe to notifications
	events, cleanup := h.notifService.Subscribe(r.Context(), userID)
	defer cleanup()

	// Send initial connection event
	fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"connected\",\"user_id\":\"%s\"}\n\n", userID)
	flusher.Flush()

	// Stream events
	keepalive := time.NewTicker(30 * time.Second)
	defer keepalive.Stop()

	for {
		select {
		case event, ok := <-events:
			if !ok {
				return
			}
			data, err := json.Marshal(event.Data)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Event, data)
			flusher.Flush()

		case <-keepalive.C:
			// Send keepalive ping
			fmt.Fprintf(w, "event: ping\ndata: {\"timestamp\":%d}\n\n", time.Now().Unix())
			flusher.Flush()

		case <-r.Context().Done():
			return
		}
	}
}
