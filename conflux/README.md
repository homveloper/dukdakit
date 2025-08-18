# Conflux

**낙관적 동시성 제어 + IoC 추상화**를 제공하는 Go 라이브러리입니다.

Conflux는 Race Condition을 방지하면서도 높은 성능을 유지할 수 있는 낙관적 동시성 제어(Optimistic Concurrency Control)를 다양한 데이터베이스에서 쉽게 사용할 수 있도록 추상화한 라이브러리입니다.

## 🚀 주요 특징

- **🔒 동시성 안전**: 버전 기반 낙관적 잠금으로 Race Condition 방지
- **🎯 IoC 패턴**: 비즈니스 로직만 작성하면 동시성 제어는 자동 처리
- **🔧 다중 인프라 지원**: Memory, Redis, MongoDB, PostgreSQL 등 다양한 데이터베이스
- **⚡ 타입 안전**: Go 제네릭을 활용한 컴파일 타임 타입 검증
- **🎨 직관적 API**: 연산별 전용 Result 타입과 함수형/인터페이스 양방향 지원

## 📦 설치

```bash
go get github.com/homveloper/dukdakit/conflux
```

## 🏗️ 아키텍처

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                        │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │   CreateFunc    │  │   UpdateFunc    │  │   UpsertFunc    │  │
│  │ (비즈니스 로직)    │  │ (비즈니스 로직)    │  │ (비즈니스 로직)    │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                                │
                                │ IoC Abstraction
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                     Conflux Core                           │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │ InsertResult[T] │  │ UpsertResult[T] │  │ UpdateResult[T] │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │   Repository    │  │  Filter System  │  │ Conflict Resolver │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                                │
                                │ Adapter Pattern
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                  Infrastructure Layer                       │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │ Memory Adapter  │  │ Redis Adapter   │  │MongoDB Adapter  │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## 🎯 핵심 개념

### 1. 연산별 전용 Result 타입

```go
// 각 연산의 특성에 맞는 결과 타입
type InsertResult[T any]  // 생성 전용 - IsDuplicate(), IsSuccess()
type UpsertResult[T any]  // 생성/수정 - WasCreated(), WasUpdated()  
type UpdateResult[T any]  // 수정 전용 - HasVersionConflict(), IsNotFound()
```

### 2. 함수형 + 인터페이스 지원

```go
// 함수형 스타일 (간단한 로직)
createFn := func(ctx context.Context) (*User, error) {
    return &User{ID: "123", Name: "John"}, nil
}
createFunc := conflux.NewCreateFunc(createFn)

// 인터페이스 스타일 (복잡한 로직)
type UserCreateLogic struct { /* fields */ }
func (u *UserCreateLogic) CreateFn(ctx context.Context) (*User, error) {
    // 복잡한 비즈니스 로직
}
```

### 3. 다중 인프라 호환 필터 시스템

```go
// MongoDB용 MapFilter
filter := conflux.NewMapFilter().And("email", "john@example.com")

// PostgreSQL용 SQLFilter  
filter := conflux.NewSQLFilter("email = ? AND status = ?", "john@example.com", "active")

// Redis용 RedisFilter
filter := conflux.NewRedisFilter("user:email:john@example.com")
```

## 🛠️ 기본 사용법

### 1. 엔터티 정의

```go
type User struct {
    conflux.BaseEntity  // 버전 및 타임스탬프 자동 관리
    ID       string `json:"id"`
    Email    string `json:"email"`
    Username string `json:"username"`
    Status   string `json:"status"`
    Credits  int    `json:"credits"`
}
```

### 2. Repository 생성

```go
// 메모리 기반 (개발/테스트용)
repo := memory.NewMemoryRepository[*User](func() *User {
    return &User{}
})

// Redis 기반 (캐싱/세션)
client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
config := &redis.RedisRepositoryConfig{KeyPrefix: "user:", TTL: time.Hour}
repo := redis.NewRedisRepository[*User](client, config, func() *User {
    return &User{}
})
```

### 3. CRUD 연산

#### 생성 (중복 방지)

```go
// 이메일 중복 검사
duplicateFilter := conflux.NewMapFilter().And("Email", "john@example.com")

createFunc := conflux.NewCreateFunc(func(ctx context.Context) (*User, error) {
    return &User{
        ID:       "user123",
        Email:    "john@example.com",
        Username: "john",
        Status:   "active",
        Credits:  100,
    }, nil
})

result, err := repo.FindOneAndInsert(ctx, duplicateFilter, createFunc)
if result.IsDuplicate() {
    fmt.Println("사용자가 이미 존재합니다")
} else {
    fmt.Println("사용자 생성 성공:", result.GetEntity().ID)
}
```

#### 생성/수정 (Upsert)

```go
lookupFilter := conflux.NewMapFilter().And("ID", "user123")

createFn := func(ctx context.Context) (*User, error) {
    return &User{ID: "user123", Credits: 50}, nil
}

updateFn := func(ctx context.Context, existing *User) (*User, error) {
    existing.Credits += 25  // 크레딧 추가
    return existing, nil
}

upsertFunc := conflux.NewUpsertFunc(createFn, updateFn)
result, err := repo.FindOneAndUpsert(ctx, lookupFilter, upsertFunc)

if result.WasCreated() {
    fmt.Println("신규 사용자 생성")
} else {
    fmt.Printf("기존 사용자 수정: %d 크레딧\n", result.GetEntity().Credits)
}
```

#### 수정 (버전 기반 충돌 감지)

```go
updateFunc := conflux.NewUpdateFunc(func(ctx context.Context, existing *User) (*User, error) {
    if existing.Credits < 100 {
        return existing, fmt.Errorf("크레딧 부족")
    }
    existing.Credits -= 100
    existing.Status = "premium"
    return existing, nil
})

result, err := repo.FindOneAndUpdate(ctx, lookupFilter, currentVersion, updateFunc)

switch {
case result.HasVersionConflict():
    fmt.Printf("버전 충돌: 기대값 %d, 실제값 %d\n", currentVersion, result.GetVersion())
case result.IsNotFound():
    fmt.Println("사용자를 찾을 수 없습니다")
case result.IsSuccess():
    fmt.Println("업데이트 성공:", result.GetEntity().Status)
}
```

## 🔌 어댑터

### Memory Adapter (내장)

```go
import "github.com/homveloper/dukdakit/conflux/adapters/memory"

repo := memory.NewMemoryRepository[*User](func() *User { return &User{} })
```

### Redis Adapter

```go
import redisadapter "github.com/homveloper/dukdakit/conflux/adapters/redis"

config := &redisadapter.RedisRepositoryConfig{
    KeyPrefix: "user:",
    TTL:       time.Hour * 24,
}
repo := redisadapter.NewRedisRepository[*User](redisClient, config, newUserFn)
```

### 사용자 정의 어댑터

```go
type MyDatabaseRepository[T any] struct {
    // 구현 필요
}

func (r *MyDatabaseRepository[T]) FindOneAndInsert(
    ctx context.Context,
    filter MyFilter, 
    createFunc conflux.CreateFunc[T],
) (*conflux.InsertResult[T], error) {
    // 사용자 정의 구현
}

// Repository 인터페이스의 모든 메서드 구현...
```

## 🎨 고급 기능

### 1. 조건부 로직

```go
type ConditionalCreateFunc[T any] struct {
    shouldCreate func(ctx context.Context) (bool, error)
    createFn     func(ctx context.Context) (T, error)
}

func (c *ConditionalCreateFunc[T]) CreateFn(ctx context.Context) (T, error) {
    should, err := c.shouldCreate(ctx)
    if err != nil || !should {
        var empty T
        return empty, err
    }
    return c.createFn(ctx)
}
```

### 2. 검증 기능

```go
validatedFactory := conflux.NewFactoryBuilder[*User]().
    WithCreate(createFn).
    WithValidation(func(ctx context.Context, user *User) error {
        if user.Email == "" {
            return fmt.Errorf("이메일은 필수입니다")
        }
        return nil
    }).
    BuildValidated()
```

### 3. 배치 연산

```go
// 여러 엔터티 조회
users, err := repo.FindMany(ctx, 
    conflux.NewMapFilter().And("Status", "active"), 
    100)

// 배치 삽입 (어댑터가 지원하는 경우)
results, err := batchRepo.InsertMany(ctx, duplicateFilters, createFuncs)
```

## ⚡ 성능 고려사항

### 동시성 처리

```go
// 동시 업데이트 시나리오
func incrementUserCredits(repo Repository, userID string, amount int) error {
    for retries := 0; retries < 3; retries++ {
        // 현재 사용자 조회
        user, err := repo.FindOne(ctx, filter)
        if err != nil {
            return err
        }
        
        // 업데이트 시도
        result, err := repo.FindOneAndUpdate(ctx, filter, user.Version, 
            conflux.NewUpdateFunc(func(ctx context.Context, existing *User) (*User, error) {
                existing.Credits += amount
                return existing, nil
            }))
            
        if err != nil {
            return err
        }
        
        if result.HasVersionConflict() {
            // 버전 충돌 시 재시도
            time.Sleep(time.Millisecond * time.Duration(retries * 10))
            continue
        }
        
        return nil  // 성공
    }
    return fmt.Errorf("최대 재시도 횟수 초과")
}
```

## 🧪 테스트

```bash
# 모든 테스트 실행
go test ./...

# 특정 패키지 테스트
go test ./adapters/memory
go test ./adapters/redis

# 벤치마크 테스트
go test -bench=. ./...
```

## 🤝 기여하기

1. 이슈 생성 또는 기존 이슈 확인
2. Feature Branch 생성
3. 코드 작성 및 테스트
4. Pull Request 생성

## 📄 라이선스

MIT License - 자세한 내용은 [LICENSE](LICENSE) 파일 참조

## 🔗 관련 프로젝트

- [DukDakit](../README.md) - 게임 서버 프레임워크
- [Friendit](../friendit/README.md) - 소셜 네트워킹 SDK

---

**Conflux**는 복잡한 동시성 제어를 간단하게 만들어 개발자가 비즈니스 로직에만 집중할 수 있도록 도와줍니다! 🚀