package memory

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/homveloper/dukdakit/conflux"
)

// ============================================================================
// 메모리 기반 동시성 안전 레포지토리 구현
// ============================================================================

// MemoryRepository 메모리 기반 낙관적 동시성 제어 레포지토리
// MapFilter를 사용하는 범용 인메모리 구현체입니다
type MemoryRepository[T any] struct {
	mu          sync.RWMutex
	entities    map[string]*versionedEntity[T] // ID를 string으로 통일
	versions    map[string]int64
	newEntityFn func() T
}

// versionedEntity 버전 정보를 포함한 엔터티 래퍼
type versionedEntity[T any] struct {
	Entity    T
	Version   int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewMemoryRepository 새 메모리 레포지토리 생성
func NewMemoryRepository[T any](newEntityFn func() T) interface {
	conflux.Repository[T, conflux.MapFilter]
	conflux.ReadRepository[T, conflux.MapFilter]
} {
	return &MemoryRepository[T]{
		entities:    make(map[string]*versionedEntity[T]),
		versions:    make(map[string]int64),
		newEntityFn: newEntityFn,
	}
}

// ============================================================================
// Repository 인터페이스 구현
// ============================================================================

// FindOneAndInsert 원자적 엔터티 생성
func (r *MemoryRepository[T]) FindOneAndInsert(
	ctx context.Context,
	duplicateCheckFilter conflux.MapFilter,
	createFunc conflux.CreateFunc[T],
	options ...conflux.InsertOption,
) (*conflux.InsertResult[T], error) {
	// 옵션 처리
	config := conflux.NewInsertConfig()
	for _, opt := range options {
		opt(config)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// 필터 유효성 검증
	if err := duplicateCheckFilter.Validate(); err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}

	// 필터 조건에 맞는 기존 엔터티 검사
	existingID, exists := r.findByMapFilter(duplicateCheckFilter)
	if exists {
		// 중복 처리 전략에 따라 동작
		existing := r.entities[existingID]
		
		switch config.OnDuplicate {
		case conflux.FailOnDuplicate:
			return conflux.NewDuplicateInsertResult(existing.Entity, existing.Version), nil
		case conflux.ReturnExistingOnDuplicate:
			return conflux.NewDuplicateInsertResult(existing.Entity, existing.Version), nil
		case conflux.UpdateOnDuplicate:
			// UpdateOnDuplicate는 여기서는 단순히 기존 엔터티 반환
			// 실제로는 더 복잡한 로직이 필요할 수 있음
			return conflux.NewDuplicateInsertResult(existing.Entity, existing.Version), nil
		default:
			return conflux.NewDuplicateInsertResult(existing.Entity, existing.Version), nil
		}
	}

	// 새 엔터티 생성
	newEntity, err := createFunc.CreateFn(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}

	// ID 추출
	id, err := r.extractID(newEntity)
	if err != nil {
		return nil, fmt.Errorf("failed to extract ID from entity: %w", err)
	}

	// 버전 및 타임스탬프 설정
	version := int64(1)
	now := time.Now()
	r.setEntityVersion(newEntity, version)
	r.setEntityTimestamps(newEntity, now, now)

	// 저장
	versioned := &versionedEntity[T]{
		Entity:    newEntity,
		Version:   version,
		CreatedAt: now,
		UpdatedAt: now,
	}

	r.entities[id] = versioned
	r.versions[id] = version

	return conflux.NewInsertResult(newEntity, version), nil
}

// FindOneAndUpsert 원자적 엔터티 생성/업데이트
func (r *MemoryRepository[T]) FindOneAndUpsert(
	ctx context.Context,
	lookupFilter conflux.MapFilter,
	upsertFunc conflux.UpsertFunc[T],
	options ...conflux.UpsertOption,
) (*conflux.UpsertResult[T], error) {
	// 옵션 처리
	config := conflux.NewUpsertConfig()
	for _, opt := range options {
		opt(config)
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	// 필터 유효성 검증
	if err := lookupFilter.Validate(); err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}

	// 필터로 기존 엔터티 조회
	existingID, exists := r.findByMapFilter(lookupFilter)

	if !exists {
		// 새로 생성
		newEntity, err := upsertFunc.CreateFn(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create entity: %w", err)
		}

		// ID 추출
		id, err := r.extractID(newEntity)
		if err != nil {
			return nil, fmt.Errorf("failed to extract ID from entity: %w", err)
		}

		// 버전 및 타임스탬프 설정
		version := int64(1)
		now := time.Now()
		r.setEntityVersion(newEntity, version)
		r.setEntityTimestamps(newEntity, now, now)

		// 저장
		versioned := &versionedEntity[T]{
			Entity:    newEntity,
			Version:   version,
			CreatedAt: now,
			UpdatedAt: now,
		}

		r.entities[id] = versioned
		r.versions[id] = version

		return conflux.NewUpsertResult(newEntity, true, version), nil
	}

	// 기존 엔터티 업데이트
	existing := r.entities[existingID]
	updatedEntity, err := upsertFunc.UpdateFn(ctx, existing.Entity)
	if err != nil {
		return nil, fmt.Errorf("failed to update entity: %w", err)
	}

	// 새 버전 설정
	newVersion := existing.Version + 1
	now := time.Now()
	r.setEntityVersion(updatedEntity, newVersion)
	r.setEntityTimestamp(updatedEntity, now)

	// 저장
	versioned := &versionedEntity[T]{
		Entity:    updatedEntity,
		Version:   newVersion,
		CreatedAt: existing.CreatedAt,
		UpdatedAt: now,
	}

	r.entities[existingID] = versioned
	r.versions[existingID] = newVersion

	return conflux.NewUpsertResult(updatedEntity, false, newVersion), nil
}

// FindOneAndUpdate 원자적 엔터티 업데이트
func (r *MemoryRepository[T]) FindOneAndUpdate(
	ctx context.Context,
	lookupFilter conflux.MapFilter,
	expectedVersion int64,
	updateFunc conflux.UpdateFunc[T],
	options ...conflux.UpdateOption,
) (*conflux.UpdateResult[T], error) {
	// 옵션 처리
	config := conflux.NewUpdateConfig()
	for _, opt := range options {
		opt(config)
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	// 필터 유효성 검증
	if err := lookupFilter.Validate(); err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}

	// 필터로 기존 엔터티 조회
	existingID, exists := r.findByMapFilter(lookupFilter)
	if !exists {
		return conflux.NewNotFoundResult[T](), nil
	}

	existing := r.entities[existingID]

	// 버전 충돌 검사
	if existing.Version != expectedVersion {
		// 충돌 해결자가 있는 경우 사용
		if config.ConflictResolver != nil {
			// 컨텍스트 생성
			conflictCtx := &conflux.ConflictContext[T]{
				Operation:       "update",
				ExpectedVersion: expectedVersion,
				CurrentVersion:  existing.Version,
				AttemptNumber:   1, // 단순화
				Metadata:        config.Metadata,
			}
			
			// 향상된 해결자 시도
			if enhancedResolver, ok := config.ConflictResolver.(conflux.EnhancedConflictResolver[T]); ok {
				var zeroValue T
				resolved, resolution, err := enhancedResolver.ResolveConflictWithContext(
					ctx, conflictCtx, existing.Entity, zeroValue)
				
				if err != nil {
					return nil, fmt.Errorf("conflict resolution failed: %w", err)
				}
				
				switch resolution {
				case conflux.ResolveWithCurrent:
					return conflux.NewUpdateResult(existing.Entity, existing.Version), nil
				case conflux.ResolveWithRetry:
					return conflux.NewVersionConflictResult(existing.Entity, existing.Version), nil
				case conflux.ResolveWithFail:
					return nil, fmt.Errorf("conflict resolution failed")
				default:
					// ResolveWithMerged 또는 기타
					// resolved는 이미 T 타입이므로 직접 사용
					newVersion := existing.Version + 1
					now := time.Now()
					r.setEntityVersion(resolved, newVersion)
					r.setEntityTimestamp(resolved, now)
					
					versioned := &versionedEntity[T]{
						Entity:    resolved,
						Version:   newVersion,
						CreatedAt: existing.CreatedAt,
						UpdatedAt: now,
					}
					
					r.entities[existingID] = versioned
					r.versions[existingID] = newVersion
					
					return conflux.NewUpdateResult(resolved, newVersion), nil
				}
			}
		}
		
		// 기본 충돌 처리 전략
		switch config.OnConflict {
		case conflux.OverwriteOnConflict:
			// 버전 무시하고 강제 업데이트 (주의: 데이터 손실 가능)
			// 이 경우에도 updateFunc을 실행해야 함
		case conflux.RetryOnConflict:
			return conflux.NewVersionConflictResult(existing.Entity, existing.Version), nil
		default: // FailOnConflict
			return conflux.NewVersionConflictResult(existing.Entity, existing.Version), nil
		}
	}

	// 엔터티 업데이트
	updatedEntity, err := updateFunc.UpdateFn(ctx, existing.Entity)
	if err != nil {
		return nil, fmt.Errorf("failed to update entity: %w", err)
	}

	// 새 버전 설정
	newVersion := expectedVersion + 1
	now := time.Now()
	r.setEntityVersion(updatedEntity, newVersion)
	r.setEntityTimestamp(updatedEntity, now)

	// 저장
	versioned := &versionedEntity[T]{
		Entity:    updatedEntity,
		Version:   newVersion,
		CreatedAt: existing.CreatedAt,
		UpdatedAt: now,
	}

	r.entities[existingID] = versioned
	r.versions[existingID] = newVersion

	return conflux.NewUpdateResult(updatedEntity, newVersion), nil
}

// NewEntity 새 엔터티 인스턴스 생성
func (r *MemoryRepository[T]) NewEntity() T {
	if r.newEntityFn != nil {
		return r.newEntityFn()
	}

	var entity T
	return entity
}

// ============================================================================
// ReadRepository 인터페이스 구현
// ============================================================================

// FindOne 필터 조건으로 엔터티 조회
func (r *MemoryRepository[T]) FindOne(ctx context.Context, filter conflux.MapFilter) (T, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var empty T

	if err := filter.Validate(); err != nil {
		return empty, fmt.Errorf("invalid filter: %w", err)
	}

	entityID, exists := r.findByMapFilter(filter)
	if !exists {
		return empty, fmt.Errorf("entity not found")
	}

	return r.entities[entityID].Entity, nil
}

// Exists 엔터티 존재 여부 확인
func (r *MemoryRepository[T]) Exists(ctx context.Context, filter conflux.MapFilter) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := filter.Validate(); err != nil {
		return false, fmt.Errorf("invalid filter: %w", err)
	}

	_, exists := r.findByMapFilter(filter)
	return exists, nil
}

// GetVersion 엔터티의 현재 버전 조회
func (r *MemoryRepository[T]) GetVersion(ctx context.Context, filter conflux.MapFilter) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := filter.Validate(); err != nil {
		return 0, fmt.Errorf("invalid filter: %w", err)
	}

	entityID, exists := r.findByMapFilter(filter)
	if !exists {
		return 0, fmt.Errorf("entity not found")
	}

	return r.entities[entityID].Version, nil
}

// ============================================================================
// BatchRepository 인터페이스 구현
// ============================================================================

// FindMany 필터 조건으로 여러 엔터티들 조회
func (r *MemoryRepository[T]) FindMany(ctx context.Context, filter conflux.MapFilter, limit int) ([]T, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := filter.Validate(); err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}

	var results []T
	count := 0

	for _, versioned := range r.entities {
		if limit > 0 && count >= limit {
			break
		}

		if r.matchesMapFilter(versioned.Entity, filter) {
			results = append(results, versioned.Entity)
			count++
		}
	}

	return results, nil
}

// ============================================================================
// 헬퍼 메서드들
// ============================================================================

// findByMapFilter MapFilter 조건에 맞는 엔터티 ID 찾기
func (r *MemoryRepository[T]) findByMapFilter(filter conflux.MapFilter) (string, bool) {
	for id, versioned := range r.entities {
		if r.matchesMapFilter(versioned.Entity, filter) {
			return id, true
		}
	}
	return "", false
}

// matchesMapFilter 엔터티가 MapFilter 조건에 맞는지 확인
func (r *MemoryRepository[T]) matchesMapFilter(entity T, filter conflux.MapFilter) bool {
	entityValue := reflect.ValueOf(entity)
	if entityValue.Kind() == reflect.Ptr {
		entityValue = entityValue.Elem()
	}

	for key, expectedValue := range filter {
		fieldValue := entityValue.FieldByName(key)
		if !fieldValue.IsValid() {
			return false
		}

		actualValue := fieldValue.Interface()
		if !reflect.DeepEqual(actualValue, expectedValue) {
			return false
		}
	}

	return true
}

// extractID 엔터티에서 ID 추출 (string으로 변환)
func (r *MemoryRepository[T]) extractID(entity T) (string, error) {
	entityValue := reflect.ValueOf(entity)
	if entityValue.Kind() == reflect.Ptr {
		entityValue = entityValue.Elem()
	}

	// ID 필드 찾기
	idField := entityValue.FieldByName("ID")
	if !idField.IsValid() {
		return "", fmt.Errorf("entity does not have ID field")
	}

	if !idField.CanInterface() {
		return "", fmt.Errorf("ID field is not accessible")
	}

	// ID를 string으로 변환
	id := fmt.Sprintf("%v", idField.Interface())
	return id, nil
}

// setEntityVersion 엔터티의 버전 설정 (Versioned 인터페이스 구현 시)
func (r *MemoryRepository[T]) setEntityVersion(entity T, version int64) {
	if versioned, ok := any(entity).(conflux.Versioned); ok {
		versioned.SetVersion(version)
	}
}

// setEntityTimestamps 엔터티의 타임스탬프 설정 (Timestamped 인터페이스 구현 시)
func (r *MemoryRepository[T]) setEntityTimestamps(entity T, createdAt, updatedAt time.Time) {
	if timestamped, ok := any(entity).(conflux.Timestamped); ok {
		timestamped.SetCreatedAt(createdAt)
		timestamped.SetUpdatedAt(updatedAt)
	}
}

// setEntityTimestamp 엔터티의 업데이트 타임스탬프 설정 (Timestamped 인터페이스 구현 시)
func (r *MemoryRepository[T]) setEntityTimestamp(entity T, updatedAt time.Time) {
	if timestamped, ok := any(entity).(conflux.Timestamped); ok {
		timestamped.SetUpdatedAt(updatedAt)
	}
}