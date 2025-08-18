package friendit

import (
	"context"
	"fmt"
	"time"
)

// ============================================================================
// Friend Request Service Interface and Implementation
// ============================================================================

// FriendRequestService handles friend request operations
type FriendRequestService[FR FriendRequestEntity] interface {
	// Friend Request Management
	SendRequest(ctx context.Context, senderID, receiverID UserID, options ...RequestOption) (FR, error)
	AcceptRequest(ctx context.Context, requestID RequestID, options ...AcceptOption) error
	RejectRequest(ctx context.Context, requestID RequestID, options ...RejectOption) error
	CancelRequest(ctx context.Context, requestID RequestID) error
	
	// Query Operations
	GetPendingRequests(ctx context.Context, userID UserID) ([]FR, error)
	GetSentRequests(ctx context.Context, userID UserID) ([]FR, error)
	GetRequestByID(ctx context.Context, requestID RequestID) (FR, error)
}

// ============================================================================
// Basic Implementation
// ============================================================================

// BasicFriendRequestService implements FriendRequestService
type BasicFriendRequestService[FR FriendRequestEntity] struct {
	repo   FriendRequestRepository[FR]
	config *ServiceConfig
}

// NewFriendRequestService creates a new friend request service
func NewFriendRequestService[FR FriendRequestEntity](
	repo FriendRequestRepository[FR],
	config *ServiceConfig,
) FriendRequestService[FR] {
	return &BasicFriendRequestService[FR]{
		repo:   repo,
		config: config,
	}
}

// ============================================================================
// Service Method Implementations
// ============================================================================

// SendRequest implements FriendRequestService.SendRequest with concurrency safety
func (s *BasicFriendRequestService[FR]) SendRequest(ctx context.Context, senderID, receiverID UserID, options ...RequestOption) (FR, error) {
	// Apply options
	config := &RequestConfig{}
	for _, opt := range options {
		opt(config)
	}
	
	// Validation logic
	if senderID == receiverID && !s.config.AllowSelfRequests {
		var empty FR
		return empty, fmt.Errorf("cannot send friend request to self")
	}
	
	// Create factory for atomic upsert operation
	factory := &createFriendRequestEntityFactory[FR]{
		senderID:   senderID,
		receiverID: receiverID,
		config:     config,
		repo:       s.repo,
	}
	
	// Atomic upsert prevents race conditions and duplicate requests
	result, err := s.repo.FindOneAndUpsert(ctx, senderID, receiverID, factory)
	if err != nil {
		var empty FR
		return empty, fmt.Errorf("failed to send friend request: %w", err)
	}
	
	return result.Entity, nil
}

// AcceptRequest implements FriendRequestService.AcceptRequest with concurrency safety
func (s *BasicFriendRequestService[FR]) AcceptRequest(ctx context.Context, requestID RequestID, options ...AcceptOption) error {
	// Apply options
	config := &AcceptConfig{}
	for _, opt := range options {
		opt(config)
	}
	
	// Create factory for atomic accept operation
	factory := &acceptFriendRequestEntityFactory[FR]{
		config: config,
	}
	
	// Atomic accept operation prevents concurrent modifications
	_, err := s.repo.AcceptIfPending(ctx, requestID, factory)
	if err != nil {
		return fmt.Errorf("failed to accept request: %w", err)
	}
	
	return nil
}

// RejectRequest implements FriendRequestService.RejectRequest with concurrency safety
func (s *BasicFriendRequestService[FR]) RejectRequest(ctx context.Context, requestID RequestID, options ...RejectOption) error {
	// Apply options
	config := &RejectConfig{}
	for _, opt := range options {
		opt(config)
	}
	
	// Create factory for atomic reject operation
	factory := &rejectFriendRequestEntityFactory[FR]{
		reason: config.Reason,
	}
	
	// Atomic reject operation prevents concurrent modifications
	_, err := s.repo.RejectIfPending(ctx, requestID, factory)
	if err != nil {
		return fmt.Errorf("failed to reject request: %w", err)
	}
	
	return nil
}

// CancelRequest implements FriendRequestService.CancelRequest with concurrency safety
func (s *BasicFriendRequestService[FR]) CancelRequest(ctx context.Context, requestID RequestID) error {
	// Atomic cancel operation
	err := s.repo.CancelIfPending(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to cancel request: %w", err)
	}
	
	return nil
}

// GetPendingRequests implements FriendRequestService.GetPendingRequests
func (s *BasicFriendRequestService[FR]) GetPendingRequests(ctx context.Context, userID UserID) ([]FR, error) {
	return s.repo.GetByStatus(ctx, userID, "pending")
}

// GetSentRequests implements FriendRequestService.GetSentRequests
func (s *BasicFriendRequestService[FR]) GetSentRequests(ctx context.Context, userID UserID) ([]FR, error) {
	return s.repo.GetBySenderID(ctx, userID)
}

// GetRequestByID implements FriendRequestService.GetRequestByID
func (s *BasicFriendRequestService[FR]) GetRequestByID(ctx context.Context, requestID RequestID) (FR, error) {
	return s.repo.GetByID(ctx, requestID)
}

// ============================================================================
// Entity Factories for Atomic Operations
// ============================================================================

// createFriendRequestEntityFactory creates new friend request entities
type createFriendRequestEntityFactory[FR FriendRequestEntity] struct {
	senderID   UserID
	receiverID UserID
	config     *RequestConfig
	repo       FriendRequestRepository[FR]
}

func (f *createFriendRequestEntityFactory[FR]) CreateFn(ctx context.Context) (FR, error) {
	request := f.repo.NewEntity()
	request.SetSenderID(f.senderID)
	request.SetReceiverID(f.receiverID)
	request.SetStatus("pending")
	request.SetCreatedAt(time.Now())
	request.SetUpdatedAt(time.Now())
	
	// Apply configuration from options
	if f.config.Message != "" {
		request.SetMessage(f.config.Message)
	}
	if f.config.ExpiresAt != nil {
		request.SetExpiresAt(*f.config.ExpiresAt)
	} else {
		// Default expiration: 7 days
		expiresAt := time.Now().Add(7 * 24 * time.Hour)
		request.SetExpiresAt(expiresAt)
	}
	if f.config.Priority != "" {
		request.SetPriority(f.config.Priority)
	}
	if len(f.config.Metadata) > 0 {
		request.SetMetadata(f.config.Metadata)
	}
	
	return request, nil
}

func (f *createFriendRequestEntityFactory[FR]) UpdateFn(ctx context.Context, existing FR) (FR, error) {
	// If already exists and pending, return as-is or update message
	if existing.GetStatus() == "pending" && !existing.IsExpired() {
		if f.config.Message != "" {
			existing.SetMessage(f.config.Message)
			existing.SetUpdatedAt(time.Now())
		}
		return existing, nil
	}
	
	// If expired or rejected, create new request
	return f.CreateFn(ctx)
}

// acceptFriendRequestEntityFactory accepts pending requests
type acceptFriendRequestEntityFactory[FR FriendRequestEntity] struct {
	config *AcceptConfig
}

func (f *acceptFriendRequestEntityFactory[FR]) CreateFn(ctx context.Context) (FR, error) {
	var empty FR
	return empty, fmt.Errorf("cannot create new friend request with accept factory")
}

func (f *acceptFriendRequestEntityFactory[FR]) UpdateFn(ctx context.Context, existing FR) (FR, error) {
	if existing.GetStatus() != "pending" {
		return existing, fmt.Errorf("cannot accept non-pending request: status=%s", existing.GetStatus())
	}
	if existing.IsExpired() {
		return existing, fmt.Errorf("cannot accept expired request")
	}
	
	existing.SetStatus("accepted")
	existing.SetUpdatedAt(time.Now())
	
	if len(f.config.Metadata) > 0 {
		existing.SetMetadata(f.config.Metadata)
	}
	
	return existing, nil
}

// rejectFriendRequestEntityFactory rejects pending requests
type rejectFriendRequestEntityFactory[FR FriendRequestEntity] struct {
	reason string
}

func (f *rejectFriendRequestEntityFactory[FR]) CreateFn(ctx context.Context) (FR, error) {
	var empty FR
	return empty, fmt.Errorf("cannot create new friend request with reject factory")
}

func (f *rejectFriendRequestEntityFactory[FR]) UpdateFn(ctx context.Context, existing FR) (FR, error) {
	if existing.GetStatus() != "pending" {
		return existing, fmt.Errorf("cannot reject non-pending request: status=%s", existing.GetStatus())
	}
	
	existing.SetStatus("rejected")
	existing.SetUpdatedAt(time.Now())
	
	if f.reason != "" {
		// Store rejection reason in metadata
		metadata := make(map[string]any)
		metadata["rejection_reason"] = f.reason
		existing.SetMetadata(metadata)
	}
	
	return existing, nil
}

// ============================================================================
// Fluent API Builder
// ============================================================================

// Request returns a friend request builder
func (s *BasicFriendRequestService[FR]) Request() *RequestBuilder[FR] {
	return &RequestBuilder[FR]{service: s}
}

// RequestBuilder provides fluent interface for friend requests
type RequestBuilder[FR FriendRequestEntity] struct {
	service    FriendRequestService[FR]
	senderID   UserID
	receiverID UserID
	message    string
	metadata   map[string]any
	expiresAt  *time.Time
	priority   string
}

// From sets the sender ID
func (rb *RequestBuilder[FR]) From(senderID UserID) *RequestBuilder[FR] {
	rb.senderID = senderID
	return rb
}

// To sets the receiver ID  
func (rb *RequestBuilder[FR]) To(receiverID UserID) *RequestBuilder[FR] {
	rb.receiverID = receiverID
	return rb
}

// WithMessage sets the request message
func (rb *RequestBuilder[FR]) WithMessage(message string) *RequestBuilder[FR] {
	rb.message = message
	return rb
}

// WithMetadata adds metadata to the request
func (rb *RequestBuilder[FR]) WithMetadata(metadata map[string]any) *RequestBuilder[FR] {
	if rb.metadata == nil {
		rb.metadata = make(map[string]any)
	}
	for k, v := range metadata {
		rb.metadata[k] = v
	}
	return rb
}

// WithExpiry sets when the request expires
func (rb *RequestBuilder[FR]) WithExpiry(expiresAt time.Time) *RequestBuilder[FR] {
	rb.expiresAt = &expiresAt
	return rb
}

// WithPriority sets request priority
func (rb *RequestBuilder[FR]) WithPriority(priority string) *RequestBuilder[FR] {
	rb.priority = priority
	return rb
}

// Send creates and sends the friend request
func (rb *RequestBuilder[FR]) Send(ctx context.Context) (FR, error) {
	// Build request options from builder state
	options := []RequestOption{}
	
	if rb.message != "" {
		options = append(options, WithMessage(rb.message))
	}
	if rb.priority != "" {
		options = append(options, WithPriority(rb.priority))
	}
	if rb.expiresAt != nil {
		options = append(options, WithExpiry(*rb.expiresAt))
	}
	if len(rb.metadata) > 0 {
		options = append(options, WithMetadata(rb.metadata))
	}
	
	return rb.service.SendRequest(ctx, rb.senderID, rb.receiverID, options...)
}