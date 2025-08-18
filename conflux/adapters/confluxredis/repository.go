package confluxredis

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/homveloper/dukdakit/conflux"
	"github.com/redis/go-redis/v9"
)

// ============================================================================
// Redis 기반 동시성 안전 레포지토리 구현
// ============================================================================

// RedisRepository Redis 기반 낙관적 동시성 제어 레포지토리
// RedisFilter를 사용하는 Redis 최적화 구현체입니다
type RedisRepository[T any] struct {
	client      *redis.Client
	keyPrefix   string        // 키 접두사 (예: "user:", "product:")
	ttl         time.Duration // 기본 TTL (0이면 영구 보존)
	newEntityFn func() T
}

// RedisRepositoryConfig Redis 레포지토리 설정
type RedisRepositoryConfig struct {
	KeyPrefix string        // Redis 키 접두사
	TTL       time.Duration // 기본 TTL
}

// NewRedisRepository 새 Redis 레포지토리 생성
func NewRedisRepository[T any](
	client *redis.Client,
	config *RedisRepositoryConfig,
	newEntityFn func() T,
) *RedisRepository[T] {
	if config == nil {
		config = &RedisRepositoryConfig{
			KeyPrefix: "entity:",
			TTL:       0, // 영구 보존
		}
	}

	return &RedisRepository[T]{
		client:      client,
		keyPrefix:   config.KeyPrefix,
		ttl:         config.TTL,
		newEntityFn: newEntityFn,
	}
}

// ============================================================================
// Repository 인터페이스 구현
// ============================================================================

// FindOneAndInsert 원자적 엔터티 생성
func (r *RedisRepository[T]) FindOneAndInsert(
	ctx context.Context,
	duplicateCheckFilter *conflux.RedisFilter,
	createFunc conflux.CreateFunc[T],
) (*conflux.InsertResult[T], error) {
	// 필터 유효성 검증
	if err := duplicateCheckFilter.Validate(); err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}

	// Redis 트랜잭션을 위한 파이프라인 사용
	txf := func(tx *redis.Tx) error {
		// 중복 검사
		exists, err := r.existsByFilter(ctx, tx, duplicateCheckFilter)
		if err != nil {
			return fmt.Errorf("failed to check duplicate: %w", err)
		}

		if exists {
			return fmt.Errorf("duplicate entity exists")
		}

		// 새 엔터티 생성
		newEntity, err := createFunc.CreateFn(ctx)
		if err != nil {
			return fmt.Errorf("failed to create entity: %w", err)
		}

		// 엔터티를 Redis에 저장
		return r.storeEntity(ctx, tx, newEntity, 1)
	}

	// Watch된 키들로 트랜잭션 실행
	watchKeys := r.getWatchKeys(duplicateCheckFilter)

	err := r.client.Watch(ctx, txf, watchKeys...)
	if err != nil {
		// 중복 검사 실패인 경우, 기존 엔터티 반환
		if strings.Contains(err.Error(), "duplicate entity exists") {
			existingEntity, findErr := r.findByFilter(ctx, duplicateCheckFilter)
			if findErr != nil {
				return nil, fmt.Errorf("failed to find existing entity: %w", findErr)
			}
			version, _ := r.getEntityVersion(ctx, existingEntity)
			return conflux.NewDuplicateInsertResult(existingEntity, version), nil
		}
		return nil, fmt.Errorf("failed to insert entity: %w", err)
	}

	// 성공적으로 생성된 엔터티 반환
	newEntity, err := createFunc.CreateFn(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to recreate entity: %w", err)
	}

	return conflux.NewInsertResult(newEntity, 1), nil
}

// FindOneAndUpsert 원자적 엔터티 생성/업데이트
func (r *RedisRepository[T]) FindOneAndUpsert(
	ctx context.Context,
	lookupFilter *conflux.RedisFilter,
	upsertFunc conflux.UpsertFunc[T],
) (*conflux.UpsertResult[T], error) {
	// 필터 유효성 검증
	if err := lookupFilter.Validate(); err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}

	var result *conflux.UpsertResult[T]
	var txErr error

	txf := func(tx *redis.Tx) error {
		// 기존 엔터티 조회
		existing, err := r.findByFilterTx(ctx, tx, lookupFilter)
		if err != nil && !isNotFoundError(err) {
			return fmt.Errorf("failed to find existing entity: %w", err)
		}

		if isNotFoundError(err) {
			// 새로 생성
			newEntity, err := upsertFunc.CreateFn(ctx)
			if err != nil {
				return fmt.Errorf("failed to create entity: %w", err)
			}

			if err := r.storeEntity(ctx, tx, newEntity, 1); err != nil {
				return fmt.Errorf("failed to store new entity: %w", err)
			}

			result = conflux.NewUpsertResult(newEntity, true, 1)
			return nil
		}

		// 기존 엔터티 업데이트
		currentVersion, err := r.getEntityVersionTx(ctx, tx, existing)
		if err != nil {
			return fmt.Errorf("failed to get entity version: %w", err)
		}

		updatedEntity, err := upsertFunc.UpdateFn(ctx, existing)
		if err != nil {
			return fmt.Errorf("failed to update entity: %w", err)
		}

		newVersion := currentVersion + 1
		if err := r.storeEntity(ctx, tx, updatedEntity, newVersion); err != nil {
			return fmt.Errorf("failed to store updated entity: %w", err)
		}

		result = conflux.NewUpsertResult(updatedEntity, false, newVersion)
		return nil
	}

	// Watch된 키들로 트랜잭션 실행
	watchKeys := r.getWatchKeys(lookupFilter)
	txErr = r.client.Watch(ctx, txf, watchKeys...)

	if txErr != nil {
		return nil, fmt.Errorf("failed to upsert entity: %w", txErr)
	}

	return result, nil
}

// FindOneAndUpdate 원자적 엔터티 업데이트
func (r *RedisRepository[T]) FindOneAndUpdate(
	ctx context.Context,
	lookupFilter *conflux.RedisFilter,
	expectedVersion int64,
	updateFunc conflux.UpdateFunc[T],
) (*conflux.UpdateResult[T], error) {
	// 필터 유효성 검증
	if err := lookupFilter.Validate(); err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}

	var result *conflux.UpdateResult[T]
	var txErr error

	txf := func(tx *redis.Tx) error {
		// 기존 엔터티 조회
		existing, err := r.findByFilterTx(ctx, tx, lookupFilter)
		if err != nil {
			if isNotFoundError(err) {
				result = conflux.NewNotFoundResult[T]()
				return nil
			}
			return fmt.Errorf("failed to find existing entity: %w", err)
		}

		// 현재 버전 확인
		currentVersion, err := r.getEntityVersionTx(ctx, tx, existing)
		if err != nil {
			return fmt.Errorf("failed to get entity version: %w", err)
		}

		// 버전 충돌 검사
		if currentVersion != expectedVersion {
			result = conflux.NewVersionConflictResult(existing, currentVersion)
			return nil
		}

		// 엔터티 업데이트
		updatedEntity, err := updateFunc.UpdateFn(ctx, existing)
		if err != nil {
			return fmt.Errorf("failed to update entity: %w", err)
		}

		newVersion := expectedVersion + 1
		if err := r.storeEntity(ctx, tx, updatedEntity, newVersion); err != nil {
			return fmt.Errorf("failed to store updated entity: %w", err)
		}

		result = conflux.NewUpdateResult(updatedEntity, newVersion)
		return nil
	}

	// Watch된 키들로 트랜잭션 실행
	watchKeys := r.getWatchKeys(lookupFilter)
	txErr = r.client.Watch(ctx, txf, watchKeys...)

	if txErr != nil {
		return nil, fmt.Errorf("failed to update entity: %w", txErr)
	}

	return result, nil
}

// NewEntity 새 엔터티 인스턴스 생성
func (r *RedisRepository[T]) NewEntity() T {
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
func (r *RedisRepository[T]) FindOne(ctx context.Context, filter *conflux.RedisFilter) (T, error) {
	var empty T

	if err := filter.Validate(); err != nil {
		return empty, fmt.Errorf("invalid filter: %w", err)
	}

	return r.findByFilter(ctx, filter)
}

// Exists 엔터티 존재 여부 확인
func (r *RedisRepository[T]) Exists(ctx context.Context, filter *conflux.RedisFilter) (bool, error) {
	if err := filter.Validate(); err != nil {
		return false, fmt.Errorf("invalid filter: %w", err)
	}

	return r.existsByFilter(ctx, r.client, filter)
}

// GetVersion 엔터티의 현재 버전 조회
func (r *RedisRepository[T]) GetVersion(ctx context.Context, filter *conflux.RedisFilter) (int64, error) {
	if err := filter.Validate(); err != nil {
		return 0, fmt.Errorf("invalid filter: %w", err)
	}

	entity, err := r.findByFilter(ctx, filter)
	if err != nil {
		return 0, err
	}

	return r.getEntityVersion(ctx, entity)
}

// ============================================================================
// BatchRepository 인터페이스 구현
// ============================================================================

// FindMany 필터 조건으로 여러 엔터티들 조회
func (r *RedisRepository[T]) FindMany(ctx context.Context, filter *conflux.RedisFilter, limit int) ([]T, error) {
	if err := filter.Validate(); err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}

	// Redis 키 패턴으로 검색
	keys, err := r.findKeysByFilter(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find keys: %w", err)
	}

	// 제한 적용
	if limit > 0 && len(keys) > limit {
		keys = keys[:limit]
	}

	// 배치로 엔터티들 조회
	entities, err := r.getEntitiesByKeys(ctx, keys)
	if err != nil {
		return nil, fmt.Errorf("failed to get entities: %w", err)
	}

	return entities, nil
}

// ============================================================================
// Redis 전용 헬퍼 메서드들
// ============================================================================

// findByFilter RedisFilter로 엔터티 조회
func (r *RedisRepository[T]) findByFilter(ctx context.Context, filter *conflux.RedisFilter) (T, error) {
	var empty T

	keys, err := r.findKeysByFilter(ctx, filter)
	if err != nil {
		return empty, err
	}

	if len(keys) == 0 {
		return empty, fmt.Errorf("entity not found")
	}

	// 첫 번째 매칭 키로 엔터티 조회
	entityData, err := r.client.HGet(ctx, keys[0], "data").Result()
	if err != nil {
		if err == redis.Nil {
			return empty, fmt.Errorf("entity not found")
		}
		return empty, fmt.Errorf("failed to get entity data: %w", err)
	}

	var entity T
	if err := json.Unmarshal([]byte(entityData), &entity); err != nil {
		return empty, fmt.Errorf("failed to unmarshal entity: %w", err)
	}

	return entity, nil
}

// findByFilterTx 트랜잭션 내에서 RedisFilter로 엔터티 조회
func (r *RedisRepository[T]) findByFilterTx(ctx context.Context, tx *redis.Tx, filter *conflux.RedisFilter) (T, error) {
	var empty T

	keys, err := r.findKeysByFilterTx(ctx, tx, filter)
	if err != nil {
		return empty, err
	}

	if len(keys) == 0 {
		return empty, fmt.Errorf("entity not found")
	}

	// 첫 번째 매칭 키로 엔터티 조회
	entityData, err := tx.HGet(ctx, keys[0], "data").Result()
	if err != nil {
		if err == redis.Nil {
			return empty, fmt.Errorf("entity not found")
		}
		return empty, fmt.Errorf("failed to get entity data: %w", err)
	}

	var entity T
	if err := json.Unmarshal([]byte(entityData), &entity); err != nil {
		return empty, fmt.Errorf("failed to unmarshal entity: %w", err)
	}

	return entity, nil
}

// existsByFilter 필터 조건에 맞는 엔터티 존재 여부 확인
func (r *RedisRepository[T]) existsByFilter(ctx context.Context, cmdable redis.Cmdable, filter *conflux.RedisFilter) (bool, error) {
	keys, err := r.findKeysByFilterCmdable(ctx, cmdable, filter)
	if err != nil {
		return false, err
	}
	return len(keys) > 0, nil
}

// findKeysByFilter 필터 조건에 맞는 Redis 키들 조회
func (r *RedisRepository[T]) findKeysByFilter(ctx context.Context, filter *conflux.RedisFilter) ([]string, error) {
	return r.findKeysByFilterCmdable(ctx, r.client, filter)
}

// findKeysByFilterTx 트랜잭션 내에서 필터 조건에 맞는 Redis 키들 조회
func (r *RedisRepository[T]) findKeysByFilterTx(ctx context.Context, tx *redis.Tx, filter *conflux.RedisFilter) ([]string, error) {
	return r.findKeysByFilterCmdable(ctx, tx, filter)
}

// findKeysByFilterCmdable Redis Cmdable로 필터 조건에 맞는 키들 조회
func (r *RedisRepository[T]) findKeysByFilterCmdable(ctx context.Context, cmdable redis.Cmdable, filter *conflux.RedisFilter) ([]string, error) {
	var pattern string

	if filter.Pattern != "" {
		pattern = filter.Pattern
	} else if filter.KeyPrefix != "" {
		pattern = filter.KeyPrefix + "*"
		if filter.Suffix != "" {
			pattern = filter.KeyPrefix + "*" + filter.Suffix
		}
	} else {
		pattern = r.keyPrefix + "*"
	}

	return cmdable.Keys(ctx, pattern).Result()
}

// storeEntity 엔터티를 Redis에 저장
func (r *RedisRepository[T]) storeEntity(ctx context.Context, cmdable redis.Cmdable, entity T, version int64) error {
	// 엔터티 ID 추출
	id, err := r.extractID(entity)
	if err != nil {
		return fmt.Errorf("failed to extract entity ID: %w", err)
	}

	key := r.keyPrefix + id

	// 엔터티 직렬화
	entityData, err := json.Marshal(entity)
	if err != nil {
		return fmt.Errorf("failed to marshal entity: %w", err)
	}

	// Redis Hash로 저장 (데이터와 메타데이터 분리)
	pipe := cmdable.TxPipeline()
	pipe.HSet(ctx, key, map[string]interface{}{
		"data":       string(entityData),
		"version":    version,
		"created_at": time.Now().Unix(),
		"updated_at": time.Now().Unix(),
	})

	if r.ttl > 0 {
		pipe.Expire(ctx, key, r.ttl)
	}

	_, err = pipe.Exec(ctx)
	return err
}

// getEntitiesByKeys 키 목록으로 엔터티들 일괄 조회
func (r *RedisRepository[T]) getEntitiesByKeys(ctx context.Context, keys []string) ([]T, error) {
	if len(keys) == 0 {
		return []T{}, nil
	}

	pipe := r.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))

	for i, key := range keys {
		cmds[i] = pipe.HGet(ctx, key, "data")
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get entities: %w", err)
	}

	var entities []T
	for _, cmd := range cmds {
		result, err := cmd.Result()
		if err != nil {
			if err == redis.Nil {
				continue // 키가 존재하지 않음, 건너뛰기
			}
			return nil, fmt.Errorf("failed to get entity data: %w", err)
		}

		var entity T
		if err := json.Unmarshal([]byte(result), &entity); err != nil {
			return nil, fmt.Errorf("failed to unmarshal entity: %w", err)
		}

		entities = append(entities, entity)
	}

	return entities, nil
}

// getEntityVersion 엔터티의 버전 조회
func (r *RedisRepository[T]) getEntityVersion(ctx context.Context, entity T) (int64, error) {
	id, err := r.extractID(entity)
	if err != nil {
		return 0, fmt.Errorf("failed to extract entity ID: %w", err)
	}

	key := r.keyPrefix + id
	return r.client.HGet(ctx, key, "version").Int64()
}

// getEntityVersionTx 트랜잭션 내에서 엔터티의 버전 조회
func (r *RedisRepository[T]) getEntityVersionTx(ctx context.Context, tx *redis.Tx, entity T) (int64, error) {
	id, err := r.extractID(entity)
	if err != nil {
		return 0, fmt.Errorf("failed to extract entity ID: %w", err)
	}

	key := r.keyPrefix + id
	return tx.HGet(ctx, key, "version").Int64()
}

// getWatchKeys 필터에서 감시할 키들 추출
func (r *RedisRepository[T]) getWatchKeys(filter *conflux.RedisFilter) []string {
	// 실제 구현에서는 필터 조건에 맞는 구체적인 키들을 반환해야 합니다
	// 여기서는 간단한 구현으로 패턴 기반으로 키를 생성
	if filter.Pattern != "" {
		// 패턴에서 와일드카드를 제거하여 구체적인 키 생성
		baseKey := strings.ReplaceAll(filter.Pattern, "*", "")
		return []string{baseKey}
	}

	if filter.KeyPrefix != "" {
		return []string{filter.KeyPrefix}
	}

	return []string{r.keyPrefix}
}

// extractID 엔터티에서 ID 추출
func (r *RedisRepository[T]) extractID(entity T) (string, error) {
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

// isNotFoundError 에러가 not found 에러인지 확인
func isNotFoundError(err error) bool {
	return err != nil && (err == redis.Nil || strings.Contains(err.Error(), "not found"))
}
