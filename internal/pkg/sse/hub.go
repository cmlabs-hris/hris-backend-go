package sse

import (
	"sync"
)

// Event represents an SSE event to be sent to subscribers
type Event struct {
	UserID string
	Event  string
	Data   interface{}
}

// Hub manages SSE subscribers and event broadcasting
type Hub struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan Event]struct{}
}

// NewHub creates a new SSE Hub instance
func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[string]map[chan Event]struct{}),
	}
}

// Subscribe registers a new subscriber for a user and returns the event channel and cleanup function
func (h *Hub) Subscribe(userID string) (chan Event, func()) {
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

// Publish sends an event to all subscribers of a specific user
func (h *Hub) Publish(userID string, event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if subs, ok := h.subscribers[userID]; ok {
		for ch := range subs {
			select {
			case ch <- event:
			default:
				// Skip if channel is full (non-blocking to prevent deadlock)
			}
		}
	}
}

// PublishToMany sends an event to multiple users
func (h *Hub) PublishToMany(userIDs []string, event Event) {
	for _, userID := range userIDs {
		eventCopy := event
		eventCopy.UserID = userID
		h.Publish(userID, eventCopy)
	}
}

// SubscriberCount returns the number of active subscribers for a user
func (h *Hub) SubscriberCount(userID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if subs, ok := h.subscribers[userID]; ok {
		return len(subs)
	}
	return 0
}

// TotalSubscribers returns the total number of active subscribers across all users
func (h *Hub) TotalSubscribers() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	total := 0
	for _, subs := range h.subscribers {
		total += len(subs)
	}
	return total
}
