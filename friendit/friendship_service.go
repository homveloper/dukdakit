package friendit

import (
	"context"
	"fmt"
	"time"
)

// ============================================================================
// Friendship Service Interface and Implementation
// ============================================================================

// FriendshipService handles friendship operations
type FriendshipService[U UserEntity, F FriendshipEntity] interface {
	// Friendship Management
	CreateFriendship(ctx context.Context, user1ID, user2ID UserID, options ...FriendshipOption) (F, error)
	RemoveFriendship(ctx context.Context, userID, friendID UserID, options ...RemoveOption) error
	GetFriendship(ctx context.Context, user1ID, user2ID UserID) (F, error)
	
	// Query Operations
	GetFriends(ctx context.Context, userID UserID, filters ...Filter) ([]Friend[U, F], error)
	GetMutualFriends(ctx context.Context, user1ID, user2ID UserID) ([]Friend[U, F], error)
	GetOnlineFriends(ctx context.Context, userID UserID) ([]Friend[U, F], error)
	
	// Statistics
	GetFriendStats(ctx context.Context, userID UserID) (*FriendStats, error)
}

// ============================================================================
// Basic Implementation
// ============================================================================

// BasicFriendshipService implements FriendshipService
type BasicFriendshipService[U UserEntity, F FriendshipEntity] struct {
	userRepo       UserRepository[U]
	friendshipRepo FriendshipRepository[F]
	config         *ServiceConfig
}

// NewFriendshipService creates a new friendship service
func NewFriendshipService[U UserEntity, F FriendshipEntity](
	userRepo UserRepository[U],
	friendshipRepo FriendshipRepository[F],
	config *ServiceConfig,
) FriendshipService[U, F] {
	return &BasicFriendshipService[U, F]{
		userRepo:       userRepo,
		friendshipRepo: friendshipRepo,
		config:         config,
	}
}

// ============================================================================
// Service Method Implementations
// ============================================================================

// CreateFriendship implements FriendshipService.CreateFriendship with concurrency safety
func (s *BasicFriendshipService[U, F]) CreateFriendship(ctx context.Context, user1ID, user2ID UserID, options ...FriendshipOption) (F, error) {
	// Apply options
	config := &FriendshipConfig{}
	for _, opt := range options {
		opt(config)
	}
	
	// Validation
	if user1ID == user2ID {
		var empty F
		return empty, fmt.Errorf("cannot create friendship with self")
	}
	
	// Create factory for atomic upsert operation
	factory := &createFriendshipEntityFactory[F]{
		user1ID: user1ID,
		user2ID: user2ID,
		config:  config,
		repo:    s.friendshipRepo,
	}
	
	// Atomic upsert prevents race conditions and duplicate friendships
	result, err := s.friendshipRepo.FindOneAndUpsert(ctx, user1ID, user2ID, factory)
	if err != nil {
		var empty F
		return empty, fmt.Errorf("failed to create friendship: %w", err)
	}
	
	return result.Entity, nil
}

// RemoveFriendship implements FriendshipService.RemoveFriendship with concurrency safety
func (s *BasicFriendshipService[U, F]) RemoveFriendship(ctx context.Context, userID, friendID UserID, options ...RemoveOption) error {
	// Apply options
	config := &RemoveConfig{}
	for _, opt := range options {
		opt(config)
	}
	
	// Atomic delete operation prevents race conditions
	err := s.friendshipRepo.DeleteByUsers(ctx, userID, friendID)
	if err != nil {
		return fmt.Errorf("failed to remove friendship: %w", err)
	}
	
	// Note: Notification logic would be handled by event system
	// if config.Notify { /* trigger notification event */ }
	
	return nil
}

// GetFriendship implements FriendshipService.GetFriendship
func (s *BasicFriendshipService[U, F]) GetFriendship(ctx context.Context, user1ID, user2ID UserID) (F, error) {
	return s.friendshipRepo.GetByUsers(ctx, user1ID, user2ID)
}

// GetFriends implements FriendshipService.GetFriends
func (s *BasicFriendshipService[U, F]) GetFriends(ctx context.Context, userID UserID, filters ...Filter) ([]Friend[U, F], error) {
	// Get friendships
	friendships, err := s.friendshipRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get friendships: %w", err)
	}
	
	// Convert to Friend structs with user data
	friends := make([]Friend[U, F], 0, len(friendships))
	for _, friendship := range friendships {
		// Determine which user is the friend (not the requesting user)
		friendUserID := friendship.GetUser1ID()
		if friendUserID == userID {
			friendUserID = friendship.GetUser2ID()
		}
		
		// Get friend user data
		friendUser, err := s.userRepo.GetByID(ctx, friendUserID)
		if err != nil {
			continue // Skip if user not found
		}
		
		// Create Friend struct
		friend := Friend[U, F]{
			User:       friendUser,
			Friendship: friendship,
			IsOnline:   friendUser.GetStatus() == "online",
			IsMutual:   true, // All friendships are mutual by default
		}
		
		friends = append(friends, friend)
	}
	
	// Apply filters if provided
	for _, filter := range filters {
		// Users can implement custom filtering logic
		_ = filter // Placeholder for filter application
	}
	
	return friends, nil
}

// GetMutualFriends implements FriendshipService.GetMutualFriends
func (s *BasicFriendshipService[U, F]) GetMutualFriends(ctx context.Context, user1ID, user2ID UserID) ([]Friend[U, F], error) {
	// Get friends of both users
	user1Friends, err := s.GetFriends(ctx, user1ID)
	if err != nil {
		return nil, err
	}
	
	user2Friends, err := s.GetFriends(ctx, user2ID)
	if err != nil {
		return nil, err
	}
	
	// Find mutual friends
	mutualFriends := []Friend[U, F]{}
	user2FriendMap := make(map[UserID]Friend[U, F])
	
	for _, friend := range user2Friends {
		user2FriendMap[friend.User.GetID()] = friend
	}
	
	for _, friend := range user1Friends {
		if mutualFriend, exists := user2FriendMap[friend.User.GetID()]; exists {
			mutualFriend.IsMutual = true
			mutualFriends = append(mutualFriends, mutualFriend)
		}
	}
	
	return mutualFriends, nil
}

// GetOnlineFriends implements FriendshipService.GetOnlineFriends
func (s *BasicFriendshipService[U, F]) GetOnlineFriends(ctx context.Context, userID UserID) ([]Friend[U, F], error) {
	// Get all friends
	friends, err := s.GetFriends(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	// Filter for online friends
	onlineFriends := []Friend[U, F]{}
	for _, friend := range friends {
		if friend.User.GetStatus() == "online" {
			friend.IsOnline = true
			onlineFriends = append(onlineFriends, friend)
		}
	}
	
	return onlineFriends, nil
}

// GetFriendStats implements FriendshipService.GetFriendStats
func (s *BasicFriendshipService[U, F]) GetFriendStats(ctx context.Context, userID UserID) (*FriendStats, error) {
	// Get all friends
	friends, err := s.GetFriends(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	// Count online friends
	onlineCount := 0
	for _, friend := range friends {
		if friend.User.GetStatus() == "online" {
			onlineCount++
		}
	}
	
	// Note: Other stats would require additional repository methods
	// This is a basic implementation that users can extend
	stats := &FriendStats{
		TotalFriends:    len(friends),
		OnlineFriends:   onlineCount,
		PendingRequests: 0, // Would need FriendRequestService integration
		SentRequests:    0, // Would need FriendRequestService integration
		BlockedUsers:    0, // Would need BlockService integration
		MutualFriends:   len(friends), // All friendships are mutual by default
		RecentActivity:  time.Now(),   // Would be updated with actual activity tracking
	}
	
	return stats, nil
}

// ============================================================================
// Entity Factories for Atomic Operations
// ============================================================================

// createFriendshipEntityFactory creates new friendship entities
type createFriendshipEntityFactory[F FriendshipEntity] struct {
	user1ID UserID
	user2ID UserID
	config  *FriendshipConfig
	repo    FriendshipRepository[F]
}

func (f *createFriendshipEntityFactory[F]) CreateFn(ctx context.Context) (F, error) {
	friendship := f.repo.NewEntity()
	
	// Normalize user order for consistent indexing
	user1ID, user2ID := f.user1ID, f.user2ID
	if user1ID > user2ID {
		user1ID, user2ID = user2ID, user1ID
	}
	
	friendship.SetUser1ID(user1ID)
	friendship.SetUser2ID(user2ID)
	friendship.SetStatus("active")
	friendship.SetCreatedAt(time.Now())
	friendship.SetUpdatedAt(time.Now())
	
	if f.config.Source != "" {
		friendship.SetSource(f.config.Source)
	} else {
		friendship.SetSource("direct")
	}
	
	if len(f.config.Metadata) > 0 {
		friendship.SetMetadata(f.config.Metadata)
	}
	
	return friendship, nil
}

func (f *createFriendshipEntityFactory[F]) UpdateFn(ctx context.Context, existing F) (F, error) {
	// If friendship exists and is active, return as-is
	if existing.GetStatus() == "active" {
		return existing, nil
	}
	
	// If friendship exists but inactive, reactivate it
	existing.SetStatus("active")
	existing.SetUpdatedAt(time.Now())
	
	// Update source and metadata if provided
	if f.config.Source != "" {
		existing.SetSource(f.config.Source)
	}
	if len(f.config.Metadata) > 0 {
		existing.SetMetadata(f.config.Metadata)
	}
	
	return existing, nil
}

// ============================================================================
// Fluent API Builder
// ============================================================================

// Filter returns a friend filter builder
func (s *BasicFriendshipService[U, F]) Filter() *FilterBuilder[U, F] {
	return &FilterBuilder[U, F]{service: s}
}

// FilterBuilder provides fluent interface for filtering friends
type FilterBuilder[U UserEntity, F FriendshipEntity] struct {
	service       FriendshipService[U, F]
	userID        UserID
	statuses      []string
	onlineOnly    bool
	limit         int
	offset        int
	customFilters map[string]any
}

// User sets the user ID to get friends for
func (fb *FilterBuilder[U, F]) User(userID UserID) *FilterBuilder[U, F] {
	fb.userID = userID
	return fb
}

// Status filters by friendship status
func (fb *FilterBuilder[U, F]) Status(statuses ...string) *FilterBuilder[U, F] {
	fb.statuses = append(fb.statuses, statuses...)
	return fb
}

// Online filters for online friends only
func (fb *FilterBuilder[U, F]) Online() *FilterBuilder[U, F] {
	fb.onlineOnly = true
	return fb
}

// Limit sets the maximum number of results
func (fb *FilterBuilder[U, F]) Limit(limit int) *FilterBuilder[U, F] {
	fb.limit = limit
	return fb
}

// Get executes the filter and returns results
func (fb *FilterBuilder[U, F]) Get(ctx context.Context) ([]Friend[U, F], error) {
	// Build filters from builder state
	filters := []Filter{}
	// Note: Custom filter implementation would go here
	// Users can implement specific Filter types based on their needs
	
	return fb.service.GetFriends(ctx, fb.userID, filters...)
}

// ============================================================================
// Helper Types
// ============================================================================

// Friend represents a user in the context of friendship
type Friend[U UserEntity, F FriendshipEntity] struct {
	User       U
	Friendship F
	IsMutual   bool
	IsOnline   bool
}

// FriendStats provides statistics about a user's friendships
type FriendStats struct {
	TotalFriends     int
	OnlineFriends    int
	PendingRequests  int
	SentRequests     int
	BlockedUsers     int
	MutualFriends    int
	RecentActivity   time.Time
}