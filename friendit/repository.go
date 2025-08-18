package friendit

import (
	"context"
)

// ============================================================================
// Repository 인터페이스들 - 사용자가 자신의 저장소에 맞게 구현
// ============================================================================

// UserRepository defines the interface for user storage operations
// 사용자가 자신의 저장소(MongoDB, PostgreSQL, Redis 등)에 맞게 구현
type UserRepository[U UserEntity] interface {
	// Basic CRUD operations
	Create(ctx context.Context, user U) error
	GetByID(ctx context.Context, id UserID) (U, error)
	Update(ctx context.Context, user U) error
	Delete(ctx context.Context, id UserID) error
	
	// Query operations
	GetByIDs(ctx context.Context, ids []UserID) ([]U, error)
	GetByStatus(ctx context.Context, status string) ([]U, error)
	Search(ctx context.Context, query string, limit int) ([]U, error)
	GetRecent(ctx context.Context, limit int) ([]U, error)
	
	// Status operations
	UpdateStatus(ctx context.Context, id UserID, status string) error
	GetOnlineUsers(ctx context.Context) ([]U, error)
	
	// Atomic operations for concurrency safety
	FindOneAndUpsert(ctx context.Context, filter map[string]any, factory EntityFactory[U]) (*AtomicResult[U], error)
	FindOneAndInsert(ctx context.Context, factory EntityFactory[U]) (*AtomicResult[U], error)
	FindOneAndUpdate(ctx context.Context, id UserID, factory EntityFactory[U]) (*AtomicResult[U], error)
	UpdateIfVersion(ctx context.Context, id UserID, expectedVersion int64, factory EntityFactory[U]) (*AtomicResult[U], error)
	SoftDelete(ctx context.Context, id UserID) error
	HardDelete(ctx context.Context, id UserID) error
	
	// Entity factory method - users implement this
	NewEntity() U
}

// FriendshipRepository defines the interface for friendship storage operations
type FriendshipRepository[F FriendshipEntity] interface {
	// Basic CRUD operations
	Create(ctx context.Context, friendship F) error
	GetByID(ctx context.Context, id FriendshipID) (F, error)
	Update(ctx context.Context, friendship F) error
	Delete(ctx context.Context, id FriendshipID) error
	
	// Query operations
	GetByUserID(ctx context.Context, userID UserID) ([]F, error)
	GetByUsers(ctx context.Context, user1ID, user2ID UserID) (F, error)
	GetMutualFriends(ctx context.Context, user1ID, user2ID UserID) ([]F, error)
	
	// Status operations
	GetByStatus(ctx context.Context, userID UserID, status string) ([]F, error)
	UpdateStatus(ctx context.Context, id FriendshipID, status string) error
	
	// Batch operations
	GetFriendships(ctx context.Context, userIDs []UserID) ([]F, error)
	DeleteByUserID(ctx context.Context, userID UserID) error
	
	// Atomic operations for concurrency safety
	FindOneAndUpsert(ctx context.Context, user1ID, user2ID UserID, factory EntityFactory[F]) (*AtomicResult[F], error)
	FindOneAndInsert(ctx context.Context, factory EntityFactory[F]) (*AtomicResult[F], error)
	FindOneAndUpdate(ctx context.Context, id FriendshipID, factory EntityFactory[F]) (*AtomicResult[F], error)
	UpdateStatusIfCurrent(ctx context.Context, id FriendshipID, currentStatus, newStatus string) (*AtomicResult[F], error)
	DeleteByUsers(ctx context.Context, user1ID, user2ID UserID) error
	
	// Entity factory method - users implement this
	NewEntity() F
}

// FriendRequestRepository defines the interface for friend request storage operations
type FriendRequestRepository[FR FriendRequestEntity] interface {
	// Basic CRUD operations
	Create(ctx context.Context, request FR) error
	GetByID(ctx context.Context, id RequestID) (FR, error)
	Update(ctx context.Context, request FR) error
	Delete(ctx context.Context, id RequestID) error
	
	// Query operations
	GetBySenderID(ctx context.Context, senderID UserID) ([]FR, error)
	GetByReceiverID(ctx context.Context, receiverID UserID) ([]FR, error)
	GetPendingRequests(ctx context.Context, userID UserID) ([]FR, error)
	GetRequest(ctx context.Context, senderID, receiverID UserID) (FR, error)
	
	// Status operations
	UpdateStatus(ctx context.Context, id RequestID, status string) error
	GetByStatus(ctx context.Context, userID UserID, status string) ([]FR, error)
	
	// Cleanup operations
	DeleteExpiredRequests(ctx context.Context) error
	DeleteByUserID(ctx context.Context, userID UserID) error
	
	// Atomic operations for concurrency safety
	FindOneAndUpsert(ctx context.Context, senderID, receiverID UserID, factory EntityFactory[FR]) (*AtomicResult[FR], error)
	FindOneAndInsert(ctx context.Context, factory EntityFactory[FR]) (*AtomicResult[FR], error)
	FindOneAndUpdate(ctx context.Context, id RequestID, factory EntityFactory[FR]) (*AtomicResult[FR], error)
	AcceptIfPending(ctx context.Context, id RequestID, factory EntityFactory[FR]) (*AtomicResult[FR], error)
	RejectIfPending(ctx context.Context, id RequestID, factory EntityFactory[FR]) (*AtomicResult[FR], error)
	CancelIfPending(ctx context.Context, id RequestID) error
	DeleteExpired(ctx context.Context) (int64, error)
	
	// Entity factory method - users implement this
	NewEntity() FR
}

// BlockRelationRepository defines the interface for block relation storage operations
type BlockRelationRepository[BR BlockRelationEntity] interface {
	// Basic CRUD operations
	Create(ctx context.Context, block BR) error
	GetByID(ctx context.Context, id BlockID) (BR, error)
	Delete(ctx context.Context, id BlockID) error
	
	// Query operations
	GetByBlockerID(ctx context.Context, blockerID UserID) ([]BR, error)
	GetByBlockedID(ctx context.Context, blockedID UserID) ([]BR, error)
	GetByUsers(ctx context.Context, blockerID, blockedID UserID) (BR, error)
	IsBlocked(ctx context.Context, blockerID, blockedID UserID) (bool, error)
	
	// Batch operations
	GetBlockedUsers(ctx context.Context, userID UserID) ([]UserID, error)
	GetBlockingUsers(ctx context.Context, userID UserID) ([]UserID, error)
	DeleteByUserID(ctx context.Context, userID UserID) error
	
	// Atomic operations for concurrency safety
	FindOneAndUpsert(ctx context.Context, blockerID, blockedID UserID, factory EntityFactory[BR]) (*AtomicResult[BR], error)
	FindOneAndInsert(ctx context.Context, factory EntityFactory[BR]) (*AtomicResult[BR], error)
	DeleteByUsers(ctx context.Context, blockerID, blockedID UserID) error
	
	// Entity factory method - users implement this
	NewEntity() BR
}

// ============================================================================
// 통합 Repository 인터페이스
// ============================================================================

// FriendService repository dependencies
// 서비스에서 필요한 모든 repository들을 하나로 묶은 인터페이스
type Repositories[U UserEntity, F FriendshipEntity, FR FriendRequestEntity, BR BlockRelationEntity] struct {
	Users         UserRepository[U]
	Friendships   FriendshipRepository[F] 
	FriendRequests FriendRequestRepository[FR]
	BlockRelations BlockRelationRepository[BR]
}

// RepositoryFactory는 사용자가 자신의 저장소 구현을 제공할 수 있도록 하는 팩토리
type RepositoryFactory[U UserEntity, F FriendshipEntity, FR FriendRequestEntity, BR BlockRelationEntity] interface {
	CreateRepositories() (*Repositories[U, F, FR, BR], error)
	Close() error
}