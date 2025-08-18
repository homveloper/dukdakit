# Conflux Redis Adapter

Redis 기반 낙관적 동시성 제어 레포지토리 구현체입니다.

## 특징

- **Redis 최적화**: Redis의 Hash, Pipeline, Transaction 기능 활용
- **동시성 안전**: Redis WATCH를 이용한 낙관적 잠금
- **유연한 필터링**: RedisFilter를 통한 패턴 기반 키 검색
- **TTL 지원**: 자동 데이터 만료 기능
- **배치 연산**: Redis Pipeline을 활용한 효율적인 배치 처리

## 설치

```bash
go get github.com/homveloper/dukdakit/conflux/adapters/redis
```

## 기본 사용법

### 1. Repository 생성

```go
import (
    "github.com/go-redis/redis/v8"
    "github.com/homveloper/dukdakit/conflux"
    confluxredis "github.com/homveloper/dukdakit/conflux/adapters/redis"
)

// Redis 클라이언트 생성
client := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
})

// Repository 설정
config := &confluxredis.RedisRepositoryConfig{
    KeyPrefix: "user:",           // Redis 키 접두사
    TTL:       time.Hour * 24,    // 24시간 TTL (0이면 영구 보존)
}

// Repository 생성
repo := confluxredis.NewRedisRepository[*User](client, config, func() *User {
    return &User{}
})
```

### 2. 엔터티 정의

```go
type User struct {
    conflux.BaseEntity // 버전 및 타임스탬프 자동 관리
    ID       string `json:"id"`
    Email    string `json:"email"`
    Username string `json:"username"`
    Status   string `json:"status"`
    Credits  int    `json:"credits"`
}
```

### 3. CRUD 연산

#### 생성 (FindOneAndInsert)

```go
// 중복 검사용 필터 (이메일 기반)
duplicateFilter := conflux.NewRedisFilterWithPrefix("user:email:" + user.Email)

// 생성 로직
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
if err != nil {
    return err
}

if result.IsDuplicate() {
    fmt.Println("User already exists:", result.GetEntity().Email)
} else {
    fmt.Println("User created:", result.GetEntity().ID)
}
```

#### 생성/업데이트 (FindOneAndUpsert)

```go
// 조회용 필터 (ID 기반)
lookupFilter := conflux.NewRedisFilter("user:" + userID)

// Upsert 로직
createFn := func(ctx context.Context) (*User, error) {
    return &User{
        ID:      userID,
        Email:   "jane@example.com",
        Status:  "active",
        Credits: 50,
    }, nil
}

updateFn := func(ctx context.Context, existing *User) (*User, error) {
    existing.Credits += 25  // 크레딧 추가
    return existing, nil
}

upsertFunc := conflux.NewUpsertFunc(createFn, updateFn)
result, err := repo.FindOneAndUpsert(ctx, lookupFilter, upsertFunc)

if result.WasCreated() {
    fmt.Println("User created")
} else {
    fmt.Println("User updated")
}
```

#### 업데이트 (FindOneAndUpdate)

```go
// 현재 버전으로 업데이트 (낙관적 잠금)
updateFunc := conflux.NewUpdateFunc(func(ctx context.Context, existing *User) (*User, error) {
    existing.Status = "premium"
    existing.Credits *= 2
    return existing, nil
})

result, err := repo.FindOneAndUpdate(ctx, lookupFilter, currentVersion, updateFunc)

if result.HasVersionConflict() {
    fmt.Printf("Version conflict: expected %d, got %d\n", 
               currentVersion, result.GetVersion())
} else if result.IsSuccess() {
    fmt.Println("User updated successfully")
}
```

### 4. Redis 특화 필터링

#### 패턴 기반 검색

```go
// 모든 admin 사용자 검색
adminFilter := conflux.NewRedisFilter("user:admin*")
adminUsers, err := repo.FindMany(ctx, adminFilter, 100)

// 특정 접두사를 가진 키들 검색
prefixFilter := conflux.NewRedisFilterWithPrefix("user:vip:")
vipUsers, err := repo.FindMany(ctx, prefixFilter, 50)
```

#### 조합 필터 (향후 확장)

```go
// 복합 조건 검색 예제
compositeFilter := conflux.NewFilterBuilder().And().
    Add(conflux.NewRedisFilter("user:*")).
    Add(conflux.NewTimeRangeFilter("created_at").After(time.Now().Add(-24 * time.Hour)))

recentUsers, err := repo.FindMany(ctx, compositeFilter, 100)
```

## Redis 데이터 구조

각 엔터티는 Redis Hash로 저장됩니다:

```
user:123 {
    "data":       "{\"id\":\"123\",\"email\":\"john@example.com\",...}",
    "version":    "1",
    "created_at": "1640995200",
    "updated_at": "1640995200"
}
```

## 동시성 제어

Redis WATCH 명령을 사용하여 낙관적 잠금을 구현:

1. **트랜잭션 시작**: WATCH로 키 감시
2. **버전 확인**: 현재 버전과 기대 버전 비교
3. **원자적 업데이트**: MULTI/EXEC로 트랜잭션 실행
4. **충돌 감지**: 트랜잭션 실패 시 버전 충돌로 처리

## 성능 최적화

- **Pipeline 사용**: 배치 연산에서 네트워크 라운드트립 최소화
- **Hash 구조**: 메타데이터와 데이터를 효율적으로 분리 저장
- **TTL 활용**: 자동 데이터 정리로 메모리 사용량 최적화

## 한계사항

1. **복잡한 쿼리**: SQL과 달리 복잡한 조인이나 집계 연산 제한
2. **메모리 사용량**: 모든 데이터가 메모리에 저장됨
3. **트랜잭션 범위**: Redis 트랜잭션은 단일 인스턴스로 제한

## 실제 운영 고려사항

```go
// 프로덕션 환경 설정 예제
config := &confluxredis.RedisRepositoryConfig{
    KeyPrefix: "myapp:user:",
    TTL:       time.Hour * 24 * 7, // 7일 TTL
}

// Redis 클러스터 지원
client := redis.NewClusterClient(&redis.ClusterOptions{
    Addrs: []string{"localhost:7000", "localhost:7001", "localhost:7002"},
})

// 연결 풀 및 타임아웃 설정
client = redis.NewClient(&redis.Options{
    Addr:         "localhost:6379",
    PoolSize:     20,
    MinIdleConns: 5,
    ReadTimeout:  time.Second * 3,
    WriteTimeout: time.Second * 3,
})
```

Redis 어댑터는 고성능이 필요한 캐싱, 세션 관리, 실시간 데이터 처리에 적합합니다!