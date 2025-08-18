package frienditredis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/homveloper/dukdakit/friendit"
)

// ============================================================================
// Redis Adapter Example - 사용자가 참고할 수 있는 Redis 구현 예제
// ============================================================================

// RedisAdapter provides Redis implementation of repositories
// 이것은 예제 구현입니다. 사용자가 자신의 요구사항에 맞게 수정할 수 있습니다.
type RedisAdapter struct {
	client *redis.Client
}

// NewRedisAdapter creates a new Redis adapter
func NewRedisAdapter(addr, password string, db int) (*RedisAdapter, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisAdapter{
		client: rdb,
	}, nil
}

// Close closes the Redis connection
func (ra *RedisAdapter) Close() error {
	return ra.client.Close()
}

// ============================================================================
// User Repository Implementation (예제)
// ============================================================================

// RedisUserRepository implements UserRepository for Redis with BasicUser
type RedisUserRepository struct {
	client *redis.Client
	keyPrefix string
}

// NewRedisUserRepository creates a new Redis user repository
func (ra *RedisAdapter) NewRedisUserRepository() *RedisUserRepository {
	return &RedisUserRepository{
		client:    ra.client,
		keyPrefix: "friendit:users:",
	}
}

// Create implements UserRepository.Create for BasicUser
func (r *RedisUserRepository) Create(ctx context.Context, user friendit.BasicUser) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}
	
	key := r.keyPrefix + string(user.ID)
	if err := r.client.Set(ctx, key, data, 0).Err(); err != nil {
		return fmt.Errorf("failed to create user in Redis: %w", err)
	}
	
	return nil
}

// GetByID implements UserRepository.GetByID for BasicUser
func (r *RedisUserRepository) GetByID(ctx context.Context, id friendit.UserID) (friendit.BasicUser, error) {
	var user friendit.BasicUser
	
	key := r.keyPrefix + string(id)
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return user, fmt.Errorf("user not found: %s", id)
		}
		return user, fmt.Errorf("failed to get user from Redis: %w", err)
	}
	
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return user, fmt.Errorf("failed to unmarshal user: %w", err)
	}
	
	return user, nil
}

// Update implements UserRepository.Update for BasicUser
func (r *RedisUserRepository) Update(ctx context.Context, user friendit.BasicUser) error {
	user.UpdatedAt = time.Now()
	
	// Check if user exists
	key := r.keyPrefix + string(user.ID)
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("user not found: %s", user.ID)
	}
	
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}
	
	if err := r.client.Set(ctx, key, data, 0).Err(); err != nil {
		return fmt.Errorf("failed to update user in Redis: %w", err)
	}
	
	return nil
}

// Delete implements UserRepository.Delete for BasicUser
func (r *RedisUserRepository) Delete(ctx context.Context, id friendit.UserID) error {
	key := r.keyPrefix + string(id)
	result, err := r.client.Del(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to delete user from Redis: %w", err)
	}
	
	if result == 0 {
		return fmt.Errorf("user not found: %s", id)
	}
	
	return nil
}

// GetByIDs implements UserRepository.GetByIDs for BasicUser  
func (r *RedisUserRepository) GetByIDs(ctx context.Context, ids []friendit.UserID) ([]friendit.BasicUser, error) {
	if len(ids) == 0 {
		return []friendit.BasicUser{}, nil
	}
	
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = r.keyPrefix + string(id)
	}
	
	values, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get users from Redis: %w", err)
	}
	
	users := make([]friendit.BasicUser, 0, len(ids))
	for _, value := range values {
		if value != nil {
			var user friendit.BasicUser
			if err := json.Unmarshal([]byte(value.(string)), &user); err == nil {
				users = append(users, user)
			}
		}
	}
	
	return users, nil
}

// FindByStatus implements UserRepository.FindByStatus for BasicUser
// Redis에서는 인덱스가 없으므로 키 스캐닝을 사용 (프로덕션에서는 적절한 인덱싱 필요)
func (r *RedisUserRepository) FindByStatus(ctx context.Context, status string) ([]friendit.BasicUser, error) {
	// 이것은 예제 구현입니다. 실제 프로덕션에서는 Redis의 SET이나 다른 자료구조를 사용하여
	// 상태별 인덱스를 구성하는 것이 좋습니다.
	pattern := r.keyPrefix + "*"
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to scan keys: %w", err)
	}
	
	var users []friendit.BasicUser
	for _, key := range keys {
		data, err := r.client.Get(ctx, key).Result()
		if err != nil {
			continue // Skip errors for individual keys
		}
		
		var user friendit.BasicUser
		if err := json.Unmarshal([]byte(data), &user); err != nil {
			continue
		}
		
		if user.Status == status {
			users = append(users, user)
		}
	}
	
	return users, nil
}

// Search implements UserRepository.Search for BasicUser
func (r *RedisUserRepository) Search(ctx context.Context, query string, limit int) ([]friendit.BasicUser, error) {
	// 이것은 예제 구현입니다. 실제로는 Redis Search 모듈이나 별도의 인덱싱이 필요합니다.
	return nil, fmt.Errorf("search not implemented for Redis adapter - use Redis Search module")
}

// UpdateStatus implements UserRepository.UpdateStatus for BasicUser
func (r *RedisUserRepository) UpdateStatus(ctx context.Context, id friendit.UserID, status string) error {
	user, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	
	user.Status = status
	now := time.Now()
	user.LastSeen = &now
	user.UpdatedAt = now
	
	return r.Update(ctx, user)
}

// GetOnlineUsers implements UserRepository.GetOnlineUsers for BasicUser
func (r *RedisUserRepository) GetOnlineUsers(ctx context.Context) ([]friendit.BasicUser, error) {
	return r.FindByStatus(ctx, "online")
}