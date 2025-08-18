package friendit

import (
	"context"
	"fmt"
	"time"
)

// ============================================================================
// User Service Interface and Implementation
// ============================================================================

// UserService handles user-related operations
type UserService[U UserEntity] interface {
	// User Management
	GetUser(ctx context.Context, userID UserID) (U, error)
	UpdateUserStatus(ctx context.Context, userID UserID, status string) error
	
	// Search and Discovery
	SearchUsers(ctx context.Context, query string, options ...SearchOption) ([]U, error)
	GetRecommendations(ctx context.Context, userID UserID, options ...RecommendOption) ([]U, error)
	
	// Status Operations
	GetOnlineUsers(ctx context.Context) ([]U, error)
	GetUsersByStatus(ctx context.Context, status string) ([]U, error)
}

// ============================================================================
// Basic Implementation
// ============================================================================

// BasicUserService implements UserService
type BasicUserService[U UserEntity] struct {
	repo   UserRepository[U]
	config *ServiceConfig
}

// NewUserService creates a new user service
func NewUserService[U UserEntity](
	repo UserRepository[U],
	config *ServiceConfig,
) UserService[U] {
	return &BasicUserService[U]{
		repo:   repo,
		config: config,
	}
}

// ============================================================================
// Service Method Implementations
// ============================================================================

// GetUser implements UserService.GetUser
func (s *BasicUserService[U]) GetUser(ctx context.Context, userID UserID) (U, error) {
	return s.repo.GetByID(ctx, userID)
}

// UpdateUserStatus implements UserService.UpdateUserStatus with concurrency safety
func (s *BasicUserService[U]) UpdateUserStatus(ctx context.Context, userID UserID, status string) error {
	// Create factory for atomic status update
	factory := &updateUserStatusEntityFactory[U]{
		newStatus: status,
		repo:      s.repo,
	}
	
	// Atomic update prevents race conditions
	_, err := s.repo.FindOneAndUpdate(ctx, userID, factory)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}
	
	return nil
}

// SearchUsers implements UserService.SearchUsers
func (s *BasicUserService[U]) SearchUsers(ctx context.Context, query string, options ...SearchOption) ([]U, error) {
	// Apply options
	config := &SearchConfig{
		Limit: 50, // Default limit
	}
	for _, opt := range options {
		opt(config)
	}
	
	// Perform search - this is a basic implementation
	// Users would typically implement more sophisticated search logic
	users, err := s.repo.Search(ctx, query, config.Limit)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	
	// Apply filters
	filteredUsers := []U{}
	for _, user := range users {
		// Skip blocked users if requested
		if config.ExcludeBlocked {
			// Note: This would require BlockService integration
			// For now, just include all users
		}
		
		// Filter for online only if requested
		if config.OnlineOnly && user.GetStatus() != "online" {
			continue
		}
		
		filteredUsers = append(filteredUsers, user)
	}
	
	return filteredUsers, nil
}

// GetRecommendations implements UserService.GetRecommendations
func (s *BasicUserService[U]) GetRecommendations(ctx context.Context, userID UserID, options ...RecommendOption) ([]U, error) {
	// Apply options
	config := &RecommendConfig{
		Limit:     10, // Default limit
		Algorithm: "basic", // Default algorithm
	}
	for _, opt := range options {
		opt(config)
	}
	
	// Basic recommendation logic - users can implement more sophisticated algorithms
	// This is a placeholder implementation that returns recent users
	users, err := s.repo.GetRecent(ctx, config.Limit*2) // Get more to allow filtering
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendations: %w", err)
	}
	
	// Filter out the requesting user and apply limits
	recommendations := []U{}
	for _, user := range users {
		if user.GetID() == userID {
			continue // Skip self
		}
		
		// Apply custom filters based on algorithm
		if config.Algorithm == "online_only" && user.GetStatus() != "online" {
			continue
		}
		
		recommendations = append(recommendations, user)
		if len(recommendations) >= config.Limit {
			break
		}
	}
	
	return recommendations, nil
}

// GetOnlineUsers implements UserService.GetOnlineUsers
func (s *BasicUserService[U]) GetOnlineUsers(ctx context.Context) ([]U, error) {
	return s.repo.GetByStatus(ctx, "online")
}

// GetUsersByStatus implements UserService.GetUsersByStatus
func (s *BasicUserService[U]) GetUsersByStatus(ctx context.Context, status string) ([]U, error) {
	return s.repo.GetByStatus(ctx, status)
}

// ============================================================================
// Entity Factories for Atomic Operations
// ============================================================================

// updateUserStatusEntityFactory updates user status atomically
type updateUserStatusEntityFactory[U UserEntity] struct {
	newStatus string
	repo      UserRepository[U]
}

func (f *updateUserStatusEntityFactory[U]) CreateFn(ctx context.Context) (U, error) {
	var empty U
	return empty, fmt.Errorf("cannot create new user with status update factory")
}

func (f *updateUserStatusEntityFactory[U]) UpdateFn(ctx context.Context, existing U) (U, error) {
	existing.SetStatus(f.newStatus)
	existing.SetLastSeen(time.Now())
	return existing, nil
}

// createUserEntityFactory creates new user entities
type createUserEntityFactory[U UserEntity] struct {
	userID UserID
	status string
	repo   UserRepository[U]
}

func (f *createUserEntityFactory[U]) CreateFn(ctx context.Context) (U, error) {
	user := f.repo.NewEntity()
	// Note: This assumes UserEntity has some way to set ID
	// Users may need to implement this based on their entity structure
	return user, nil
}

func (f *createUserEntityFactory[U]) UpdateFn(ctx context.Context, existing U) (U, error) {
	// If user exists, just update status
	existing.SetStatus(f.status)
	existing.SetLastSeen(time.Now())
	return existing, nil
}

// ============================================================================
// Fluent API Builder
// ============================================================================

// Search returns a search builder
func (s *BasicUserService[U]) Search() *SearchBuilder[U] {
	return &SearchBuilder[U]{service: s}
}

// SearchBuilder provides fluent interface for searching users
type SearchBuilder[U UserEntity] struct {
	service UserService[U]
	query   string
	limit   int
	filters map[string]any
}

// Query sets the search query
func (sb *SearchBuilder[U]) Query(query string) *SearchBuilder[U] {
	sb.query = query
	return sb
}

// Limit sets the maximum number of results
func (sb *SearchBuilder[U]) Limit(limit int) *SearchBuilder[U] {
	sb.limit = limit
	return sb
}

// Where adds search filters
func (sb *SearchBuilder[U]) Where(key string, value any) *SearchBuilder[U] {
	if sb.filters == nil {
		sb.filters = make(map[string]any)
	}
	sb.filters[key] = value
	return sb
}

// Execute performs the search
func (sb *SearchBuilder[U]) Execute(ctx context.Context) ([]U, error) {
	// Build search options from builder state
	options := []SearchOption{}
	
	if sb.limit > 0 {
		options = append(options, func(c *SearchConfig) { c.Limit = sb.limit })
	}
	
	// Apply custom filters
	for key, value := range sb.filters {
		switch key {
		case "online_only":
			if online, ok := value.(bool); ok && online {
				options = append(options, func(c *SearchConfig) { c.OnlineOnly = true })
			}
		case "exclude_blocked":
			if exclude, ok := value.(bool); ok && exclude {
				options = append(options, func(c *SearchConfig) { c.ExcludeBlocked = true })
			}
		}
	}
	
	return sb.service.SearchUsers(ctx, sb.query, options...)
}