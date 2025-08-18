package friendit

import (
	"time"
)

// ============================================================================
// 통합 서비스 매니저 (선택적 사용)
// ============================================================================

// ServiceManager provides access to all domain services
// 사용자가 원하면 개별 서비스를 직접 사용하거나, 이 매니저를 통해 접근할 수 있습니다
type ServiceManager[U UserEntity, F FriendshipEntity, FR FriendRequestEntity, BR BlockRelationEntity] struct {
	FriendRequests FriendRequestService[FR]
	Friendships    FriendshipService[U, F]
	Blocks         BlockService[BR]
	Users          UserService[U]
}

// NewServiceManager creates a new service manager
func NewServiceManager[U UserEntity, F FriendshipEntity, FR FriendRequestEntity, BR BlockRelationEntity](
	repos *Repositories[U, F, FR, BR],
	options ...ServiceOption,
) *ServiceManager[U, F, FR, BR] {
	config := defaultServiceConfig()
	for _, opt := range options {
		opt(config)
	}
	
	return &ServiceManager[U, F, FR, BR]{
		FriendRequests: NewFriendRequestService(repos.FriendRequests, config),
		Friendships:    NewFriendshipService(repos.Users, repos.Friendships, config),
		Blocks:         NewBlockService(repos.BlockRelations, config),
		Users:          NewUserService(repos.Users, config),
	}
}

// ============================================================================
// ServiceManager의 편의 메서드들 (Fluent API 접근)
// ============================================================================

// Request returns a friend request builder
func (sm *ServiceManager[U, F, FR, BR]) Request() *RequestBuilder[FR] {
	if reqService, ok := sm.FriendRequests.(*BasicFriendRequestService[FR]); ok {
		return reqService.Request()
	}
	return &RequestBuilder[FR]{} // fallback
}

// Filter returns a friend filter builder
func (sm *ServiceManager[U, F, FR, BR]) Filter() *FilterBuilder[U, F] {
	if friendService, ok := sm.Friendships.(*BasicFriendshipService[U, F]); ok {
		return friendService.Filter()
	}
	return &FilterBuilder[U, F]{} // fallback
}

// Block returns a block builder
func (sm *ServiceManager[U, F, FR, BR]) Block() *BlockBuilder[BR] {
	if blockService, ok := sm.Blocks.(*BasicBlockService[BR]); ok {
		return blockService.Block()
	}
	return &BlockBuilder[BR]{} // fallback
}

// Search returns a search builder
func (sm *ServiceManager[U, F, FR, BR]) Search() *SearchBuilder[U] {
	if userService, ok := sm.Users.(*BasicUserService[U]); ok {
		return userService.Search()
	}
	return &SearchBuilder[U]{} // fallback
}

// ============================================================================
// Service Configuration
// ============================================================================

// ServiceConfig holds configuration for the friend service
type ServiceConfig struct {
	MaxFriends          int
	MaxPendingRequests  int
	RequestExpiration   time.Duration
	AutoCleanupExpired  bool
	BlockDuration       *time.Duration
	AllowSelfRequests   bool
	RequireMessage      bool
	EnableRecommendations bool
}

// ServiceOption configures the service
type ServiceOption func(*ServiceConfig)

// WithMaxFriends sets the maximum number of friends a user can have
func WithMaxFriends(max int) ServiceOption {
	return func(c *ServiceConfig) { c.MaxFriends = max }
}

// WithMaxPendingRequests sets the maximum number of pending requests
func WithMaxPendingRequests(max int) ServiceOption {
	return func(c *ServiceConfig) { c.MaxPendingRequests = max }
}

// WithRequestExpiration sets how long friend requests remain valid
func WithRequestExpiration(duration time.Duration) ServiceOption {
	return func(c *ServiceConfig) { c.RequestExpiration = duration }
}

// WithAutoCleanup enables automatic cleanup of expired requests
func WithAutoCleanup(enabled bool) ServiceOption {
	return func(c *ServiceConfig) { c.AutoCleanupExpired = enabled }
}

// WithAllowSelfRequests allows users to send requests to themselves
func WithAllowSelfRequests(allow bool) ServiceOption {
	return func(c *ServiceConfig) { c.AllowSelfRequests = allow }
}

// WithRequireMessage requires a message for friend requests
func WithRequireMessage(require bool) ServiceOption {
	return func(c *ServiceConfig) { c.RequireMessage = require }
}

func defaultServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		MaxFriends:         500,
		MaxPendingRequests: 50,
		RequestExpiration:  24 * time.Hour * 7, // 7 days
		AutoCleanupExpired: true,
		BlockDuration:      nil, // permanent by default
		AllowSelfRequests:  false,
		RequireMessage:     false,
		EnableRecommendations: true,
	}
}