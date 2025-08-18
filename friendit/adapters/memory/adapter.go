package frienditmemory

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/homveloper/dukdakit/friendit"
)

// ============================================================================
// Memory Adapter Example - 개발/테스트용 인메모리 구현 예제
// ============================================================================

// MemoryAdapter provides in-memory implementation of repositories
// 이것은 개발과 테스트를 위한 예제 구현입니다.
type MemoryAdapter struct {
	users         map[friendit.UserID]friendit.BasicUser
	friendships   map[friendit.FriendshipID]friendit.BasicFriendship
	friendRequests map[friendit.RequestID]friendit.BasicFriendRequest
	blockRelations map[friendit.BlockID]friendit.BasicBlockRelation
	mu            sync.RWMutex
}

// NewMemoryAdapter creates a new in-memory adapter
func NewMemoryAdapter() *MemoryAdapter {
	return &MemoryAdapter{
		users:         make(map[friendit.UserID]friendit.BasicUser),
		friendships:   make(map[friendit.FriendshipID]friendit.BasicFriendship),
		friendRequests: make(map[friendit.RequestID]friendit.BasicFriendRequest),
		blockRelations: make(map[friendit.BlockID]friendit.BasicBlockRelation),
	}
}

// Close is a no-op for memory adapter
func (ma *MemoryAdapter) Close() error {
	return nil
}

// ============================================================================
// User Repository Implementation
// ============================================================================

// MemoryUserRepository implements UserRepository for in-memory with BasicUser
type MemoryUserRepository struct {
	adapter *MemoryAdapter
}

// NewMemoryUserRepository creates a new in-memory user repository
func (ma *MemoryAdapter) NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{adapter: ma}
}

// Create implements UserRepository.Create for BasicUser
func (r *MemoryUserRepository) Create(ctx context.Context, user friendit.BasicUser) error {
	r.adapter.mu.Lock()
	defer r.adapter.mu.Unlock()

	if _, exists := r.adapter.users[user.ID]; exists {
		return fmt.Errorf("user already exists: %s", user.ID)
	}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	r.adapter.users[user.ID] = user

	return nil
}

// GetByID implements UserRepository.GetByID for BasicUser
func (r *MemoryUserRepository) GetByID(ctx context.Context, id friendit.UserID) (friendit.BasicUser, error) {
	r.adapter.mu.RLock()
	defer r.adapter.mu.RUnlock()

	user, exists := r.adapter.users[id]
	if !exists {
		return friendit.BasicUser{}, fmt.Errorf("user not found: %s", id)
	}

	return user, nil
}

// Update implements UserRepository.Update for BasicUser
func (r *MemoryUserRepository) Update(ctx context.Context, user friendit.BasicUser) error {
	r.adapter.mu.Lock()
	defer r.adapter.mu.Unlock()

	if _, exists := r.adapter.users[user.ID]; !exists {
		return fmt.Errorf("user not found: %s", user.ID)
	}

	user.UpdatedAt = time.Now()
	r.adapter.users[user.ID] = user

	return nil
}

// Delete implements UserRepository.Delete for BasicUser
func (r *MemoryUserRepository) Delete(ctx context.Context, id friendit.UserID) error {
	r.adapter.mu.Lock()
	defer r.adapter.mu.Unlock()

	if _, exists := r.adapter.users[id]; !exists {
		return fmt.Errorf("user not found: %s", id)
	}

	delete(r.adapter.users, id)
	return nil
}

// GetByIDs implements UserRepository.GetByIDs for BasicUser
func (r *MemoryUserRepository) GetByIDs(ctx context.Context, ids []friendit.UserID) ([]friendit.BasicUser, error) {
	r.adapter.mu.RLock()
	defer r.adapter.mu.RUnlock()

	users := make([]friendit.BasicUser, 0, len(ids))
	for _, id := range ids {
		if user, exists := r.adapter.users[id]; exists {
			users = append(users, user)
		}
	}

	return users, nil
}

// FindByStatus implements UserRepository.FindByStatus for BasicUser
func (r *MemoryUserRepository) FindByStatus(ctx context.Context, status string) ([]friendit.BasicUser, error) {
	r.adapter.mu.RLock()
	defer r.adapter.mu.RUnlock()

	var users []friendit.BasicUser
	for _, user := range r.adapter.users {
		if user.Status == status {
			users = append(users, user)
		}
	}

	return users, nil
}

// Search implements UserRepository.Search for BasicUser
func (r *MemoryUserRepository) Search(ctx context.Context, query string, limit int) ([]friendit.BasicUser, error) {
	r.adapter.mu.RLock()
	defer r.adapter.mu.RUnlock()

	query = strings.ToLower(query)
	var users []friendit.BasicUser
	count := 0

	for _, user := range r.adapter.users {
		if count >= limit {
			break
		}

		username := strings.ToLower(user.Username)
		displayName := strings.ToLower(user.DisplayName)

		if strings.Contains(username, query) || strings.Contains(displayName, query) {
			users = append(users, user)
			count++
		}
	}

	return users, nil
}

// UpdateStatus implements UserRepository.UpdateStatus for BasicUser
func (r *MemoryUserRepository) UpdateStatus(ctx context.Context, id friendit.UserID, status string) error {
	r.adapter.mu.Lock()
	defer r.adapter.mu.Unlock()

	user, exists := r.adapter.users[id]
	if !exists {
		return fmt.Errorf("user not found: %s", id)
	}

	user.Status = status
	now := time.Now()
	user.LastSeen = &now
	user.UpdatedAt = now
	r.adapter.users[id] = user

	return nil
}

// GetOnlineUsers implements UserRepository.GetOnlineUsers for BasicUser
func (r *MemoryUserRepository) GetOnlineUsers(ctx context.Context) ([]friendit.BasicUser, error) {
	return r.FindByStatus(ctx, "online")
}

// ============================================================================
// Friendship Repository Implementation
// ============================================================================

// MemoryFriendshipRepository implements FriendshipRepository for in-memory with BasicFriendship
type MemoryFriendshipRepository struct {
	adapter *MemoryAdapter
}

// NewMemoryFriendshipRepository creates a new in-memory friendship repository
func (ma *MemoryAdapter) NewMemoryFriendshipRepository() *MemoryFriendshipRepository {
	return &MemoryFriendshipRepository{adapter: ma}
}

// Create implements FriendshipRepository.Create for BasicFriendship
func (r *MemoryFriendshipRepository) Create(ctx context.Context, friendship friendit.BasicFriendship) error {
	r.adapter.mu.Lock()
	defer r.adapter.mu.Unlock()

	if _, exists := r.adapter.friendships[friendship.ID]; exists {
		return fmt.Errorf("friendship already exists: %s", friendship.ID)
	}

	friendship.CreatedAt = time.Now()
	friendship.UpdatedAt = time.Now()
	r.adapter.friendships[friendship.ID] = friendship

	return nil
}

// GetByID implements FriendshipRepository.GetByID for BasicFriendship
func (r *MemoryFriendshipRepository) GetByID(ctx context.Context, id friendit.FriendshipID) (friendit.BasicFriendship, error) {
	r.adapter.mu.RLock()
	defer r.adapter.mu.RUnlock()

	friendship, exists := r.adapter.friendships[id]
	if !exists {
		return friendit.BasicFriendship{}, fmt.Errorf("friendship not found: %s", id)
	}

	return friendship, nil
}

// GetFriendship implements FriendshipRepository.GetFriendship for BasicFriendship
func (r *MemoryFriendshipRepository) GetFriendship(ctx context.Context, user1ID, user2ID friendit.UserID) (friendit.BasicFriendship, error) {
	r.adapter.mu.RLock()
	defer r.adapter.mu.RUnlock()

	for _, friendship := range r.adapter.friendships {
		if (friendship.User1ID == user1ID && friendship.User2ID == user2ID) ||
			(friendship.User1ID == user2ID && friendship.User2ID == user1ID) {
			return friendship, nil
		}
	}

	return friendit.BasicFriendship{}, fmt.Errorf("friendship not found between %s and %s", user1ID, user2ID)
}

// GetByUserID implements FriendshipRepository.GetByUserID for BasicFriendship
func (r *MemoryFriendshipRepository) GetByUserID(ctx context.Context, userID friendit.UserID) ([]friendit.BasicFriendship, error) {
	r.adapter.mu.RLock()
	defer r.adapter.mu.RUnlock()

	var friendships []friendit.BasicFriendship
	for _, friendship := range r.adapter.friendships {
		if friendship.User1ID == userID || friendship.User2ID == userID {
			friendships = append(friendships, friendship)
		}
	}

	return friendships, nil
}

// Update implements FriendshipRepository.Update for BasicFriendship
func (r *MemoryFriendshipRepository) Update(ctx context.Context, friendship friendit.BasicFriendship) error {
	r.adapter.mu.Lock()
	defer r.adapter.mu.Unlock()

	if _, exists := r.adapter.friendships[friendship.ID]; !exists {
		return fmt.Errorf("friendship not found: %s", friendship.ID)
	}

	friendship.UpdatedAt = time.Now()
	r.adapter.friendships[friendship.ID] = friendship

	return nil
}

// Delete implements FriendshipRepository.Delete for BasicFriendship
func (r *MemoryFriendshipRepository) Delete(ctx context.Context, id friendit.FriendshipID) error {
	r.adapter.mu.Lock()
	defer r.adapter.mu.Unlock()

	if _, exists := r.adapter.friendships[id]; !exists {
		return fmt.Errorf("friendship not found: %s", id)
	}

	delete(r.adapter.friendships, id)
	return nil
}

// GetMutualFriends implements FriendshipRepository.GetMutualFriends for BasicFriendship
func (r *MemoryFriendshipRepository) GetMutualFriends(ctx context.Context, user1ID, user2ID friendit.UserID) ([]friendit.BasicFriendship, error) {
	r.adapter.mu.RLock()
	defer r.adapter.mu.RUnlock()

	// Get friends of user1
	user1Friends := make(map[friendit.UserID]bool)
	for _, friendship := range r.adapter.friendships {
		if friendship.Status == "active" {
			if friendship.User1ID == user1ID {
				user1Friends[friendship.User2ID] = true
			} else if friendship.User2ID == user1ID {
				user1Friends[friendship.User1ID] = true
			}
		}
	}

	// Find mutual friends
	var mutualFriends []friendit.BasicFriendship
	for _, friendship := range r.adapter.friendships {
		if friendship.Status == "active" {
			var friendOfUser2 friendit.UserID
			if friendship.User1ID == user2ID {
				friendOfUser2 = friendship.User2ID
			} else if friendship.User2ID == user2ID {
				friendOfUser2 = friendship.User1ID
			} else {
				continue
			}

			if user1Friends[friendOfUser2] {
				mutualFriends = append(mutualFriends, friendship)
			}
		}
	}

	return mutualFriends, nil
}

// GetByStatus implements FriendshipRepository.GetByStatus for BasicFriendship
func (r *MemoryFriendshipRepository) GetByStatus(ctx context.Context, userID friendit.UserID, status string) ([]friendit.BasicFriendship, error) {
	r.adapter.mu.RLock()
	defer r.adapter.mu.RUnlock()

	var friendships []friendit.BasicFriendship
	for _, friendship := range r.adapter.friendships {
		if friendship.Status == status &&
			(friendship.User1ID == userID || friendship.User2ID == userID) {
			friendships = append(friendships, friendship)
		}
	}

	return friendships, nil
}

// UpdateStatus implements FriendshipRepository.UpdateStatus for BasicFriendship
func (r *MemoryFriendshipRepository) UpdateStatus(ctx context.Context, id friendit.FriendshipID, status string) error {
	r.adapter.mu.Lock()
	defer r.adapter.mu.Unlock()

	friendship, exists := r.adapter.friendships[id]
	if !exists {
		return fmt.Errorf("friendship not found: %s", id)
	}

	friendship.Status = status
	friendship.UpdatedAt = time.Now()
	r.adapter.friendships[id] = friendship

	return nil
}

// GetFriendships implements FriendshipRepository.GetFriendships for BasicFriendship
func (r *MemoryFriendshipRepository) GetFriendships(ctx context.Context, userIDs []friendit.UserID) ([]friendit.BasicFriendship, error) {
	r.adapter.mu.RLock()
	defer r.adapter.mu.RUnlock()

	userSet := make(map[friendit.UserID]bool)
	for _, id := range userIDs {
		userSet[id] = true
	}

	var friendships []friendit.BasicFriendship
	for _, friendship := range r.adapter.friendships {
		if userSet[friendship.User1ID] || userSet[friendship.User2ID] {
			friendships = append(friendships, friendship)
		}
	}

	return friendships, nil
}

// DeleteByUserID implements FriendshipRepository.DeleteByUserID for BasicFriendship
func (r *MemoryFriendshipRepository) DeleteByUserID(ctx context.Context, userID friendit.UserID) error {
	r.adapter.mu.Lock()
	defer r.adapter.mu.Unlock()

	for id, friendship := range r.adapter.friendships {
		if friendship.User1ID == userID || friendship.User2ID == userID {
			delete(r.adapter.friendships, id)
		}
	}

	return nil
}