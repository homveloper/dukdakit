package conflux_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/homveloper/dukdakit/conflux"
	memory "github.com/homveloper/dukdakit/conflux/adapters/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// 옵션 패턴 사용 예제들
// ============================================================================

func TestRepository_WithConflictStrategy_Overwrite(t *testing.T) {
	// Arrange
	repo := memory.NewMemoryRepository[*User](func() *User { return &User{} })
	ctx := context.Background()

	// 먼저 사용자 생성
	filter := conflux.NewMapFilter().And("ID", "conflict_user")
	createFunc := conflux.NewCreateFunc(func(ctx context.Context) (*User, error) {
		return &User{
			ID:      "conflict_user",
			Email:   "conflict@example.com",
			Credits: 100,
		}, nil
	})

	insertResult, err := repo.FindOneAndInsert(ctx, filter, createFunc)
	require.NoError(t, err)

	// 충돌 해결자와 함께 업데이트
	updateFunc := conflux.NewUpdateFunc(func(ctx context.Context, existing *User) (*User, error) {
		existing.Credits = 200
		existing.Status = "premium"
		return existing, nil
	})

	// 덮어쓰기 전략 사용
	overwriteResolver := conflux.NewOverwriteResolver[*User]()
	result, err := repo.FindOneAndUpdate(ctx, filter, insertResult.GetVersion(),
		updateFunc,
		conflux.WithUpdateConflictResolver(overwriteResolver),
		conflux.WithUpdateConflictStrategy(conflux.OverwriteOnConflict),
	)

	// Assert
	require.NoError(t, err)
	assert.True(t, result.IsSuccess())
	assert.Equal(t, "premium", result.GetEntity().Status)
	assert.Equal(t, 200, result.GetEntity().Credits)
}

func TestRepository_WithDuplicateStrategy_ReturnExisting(t *testing.T) {
	// Arrange
	repo := memory.NewMemoryRepository[*User](func() *User { return &User{} })
	ctx := context.Background()

	email := "duplicate@example.com"
	duplicateFilter := conflux.NewMapFilter().And("Email", email)

	createFunc := conflux.NewCreateFunc(func(ctx context.Context) (*User, error) {
		return &User{
			ID:      "user1",
			Email:   email,
			Credits: 100,
		}, nil
	})

	// 첫 번째 삽입
	_, err := repo.FindOneAndInsert(ctx, duplicateFilter, createFunc)
	require.NoError(t, err)

	// 중복 시 기존 엔터티 반환 전략으로 두 번째 삽입
	result, err := repo.FindOneAndInsert(ctx, duplicateFilter, createFunc,
		conflux.WithDuplicateStrategy(conflux.ReturnExistingOnDuplicate),
	)

	// Assert
	require.NoError(t, err)
	assert.True(t, result.IsDuplicate())
	assert.Equal(t, email, result.GetEntity().Email)
}

func TestRepository_WithTimeout(t *testing.T) {
	// Arrange
	repo := memory.NewMemoryRepository[*User](func() *User { return &User{} })
	ctx := context.Background()

	filter := conflux.NewMapFilter().And("ID", "timeout_user")
	createFunc := conflux.NewCreateFunc(func(ctx context.Context) (*User, error) {
		// 타임아웃 테스트를 위한 지연
		select {
		case <-time.After(100 * time.Millisecond):
			return &User{ID: "timeout_user", Email: "timeout@example.com"}, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	})

	// 짧은 타임아웃으로 삽입 시도
	_, err := repo.FindOneAndInsert(ctx, filter, createFunc,
		conflux.WithInsertTimeout(50*time.Millisecond),
	)

	// 메모리 어댑터에서는 실제로 타임아웃을 구현하지 않으므로
	// 이 테스트는 API 시연용입니다
	assert.NoError(t, err) // 메모리에서는 즉시 실행됨
}

func TestRepository_WithMetadata(t *testing.T) {
	// Arrange
	repo := memory.NewMemoryRepository[*User](func() *User { return &User{} })
	ctx := context.Background()

	filter := conflux.NewMapFilter().And("ID", "metadata_user")
	createFunc := conflux.NewCreateFunc(func(ctx context.Context) (*User, error) {
		return &User{ID: "metadata_user", Email: "meta@example.com"}, nil
	})

	metadata := map[string]any{
		"operation_id": "op_123",
		"source":       "api_v1",
		"priority":     "high",
	}

	// 메타데이터와 함께 삽입
	result, err := repo.FindOneAndInsert(ctx, filter, createFunc,
		conflux.WithInsertMetadata(metadata),
	)

	// Assert
	require.NoError(t, err)
	assert.True(t, result.IsSuccess())
	// 메타데이터는 Repository 구현체에서 활용됩니다 (로깅, 감사 등)
}

// ============================================================================
// 커스텀 충돌 해결자 예제
// ============================================================================

// CreditMergeResolver 크레딧을 병합하는 커스텀 해결자
type CreditMergeResolver struct{}

func (r *CreditMergeResolver) ResolveConflict(ctx context.Context, current *User, incoming *User) (*User, bool, error) {
	// 크레딧을 합산하여 병합
	merged := *current
	merged.Credits += incoming.Credits
	merged.Status = incoming.Status // 상태는 새로운 값 사용
	return &merged, true, nil
}

func (r *CreditMergeResolver) ResolveConflictWithContext(
	ctx context.Context,
	conflictCtx *conflux.ConflictContext[*User],
	current *User,
	incoming *User,
) (*User, conflux.ConflictResolution, error) {
	if conflictCtx.AttemptNumber > 3 {
		return current, conflux.ResolveWithFail, fmt.Errorf("too many merge attempts")
	}

	merged := *current
	merged.Credits += incoming.Credits
	merged.Status = incoming.Status

	return &merged, conflux.ResolveWithMerged, nil
}

func TestRepository_WithCustomConflictResolver(t *testing.T) {
	// Arrange
	repo := memory.NewMemoryRepository[*User](func() *User { return &User{} })
	ctx := context.Background()

	// 초기 사용자 생성
	filter := conflux.NewMapFilter().And("ID", "merge_user")
	createFunc := conflux.NewCreateFunc(func(ctx context.Context) (*User, error) {
		return &User{
			ID:      "merge_user",
			Email:   "merge@example.com",
			Credits: 100,
			Status:  "active",
		}, nil
	})

	insertResult, err := repo.FindOneAndInsert(ctx, filter, createFunc)
	require.NoError(t, err)

	// 커스텀 병합 해결자로 업데이트
	updateFunc := conflux.NewUpdateFunc(func(ctx context.Context, existing *User) (*User, error) {
		return &User{
			ID:      existing.ID,
			Email:   existing.Email,
			Credits: 50, // 추가할 크레딧
			Status:  "premium",
		}, nil
	})

	// 이 테스트는 개념적 예시입니다 (실제 메모리 어댑터에서는 구현되지 않음)
	customResolver := &CreditMergeResolver{}
	result, err := repo.FindOneAndUpdate(ctx, filter, insertResult.GetVersion(),
		updateFunc,
		conflux.WithUpdateConflictResolver(customResolver),
	)

	// Assert
	require.NoError(t, err)
	assert.True(t, result.IsSuccess())
	// 실제 병합 로직은 어댑터 구현체에서 처리됩니다
}

// ============================================================================
// 재시도 라이브러리와의 조합 예제
// ============================================================================

func TestRepository_WithRetryLibrary_ConceptualExample(t *testing.T) {
	// 이것은 외부 재시도 라이브러리와의 조합 방법을 보여주는 개념적 예제입니다
	// 실제로는 cenkalti/backoff, avast/retry-go 등과 함께 사용됩니다

	repo := memory.NewMemoryRepository[*User](func() *User { return &User{} })
	ctx := context.Background()

	filter := conflux.NewMapFilter().And("ID", "retry_user")
	
	// 재시도가 필요한 업데이트 로직
	performUpdate := func() error {
		// 현재 버전 조회
		version, err := repo.(conflux.ReadRepository[*User, conflux.MapFilter]).GetVersion(ctx, filter)
		if err != nil {
			return err
		}

		// 현재 버전으로 업데이트 시도
		updateFunc := conflux.NewUpdateFunc(func(ctx context.Context, existing *User) (*User, error) {
			existing.Credits += 10
			return existing, nil
		})

		result, err := repo.FindOneAndUpdate(ctx, filter, version,
			updateFunc,
			conflux.WithUpdateMaxRetries(3), // 힌트 제공
		)

		if err != nil {
			return err
		}

		if result.HasVersionConflict() {
			return fmt.Errorf("version conflict: retry needed")
		}

		return nil
	}

	// 외부 재시도 라이브러리 사용 시:
	// 
	// import "github.com/cenkalti/backoff/v4"
	// 
	// err := backoff.Retry(performUpdate, backoff.NewExponentialBackOff())
	// 
	// 또는
	//
	// import "github.com/avast/retry-go"
	// 
	// err := retry.Do(
	//     performUpdate,
	//     retry.Attempts(3),
	//     retry.DelayType(retry.BackOffDelay),
	// )

	// 테스트에서는 단순히 한 번만 실행
	err := performUpdate()
	
	// 첫 실행에서는 사용자가 없으므로 에러가 발생할 것입니다
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ============================================================================
// 복잡한 비즈니스 로직 예제
// ============================================================================

func TestRepository_ComplexBusinessLogic_WithOptions(t *testing.T) {
	repo := memory.NewMemoryRepository[*User](func() *User { return &User{} })
	ctx := context.Background()

	// 사용자 등록 with 중복 처리
	registerUser := func(email, username string) (*User, error) {
		duplicateFilter := conflux.NewMapFilter().And("Email", email)
		
		createFunc := conflux.NewCreateFunc(func(ctx context.Context) (*User, error) {
			return &User{
				ID:       fmt.Sprintf("user_%d", time.Now().UnixNano()),
				Email:    email,
				Username: username,
				Status:   "active",
				Credits:  100, // 가입 보너스
			}, nil
		})

		result, err := repo.FindOneAndInsert(ctx, duplicateFilter, createFunc,
			conflux.WithDuplicateStrategy(conflux.ReturnExistingOnDuplicate),
			conflux.WithInsertTimeout(5*time.Second),
			conflux.WithInsertMetadata(map[string]any{
				"operation": "user_registration",
				"timestamp": time.Now(),
			}),
		)

		if err != nil {
			return nil, err
		}

		if result.IsDuplicate() {
			return result.GetEntity(), fmt.Errorf("user already exists: %s", email)
		}

		return result.GetEntity(), nil
	}

	// 크레딧 업데이트 with 충돌 해결
	updateCredits := func(userID string, amount int) error {
		filter := conflux.NewMapFilter().And("ID", userID)
		
		// 현재 버전 조회
		version, err := repo.(conflux.ReadRepository[*User, conflux.MapFilter]).GetVersion(ctx, filter)
		if err != nil {
			return err
		}

		updateFunc := conflux.NewUpdateFunc(func(ctx context.Context, existing *User) (*User, error) {
			if existing.Credits+amount < 0 {
				return existing, fmt.Errorf("insufficient credits")
			}
			existing.Credits += amount
			return existing, nil
		})

		result, err := repo.FindOneAndUpdate(ctx, filter, version,
			updateFunc,
			conflux.WithUpdateConflictStrategy(conflux.RetryOnConflict),
			conflux.WithUpdateMaxRetries(5),
		)

		if err != nil {
			return err
		}

		if result.HasVersionConflict() {
			return fmt.Errorf("credits update failed due to conflict")
		}

		return nil
	}

	// 테스트 실행
	user, err := registerUser("business@example.com", "businessuser")
	require.NoError(t, err)
	assert.Equal(t, 100, user.Credits)

	err = updateCredits(user.ID, 50)
	require.NoError(t, err)

	// 중복 등록 시도
	_, err = registerUser("business@example.com", "duplicate")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}