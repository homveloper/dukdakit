# 🚀 DukDakit Usage Guide

## Category-Based API Structure

DukDakit organizes features into categories accessible via dot notation:

```go
import "github.com/homveloper/dukdakit"

// Distributed computing features
controller := dukdakit.Distributed.NewOptimistic()

// Retry mechanisms
retrier := dukdakit.Retry.New()
```

## 🔄 Distributed Category

### Optimistic Concurrency Control

```go
package main

import (
    "context"
    "fmt"
    "github.com/homveloper/dukdakit"
    "github.com/homveloper/dukdakit/internal/distributed"
)

// Your entity must implement VersionedEntity interface
type GameEntity struct {
    ID      string
    Data    map[string]interface{}
    version int64
}

func (e *GameEntity) GetID() string { return e.ID }
func (e *GameEntity) GetVersion() int64 { return e.version }
func (e *GameEntity) SetVersion(v int64) { e.version = v }

func main() {
    // Create optimistic controller
    controller := dukdakit.Distributed.NewOptimistic()
    
    entity := &GameEntity{ID: "player-123", Data: make(map[string]interface{})}
    
    // Update with optimistic locking
    err := controller.UpdateWithOptimisticLock(
        context.Background(),
        entity,
        func(e distributed.VersionedEntity) error {
            // Your update logic here
            player := e.(*GameEntity)
            player.Data["score"] = 1000
            return nil
        },
        func(e distributed.VersionedEntity) error {
            // Optional validation
            return nil
        },
    )
    
    if err != nil {
        fmt.Printf("Update failed: %v\n", err)
    }
}
```

### Custom Configuration

```go
import "time"

config := distributed.OptimisticConfig{
    MaxRetries:    5,
    RetryDelay:    200 * time.Millisecond,
    VersionField:  "version",
    EnableMetrics: true,
}

controller := dukdakit.Distributed.NewOptimistic(config)
```

## 🔁 Retry Category

### Basic Retry

```go
package main

import (
    "context"
    "fmt"
    "github.com/homveloper/dukdakit"
)

func main() {
    // Create retrier with default config
    retrier := dukdakit.Retry.New()
    
    // Execute with retry
    err := retrier.Execute(context.Background(), func() error {
        // Your operation that might fail
        return callExternalAPI()
    })
    
    if err != nil {
        fmt.Printf("Operation failed: %v\n", err)
    }
}

func callExternalAPI() error {
    // Simulate API call
    return nil
}
```

### Advanced Retry Configuration

```go
import (
    "time"
    "github.com/homveloper/dukdakit/internal/retry"
)

config := retry.RetryConfig{
    MaxAttempts:     5,
    BaseDelay:       100 * time.Millisecond,
    MaxDelay:        10 * time.Second,
    Multiplier:      2.0,
    Jitter:          true,
    RetryableErrors: []string{"NetworkError", "TimeoutError"},
    CircuitBreaker: &retry.CircuitBreakerConfig{
        FailureThreshold: 3,
        ResetTimeout:     30 * time.Second,
        HalfOpenTimeout:  5 * time.Second,
    },
}

retrier := dukdakit.Retry.New(config)
```

### Retry with Results

```go
var result interface{}
err := retrier.ExecuteWithResult(context.Background(), func() (interface{}, error) {
    data, err := fetchImportantData()
    return data, err
})

if err == nil {
    // Use result
    fmt.Printf("Got result: %v\n", result)
}
```

## 📊 Metrics

Both categories provide metrics for monitoring:

```go
// Optimistic concurrency metrics
optimisticMetrics := controller.GetMetrics()
fmt.Printf("Success rate: %.2f%%\n", 
    float64(optimisticMetrics.SuccessfulUpdates) / 
    float64(optimisticMetrics.TotalOperations) * 100)

// Retry metrics
retryMetrics := retrier.GetMetrics()
fmt.Printf("Average attempts: %.1f\n", retryMetrics.AverageAttempts)
```

## 🔧 Error Handling

### Optimistic Concurrency Conflicts

```go
import "github.com/homveloper/dukdakit/internal/distributed"

err := controller.UpdateWithOptimisticLock(ctx, entity, updateFn, nil)
if err != nil {
    if conflictErr, ok := err.(*distributed.ConflictError); ok {
        fmt.Printf("Version conflict: expected %d, got %d\n", 
            conflictErr.ExpectedVersion, conflictErr.ActualVersion)
        // Handle conflict (reload entity, merge changes, etc.)
    }
}
```

### Circuit Breaker States

```go
circuitBreaker := dukdakit.Retry.NewCircuitBreaker(retry.CircuitBreakerConfig{
    FailureThreshold: 5,
    ResetTimeout:     60 * time.Second,
})

switch circuitBreaker.GetState() {
case retry.CircuitClosed:
    // Normal operation
case retry.CircuitOpen:
    // Circuit is open, requests are failing fast
case retry.CircuitHalfOpen:
    // Testing if service is recovered
}
```

## 🎯 Best Practices

### 1. Use Context for Cancellation

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := retrier.Execute(ctx, func() error {
    return longRunningOperation()
})
```

### 2. Configure Appropriately for Your Use Case

```go
// For critical operations
criticalRetrier := dukdakit.Retry.New(retry.RetryConfig{
    MaxAttempts: 10,
    BaseDelay:   500 * time.Millisecond,
    MaxDelay:    30 * time.Second,
    Multiplier:  1.5,
})

// For quick operations
quickRetrier := dukdakit.Retry.New(retry.RetryConfig{
    MaxAttempts: 3,
    BaseDelay:   50 * time.Millisecond,
    MaxDelay:    1 * time.Second,
    Multiplier:  2.0,
})
```

### 3. Monitor Metrics

```go
// Log metrics periodically
go func() {
    ticker := time.NewTicker(1 * time.Minute)
    for range ticker.C {
        metrics := retrier.GetMetrics()
        log.Printf("Retry metrics: %+v", metrics)
    }
}()
```

## 🧪 Testing

Run the examples to see DukDakit in action:

```bash
go run examples/basic_usage.go
```

This will demonstrate:
- Basic framework usage
- Optimistic concurrency control
- Retry mechanisms with exponential backoff
- Metrics collection

---

**🔨 뚝딱! Easy, Fast, and Production-Ready!** 🎮