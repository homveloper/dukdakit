package friendit

import (
	"context"
	"time"
)

// ============================================================================
// 원자적 연산 결과와 팩토리 인터페이스 - 동시성 안전성 보장
// ============================================================================

// AtomicResult 원자적 연산의 결과
type AtomicResult[T any] struct {
	Entity     T         // 결과 엔터티
	Created    bool      // true면 생성, false면 업데이트
	Version    int64     // 낙관적 잠금을 위한 버전
	ModifiedAt time.Time // 수정 시간
}

// EntityFactory 엔터티 생성/업데이트 로직을 캡슐화하는 팩토리 인터페이스
type EntityFactory[T any] interface {
	// CreateFn: 새 엔터티 생성 시 도메인 로직 적용
	CreateFn(ctx context.Context) (T, error)
	// UpdateFn: 기존 엔터티 업데이트 시 도메인 로직 적용  
	UpdateFn(ctx context.Context, existing T) (T, error)
}

// ConflictResolver 동시성 충돌 해결 전략
type ConflictResolver[T any] interface {
	Resolve(ctx context.Context, current T, incoming T) (T, error)
}

// ============================================================================
// 헬퍼 함수들
// ============================================================================

// NewAtomicResult 원자적 연산 결과 생성 헬퍼
func NewAtomicResult[T any](entity T, created bool, version int64) *AtomicResult[T] {
	return &AtomicResult[T]{
		Entity:     entity,
		Created:    created,
		Version:    version,
		ModifiedAt: time.Now(),
	}
}

// IsCreated 엔터티가 새로 생성되었는지 확인
func (r *AtomicResult[T]) IsCreated() bool {
	return r.Created
}

// IsUpdated 엔터티가 업데이트되었는지 확인
func (r *AtomicResult[T]) IsUpdated() bool {
	return !r.Created
}

// GetVersion 현재 버전 반환
func (r *AtomicResult[T]) GetVersion() int64 {
	return r.Version
}

// ============================================================================
// 엔터티별 메타데이터 확장 인터페이스
// ============================================================================

// FriendRequestEntity 인터페이스에 누락된 메서드들 추가
type EnhancedFriendRequestEntity interface {
	FriendRequestEntity
	GetMessage() string
	GetMetadata() map[string]any
	GetReason() string // rejection reason
}

// BlockRelationEntity 인터페이스에 누락된 메서드들 추가  
type EnhancedBlockRelationEntity interface {
	BlockRelationEntity
	GetReason() string
}