package confluxredis_test

import (
	"context"
	"testing"
	"time"

	"github.com/homveloper/dukdakit/conflux"
	"github.com/homveloper/dukdakit/conflux/adapters/confluxredis"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// 테스트용 엔터티 정의
// ============================================================================

// User Redis 테스트용 사용자 엔터티
type User struct {
	conflux.BaseEntity
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Status   string `json:"status"`
	Credits  int    `json:"credits"`
}

// ============================================================================
// Redis 연결 헬퍼
// ============================================================================

func setupRedisClient() *redis.Client {
	// 테스트용 Redis 클라이언트 (실제 환경에서는 설정 조정 필요)
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1, // 테스트용 DB 사용
	})
}

func cleanupRedis(client *redis.Client, ctx context.Context) {
	// 테스트 데이터 정리
	client.FlushDB(ctx)
}

// ============================================================================
// Redis Repository 테스트들
// ============================================================================

func TestRedisRepository_FindOneAndInsert_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	client := setupRedisClient()
	defer client.Close()

	ctx := context.Background()
	defer cleanupRedis(client, ctx)

	// Repository 생성
	config := &confluxredis.RedisRepositoryConfig{
		KeyPrefix: "user:",
		TTL:       time.Hour, // 1시간 TTL
	}

	repo := confluxredis.NewRedisRepository[*User](client, config, func() *User {
		return &User{}
	})

	// 중복 검사용 필터 (이메일 기반)
	duplicateCheckFilter := conflux.NewRedisFilterWithPrefix("user:email:" + "john@example.com")

	// CreateFunc 정의
	createFunc := conflux.NewCreateFunc(func(ctx context.Context) (*User, error) {
		return &User{
			ID:       "user123",
			Email:    "john@example.com",
			Username: "john",
			Status:   "active",
			Credits:  100,
		}, nil
	})

	// Act
	result, err := repo.FindOneAndInsert(ctx, duplicateCheckFilter, createFunc)

	// Assert
	require.NoError(t, err)
	assert.True(t, result.IsSuccess())
	assert.False(t, result.IsDuplicate())
	assert.Equal(t, "john@example.com", result.GetEntity().Email)
	assert.Equal(t, int64(1), result.GetVersion())
}

func TestRedisRepository_FindOneAndUpsert_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	client := setupRedisClient()
	defer client.Close()

	ctx := context.Background()
	defer cleanupRedis(client, ctx)

	// Repository 생성
	config := &confluxredis.RedisRepositoryConfig{
		KeyPrefix: "user:",
		TTL:       0, // 영구 보존
	}

	repo := confluxredis.NewRedisRepository[*User](client, config, func() *User {
		return &User{}
	})

	// 조회용 필터
	lookupFilter := conflux.NewRedisFilter("user:user456")

	// UpsertFunc 정의
	createFn := func(ctx context.Context) (*User, error) {
		return &User{
			ID:       "user456",
			Email:    "jane@example.com",
			Username: "jane",
			Status:   "active",
			Credits:  50,
		}, nil
	}

	updateFn := func(ctx context.Context, existing *User) (*User, error) {
		existing.Credits += 25
		existing.Status = "premium"
		return existing, nil
	}

	upsertFunc := conflux.NewUpsertFunc(createFn, updateFn)

	// Act
	result, err := repo.FindOneAndUpsert(ctx, lookupFilter, upsertFunc)

	// Assert
	require.NoError(t, err)
	assert.True(t, result.WasCreated())
	assert.False(t, result.WasUpdated())
	assert.Equal(t, "active", result.GetEntity().Status)
	assert.Equal(t, 50, result.GetEntity().Credits)
	assert.Equal(t, int64(1), result.GetVersion())
}

func TestRedisRepository_FindOneAndUpsert_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	client := setupRedisClient()
	defer client.Close()

	ctx := context.Background()
	defer cleanupRedis(client, ctx)

	// Repository 생성
	config := &confluxredis.RedisRepositoryConfig{
		KeyPrefix: "user:",
		TTL:       time.Minute * 30,
	}

	repo := confluxredis.NewRedisRepository[*User](client, config, func() *User {
		return &User{}
	})

	// 먼저 사용자 생성
	lookupFilter := conflux.NewRedisFilter("user:user789")

	createFn := func(ctx context.Context) (*User, error) {
		return &User{
			ID:       "user789",
			Email:    "bob@example.com",
			Username: "bob",
			Status:   "active",
			Credits:  75,
		}, nil
	}

	updateFn := func(ctx context.Context, existing *User) (*User, error) {
		existing.Credits += 25
		existing.Status = "vip"
		return existing, nil
	}

	upsertFunc := conflux.NewUpsertFunc(createFn, updateFn)

	// 첫 번째 호출 (생성)
	_, err := repo.FindOneAndUpsert(ctx, lookupFilter, upsertFunc)
	require.NoError(t, err)

	// Act - 두 번째 호출 (업데이트)
	result, err := repo.FindOneAndUpsert(ctx, lookupFilter, upsertFunc)

	// Assert
	require.NoError(t, err)
	assert.False(t, result.WasCreated())
	assert.True(t, result.WasUpdated())
	assert.Equal(t, "vip", result.GetEntity().Status)
	assert.Equal(t, 100, result.GetEntity().Credits) // 75 + 25
	assert.Equal(t, int64(2), result.GetVersion())
}

func TestRedisRepository_FindOneAndUpdate_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	client := setupRedisClient()
	defer client.Close()

	ctx := context.Background()
	defer cleanupRedis(client, ctx)

	// Repository 생성
	config := &confluxredis.RedisRepositoryConfig{
		KeyPrefix: "user:",
		TTL:       0,
	}

	repo := confluxredis.NewRedisRepository[*User](client, config, func() *User {
		return &User{}
	})

	// 먼저 사용자 생성
	lookupFilter := conflux.NewRedisFilter("user:update_test")
	duplicateCheckFilter := conflux.NewRedisFilter("user:update_test")

	createFunc := conflux.NewCreateFunc(func(ctx context.Context) (*User, error) {
		return &User{
			ID:       "update_test",
			Email:    "update@example.com",
			Username: "updateuser",
			Status:   "active",
			Credits:  200,
		}, nil
	})

	insertResult, err := repo.FindOneAndInsert(ctx, duplicateCheckFilter, createFunc)
	require.NoError(t, err)

	// UpdateFunc 정의
	updateFunc := conflux.NewUpdateFunc(func(ctx context.Context, existing *User) (*User, error) {
		existing.Credits = 300
		existing.Status = "premium"
		return existing, nil
	})

	// Act
	result, err := repo.FindOneAndUpdate(ctx, lookupFilter, insertResult.GetVersion(), updateFunc)

	// Assert
	require.NoError(t, err)
	assert.True(t, result.IsSuccess())
	assert.False(t, result.HasVersionConflict())
	assert.False(t, result.IsNotFound())
	assert.Equal(t, "premium", result.GetEntity().Status)
	assert.Equal(t, 300, result.GetEntity().Credits)
	assert.Equal(t, int64(2), result.GetVersion())
}

func TestRedisRepository_FindOneAndUpdate_VersionConflict(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	client := setupRedisClient()
	defer client.Close()

	ctx := context.Background()
	defer cleanupRedis(client, ctx)

	// Repository 생성
	config := &confluxredis.RedisRepositoryConfig{
		KeyPrefix: "user:",
		TTL:       0,
	}

	repo := confluxredis.NewRedisRepository[*User](client, config, func() *User {
		return &User{}
	})

	// 먼저 사용자 생성
	lookupFilter := conflux.NewRedisFilter("user:conflict_test")
	duplicateCheckFilter := conflux.NewRedisFilter("user:conflict_test")

	createFunc := conflux.NewCreateFunc(func(ctx context.Context) (*User, error) {
		return &User{
			ID:       "conflict_test",
			Email:    "conflict@example.com",
			Username: "conflictuser",
			Status:   "active",
			Credits:  150,
		}, nil
	})

	_, err := repo.FindOneAndInsert(ctx, duplicateCheckFilter, createFunc)
	require.NoError(t, err)

	// UpdateFunc 정의
	updateFunc := conflux.NewUpdateFunc(func(ctx context.Context, existing *User) (*User, error) {
		existing.Credits = 250
		return existing, nil
	})

	// Act - 잘못된 버전으로 업데이트 시도
	result, err := repo.FindOneAndUpdate(ctx, lookupFilter, 999, updateFunc) // 잘못된 버전

	// Assert
	require.NoError(t, err)
	assert.False(t, result.IsSuccess())
	assert.True(t, result.HasVersionConflict())
	assert.False(t, result.IsNotFound())
	assert.Equal(t, int64(1), result.GetVersion()) // 실제 현재 버전
}

// ============================================================================
// Redis 필터 기능 테스트
// ============================================================================

func TestRedisRepository_RedisFilter_PatternMatching(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	client := setupRedisClient()
	defer client.Close()

	ctx := context.Background()
	defer cleanupRedis(client, ctx)

	// Repository 생성
	config := &confluxredis.RedisRepositoryConfig{
		KeyPrefix: "pattern_test:",
		TTL:       0,
	}

	repo := confluxredis.NewRedisRepository[*User](client, config, func() *User {
		return &User{}
	})

	// 여러 사용자 생성
	users := []*User{
		{ID: "admin1", Email: "admin1@test.com", Username: "admin1", Status: "admin"},
		{ID: "user1", Email: "user1@test.com", Username: "user1", Status: "active"},
		{ID: "user2", Email: "user2@test.com", Username: "user2", Status: "active"},
	}

	for _, user := range users {
		duplicateFilter := conflux.NewRedisFilter("pattern_test:" + user.ID)
		createFunc := conflux.NewCreateFunc(func(u *User) func(ctx context.Context) (*User, error) {
			return func(ctx context.Context) (*User, error) {
				return u, nil
			}
		}(user))

		_, err := repo.FindOneAndInsert(ctx, duplicateFilter, createFunc)
		require.NoError(t, err)
	}

	// 패턴으로 검색
	patternFilter := conflux.NewRedisFilter("pattern_test:user*")
	results, err := repo.FindMany(ctx, patternFilter, 10)

	// Assert
	require.NoError(t, err)
	assert.Len(t, results, 2) // user1, user2만 매칭되어야 함
}

// ============================================================================
// 성능 테스트
// ============================================================================

func TestRedisRepository_ConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	client := setupRedisClient()
	defer client.Close()

	ctx := context.Background()
	defer cleanupRedis(client, ctx)

	// Repository 생성
	config := &confluxredis.RedisRepositoryConfig{
		KeyPrefix: "concurrent:",
		TTL:       0,
	}

	repo := confluxredis.NewRedisRepository[*User](client, config, func() *User {
		return &User{}
	})

	userID := "concurrent_user"
	lookupFilter := conflux.NewRedisFilter("concurrent:" + userID)

	createFn := func(ctx context.Context) (*User, error) {
		return &User{
			ID:       userID,
			Email:    "concurrent@test.com",
			Username: "concurrent",
			Status:   "active",
			Credits:  0,
		}, nil
	}

	updateFn := func(ctx context.Context, existing *User) (*User, error) {
		existing.Credits += 10
		return existing, nil
	}

	upsertFunc := conflux.NewUpsertFunc(createFn, updateFn)

	// 동시에 여러 upsert 연산 실행
	numGoroutines := 10
	results := make(chan *conflux.UpsertResult[*User], numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			result, err := repo.FindOneAndUpsert(ctx, lookupFilter, upsertFunc)
			if err != nil {
				errors <- err
				return
			}
			results <- result
		}()
	}

	// 결과 수집
	var createdCount, updatedCount int
	var finalCredits int
	var finalVersion int64

	for i := 0; i < numGoroutines; i++ {
		select {
		case result := <-results:
			if result.WasCreated() {
				createdCount++
			} else {
				updatedCount++
			}
			finalCredits = result.GetEntity().Credits
			if result.GetVersion() > finalVersion {
				finalVersion = result.GetVersion()
			}
		case err := <-errors:
			t.Fatalf("Unexpected error: %v", err)
		}
	}

	// Assert
	assert.Equal(t, 1, createdCount, "Only one creation should occur")
	assert.Equal(t, numGoroutines-1, updatedCount, "All other operations should be updates")
	assert.Equal(t, (numGoroutines-1)*10, finalCredits, "Credits should be accumulated correctly")
	assert.Equal(t, int64(numGoroutines), finalVersion, "Final version should match operation count")
}
