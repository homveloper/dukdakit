package friendit

import (
	"context"
	"sync"
)

// ============================================================================
// Event System Interface
// ============================================================================

// EventEmitter defines the interface for event emission and handling
type EventEmitter interface {
	// Event registration
	On(event string, handler EventHandler)
	Once(event string, handler EventHandler) 
	Off(event string, handler EventHandler)
	
	// Event emission
	Emit(event string, data map[string]any)
	EmitAsync(event string, data map[string]any)
	
	// Event management
	ListenerCount(event string) int
	Events() []string
	RemoveAllListeners(event ...string)
}

// EventHandler defines the signature for event handlers
type EventHandler func(ctx context.Context, event string, data map[string]any)

// EventMiddleware allows processing events before they reach handlers
type EventMiddleware func(ctx context.Context, event string, data map[string]any, next func())

// ============================================================================
// Basic Event Emitter Implementation
// ============================================================================

// BasicEventEmitter provides a thread-safe event emitter implementation
type BasicEventEmitter struct {
	mu          sync.RWMutex
	handlers    map[string][]EventHandler
	onceHandlers map[string][]EventHandler
	middlewares []EventMiddleware
}

// NewEventEmitter creates a new event emitter
func NewEventEmitter() *BasicEventEmitter {
	return &BasicEventEmitter{
		handlers:     make(map[string][]EventHandler),
		onceHandlers: make(map[string][]EventHandler),
		middlewares:  make([]EventMiddleware, 0),
	}
}

// On adds a persistent event handler
func (e *BasicEventEmitter) On(event string, handler EventHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.handlers[event] = append(e.handlers[event], handler)
}

// Once adds a one-time event handler
func (e *BasicEventEmitter) Once(event string, handler EventHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.onceHandlers[event] = append(e.onceHandlers[event], handler)
}

// Off removes an event handler
func (e *BasicEventEmitter) Off(event string, handler EventHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	// Remove from persistent handlers
	handlers := e.handlers[event]
	for i, h := range handlers {
		if &h == &handler { // Compare function pointers
			e.handlers[event] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
	
	// Remove from once handlers
	onceHandlers := e.onceHandlers[event]
	for i, h := range onceHandlers {
		if &h == &handler {
			e.onceHandlers[event] = append(onceHandlers[:i], onceHandlers[i+1:]...)
			break
		}
	}
}

// Emit synchronously emits an event
func (e *BasicEventEmitter) Emit(event string, data map[string]any) {
	e.emitEvent(context.Background(), event, data, false)
}

// EmitAsync asynchronously emits an event
func (e *BasicEventEmitter) EmitAsync(event string, data map[string]any) {
	go e.emitEvent(context.Background(), event, data, true)
}

// emitEvent is the internal event emission logic
func (e *BasicEventEmitter) emitEvent(ctx context.Context, event string, data map[string]any, async bool) {
	e.mu.RLock()
	
	// Get handlers
	handlers := make([]EventHandler, len(e.handlers[event]))
	copy(handlers, e.handlers[event])
	
	onceHandlers := make([]EventHandler, len(e.onceHandlers[event]))
	copy(onceHandlers, e.onceHandlers[event])
	
	middlewares := make([]EventMiddleware, len(e.middlewares))
	copy(middlewares, e.middlewares)
	
	e.mu.RUnlock()
	
	// Clear once handlers
	if len(onceHandlers) > 0 {
		e.mu.Lock()
		delete(e.onceHandlers, event)
		e.mu.Unlock()
	}
	
	// Execute middlewares and handlers
	allHandlers := append(handlers, onceHandlers...)
	
	if len(middlewares) > 0 {
		e.executeWithMiddlewares(ctx, event, data, allHandlers, middlewares)
	} else {
		e.executeHandlers(ctx, event, data, allHandlers, async)
	}
}

// executeWithMiddlewares runs handlers through middleware chain
func (e *BasicEventEmitter) executeWithMiddlewares(
	ctx context.Context, 
	event string, 
	data map[string]any, 
	handlers []EventHandler, 
	middlewares []EventMiddleware,
) {
	var runMiddleware func(int)
	
	runMiddleware = func(index int) {
		if index >= len(middlewares) {
			e.executeHandlers(ctx, event, data, handlers, false)
			return
		}
		
		middleware := middlewares[index]
		middleware(ctx, event, data, func() {
			runMiddleware(index + 1)
		})
	}
	
	runMiddleware(0)
}

// executeHandlers runs the actual event handlers
func (e *BasicEventEmitter) executeHandlers(
	ctx context.Context, 
	event string, 
	data map[string]any, 
	handlers []EventHandler, 
	async bool,
) {
	for _, handler := range handlers {
		if async {
			go handler(ctx, event, data)
		} else {
			handler(ctx, event, data)
		}
	}
}

// ListenerCount returns the number of listeners for an event
func (e *BasicEventEmitter) ListenerCount(event string) int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.handlers[event]) + len(e.onceHandlers[event])
}

// Events returns all event names that have listeners
func (e *BasicEventEmitter) Events() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	eventSet := make(map[string]bool)
	
	for event := range e.handlers {
		eventSet[event] = true
	}
	for event := range e.onceHandlers {
		eventSet[event] = true
	}
	
	events := make([]string, 0, len(eventSet))
	for event := range eventSet {
		events = append(events, event)
	}
	
	return events
}

// RemoveAllListeners removes all listeners for specified events (or all if none specified)
func (e *BasicEventEmitter) RemoveAllListeners(events ...string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if len(events) == 0 {
		// Remove all
		e.handlers = make(map[string][]EventHandler)
		e.onceHandlers = make(map[string][]EventHandler)
	} else {
		// Remove specific events
		for _, event := range events {
			delete(e.handlers, event)
			delete(e.onceHandlers, event)
		}
	}
}

// Use adds middleware to the event processing chain
func (e *BasicEventEmitter) Use(middleware EventMiddleware) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.middlewares = append(e.middlewares, middleware)
}

// ============================================================================
// Pre-defined Event Types
// ============================================================================

const (
	// Friend Request Events
	EventFriendRequestSent     = "friend_request_sent"
	EventFriendRequestReceived = "friend_request_received"
	EventFriendRequestAccepted = "friend_request_accepted"
	EventFriendRequestRejected = "friend_request_rejected"
	EventFriendRequestCanceled = "friend_request_canceled"
	EventFriendRequestExpired  = "friend_request_expired"
	
	// Friendship Events
	EventFriendAdded   = "friend_added"
	EventFriendRemoved = "friend_removed"
	EventFriendOnline  = "friend_online"
	EventFriendOffline = "friend_offline"
	
	// Block Events
	EventUserBlocked   = "user_blocked"
	EventUserUnblocked = "user_unblocked"
	
	// Status Events
	EventStatusChanged = "status_changed"
)

// ============================================================================
// Event Data Structures
// ============================================================================

// FriendRequestEventData contains data for friend request events
type FriendRequestEventData struct {
	RequestID  RequestID `json:"request_id"`
	SenderID   UserID    `json:"sender_id"`
	ReceiverID UserID    `json:"receiver_id"`
	Message    string    `json:"message,omitempty"`
	Timestamp  int64     `json:"timestamp"`
}

// FriendshipEventData contains data for friendship events
type FriendshipEventData struct {
	FriendshipID FriendshipID `json:"friendship_id"`
	User1ID      UserID       `json:"user1_id"`
	User2ID      UserID       `json:"user2_id"`
	Timestamp    int64        `json:"timestamp"`
}

// BlockEventData contains data for block events
type BlockEventData struct {
	BlockID   BlockID `json:"block_id"`
	BlockerID UserID  `json:"blocker_id"`
	BlockedID UserID  `json:"blocked_id"`
	Reason    string  `json:"reason,omitempty"`
	Timestamp int64   `json:"timestamp"`
}

// StatusEventData contains data for status change events
type StatusEventData struct {
	UserID    UserID `json:"user_id"`
	OldStatus string `json:"old_status"`
	NewStatus string `json:"new_status"`
	Timestamp int64  `json:"timestamp"`
}

// ============================================================================
// Event Builder Pattern for Fluent API
// ============================================================================

// EventBuilder provides a fluent interface for creating events
type EventBuilder struct {
	emitter   EventEmitter
	event     string
	data      map[string]any
	condition func() bool
}

// NewEventBuilder creates a new event builder
func NewEventBuilder(emitter EventEmitter, event string) *EventBuilder {
	return &EventBuilder{
		emitter: emitter,
		event:   event,
		data:    make(map[string]any),
		condition: func() bool { return true },
	}
}

// With adds data to the event
func (eb *EventBuilder) With(key string, value any) *EventBuilder {
	eb.data[key] = value
	return eb
}

// WithData adds multiple data fields
func (eb *EventBuilder) WithData(data map[string]any) *EventBuilder {
	for k, v := range data {
		eb.data[k] = v
	}
	return eb
}

// When sets a condition for emitting the event
func (eb *EventBuilder) When(condition func() bool) *EventBuilder {
	eb.condition = condition
	return eb
}

// Emit emits the event if condition is met
func (eb *EventBuilder) Emit() {
	if eb.condition() {
		eb.emitter.Emit(eb.event, eb.data)
	}
}

// EmitAsync asynchronously emits the event if condition is met
func (eb *EventBuilder) EmitAsync() {
	if eb.condition() {
		eb.emitter.EmitAsync(eb.event, eb.data)
	}
}