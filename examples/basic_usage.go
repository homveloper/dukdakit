package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/homveloper/dukdakit"
	"github.com/homveloper/dukdakit/internal/distributed"
	"github.com/homveloper/dukdakit/internal/retry"
)

// ExampleEntity demonstrates optimistic concurrency control
type ExampleEntity struct {
	ID      string
	Name    string
	Score   int
	version int64
}

func (e *ExampleEntity) GetID() string {
	return e.ID
}

func (e *ExampleEntity) GetVersion() int64 {
	return e.version
}

func (e *ExampleEntity) SetVersion(version int64) {
	e.version = version
}

func main() {
	fmt.Println("🔨 DukDakit Basic Usage Examples")
	fmt.Println()

	// Example 1: Basic Framework Usage
	basicFrameworkExample()

	// Example 2: Optimistic Concurrency Control
	optimisticConcurrencyExample()

	// Example 3: Retry Mechanisms
	retryExample()
}

func basicFrameworkExample() {
	fmt.Println("=== Basic Framework Usage ===")
	
	// Create DukDakit instance
	server := dukdakit.New()
	
	// Start the server (placeholder implementation)
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
	
	fmt.Println()
}

func optimisticConcurrencyExample() {
	fmt.Println("=== Optimistic Concurrency Control Example ===")
	
	// Create optimistic controller with default config
	controller := dukdakit.Distributed.NewOptimistic()
	
	// Create an entity
	entity := &ExampleEntity{
		ID:    "player-123",
		Name:  "TestPlayer",
		Score: 100,
	}
	
	ctx := context.Background()
	
	// Update entity with optimistic locking
	err := controller.UpdateWithOptimisticLock(
		ctx,
		entity,
		func(e distributed.VersionedEntity) error {
			// Simulate updating the entity
			player := e.(*ExampleEntity)
			player.Score += 50
			fmt.Printf("Updated player %s score to %d (version: %d)\n", 
				player.Name, player.Score, player.GetVersion())
			return nil
		},
		func(e distributed.VersionedEntity) error {
			// Validation function
			player := e.(*ExampleEntity)
			if player.Score < 0 {
				return fmt.Errorf("score cannot be negative")
			}
			return nil
		},
	)
	
	if err != nil {
		fmt.Printf("Update failed: %v\n", err)
	} else {
		fmt.Printf("✅ Successfully updated entity with optimistic locking\n")
	}
	
	// Show metrics
	metrics := controller.GetMetrics()
	fmt.Printf("📊 Metrics - Total: %d, Success: %d, Conflicts: %d, Failed: %d\n",
		metrics.TotalOperations,
		metrics.SuccessfulUpdates,
		metrics.ConflictRetries,
		metrics.FailedOperations)
	
	fmt.Println()
}

func retryExample() {
	fmt.Println("=== Retry Mechanisms Example ===")
	
	// Create retry configuration without circuit breaker for this example
	config := retry.RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   100 * time.Millisecond,
		Multiplier:  2.0,
		Jitter:      true,
		// CircuitBreaker: nil, // Disabled for clearer demonstration
	}
	
	// Create retrier
	retrier := dukdakit.Retry.New(config)
	
	ctx := context.Background()
	attemptCount := 0
	
	// Example 1: Successful retry after failures
	fmt.Println("Example 1: Operation that succeeds after retries")
	err := retrier.Execute(ctx, func() error {
		attemptCount++
		fmt.Printf("  Attempt %d...\n", attemptCount)
		
		if attemptCount < 3 {
			return fmt.Errorf("temporary failure")
		}
		
		fmt.Printf("  ✅ Success on attempt %d!\n", attemptCount)
		return nil
	})
	
	if err != nil {
		fmt.Printf("❌ Operation failed: %v\n", err)
	}
	
	// Reset for next example
	attemptCount = 0
	
	fmt.Println("\nExample 2: Operation that always fails")
	err = retrier.Execute(ctx, func() error {
		attemptCount++
		fmt.Printf("  Attempt %d - always fails\n", attemptCount)
		return fmt.Errorf("persistent failure")
	})
	
	if err != nil {
		fmt.Printf("❌ Operation failed after all retries: %v\n", err)
	}
	
	// Show retry metrics
	metrics := retrier.GetMetrics()
	fmt.Printf("📊 Retry Metrics - Total: %d, Success: %d, Failed: %d, Avg Attempts: %.1f\n",
		metrics.TotalAttempts,
		metrics.SuccessfulCalls,
		metrics.FailedCalls,
		metrics.AverageAttempts)
	
	fmt.Println()
}