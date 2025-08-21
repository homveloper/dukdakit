# Pipit Implementation Workflow
## Go Context & Error Handling 중심 함수형 프로그래밍 라이브러리 구현 가이드

*작성일: 2025년 8월*  
*대상: Pipit 독립 패키지*

---

## 🎯 Workflow Overview

이 문서는 Go의 `context.Context`와 에러 핸들링을 중심으로 한 **Pipit** 독립 함수형 프로그래밍 라이브러리의 체계적인 구현 워크플로우를 제시합니다. Go의 관용적 패턴과 함수형 프로그래밍 원칙을 바탕으로 단계별 구현 전략을 제공합니다.

### Core Design Philosophy

```go
// Go-First Approach: 단순함과 명확성 우선
pipit.From(data).
    Filter(predicate).
    Map(transformer).
    WithContext(ctx).     // Go의 context 패턴 활용
    ToSliceE()            // 명시적 에러 반환 (Go 컨벤션)
```

---

## 📊 TODO Progress Tracker

### 🎯 Overall Progress: 16/30 (53%)
```
Phase 1: Foundation      [x] 6/6  (100%) ← COMPLETED!
Phase 2: Core Operations [x] 10/8 (125%) ← P2-T1,T2 COMPLETED!
Phase 3: Advanced Ops    [ ] 0/6  (0%)
Phase 4: Context & Mgmt  [ ] 0/4  (0%)
Phase 5: Performance     [ ] 0/6  (0%)
```

### 📋 Quick TODO Summary
- [x] **P1-T1**: Core Type System (16h) ← COMPLETED!
- [x] **P1-T2**: Error Handling Strategy (8h) ← COMPLETED!
- [x] **P1-T3**: Slice Iterator (12h) ← COMPLETED!
- [x] **P2-T1**: Filter Operation (10h) ← COMPLETED!
- [x] **P2-T2**: Map Operation (14h) ← COMPLETED! 🎯 *Learn by Doing* (+MapUnsafe/E)
- [ ] **P2-T3**: Operation Testing (8h)
- [ ] **P3-T1**: Reduce & Aggregation (14h)
- [ ] **P3-T2**: Collection Operations (10h)
- [ ] **P4-T1**: Context Integration (12h)
- [ ] **P4-T2**: Resource Cleanup (6h)
- [ ] **P5-T1**: Performance Optimization (16h)
- [ ] **P5-T2**: Public API Finalization (8h)

---

## 📋 Phase 1: Foundation Architecture (Week 1-2)
**Status**: ✅ COMPLETED | **Progress**: 6/6 TODOs | **Estimated**: 36 hours

### 🏗️ P1-T1: Core Type System Design

**Priority**: 🔴 Critical | **ID**: P1-T1 | **Time**: 16h | **Status**: ✅ COMPLETED  
**Dependencies**: Go 1.21+ generics | **Blocks**: P1-T2, P2-T1, P2-T2

#### TODO Checklist:
- [x] **P1-T1.1**: Define Query[T] struct with generics (4h)
- [x] **P1-T1.2**: Create Iterator[T] interface (3h)  
- [x] **P1-T1.3**: Design Operation interface (3h)
- [x] **P1-T1.4**: Define OperationType constants (2h)
- [x] **P1-T1.5**: Add context and error fields (2h)
- [x] **P1-T1.6**: Write unit tests for type system (2h)

#### 1.1.1 Basic Query Interface
```go
// query.go
package pipit

import (
    "context"
    "fmt"
)

// Query represents a lazy-evaluated data pipeline
type Query[T any] struct {
    source   Iterator[T]
    pipeline []Operation
    ctx      context.Context  // Context 통합
    err      error           // 에러 상태 추적
}

// Iterator interface for lazy evaluation
type Iterator[T any] interface {
    Next(ctx context.Context) (T, bool, error)
    HasNext() bool
    Close() error  // 리소스 해제
}

// Operation represents a pipeline operation
type Operation interface {
    Apply(ctx context.Context, item any) (any, error)
    Type() OperationType
}

type OperationType int

const (
    FilterOp OperationType = iota
    MapOp
    TakeOp
    DropOp
    DistinctOp
)
```

**Acceptance Criteria**:
- [x] **AC1**: Generic type safety 보장
- [x] **AC2**: Context propagation 지원  
- [x] **AC3**: Memory-efficient lazy evaluation
- [x] **AC4**: 명시적 에러 처리

---

### 🏗️ P1-T2: Error Handling Strategy

**Priority**: 🔴 Critical | **ID**: P1-T2 | **Time**: 8h | **Status**: ✅ COMPLETED  
**Dependencies**: P1-T1 | **Blocks**: P2-T1, P2-T2

#### TODO Checklist:
- [x] **P1-T2.1**: Define PipitError struct (2h)
- [x] **P1-T2.2**: Implement Error() and Unwrap() methods (2h)
- [x] **P1-T2.3**: Create withError helper function (2h)
- [x] **P1-T2.4**: Write error handling tests (2h)

#### 1.1.2 Error Handling Strategy
```go
// Error wrapping with context
type PipitError struct {
    Op      string    // 실패한 연산
    Stage   int       // 파이프라인 단계
    Cause   error     // 원본 에러
    Context context.Context
}

func (e *PipitError) Error() string {
    return fmt.Sprintf("pipit: %s at stage %d: %v", e.Op, e.Stage, e.Cause)
}

func (e *PipitError) Unwrap() error {
    return e.Cause
}

// 에러 처리 함수
func (q *Query[T]) withError(op string, err error) *Query[T] {
    if err != nil && q.err == nil {
        q.err = &PipitError{
            Op:      op,
            Stage:   len(q.pipeline),
            Cause:   err,
            Context: q.ctx,
        }
    }
    return q
}
```

---

### 🧩 P1-T3: Slice Iterator Implementation

**Priority**: 🔴 Critical | **ID**: P1-T3 | **Time**: 12h | **Status**: ⏳ Not Started  
**Dependencies**: P1-T1 | **Blocks**: P2-T1, P2-T2, P3-T1

#### TODO Checklist:
- [x] **P1-T3.1**: Define SliceIterator[T] struct (2h)
- [x] **P1-T3.2**: Implement NewSliceIterator constructor (2h)
- [x] **P1-T3.3**: Implement Next() with context support (3h)
- [x] **P1-T3.4**: Implement HasNext() method (1h)
- [x] **P1-T3.5**: Add thread-safety with mutex (2h)
- [x] **P1-T3.6**: Write comprehensive tests (2h)

```go
// slice_iterator.go
type SliceIterator[T any] struct {
    data  []T
    index int
    mutex sync.RWMutex  // Thread-safe access
}

func NewSliceIterator[T any](data []T) *SliceIterator[T] {
    return &SliceIterator[T]{
        data:  data,
        index: 0,
    }
}

func (si *SliceIterator[T]) Next(ctx context.Context) (T, bool, error) {
    select {
    case <-ctx.Done():
        var zero T
        return zero, false, ctx.Err()
    default:
    }
    
    si.mutex.RLock()
    defer si.mutex.RUnlock()
    
    if si.index >= len(si.data) {
        var zero T
        return zero, false, nil
    }
    
    item := si.data[si.index]
    si.index++
    return item, true, nil
}
```

**Test Requirements**:
```go
func TestSliceIterator_ContextCancellation(t *testing.T) {
    data := []int{1, 2, 3, 4, 5}
    iter := NewSliceIterator(data)
    
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
    defer cancel()
    
    time.Sleep(2 * time.Millisecond) // Context timeout
    
    _, _, err := iter.Next(ctx)
    assert.ErrorIs(t, err, context.DeadlineExceeded)
}
```

---

## 📋 Phase 2: Core Operations (Week 3-4)
**Status**: 🔄 In Progress | **Progress**: 10/9 TODOs | **Estimated**: 32 hours

### 🔄 P2-T1: Filter Operation

**Priority**: 🔴 Critical | **ID**: P2-T1 | **Time**: 10h | **Status**: ✅ COMPLETED  
**Dependencies**: P1-T1, P1-T2, P1-T3 | **Blocks**: P2-T3

#### TODO Checklist:
- [x] **P2-T1.1**: Define FilterOperation[T] struct (2h)
- [x] **P2-T1.2**: Implement Filter() method (2h)
- [x] **P2-T1.3**: Implement FilterE() with error handling (3h)
- [x] **P2-T1.4**: Implement Apply() method (2h)
- [x] **P2-T1.5**: Write unit tests (1h)

```go
// filter.go
type FilterOperation[T any] struct {
    predicate func(T) bool
    safePredicateWithContext func(context.Context, T) (bool, error)
}

func (q *Query[T]) Filter(predicate func(T) bool) *Query[T] {
    if q.err != nil {
        return q
    }
    
    op := &FilterOperation[T]{predicate: predicate}
    q.pipeline = append(q.pipeline, op)
    return q
}

// Context-aware filter with error handling
func (q *Query[T]) FilterE(predicate func(context.Context, T) (bool, error)) *Query[T] {
    if q.err != nil {
        return q
    }
    
    op := &FilterOperation[T]{safePredicateWithContext: predicate}
    q.pipeline = append(q.pipeline, op)
    return q
}

func (op *FilterOperation[T]) Apply(ctx context.Context, item any) (any, error) {
    typedItem, ok := item.(T)
    if !ok {
        return nil, fmt.Errorf("type assertion failed: expected %T, got %T", 
            *new(T), item)
    }
    
    if op.safePredicateWithContext != nil {
        match, err := op.safePredicateWithContext(ctx, typedItem)
        if err != nil {
            return nil, err
        }
        if match {
            return typedItem, nil
        }
        return nil, nil // Filtered out
    }
    
    if op.predicate(typedItem) {
        return typedItem, nil
    }
    return nil, nil // Filtered out
}
```

---

### 🗺️ P2-T2: Map Operation

**Priority**: 🔴 Critical | **ID**: P2-T2 | **Time**: 14h | **Status**: ✅ COMPLETED  
**Dependencies**: P1-T1, P1-T2, P1-T3 | **Blocks**: P2-T3

#### TODO Checklist:
- [x] **P2-T2.1**: Define MapOperation[T,U] struct (3h)
- [x] **P2-T2.2**: Implement Map[T,U]() function (3h)
- [x] **P2-T2.3**: Implement MapE[T,U]() with error handling (3h)
- [x] **P2-T2.4**: Implement Apply() method 🎯 **(Learn by Doing)** (2h)
- [x] **P2-T2.5**: Write comprehensive tests (1h)
- [x] **P2-T2.6**: Implement MapUnsafe() & MapUnsafeE() - 런타임 타입 변환 지원 (2h)

```go
// map.go
type MapOperation[T, U any] struct {
    mapper func(T) U
    safeMapperWithContext func(context.Context, T) (U, error)
}

func Map[T, U any](q *Query[T], mapper func(T) U) *Query[U] {
    if q.err != nil {
        return &Query[U]{err: q.err, ctx: q.ctx}
    }
    
    op := &MapOperation[T, U]{mapper: mapper}
    
    return &Query[U]{
        source:   q.source,
        pipeline: append(q.pipeline, op),
        ctx:      q.ctx,
        err:      q.err,
    }
}

// Context-aware mapping with error handling
func MapE[T, U any](q *Query[T], mapper func(context.Context, T) (U, error)) *Query[U] {
    if q.err != nil {
        return &Query[U]{err: q.err, ctx: q.ctx}
    }
    
    op := &MapOperation[T, U]{safeMapperWithContext: mapper}
    
    return &Query[U]{
        source:   q.source,
        pipeline: append(q.pipeline, op),
        ctx:      q.ctx,
        err:      q.err,
    }
}

// MapUnsafe provides runtime type conversion without compile-time safety
// This is an experimental feature for dynamic scenarios where type safety can be relaxed
func (q *Query[T]) MapUnsafe(mapper func(any) any) *Query[any] {
    if q.err != nil {
        return &Query[any]{err: q.err, ctx: q.ctx}
    }
    
    // Create unsafe operation that bypasses type checking
    op := &UnsafeMapOperation{
        unsafeMapper: mapper,
    }
    
    return &Query[any]{
        source:   nil, // Will be resolved during execution
        pipeline: append(q.pipeline, op),
        ctx:      q.ctx,
        err:      q.err,
    }
}

// UnsafeMapOperation handles runtime type conversion
type UnsafeMapOperation struct {
    unsafeMapper func(any) any
}

func (op *UnsafeMapOperation) Type() OperationType {
    return MapOp // Same as regular map
}

func (op *UnsafeMapOperation) Apply(ctx context.Context, item any) (any, error) {
    // Check context cancellation first
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    // Apply unsafe mapping - no type checking
    result := op.unsafeMapper(item)
    return result, nil
}
```

**⚠️ MapUnsafe 사용 시 주의사항:**
- 타입 안전성 완전 상실
- 런타임 패닉 가능성
- 성능상 이점 없음 (오히려 boxing/unboxing 오버헤드)
- 디버깅 어려움
- **권장 사용 케이스**: 프로토타이핑, 마이그레이션, 동적 데이터 처리


---

### 🧪 P2-T3: Operation Testing & Integration

**Priority**: 🟡 High | **ID**: P2-T3 | **Time**: 8h | **Status**: ⏳ Not Started  
**Dependencies**: P2-T1, P2-T2 | **Blocks**: P3-T1

#### TODO Checklist:
- [x] **P2-T3.1**: Integration tests for Filter+Map chains (3h)
- [ ] **P2-T3.2**: Performance benchmarks vs native loops (2h)
- [ ] **P2-T3.3**: Error propagation tests (2h)
- [ ] **P2-T3.4**: Context cancellation tests (1h)


---

## 📋 Phase 3: Advanced Operations (Week 5-6)
**Status**: ⏳ Not Started | **Progress**: 0/6 TODOs | **Estimated**: 24 hours

### 🔢 P3-T1: Aggregation Operations

**Priority**: 🟡 High | **ID**: P3-T1 | **Time**: 14h | **Status**: ⏳ Not Started  
**Dependencies**: P2-T1, P2-T2, P2-T3 | **Blocks**: P3-T2

#### TODO Checklist:
- [ ] **P3-T1.1**: Implement Reduce() with context (4h)
- [ ] **P3-T1.2**: Implement Count() operation (2h)
- [ ] **P3-T1.3**: Implement First() and Last() (3h)
- [ ] **P3-T1.4**: Implement Any() and All() predicates (3h)
- [ ] **P3-T1.5**: Write aggregation tests (2h)

```go
// reduce.go
// Reduce operation with context and error handling
func (q *Query[T]) Reduce(ctx context.Context, 
    initial T, 
    reducer func(context.Context, T, T) (T, error)) (T, error) {
    
    if q.err != nil {
        return initial, q.err
    }
    
    accumulator := initial
    iter := q.execute(ctx)
    defer iter.Close()
    
    for {
        select {
        case <-ctx.Done():
            return accumulator, ctx.Err()
        default:
        }
        
        item, hasNext, err := iter.Next(ctx)
        if err != nil {
            return accumulator, err
        }
        if !hasNext {
            break
        }
        
        accumulator, err = reducer(ctx, accumulator, item)
        if err != nil {
            return accumulator, fmt.Errorf("reduce operation failed: %w", err)
        }
    }
    
    return accumulator, nil
}

// Count with context support
func (q *Query[T]) Count(ctx context.Context) (int, error) {
    return q.Reduce(ctx, 0, func(ctx context.Context, acc int, _ T) (int, error) {
        return acc + 1, nil
    })
}
```

---

### 🎯 P3-T2: Collection Operations

**Priority**: 🟡 High | **ID**: P3-T2 | **Time**: 10h | **Status**: ⏳ Not Started  
**Dependencies**: P3-T1 | **Blocks**: P4-T1

#### TODO Checklist:
- [ ] **P3-T2.1**: Implement ToSliceE() terminal operation (3h)
- [ ] **P3-T2.2**: Implement ToSlice() convenience method (2h)
- [ ] **P3-T2.3**: Implement Take() and Skip() (3h)
- [ ] **P3-T2.4**: Write collection operation tests (2h)

```go
// Terminal operations with explicit error handling
func (q *Query[T]) ToSliceE() ([]T, error) {
    if q.err != nil {
        return nil, q.err
    }
    
    ctx := q.ctx
    if ctx == nil {
        ctx = context.Background()
    }
    
    var result []T
    iter := q.execute(ctx)
    defer iter.Close()
    
    for {
        item, hasNext, err := iter.Next(ctx)
        if err != nil {
            return nil, err
        }
        if !hasNext {
            break
        }
        result = append(result, item)
    }
    
    return result, nil
}

// Panic-free version (Go idiomatic)
func (q *Query[T]) ToSlice() []T {
    result, err := q.ToSliceE()
    if err != nil {
        // Log error but don't panic
        // In production: use proper logging
        fmt.Printf("pipit: ToSlice() failed: %v\n", err)
        return nil
    }
    return result
}
```

---

## 📋 Phase 4: Context & Resource Management (Week 7)
**Status**: ⏳ Not Started | **Progress**: 0/4 TODOs | **Estimated**: 18 hours

### 🛡️ P4-T1: Context Integration Patterns

**Priority**: 🔴 Critical | **ID**: P4-T1 | **Time**: 12h | **Status**: ⏳ Not Started  
**Dependencies**: P3-T1, P3-T2 | **Blocks**: P4-T2, P5-T1

#### TODO Checklist:
- [ ] **P4-T1.1**: Implement WithContext() method (3h)
- [ ] **P4-T1.2**: Implement WithTimeout() method (3h)
- [ ] **P4-T1.3**: Implement WithDeadline() method (2h)
- [ ] **P4-T1.4**: Add context propagation tests (2h)
- [ ] **P4-T1.5**: Context cancellation integration tests (2h)

```go
// Context methods for pipeline control
func (q *Query[T]) WithContext(ctx context.Context) *Query[T] {
    newQuery := *q  // Shallow copy
    newQuery.ctx = ctx
    return &newQuery
}

func (q *Query[T]) WithTimeout(timeout time.Duration) *Query[T] {
    ctx := q.ctx
    if ctx == nil {
        ctx = context.Background()
    }
    
    timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
    newQuery := q.WithContext(timeoutCtx)
    
    // Store cancel function for cleanup
    // This requires additional state management
    return newQuery
}

func (q *Query[T]) WithDeadline(deadline time.Time) *Query[T] {
    ctx := q.ctx
    if ctx == nil {
        ctx = context.Background()
    }
    
    deadlineCtx, cancel := context.WithDeadline(ctx, deadline)
    return q.WithContext(deadlineCtx)
}
```

---

### 🧹 P4-T2: Resource Cleanup Patterns

**Priority**: 🟡 High | **ID**: P4-T2 | **Time**: 6h | **Status**: ⏳ Not Started  
**Dependencies**: P4-T1 | **Blocks**: P5-T1

#### TODO Checklist:
- [ ] **P4-T2.1**: Define CleanupQuery struct (2h)
- [ ] **P4-T2.2**: Implement WithCleanup() method (2h)
- [ ] **P4-T2.3**: Implement Execute() with defer cleanup (2h)

```go
// Resource cleanup with defer patterns
type CleanupQuery[T any] struct {
    *Query[T]
    cleanupFuncs []func() error
}

func (q *Query[T]) WithCleanup(cleanup func() error) *CleanupQuery[T] {
    cq := &CleanupQuery[T]{
        Query:        q,
        cleanupFuncs: []func() error{cleanup},
    }
    return cq
}

func (cq *CleanupQuery[T]) Execute(ctx context.Context) error {
    defer func() {
        for _, cleanup := range cq.cleanupFuncs {
            if err := cleanup(); err != nil {
                // Log cleanup errors
                fmt.Printf("cleanup error: %v\n", err)
            }
        }
    }()
    
    _, err := cq.ToSliceE()
    return err
}
```

---

## 📋 Phase 5: Performance & Package Finalization (Week 8)
**Status**: ⏳ Not Started | **Progress**: 0/6 TODOs | **Estimated**: 24 hours

### ⚡ P5-T1: Performance Optimization

**Priority**: 🟡 High | **ID**: P5-T1 | **Time**: 16h | **Status**: ⏳ Not Started  
**Dependencies**: P4-T1, P4-T2 | **Blocks**: P5-T2

#### TODO Checklist:
- [ ] **P5-T1.1**: Implement memory pooling (6h)
- [ ] **P5-T1.2**: Add ToSlicePooled() method (3h)
- [ ] **P5-T1.3**: Create benchmarking utilities (3h)
- [ ] **P5-T1.4**: Performance profiling and optimization (4h)

```go
// Memory pooling for large datasets
var slicePool = sync.Pool{
    New: func() interface{} {
        return make([]interface{}, 0, 1000) // Pre-allocated capacity
    },
}

func (q *Query[T]) ToSlicePooled() ([]T, error) {
    if q.err != nil {
        return nil, q.err
    }
    
    // Get from pool and ensure proper type
    pooled := slicePool.Get().([]interface{})
    defer slicePool.Put(pooled[:0]) // Reset and return to pool
    
    // Implementation with pooled slice...
}

// Benchmarking utilities
func BenchmarkPipeline[T any](name string, q *Query[T]) {
    start := time.Now()
    _, err := q.ToSliceE()
    duration := time.Since(start)
    
    if err != nil {
        fmt.Printf("Benchmark %s: ERROR - %v\n", name, err)
    } else {
        fmt.Printf("Benchmark %s: %v\n", name, duration)
    }
}
```

---

### 🔗 P5-T2: Public API Design & Documentation

**Priority**: 🟡 High | **ID**: P5-T2 | **Time**: 8h | **Status**: ⏳ Not Started  
**Dependencies**: P5-T1 | **Blocks**: None

#### TODO Checklist:
- [ ] **P5-T2.1**: Finalize public API (2h)
- [ ] **P5-T2.2**: Write comprehensive documentation (3h)
- [ ] **P5-T2.3**: Create usage examples (2h)  
- [ ] **P5-T2.4**: Package README and godoc (1h)

```go
// pipit.go - Public API for independent package
package pipit

// From creates a new pipeline from a slice
func From[T any](source []T) *Query[T] {
    return &Query[T]{
        source: NewSliceIterator(source),
        ctx:    context.Background(),
    }
}

// Range creates a pipeline from numeric range
func Range(start, count int) *Query[int] {
    data := make([]int, count)
    for i := 0; i < count; i++ {
        data[i] = start + i
    }
    return From(data)
}

// Repeat creates a pipeline that repeats a value
func Repeat[T any](value T, count int) *Query[T] {
    data := make([]T, count)
    for i := range data {
        data[i] = value
    }
    return From(data)
}

// Example usage:
// result := pipit.From(players).
//     Filter(func(p Player) bool { return p.IsActive() }).
//     Map(func(p Player) PlayerSummary { return p.Summary() }).
//     Take(10).
//     ToSlice()
```

---

## 🧪 Testing Strategy

### Unit Testing Framework
```go
// pipit_test.go
func TestPipit_BasicPipeline(t *testing.T) {
    data := []int{1, 2, 3, 4, 5}
    
    result, err := pipit.From(data).
        Filter(func(x int) bool { return x > 2 }).
        Map(func(x int) int { return x * 2 }).
        ToSliceE()
    
    require.NoError(t, err)
    assert.Equal(t, []int{6, 8, 10}, result)
}

func TestPipit_ContextCancellation(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    
    // Large dataset to ensure cancellation occurs
    data := make([]int, 10000)
    for i := range data {
        data[i] = i
    }
    
    go func() {
        time.Sleep(10 * time.Millisecond)
        cancel()
    }()
    
    _, err := pipit.From(data).
        WithContext(ctx).
        Map(func(x int) int {
            time.Sleep(1 * time.Millisecond) // Slow operation
            return x * 2
        }).
        ToSliceE()
    
    assert.ErrorIs(t, err, context.Canceled)
}
```

### Performance Benchmarks
```go
func BenchmarkPipit_vs_NativeLoop(b *testing.B) {
    data := make([]int, 1000)
    for i := range data {
        data[i] = i
    }
    
    b.Run("Pipit", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            pipit.From(data).
                Filter(func(x int) bool { return x%2 == 0 }).
                Map(func(x int) int { return x * 2 }).
                ToSlice()
        }
    })
    
    b.Run("Native", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            var result []int
            for _, x := range data {
                if x%2 == 0 {
                    result = append(result, x*2)
                }
            }
        }
    })
}
```

---

## 📈 Implementation Metrics & Success Criteria

### Performance Targets
- **Memory Overhead**: <15% compared to native loops
- **Execution Time**: <25% slower than native loops
- **Context Cancellation**: <100ms response time
- **Error Recovery**: Graceful handling of all error scenarios

### API Completeness Checklist
- [ ] Core operations (Filter, Map, Reduce)
- [ ] Collection operations (ToSlice, ToMap, Count)
- [ ] Context integration (WithContext, WithTimeout)
- [ ] Error handling (explicit error returns)
- [ ] Resource cleanup (defer patterns)
- [ ] Type safety (generics validation)
- [ ] Thread safety (concurrent access)
- [ ] Memory efficiency (pooling, lazy evaluation)

### Documentation Requirements
- [ ] API reference with examples
- [ ] Performance characteristics guide
- [ ] Error handling best practices
- [ ] Context usage patterns
- [ ] Independent package documentation and examples

---

## 🚀 Deployment & Rollout Strategy

### Phase Deployment
1. **Alpha Release**: Core operations only (Filter, Map, ToSlice)
2. **Beta Release**: Full operation set with context support
3. **Production Release**: Performance optimized with comprehensive testing

### Monitoring & Observability
```go
// Metrics collection for production
type PipitMetrics struct {
    PipelineExecutions int64
    AverageLatency     time.Duration
    ErrorRate          float64
    ContextCancellations int64
}

func (q *Query[T]) WithMetrics(metrics *PipitMetrics) *Query[T] {
    // Add metrics collection to pipeline
}
```

---

---

## 📚 BACKLOG: Future Enhancements

### 🔄 B-1: QueryBuilder Pattern for Method Chaining

**Priority**: 🟡 Medium | **Status**: 🔍 Research Required | **Complexity**: High

#### Problem Analysis:
Go 제네릭의 핵심 제약사항으로 인해 메소드 체이닝이 불가능:
```go
// 불가능한 코드 - 컴파일 에러
func (qb QueryBuilder[T]) Map[U any](mapper func(T) U) QueryBuilder[U] {
    //                      ^^^^^ Error: "method must have no type parameters"
}
```

#### Investigation Results:
- **Root Cause**: Go는 메소드에서 추가 타입 파라미터를 허용하지 않음
- **Language Limitation**: 의도적인 설계 제약 (타입 추론 복잡성 방지)
- **Alternative Approaches Evaluated**:
  - ❌ QueryBuilder pattern with generic methods (컴파일 실패)
  - ❌ Interface{} based approach (타입 안전성 상실)
  - ❌ Code generation (복잡도 증가)
  - ✅ Current functional approach (권장)

#### Future Research Areas:
1. **Go 2.0+ Language Evolution**: 향후 Go 버전에서의 제네릭 확장 가능성
2. **Code Generation Tools**: 빌드 타임 코드 생성을 통한 체이닝 지원
3. **DSL Approach**: Domain Specific Language 기반 쿼리 빌더
4. **Reflection-Based Solutions**: 런타임 타입 처리 (성능 trade-off)

#### Technical Debt Notes:
```go
// 현재 구현 (권장 유지)
result := Map(
    query.Filter(predicate),
    mapper
)

// 이상적이지만 불가능한 체이닝
result := query.
    Filter(predicate).
    Map(mapper).       // <- 이 부분이 Go에서 불가능
    ToSlice(ctx)
```

#### Decision:
- **Current Status**: 현재 함수형 접근법 유지 권장
- **Monitoring**: Go 언어 evolution 및 community solutions 추적
- **Action Required**: 없음 (언어 제약사항)

---

### 🎯 B-2: Advanced Pipeline Optimization

**Priority**: 🟢 Low | **Status**: 💡 Ideas | **Complexity**: Medium

#### Potential Enhancements:
1. **Pipeline Fusion**: 연속된 Map/Filter 연산 최적화
2. **Lazy Evaluation Improvements**: 더 정교한 지연 평가
3. **Parallel Processing**: Goroutine 기반 병렬 처리
4. **Memory Pool Integration**: 고성능 메모리 관리

#### Research Topics:
- Rust Iterator의 zero-cost abstractions 벤치마킹
- Java Stream API의 spliterator 패턴 분석
- .NET LINQ의 expression tree 컴파일 최적화

---

**이 워크플로우는 Go의 idiomatic patterns를 따르면서도 함수형 프로그래밍의 장점을 최대한 활용하도록 설계되었습니다. Context와 explicit error handling을 통해 production-ready한 라이브러리를 구축할 수 있습니다.**