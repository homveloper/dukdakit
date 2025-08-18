package friendit

import (
	"time"
)

// ============================================================================
// Request/Response Options
// ============================================================================

// RequestOption configures friend request creation
type RequestOption func(*RequestConfig)

// RequestConfig holds options for creating friend requests
type RequestConfig struct {
	Message      string
	Priority     string
	ExpiresAt    *time.Time
	Metadata     map[string]any
	AutoAccept   any // 타입 안전성을 위해 any로 변경
}

// WithMessage adds a message to the friend request
func WithMessage(message string) RequestOption {
	return func(c *RequestConfig) { c.Message = message }
}

// WithPriority sets the priority of the request
func WithPriority(priority string) RequestOption {
	return func(c *RequestConfig) { c.Priority = priority }
}

// WithExpiry sets when the request expires
func WithExpiry(expiresAt time.Time) RequestOption {
	return func(c *RequestConfig) { c.ExpiresAt = &expiresAt }
}

// WithMetadata adds metadata to the request
func WithMetadata(metadata map[string]any) RequestOption {
	return func(c *RequestConfig) { c.Metadata = metadata }
}

// WithAutoAccept sets a function to determine if the request should be auto-accepted
func WithAutoAccept(fn any) RequestOption {
	return func(c *RequestConfig) { c.AutoAccept = fn }
}

// ============================================================================
// Accept/Reject Options
// ============================================================================

type AcceptOption func(*AcceptConfig)
type AcceptConfig struct {
	Metadata map[string]any
}

type RejectOption func(*RejectConfig)
type RejectConfig struct {
	Reason string
}

type RemoveOption func(*RemoveConfig)
type RemoveConfig struct {
	Reason string
	Notify bool
}

// ============================================================================
// Block Options
// ============================================================================

type BlockOption func(*BlockConfig)
type BlockConfig struct {
	Reason   string
	Duration *time.Duration
}

// ============================================================================
// Search Options
// ============================================================================

type SearchOption func(*SearchConfig)
type SearchConfig struct {
	Limit      int
	OnlineOnly bool
	ExcludeBlocked bool
}

type RecommendOption func(*RecommendConfig)
type RecommendConfig struct {
	Limit     int
	Algorithm string
	Filters   []string
}

// ============================================================================
// Friendship Options
// ============================================================================

type FriendshipOption func(*FriendshipConfig)
type FriendshipConfig struct {
	Source   string
	Metadata map[string]any
}

// ============================================================================
// Filter Interface
// ============================================================================

type Filter interface {
	Apply(query any) any
}

// ============================================================================
// Result Types
// ============================================================================

// RequestResult contains the result of a friend request operation
type RequestResult struct {
	SenderID   UserID         `json:"sender_id"`
	ReceiverID UserID         `json:"receiver_id"`
	Message    string         `json:"message,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	ExpiresAt  *time.Time     `json:"expires_at,omitempty"`
	Priority   string         `json:"priority,omitempty"`
	Validated  bool           `json:"validated"`
}