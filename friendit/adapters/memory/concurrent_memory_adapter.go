package frienditmemory

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/homveloper/dukdakit/friendit"
)

// ============================================================================
// 동시성 안전 메모리 어댑터 - 원자적 연산 지원
// ============================================================================

// ConcurrentMemoryUserRepository 동시성 안전 메모리 사용자 저장소
type ConcurrentMemoryUserRepository struct {
	mu       sync.RWMutex
	users    map[friendit.UserID]*versionedUser
	versions map[friendit.UserID]int64
}

// versionedUser 버전 정보를 포함한 사용자 엔터티
type versionedUser struct {
	*friendit.BasicUser
	Version int64
}

// NewConcurrentMemoryUserRepository 새 동시성 안전 메모리 사용자 저장소 생성
func NewConcurrentMemoryUserRepository() *ConcurrentMemoryUserRepository {
	return &ConcurrentMemoryUserRepository{
		users:    make(map[friendit.UserID]*versionedUser),
		versions: make(map[friendit.UserID]int64),
	}
}

// ============================================================================
// 기본 CRUD 연산들
// ============================================================================

func (r *ConcurrentMemoryUserRepository) Create(ctx context.Context, user *friendit.BasicUser) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.users[user.ID]; exists {
		return fmt.Errorf("user already exists: %s", user.ID)
	}
	
	versioned := &versionedUser{
		BasicUser: user,
		Version:   1,
	}
	
	r.users[user.ID] = versioned
	r.versions[user.ID] = 1
	
	return nil
}

func (r *ConcurrentMemoryUserRepository) GetByID(ctx context.Context, id friendit.UserID) (*friendit.BasicUser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	user, exists := r.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", id)
	}
	
	return r.copyUser(user.BasicUser), nil
}

func (r *ConcurrentMemoryUserRepository) Update(ctx context.Context, user *friendit.BasicUser) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	existing, exists := r.users[user.ID]
	if !exists {
		return fmt.Errorf("user not found: %s", user.ID)
	}
	
	existing.BasicUser = user
	existing.Version++
	r.versions[user.ID] = existing.Version
	
	return nil
}

func (r *ConcurrentMemoryUserRepository) Delete(ctx context.Context, id friendit.UserID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.users[id]; !exists {
		return fmt.Errorf("user not found: %s", id)
	}
	
	delete(r.users, id)
	delete(r.versions, id)
	
	return nil
}

// ============================================================================
// 원자적 연산들 (동시성 안전)
// ============================================================================

// FindOneAndUpsert 원자적 사용자 생성/업데이트 (동시성 안전)
func (r *ConcurrentMemoryUserRepository) FindOneAndUpsert(
	ctx context.Context,
	filter map[string]any,
	factory friendit.EntityFactory[*friendit.BasicUser],
) (*friendit.AtomicResult[*friendit.BasicUser], error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// 필터에서 사용자 ID 추출
	userID, ok := filter["_id"].(friendit.UserID)
	if !ok {
		return nil, fmt.Errorf("invalid filter: missing _id")
	}
	
	existing, exists := r.users[userID]
	
	if !exists {
		// 새로 생성
		newEntity, err := factory.CreateFn(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create entity: %w", err)
		}
		
		version := int64(1)
		versioned := &versionedUser{
			BasicUser: newEntity,
			Version:   version,
		}
		
		r.users[userID] = versioned
		r.versions[userID] = version
		
		return friendit.NewAtomicResult(newEntity, true, version), nil
	}
	
	// 업데이트
	updatedEntity, err := factory.UpdateFn(ctx, existing.BasicUser)
	if err != nil {
		return nil, fmt.Errorf("failed to update entity: %w", err)
	}
	
	newVersion := existing.Version + 1
	versioned := &versionedUser{
		BasicUser: updatedEntity,
		Version:   newVersion,
	}
	
	r.users[userID] = versioned
	r.versions[userID] = newVersion
	
	return friendit.NewAtomicResult(updatedEntity, false, newVersion), nil
}

// FindOneAndInsert 원자적 사용자 생성 (중복 방지)
func (r *ConcurrentMemoryUserRepository) FindOneAndInsert(
	ctx context.Context,
	factory friendit.EntityFactory[*friendit.BasicUser],
) (*friendit.AtomicResult[*friendit.BasicUser], error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	newEntity, err := factory.CreateFn(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}
	
	// 중복 검사
	if _, exists := r.users[newEntity.ID]; exists {
		return nil, fmt.Errorf("user already exists: %s", newEntity.ID)
	}
	
	version := int64(1)
	versioned := &versionedUser{
		BasicUser: newEntity,
		Version:   version,
	}
	
	r.users[newEntity.ID] = versioned
	r.versions[newEntity.ID] = version
	
	return friendit.NewAtomicResult(newEntity, true, version), nil
}

// FindOneAndUpdate 원자적 사용자 업데이트
func (r *ConcurrentMemoryUserRepository) FindOneAndUpdate(
	ctx context.Context,
	id friendit.UserID,
	factory friendit.EntityFactory[*friendit.BasicUser],
) (*friendit.AtomicResult[*friendit.BasicUser], error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	existing, exists := r.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", id)
	}
	
	updatedEntity, err := factory.UpdateFn(ctx, existing.BasicUser)
	if err != nil {
		return nil, fmt.Errorf("failed to update entity: %w", err)
	}
	
	newVersion := existing.Version + 1
	versioned := &versionedUser{
		BasicUser: updatedEntity,
		Version:   newVersion,
	}
	
	r.users[id] = versioned
	r.versions[id] = newVersion
	
	return friendit.NewAtomicResult(updatedEntity, false, newVersion), nil
}

// UpdateIfVersion 낙관적 잠금을 사용한 조건부 업데이트
func (r *ConcurrentMemoryUserRepository) UpdateIfVersion(
	ctx context.Context,
	id friendit.UserID,
	expectedVersion int64,
	factory friendit.EntityFactory[*friendit.BasicUser],
) (*friendit.AtomicResult[*friendit.BasicUser], error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	existing, exists := r.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", id)
	}
	
	if existing.Version != expectedVersion {
		return nil, fmt.Errorf("version conflict: expected %d, got %d", expectedVersion, existing.Version)
	}
	
	updatedEntity, err := factory.UpdateFn(ctx, existing.BasicUser)
	if err != nil {
		return nil, fmt.Errorf("failed to update entity: %w", err)
	}
	
	newVersion := expectedVersion + 1
	versioned := &versionedUser{
		BasicUser: updatedEntity,
		Version:   newVersion,
	}
	
	r.users[id] = versioned
	r.versions[id] = newVersion
	
	return friendit.NewAtomicResult(updatedEntity, false, newVersion), nil
}

func (r *ConcurrentMemoryUserRepository) SoftDelete(ctx context.Context, id friendit.UserID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	user, exists := r.users[id]
	if !exists {
		return fmt.Errorf("user not found: %s", id)
	}
	
	user.BasicUser.Status = "deleted"
	user.BasicUser.UpdatedAt = time.Now()
	user.Version++
	r.versions[id] = user.Version
	
	return nil
}

func (r *ConcurrentMemoryUserRepository) HardDelete(ctx context.Context, id friendit.UserID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.users[id]; !exists {
		return fmt.Errorf("user not found: %s", id)
	}
	
	delete(r.users, id)
	delete(r.versions, id)
	
	return nil
}

// ============================================================================
// 읽기 전용 연산들 (동시성 안전)
// ============================================================================

func (r *ConcurrentMemoryUserRepository) GetByIDs(ctx context.Context, ids []friendit.UserID) ([]*friendit.BasicUser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var users []*friendit.BasicUser
	for _, id := range ids {
		if user, exists := r.users[id]; exists {
			users = append(users, r.copyUser(user.BasicUser))
		}
	}
	
	return users, nil
}

func (r *ConcurrentMemoryUserRepository) GetByStatus(ctx context.Context, status string) ([]*friendit.BasicUser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var users []*friendit.BasicUser
	for _, user := range r.users {
		if user.BasicUser.Status == status {
			users = append(users, r.copyUser(user.BasicUser))
		}
	}
	
	return users, nil
}

func (r *ConcurrentMemoryUserRepository) Search(ctx context.Context, query string, limit int) ([]*friendit.BasicUser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	query = strings.ToLower(query)
	var users []*friendit.BasicUser
	count := 0
	
	for _, user := range r.users {
		// 사용자명 또는 표시명에서 검색
		if strings.Contains(strings.ToLower(user.BasicUser.Username), query) ||
		   strings.Contains(strings.ToLower(user.BasicUser.DisplayName), query) {
			users = append(users, r.copyUser(user.BasicUser))
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
	}
	
	// ID로 정렬
	sort.Slice(users, func(i, j int) bool {
		return users[i].ID < users[j].ID
	})
	
	return users, nil
}

func (r *ConcurrentMemoryUserRepository) GetRecent(ctx context.Context, limit int) ([]*friendit.BasicUser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var users []*friendit.BasicUser
	for _, user := range r.users {
		users = append(users, r.copyUser(user.BasicUser))
	}
	
	// 생성 시간으로 정렬 (최신순)
	sort.Slice(users, func(i, j int) bool {
		return users[i].CreatedAt.After(users[j].CreatedAt)
	})
	
	if limit > 0 && len(users) > limit {
		users = users[:limit]
	}
	
	return users, nil
}

func (r *ConcurrentMemoryUserRepository) UpdateStatus(ctx context.Context, id friendit.UserID, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	user, exists := r.users[id]
	if !exists {
		return fmt.Errorf("user not found: %s", id)
	}
	
	user.BasicUser.Status = status
	user.BasicUser.SetLastSeen(time.Now())
	user.Version++
	r.versions[id] = user.Version
	
	return nil
}

func (r *ConcurrentMemoryUserRepository) GetOnlineUsers(ctx context.Context) ([]*friendit.BasicUser, error) {
	return r.GetByStatus(ctx, "online")
}

// NewEntity 새 BasicUser 엔터티 생성
func (r *ConcurrentMemoryUserRepository) NewEntity() *friendit.BasicUser {
	return &friendit.BasicUser{
		Metadata: make(map[string]any),
	}
}

// ============================================================================
// 헬퍼 메서드
// ============================================================================

// copyUser 사용자 엔터티의 깊은 복사본 생성
func (r *ConcurrentMemoryUserRepository) copyUser(user *friendit.BasicUser) *friendit.BasicUser {
	// 메타데이터 복사
	metadata := make(map[string]any)
	for k, v := range user.Metadata {
		metadata[k] = v
	}
	
	// LastSeen 복사
	var lastSeen *time.Time
	if user.LastSeen != nil {
		ts := *user.LastSeen
		lastSeen = &ts
	}
	
	return &friendit.BasicUser{
		ID:          user.ID,
		Status:      user.Status,
		LastSeen:    lastSeen,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Metadata:    metadata,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}