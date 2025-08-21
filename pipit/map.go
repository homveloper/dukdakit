package pipit

import (
	"context"
	"fmt"
)

// MapOperation은 파이프라인에서 타입 변환 연산을 수행합니다.
// T 타입에서 U 타입으로의 변환을 지원하며, 단순 매퍼와 컨텍스트 인식 매퍼를 모두 지원합니다.
//
// MapOperation은 Operation 인터페이스를 구현하며 파이프라인의 각 아이템에 적용되어
// 새로운 타입의 값으로 변환합니다.
//
// 사용 예시:
//
//	data := []int{1, 2, 3, 4, 5}
//	result := Map(From(data), func(x int) string { return fmt.Sprintf("num_%d", x) })
//	// result: Query[string] -> ["num_1", "num_2", "num_3", "num_4", "num_5"]
type MapOperation[T, U any] struct {
	// mapper는 단순한 타입 변환 함수입니다
	mapper func(T) U

	// safeMapperWithContext는 컨텍스트 인식 변환 함수로 에러 처리를 지원합니다
	// 이를 통해 실패할 수 있는 연산이나 취소가 필요한 연산을 처리할 수 있습니다
	safeMapperWithContext func(context.Context, T) (U, error)
}

// Type은 MapOperation의 연산 타입 식별자를 반환합니다.
// Operation 인터페이스를 구현합니다.
func (op *MapOperation[T, U]) Type() OperationType {
	return MapOp
}

// Apply는 단일 아이템에 대해 매핑 연산을 실행합니다.
// 입력 아이템을 새로운 타입의 값으로 변환하여 반환합니다.
//
// 이 메소드는 단순 매퍼와 컨텍스트 인식 매퍼를 모두 처리하며,
// 컨텍스트 인식 매퍼의 경우 취소 및 타임아웃이 적절히 처리됩니다.
//
// TODO(human): Map operation의 Apply 메소드를 구현해주세요
// 고려사항:
// 1. Type assertion 안전성
// 2. Context cancellation 체크
// 3. 에러 처리 및 전파
// 4. safeMapperWithContext vs mapper 분기 처리
func (op *MapOperation[T, U]) Apply(ctx context.Context, item any) (any, error) {
	// Check context cancellation first for immediate response
	select {
	case <-ctx.Done():
		var zero U
		return zero, ctx.Err()
	default:
	}

	// Type assertion to ensure item is of the expected type
	typedItem, ok := item.(T)
	if !ok {
		return nil, fmt.Errorf("type assertion failed: expected %T, got %T", *new(T), item)
	}

	// Apply the mapping operation
	if op.safeMapperWithContext != nil {
		// Use the context-aware mapper
		return op.safeMapperWithContext(ctx, typedItem)
	}

	// Use the simple mapper (context cancellation checked above)
	return op.mapper(typedItem), nil
}

// Map applies a transformation function to each element, converting from type T to type U.
// This creates a new Query with the transformed type and adds the mapping operation to the pipeline.
//
// This is a lazy operation - the mapper function is not executed until a terminal operation
// is called. The transformation happens as items flow through the pipeline.
//
// Example:
//
//	data := []int{1, 2, 3, 4, 5}
//	query := Map(From(data), func(x int) string { return fmt.Sprintf("num_%d", x) })
//	result := query.ToSlice(ctx) // ["num_1", "num_2", "num_3", "num_4", "num_5"]
//
// The mapper function should be pure (no side effects) and fast, as it will
// be called for each item during pipeline execution.
//
// If the input Query already has an error, Map returns a new Query[U] with the same error.
func Map[T, U any](q *Query[T], mapper func(T) U) *Query[U] {
	if q.err != nil {
		return &Query[U]{err: q.err, ctx: q.ctx}
	}

	// Create new MapOperation with the mapper function
	op := &MapOperation[T, U]{mapper: mapper}

	// For Map operations, we need to create a new Query[U] type but preserve the original pipeline
	// The type conversion happens during execution, not at compile time
	newQuery := &Query[U]{
		source:   nil, // Will be resolved during execution through type conversion
		pipeline: append(q.pipeline, op),
		ctx:      q.ctx,
		err:      q.err,
	}

	return newQuery
}

// MapE applies a context-aware transformation function with error handling to each element.
// This is the error-safe version of Map that supports context cancellation and
// mappers that can fail during execution.
//
// The mapper function receives the current context and can return an error.
// If the mapper returns an error, the entire pipeline execution will stop and
// return that error. Context cancellation is properly handled.
//
// This is a lazy operation - the mapper is not executed until a terminal operation
// is called. Any errors are captured and propagated through the pipeline.
//
// Example:
//
//	data := []string{"1", "2", "invalid", "4"}
//	query := MapE(From(data), func(ctx context.Context, s string) (int, error) {
//	    return strconv.Atoi(s) // Can fail on "invalid"
//	})
//	result := query.ToSlice(ctx) // Error when "invalid" is encountered
//
// If the input Query already has an error, MapE returns a new Query[U] with the same error.
func MapE[T, U any](q *Query[T], mapper func(context.Context, T) (U, error)) *Query[U] {
	if q.err != nil {
		return &Query[U]{err: q.err, ctx: q.ctx}
	}

	// Create new MapOperation with context-aware mapper
	op := &MapOperation[T, U]{safeMapperWithContext: mapper}

	// Return new Query[U] with the operation added to the pipeline
	return &Query[U]{
		source:   nil, // Will be resolved during execution
		pipeline: append(q.pipeline, op),
		ctx:      q.ctx,
		err:      q.err,
	}
}
