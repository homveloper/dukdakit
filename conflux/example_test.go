package conflux_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/homveloper/dukdakit/conflux"
	memory "github.com/homveloper/dukdakit/conflux/adapters/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// 테스트용 엔터티 정의
// ============================================================================

// User 테스트용 사용자 엔터티
type User struct {
	conflux.BaseEntity
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Status   string `json:"status"`
	Credits  int    `json:"credits"`
}

// ============================================================================
// FindOneAndInsert 테스트 (새로운 API 스타일)
// ============================================================================

func TestFindOneAndInsert_Success(t *testing.T) {
	// Arrange
	repo := memory.NewMemoryRepository[*User](func() *User {
		return &User{}
	})
	
	ctx := context.Background()
	filter := map[string]any{"Email": "user@example.com"}
	
	// CreateFunc 인터페이스 사용
	createFunc := conflux.NewCreateFunc(func(ctx context.Context) (*User, error) {
		return &User{
			ID:       "user1",
			Email:    "user@example.com",
			Username: "testuser",
			Status:   "active",
			Credits:  100,
		}, nil
	})
	
	// Act
	result, err := repo.FindOneAndInsert(ctx, filter, createFunc)
	
	// Assert
	require.NoError(t, err)
	assert.True(t, result.IsSuccess())
	assert.False(t, result.IsDuplicate())
	assert.Equal(t, "user@example.com", result.GetEntity().Email)
	assert.Equal(t, int64(1), result.GetVersion())
}

func TestFindOneAndInsert_Duplicate(t *testing.T) {
	// Arrange
	repo := memory.NewMemoryRepository[*User](func() *User {
		return &User{}
	})
	
	ctx := context.Background()
	filter := map[string]any{"Email": "user@example.com"}
	
	// 함수형 스타일 사용
	createFn := func(ctx context.Context) (*User, error) {
		return &User{
			ID:       "user1",
			Email:    "user@example.com",
			Username: "testuser",
			Status:   "active",
			Credits:  100,
		}, nil
	}
	
	// 첫 번째 삽입
	_, err := repo.FindOneAndInsert(ctx, filter, conflux.NewCreateFunc(createFn))
	require.NoError(t, err)
	
	// Act - 두 번째 삽입 시도 (중복)
	result, err := repo.FindOneAndInsert(ctx, filter, conflux.NewCreateFunc(createFn))
	
	// Assert
	require.NoError(t, err)
	assert.False(t, result.IsSuccess())
	assert.True(t, result.IsDuplicate())
	assert.Equal(t, "user@example.com", result.GetEntity().Email)
}

// ============================================================================
// FindOneAndUpsert 테스트
// ============================================================================

func TestFindOneAndUpsert_Create(t *testing.T) {
	// Arrange
	repo := memory.NewMemoryRepository[*User](func() *User {
		return &User{}
	})
	
	ctx := context.Background()
	userID := "user1"
	
	// UpsertFunc 생성 - 함수형 방식
	createFn := func(ctx context.Context) (*User, error) {
		return &User{
			ID:       userID,
			Email:    "user@example.com",
			Username: "testuser",
			Status:   "active",
			Credits:  100,
		}, nil
	}
	
	updateFn := func(ctx context.Context, existing *User) (*User, error) {
		existing.Credits += 50
		existing.Status = "updated"
		return existing, nil
	}
	
	upsertFunc := conflux.NewUpsertFunc(createFn, updateFn)
	
	// Act
	filter := conflux.NewMapFilter().And("ID", userID)
	result, err := repo.FindOneAndUpsert(ctx, filter, upsertFunc)
	
	// Assert
	require.NoError(t, err)
	assert.True(t, result.WasCreated())
	assert.False(t, result.WasUpdated())
	assert.Equal(t, "active", result.GetEntity().Status)
	assert.Equal(t, 100, result.GetEntity().Credits)
	assert.Equal(t, int64(1), result.GetVersion())
}

func TestFindOneAndUpsert_Update(t *testing.T) {
	// Arrange
	repo := memory.NewMemoryRepository[*User](func() *User {
		return &User{}
	})
	
	ctx := context.Background()
	userID := "user1"
	
	// 먼저 사용자 생성
	createFn := func(ctx context.Context) (*User, error) {
		return &User{
			ID:       userID,
			Email:    "user@example.com",
			Username: "testuser",
			Status:   "active",
			Credits:  100,
		}, nil
	}
	
	updateFn := func(ctx context.Context, existing *User) (*User, error) {
		existing.Credits += 50
		existing.Status = "premium"
		return existing, nil
	}
	
	upsertFunc := conflux.NewUpsertFunc(createFn, updateFn)
	
	// 첫 번째 호출 (생성)
	filter := conflux.NewMapFilter().And("ID", userID)
	_, err := repo.FindOneAndUpsert(ctx, filter, upsertFunc)
	require.NoError(t, err)
	
	// Act - 두 번째 호출 (업데이트)
	result, err := repo.FindOneAndUpsert(ctx, filter, upsertFunc)
	
	// Assert
	require.NoError(t, err)
	assert.False(t, result.WasCreated())
	assert.True(t, result.WasUpdated())
	assert.Equal(t, "premium", result.GetEntity().Status)
	assert.Equal(t, 150, result.GetEntity().Credits)
	assert.Equal(t, int64(2), result.GetVersion())
}

// ============================================================================
// FindOneAndUpdate 테스트
// ============================================================================

func TestFindOneAndUpdate_Success(t *testing.T) {
	// Arrange
	repo := memory.NewMemoryRepository[*User](func() *User {
		return &User{}
	})
	
	ctx := context.Background()
	userID := "user1"
	
	// 먼저 사용자 생성
	createFunc := conflux.NewCreateFunc(func(ctx context.Context) (*User, error) {
		return &User{
			ID:       userID,
			Email:    "user@example.com",
			Username: "testuser",
			Status:   "active",
			Credits:  100,
		}, nil
	})
	
	filter := map[string]any{"Email": "user@example.com"}
	insertResult, err := repo.FindOneAndInsert(ctx, filter, createFunc)
	require.NoError(t, err)
	
	// UpdateFunc 생성
	updateFunc := conflux.NewUpdateFunc(func(ctx context.Context, existing *User) (*User, error) {
		existing.Credits = 200
		existing.Status = "vip"
		return existing, nil
	})
	
	// Act
	updateFilter := conflux.NewMapFilter().And("ID", userID)
	result, err := repo.FindOneAndUpdate(ctx, updateFilter, insertResult.GetVersion(), updateFunc)
	
	// Assert
	require.NoError(t, err)
	assert.True(t, result.IsSuccess())
	assert.False(t, result.HasVersionConflict())
	assert.False(t, result.IsNotFound())
	assert.Equal(t, "vip", result.GetEntity().Status)
	assert.Equal(t, 200, result.GetEntity().Credits)
	assert.Equal(t, int64(2), result.GetVersion())
}

func TestFindOneAndUpdate_VersionConflict(t *testing.T) {
	// Arrange
	repo := memory.NewMemoryRepository[*User](func() *User {
		return &User{}
	})
	
	ctx := context.Background()
	userID := "user1"
	
	// 먼저 사용자 생성
	createFunc := conflux.NewCreateFunc(func(ctx context.Context) (*User, error) {
		return &User{
			ID:       userID,
			Email:    "user@example.com",
			Username: "testuser",
			Status:   "active",
			Credits:  100,
		}, nil
	})
	
	filter := map[string]any{"Email": "user@example.com"}
	_, err := repo.FindOneAndInsert(ctx, filter, createFunc)
	require.NoError(t, err)
	
	// UpdateFunc 생성
	updateFunc := conflux.NewUpdateFunc(func(ctx context.Context, existing *User) (*User, error) {
		existing.Credits = 200
		return existing, nil
	})
	
	// Act - 잘못된 버전으로 업데이트 시도
	updateFilter := conflux.NewMapFilter().And("ID", userID)
	result, err := repo.FindOneAndUpdate(ctx, updateFilter, 999, updateFunc) // 잘못된 버전
	
	// Assert
	require.NoError(t, err)
	assert.False(t, result.IsSuccess())
	assert.True(t, result.HasVersionConflict())
	assert.False(t, result.IsNotFound())
	assert.Equal(t, int64(1), result.GetVersion()) // 현재 버전
}

func TestFindOneAndUpdate_NotFound(t *testing.T) {
	// Arrange
	repo := memory.NewMemoryRepository[*User](func() *User {
		return &User{}
	})
	
	ctx := context.Background()
	
	// UpdateFunc 생성
	updateFunc := conflux.NewUpdateFunc(func(ctx context.Context, existing *User) (*User, error) {
		existing.Credits = 200
		return existing, nil
	})
	
	// Act - 존재하지 않는 사용자 업데이트 시도
	filter := conflux.NewMapFilter().And("ID", "nonexistent")
	result, err := repo.FindOneAndUpdate(ctx, filter, 1, updateFunc)
	
	// Assert
	require.NoError(t, err)
	assert.False(t, result.IsSuccess())
	assert.False(t, result.HasVersionConflict())
	assert.True(t, result.IsNotFound())
}

// ============================================================================
// 동시성 테스트
// ============================================================================

func TestConcurrentUpsert_RaceConditionSafety(t *testing.T) {
	// Arrange
	repo := memory.NewMemoryRepository[*User](func() *User {
		return &User{}
	})
	
	ctx := context.Background()
	userID := "user1"
	numGoroutines := 10
	
	createFn := func(ctx context.Context) (*User, error) {
		return &User{
			ID:       userID,
			Email:    "user@example.com",
			Username: "testuser",
			Status:   "active",
			Credits:  0,
		}, nil
	}
	
	updateFn := func(ctx context.Context, existing *User) (*User, error) {
		existing.Credits += 10
		return existing, nil
	}
	
	upsertFunc := conflux.NewUpsertFunc(createFn, updateFn)
	
	// Act - 동시에 여러 upsert 연산 실행
	results := make(chan *conflux.UpsertResult[*User], numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func() {
			filter := conflux.NewMapFilter().And("ID", userID)
	result, err := repo.FindOneAndUpsert(ctx, filter, upsertFunc)
			require.NoError(t, err)
			results <- result
		}()
	}
	
	// 결과 수집
	var createdCount, updatedCount int
	var finalCredits int
	var finalVersion int64
	
	for i := 0; i < numGoroutines; i++ {
		result := <-results
		if result.WasCreated() {
			createdCount++
		} else {
			updatedCount++
		}
		finalCredits = result.GetEntity().Credits
		if result.GetVersion() > finalVersion {
			finalVersion = result.GetVersion()
		}
	}
	
	// Assert
	assert.Equal(t, 1, createdCount, "Only one creation should occur")
	assert.Equal(t, numGoroutines-1, updatedCount, "All other operations should be updates")
	assert.Equal(t, (numGoroutines-1)*10, finalCredits, "Credits should be accumulated correctly")
	assert.Equal(t, int64(numGoroutines), finalVersion, "Final version should match operation count")
}

// ============================================================================
// 실제 사용 시나리오 예제 테스트
// ============================================================================

func TestRealWorldScenario_UserRegistrationAndCredits(t *testing.T) {
	// 실제 사용 시나리오: 사용자 등록 및 크레딧 관리
	repo := memory.NewMemoryRepository[*User](func() *User {
		return &User{}
	})
	
	ctx := context.Background()
	
	t.Run("사용자 등록 (중복 이메일 방지)", func(t *testing.T) {
		email := "john@example.com"
		filter := map[string]any{"Email": email}
		
		createFunc := conflux.NewCreateFunc(func(ctx context.Context) (*User, error) {
			return &User{
				ID:       "john123",
				Email:    email,
				Username: "john",
				Status:   "active",
				Credits:  100, // 가입 보너스
			}, nil
		})
		
		// 첫 번째 등록 시도
		result1, err := repo.FindOneAndInsert(ctx, filter, createFunc)
		require.NoError(t, err)
		assert.True(t, result1.IsSuccess())
		assert.Equal(t, 100, result1.GetEntity().Credits)
		
		// 같은 이메일로 두 번째 등록 시도 (중복 방지)
		result2, err := repo.FindOneAndInsert(ctx, filter, createFunc)
		require.NoError(t, err)
		assert.False(t, result2.IsSuccess())
		assert.True(t, result2.IsDuplicate())
	})
	
	t.Run("크레딧 충전 및 사용", func(t *testing.T) {
		userID := "john123"
		
		// 크레딧 충전 로직
		addCreditsFunc := conflux.NewUpdateFunc(func(ctx context.Context, existing *User) (*User, error) {
			existing.Credits += 50
			return existing, nil
		})
		
		// 현재 버전 가져오기 위해 먼저 조회 (실제로는 별도 조회 메서드 필요)
		currentVersion := int64(1)
		
		filter := conflux.NewMapFilter().And("ID", userID)
		result, err := repo.FindOneAndUpdate(ctx, filter, currentVersion, addCreditsFunc)
		require.NoError(t, err)
		assert.True(t, result.IsSuccess())
		assert.Equal(t, 150, result.GetEntity().Credits)
		
		// 크레딧 사용 로직
		useCreditsFunc := conflux.NewUpdateFunc(func(ctx context.Context, existing *User) (*User, error) {
			if existing.Credits < 30 {
				return existing, fmt.Errorf("insufficient credits")
			}
			existing.Credits -= 30
			return existing, nil
		})
		
		filter2 := conflux.NewMapFilter().And("ID", userID)
		result2, err := repo.FindOneAndUpdate(ctx, filter2, result.GetVersion(), useCreditsFunc)
		require.NoError(t, err)
		assert.True(t, result2.IsSuccess())
		assert.Equal(t, 120, result2.GetEntity().Credits)
	})
}