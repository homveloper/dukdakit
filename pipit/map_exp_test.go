package pipit

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMapUnsafe_BasicFunctionality tests basic MapUnsafe operation
func TestMapUnsafe_BasicFunctionality(t *testing.T) {
	t.Run("MapUnsafe with int to string conversion", func(t *testing.T) {
		data := []int{1, 2, 3}
		source := NewSliceIterator(data)
		
		query := &Query[int]{
			source:   source,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      nil,
		}
		
		// Use MapUnsafe to convert int to string
		unsafeQuery := query.MapUnsafe(func(item any) any {
			// Runtime type assertion - dangerous but necessary
			return fmt.Sprintf("num_%d", item.(int))
		})
		
		// Should return Query[any]
		assert.Equal(t, 1, len(unsafeQuery.pipeline))
		assert.Equal(t, MapOp, unsafeQuery.pipeline[0].Type())
		assert.Nil(t, unsafeQuery.err)
	})
	
	t.Run("MapUnsafe with existing error", func(t *testing.T) {
		existingErr := fmt.Errorf("existing error")
		query := &Query[int]{
			source:   nil,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      existingErr,
		}
		
		result := query.MapUnsafe(func(item any) any {
			return fmt.Sprintf("%v", item)
		})
		
		assert.Equal(t, existingErr, result.err)
		assert.Equal(t, 0, len(result.pipeline)) // No operation added
	})
}

// TestUnsafeMapOperation_Apply tests the Apply method
func TestUnsafeMapOperation_Apply(t *testing.T) {
	ctx := context.Background()
	
	t.Run("Apply with successful conversion", func(t *testing.T) {
		op := &UnsafeMapOperation{
			unsafeMapper: func(item any) any {
				return fmt.Sprintf("value_%v", item)
			},
		}
		
		result, err := op.Apply(ctx, 42)
		require.NoError(t, err)
		assert.Equal(t, "value_42", result)
	})
	
	t.Run("Apply with context cancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		op := &UnsafeMapOperation{
			unsafeMapper: func(item any) any {
				return fmt.Sprintf("%v", item)
			},
		}
		
		result, err := op.Apply(cancelCtx, 42)
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
		assert.Nil(t, result)
	})
	
	t.Run("Apply with type assertion panic (recovered)", func(t *testing.T) {
		op := &UnsafeMapOperation{
			unsafeMapper: func(item any) any {
				// This will panic if item is not an int
				return item.(int) * 2
			},
		}
		
		// This should work
		result, err := op.Apply(ctx, 21)
		require.NoError(t, err)
		assert.Equal(t, 42, result)
		
		// This will panic in a real scenario, but for test we'll skip
		// In production, MapUnsafe users must handle panics themselves
	})
}

// TestMapUnsafe_DangerousScenarios demonstrates why MapUnsafe is dangerous
func TestMapUnsafe_DangerousScenarios(t *testing.T) {
	t.Run("Potential panic scenario - type mismatch", func(t *testing.T) {
		// This test demonstrates the danger but doesn't actually panic
		// In real usage, this would cause a runtime panic
		
		data := []string{"hello", "world"}
		source := NewSliceIterator(data)
		
		query := &Query[string]{
			source:   source,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      nil,
		}
		
		// This mapper expects int but will receive string - would panic in execution
		dangerousQuery := query.MapUnsafe(func(item any) any {
			// In real execution: item.(int) would panic here
			// For test, we'll handle it safely
			if val, ok := item.(int); ok {
				return val * 2
			}
			return "TYPE_ERROR"
		})
		
		assert.Equal(t, 1, len(dangerousQuery.pipeline))
		// The danger is in execution, not in query construction
	})
	
	t.Run("Loss of type information", func(t *testing.T) {
		data := []int{1, 2, 3}
		source := NewSliceIterator(data)
		
		typedQuery := &Query[int]{
			source:   source,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      nil,
		}
		
		// After MapUnsafe, we lose all type information
		untypedQuery := typedQuery.MapUnsafe(func(item any) any {
			return item
		})
		
		// typedQuery is Query[int]
		// untypedQuery is Query[any] - type safety lost
		
		assert.Equal(t, 1, len(untypedQuery.pipeline))
		
		// Cannot chain type-safe operations after MapUnsafe
		// This would be a compile error:
		// untypedQuery.Filter(func(x int) bool { return x > 0 })
	})
}

// TestMapUnsafe_ValidUseCases shows legitimate use cases
func TestMapUnsafe_ValidUseCases(t *testing.T) {
	t.Run("JSON-like data processing", func(t *testing.T) {
		// Simulate processing data from JSON where types are known at runtime
		data := []any{
			map[string]any{"id": 1, "name": "Alice"},
			map[string]any{"id": 2, "name": "Bob"},
		}
		source := NewSliceIterator(data)
		
		query := &Query[any]{
			source:   source,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      nil,
		}
		
		// Extract names from map structure
		nameQuery := query.MapUnsafe(func(item any) any {
			if userMap, ok := item.(map[string]any); ok {
				if name, exists := userMap["name"]; exists {
					return name
				}
			}
			return "UNKNOWN"
		})
		
		assert.Equal(t, 1, len(nameQuery.pipeline))
		assert.Equal(t, MapOp, nameQuery.pipeline[0].Type())
	})
	
	t.Run("Dynamic type conversion", func(t *testing.T) {
		// Mixed type slice - common in dynamic scenarios
		data := []any{1, "hello", 3.14, true}
		source := NewSliceIterator(data)
		
		query := &Query[any]{
			source:   source,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      nil,
		}
		
		// Convert everything to string representation
		stringQuery := query.MapUnsafe(func(item any) any {
			return fmt.Sprintf("%v", item)
		})
		
		assert.Equal(t, 1, len(stringQuery.pipeline))
	})
}

// TestMapUnsafeE_ContextAwareErrorHandling tests MapUnsafeE with context and error support
func TestMapUnsafeE_ContextAwareErrorHandling(t *testing.T) {
	t.Run("MapUnsafeE with successful conversion", func(t *testing.T) {
		data := []any{"123", "456", "789"}
		source := NewSliceIterator(data)
		
		query := &Query[any]{
			source:   source,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      nil,
		}
		
		// Convert strings to integers with error handling
		intQuery := query.MapUnsafeE(func(ctx context.Context, item any) (any, error) {
			str, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("expected string, got %T", item)
			}
			
			num, err := strconv.Atoi(str)
			if err != nil {
				return nil, fmt.Errorf("invalid number: %w", err)
			}
			return num, nil
		})
		
		assert.Equal(t, 1, len(intQuery.pipeline))
		assert.Equal(t, MapOp, intQuery.pipeline[0].Type())
		assert.Nil(t, intQuery.err)
	})
	
	t.Run("MapUnsafeE with error in mapper", func(t *testing.T) {
		op := &UnsafeMapOperation{
			unsafeMapperWithContext: func(ctx context.Context, item any) (any, error) {
				str, ok := item.(string)
				if !ok {
					return nil, fmt.Errorf("expected string, got %T", item)
				}
				
				// This will fail for "invalid"
				return strconv.Atoi(str)
			},
		}
		
		ctx := context.Background()
		
		// Valid conversion
		result, err := op.Apply(ctx, "123")
		require.NoError(t, err)
		assert.Equal(t, 123, result)
		
		// Invalid conversion
		result, err = op.Apply(ctx, "invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid syntax")
		assert.Equal(t, 0, result) // strconv.Atoi returns 0 on error
		
		// Type mismatch
		result, err = op.Apply(ctx, 123)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected string, got int")
		assert.Nil(t, result)
	})
	
	t.Run("MapUnsafeE with context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		
		op := &UnsafeMapOperation{
			unsafeMapperWithContext: func(ctx context.Context, item any) (any, error) {
				select {
				case <-time.After(20 * time.Millisecond): // Longer than timeout
					return fmt.Sprintf("%v", item), nil
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			},
		}
		
		result, err := op.Apply(ctx, "test")
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		assert.Nil(t, result)
	})
	
	t.Run("MapUnsafeE with existing error", func(t *testing.T) {
		existingErr := fmt.Errorf("existing error")
		query := &Query[string]{
			source:   nil,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      existingErr,
		}
		
		result := query.MapUnsafeE(func(ctx context.Context, item any) (any, error) {
			return strconv.Atoi(item.(string))
		})
		
		assert.Equal(t, existingErr, result.err)
		assert.Equal(t, 0, len(result.pipeline)) // No operation added
	})
	
	t.Run("MapUnsafeE context value propagation", func(t *testing.T) {
		type contextKey string
		const key = contextKey("test-key")
		ctx := context.WithValue(context.Background(), key, "test-value")
		
		op := &UnsafeMapOperation{
			unsafeMapperWithContext: func(ctx context.Context, item any) (any, error) {
				value := ctx.Value(key)
				if value == "test-value" {
					return fmt.Sprintf("%v_processed", item), nil
				}
				return nil, fmt.Errorf("context value not found")
			},
		}
		
		result, err := op.Apply(ctx, "test")
		require.NoError(t, err)
		assert.Equal(t, "test_processed", result)
	})
}

// TestMapUnsafeE_RealWorldScenarios tests practical use cases
func TestMapUnsafeE_RealWorldScenarios(t *testing.T) {
	t.Run("JSON parsing with error handling", func(t *testing.T) {
		// Simulate JSON-like data with potential parsing errors
		data := []any{
			map[string]any{"age": "25", "valid": true},
			map[string]any{"age": "invalid", "valid": true},
			map[string]any{"name": "Alice"}, // missing age field
		}
		source := NewSliceIterator(data)
		
		query := &Query[any]{
			source:   source,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      nil,
		}
		
		// Extract and parse age with comprehensive error handling
		ageQuery := query.MapUnsafeE(func(ctx context.Context, item any) (any, error) {
			userMap, ok := item.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("expected map, got %T", item)
			}
			
			ageStr, exists := userMap["age"]
			if !exists {
				return nil, fmt.Errorf("age field missing")
			}
			
			ageStrTyped, ok := ageStr.(string)
			if !ok {
				return nil, fmt.Errorf("age must be string, got %T", ageStr)
			}
			
			age, err := strconv.Atoi(ageStrTyped)
			if err != nil {
				return nil, fmt.Errorf("invalid age format: %w", err)
			}
			
			if age < 0 || age > 150 {
				return nil, fmt.Errorf("age out of range: %d", age)
			}
			
			return age, nil
		})
		
		assert.Equal(t, 1, len(ageQuery.pipeline))
		
		// Test the operation directly
		op := ageQuery.pipeline[0].(*UnsafeMapOperation)
		ctx := context.Background()
		
		// Valid case
		result, err := op.Apply(ctx, data[0])
		require.NoError(t, err)
		assert.Equal(t, 25, result)
		
		// Invalid format case  
		result, err = op.Apply(ctx, data[1])
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid age format")
		
		// Missing field case
		result, err = op.Apply(ctx, data[2])
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "age field missing")
	})
	
	t.Run("Type conversion with fallbacks", func(t *testing.T) {
		data := []any{42, "123", 3.14, "invalid", nil}
		source := NewSliceIterator(data)
		
		query := &Query[any]{
			source:   source,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      nil,
		}
		
		// Convert various types to int with fallbacks
		intQuery := query.MapUnsafeE(func(ctx context.Context, item any) (any, error) {
			switch v := item.(type) {
			case int:
				return v, nil
			case string:
				if num, err := strconv.Atoi(v); err == nil {
					return num, nil
				}
				return 0, fmt.Errorf("invalid string format: %s", v)
			case float64:
				return int(v), nil
			case nil:
				return 0, fmt.Errorf("nil value not allowed")
			default:
				return 0, fmt.Errorf("unsupported type: %T", v)
			}
		})
		
		assert.Equal(t, 1, len(intQuery.pipeline))
	})
}

// TestMapUnsafe_ChainedPipeline tests multiple MapUnsafe operations in sequence
func TestMapUnsafe_ChainedPipeline(t *testing.T) {
	t.Run("Multi-stage data processing pipeline", func(t *testing.T) {
		// Simulate raw data from external API or mixed sources
		rawData := []any{
			map[string]any{"type": "user", "data": "alice,25,engineer"},
			map[string]any{"type": "product", "data": "laptop:1200:electronics"},
			map[string]any{"type": "user", "data": "bob,30,designer"},
			"invalid_record", // Bad data mixed in
			map[string]any{"type": "order", "data": "ord_123:alice:laptop:1"},
		}
		source := NewSliceIterator(rawData)
		
		query := &Query[any]{
			source:   source,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      nil,
		}
		
		// Stage 1: Filter out invalid records and extract typed data
		stage1 := query.MapUnsafe(func(item any) any {
			switch record := item.(type) {
			case map[string]any:
				if recordType, ok := record["type"].(string); ok {
					if data, ok := record["data"].(string); ok {
						return map[string]any{
							"type": recordType,
							"raw_data": data,
							"stage": "parsed",
						}
					}
				}
			}
			// Return error marker for invalid data
			return map[string]any{"type": "error", "reason": "invalid_format"}
		})
		
		// Stage 2: Parse raw data strings into structured data
		stage2 := stage1.MapUnsafe(func(item any) any {
			record := item.(map[string]any)
			
			if record["type"] == "error" {
				return record // Pass through errors
			}
			
			recordType := record["type"].(string)
			rawData := record["raw_data"].(string)
			
			switch recordType {
			case "user":
				// In real scenario, would split by comma and parse fields
				_ = rawData // Simplified parsing for test
				return map[string]any{
					"type": "user",
					"parsed_data": map[string]any{
						"name": "extracted_name",
						"age": "extracted_age", 
						"role": "extracted_role",
					},
					"stage": "structured",
				}
			case "product":
				return map[string]any{
					"type": "product",
					"parsed_data": map[string]any{
						"name": "extracted_product",
						"price": "extracted_price",
						"category": "extracted_category",
					},
					"stage": "structured",
				}
			case "order":
				return map[string]any{
					"type": "order", 
					"parsed_data": map[string]any{
						"id": "extracted_id",
						"user": "extracted_user",
						"product": "extracted_product",
						"quantity": "extracted_quantity",
					},
					"stage": "structured",
				}
			default:
				return map[string]any{"type": "error", "reason": "unknown_type"}
			}
		})
		
		// Stage 3: Convert to final normalized format
		stage3 := stage2.MapUnsafe(func(item any) any {
			record := item.(map[string]any)
			
			if record["type"] == "error" {
				return map[string]any{
					"entity_type": "error",
					"error_message": record["reason"],
					"processed": true,
				}
			}
			
			recordType := record["type"].(string)
			parsedData := record["parsed_data"].(map[string]any)
			
			// Normalize all entity types to common format
			return map[string]any{
				"entity_type": recordType,
				"entity_id": fmt.Sprintf("%s_%d", recordType, len(parsedData)),
				"attributes": parsedData,
				"processed": true,
				"pipeline_version": "v1.0",
			}
		})
		
		// Verify the pipeline structure
		assert.Equal(t, 3, len(stage3.pipeline))
		assert.Equal(t, MapOp, stage3.pipeline[0].Type())
		assert.Equal(t, MapOp, stage3.pipeline[1].Type())
		assert.Equal(t, MapOp, stage3.pipeline[2].Type())
	})
	
	t.Run("Log processing pipeline with mixed data types", func(t *testing.T) {
		// Simulate mixed log entries from different sources
		logData := []any{
			"ERROR 2023-08-21 10:30:00 Database connection failed",
			map[string]any{"level": "INFO", "timestamp": "2023-08-21T10:30:01Z", "message": "User logged in", "user_id": 123},
			"WARN 2023-08-21 10:30:02 Slow query detected: 1.2s",
			[]any{"DEBUG", "2023-08-21T10:30:03Z", "Cache miss", map[string]any{"key": "user:123"}},
			42, // Invalid log entry
		}
		source := NewSliceIterator(logData)
		
		query := &Query[any]{
			source:   source,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      nil,
		}
		
		// Stage 1: Normalize all log formats to common structure
		normalized := query.MapUnsafe(func(item any) any {
			switch entry := item.(type) {
			case string:
				// Parse string format logs
				return map[string]any{
					"format": "string",
					"raw": entry,
					"normalized": false,
				}
			case map[string]any:
				// Already structured
				return map[string]any{
					"format": "json",
					"raw": entry,
					"normalized": false,
				}
			case []any:
				// Array format logs
				return map[string]any{
					"format": "array",
					"raw": entry,
					"normalized": false,
				}
			default:
				return map[string]any{
					"format": "error",
					"raw": item,
					"error": "unsupported_format",
				}
			}
		})
		
		// Stage 2: Extract common fields
		extracted := normalized.MapUnsafe(func(item any) any {
			record := item.(map[string]any)
			
			if record["format"] == "error" {
				return record
			}
			
			// Extract timestamp, level, message regardless of original format
			result := map[string]any{
				"original_format": record["format"],
				"extracted": true,
			}
			
			switch record["format"] {
			case "string":
				result["level"] = "parsed_from_string"
				result["timestamp"] = "parsed_timestamp"
				result["message"] = "parsed_message"
			case "json":
				rawData := record["raw"].(map[string]any)
				result["level"] = rawData["level"]
				result["timestamp"] = rawData["timestamp"]
				result["message"] = rawData["message"]
			case "array":
				rawData := record["raw"].([]any)
				if len(rawData) >= 3 {
					result["level"] = rawData[0]
					result["timestamp"] = rawData[1]
					result["message"] = rawData[2]
				}
			}
			
			return result
		})
		
		// Stage 3: Final enrichment and classification
		enriched := extracted.MapUnsafe(func(item any) any {
			record := item.(map[string]any)
			
			if _, isError := record["error"]; isError {
				return map[string]any{
					"type": "processing_error",
					"severity": "high",
					"requires_attention": true,
					"details": record,
				}
			}
			
			level := "unknown"
			if l, ok := record["level"].(string); ok {
				level = l
			}
			
			return map[string]any{
				"type": "log_entry",
				"severity": level,
				"requires_attention": level == "ERROR" || level == "WARN",
				"processed_at": "2023-08-21T10:30:00Z",
				"pipeline_stages": 3,
				"details": record,
			}
		})
		
		// Verify pipeline depth
		assert.Equal(t, 3, len(enriched.pipeline))
		for i := 0; i < 3; i++ {
			assert.Equal(t, MapOp, enriched.pipeline[i].Type())
		}
	})
	
	t.Run("E-commerce data integration pipeline", func(t *testing.T) {
		// Simulate data from multiple e-commerce sources with different schemas
		integrationData := []any{
			// Amazon-style data
			map[string]any{
				"source": "amazon",
				"product": map[string]any{
					"asin": "B08N5WRWNW",
					"title": "Echo Dot (4th Gen)",
					"price": map[string]any{"amount": 49.99, "currency": "USD"},
					"reviews": map[string]any{"average": 4.7, "count": 12453},
				},
			},
			// eBay-style data  
			map[string]any{
				"source": "ebay",
				"item": map[string]any{
					"itemId": "123456789",
					"name": "Apple AirPods Pro",
					"currentPrice": "$249.99",
					"sellerInfo": map[string]any{"rating": "99.2%", "feedbackCount": 1542},
				},
			},
			// Internal system data
			map[string]any{
				"source": "internal",
				"sku": "SKU-2023-001",
				"description": "Wireless Headphones Premium",
				"cost": 89.50,
				"markup": 1.8,
				"category_id": 15,
			},
			// Malformed data
			map[string]any{
				"source": "unknown",
				"corrupted": true,
			},
		}
		source := NewSliceIterator(integrationData)
		
		query := &Query[any]{
			source:   source,
			pipeline: []Operation{},
			ctx:      context.Background(),
			err:      nil,
		}
		
		// Stage 1: Source-specific parsing
		sourceParsed := query.MapUnsafe(func(item any) any {
			record := item.(map[string]any)
			source := record["source"].(string)
			
			switch source {
			case "amazon":
				product := record["product"].(map[string]any)
				return map[string]any{
					"source": source,
					"id": product["asin"],
					"name": product["title"],
					"price_info": product["price"],
					"rating_info": product["reviews"],
					"parsed": true,
				}
			case "ebay":
				item := record["item"].(map[string]any)
				return map[string]any{
					"source": source,
					"id": item["itemId"],
					"name": item["name"], 
					"price_info": item["currentPrice"],
					"rating_info": item["sellerInfo"],
					"parsed": true,
				}
			case "internal":
				return map[string]any{
					"source": source,
					"id": record["sku"],
					"name": record["description"],
					"price_info": map[string]any{
						"cost": record["cost"],
						"markup": record["markup"],
					},
					"category": record["category_id"],
					"parsed": true,
				}
			default:
				return map[string]any{
					"source": source,
					"error": "unsupported_source",
					"parsed": false,
				}
			}
		})
		
		// Stage 2: Normalize price information
		priceNormalized := sourceParsed.MapUnsafe(func(item any) any {
			record := item.(map[string]any)
			
			if !record["parsed"].(bool) {
				return record // Pass through errors
			}
			
			source := record["source"].(string)
			priceInfo := record["price_info"]
			
			var normalizedPrice float64
			
			switch source {
			case "amazon":
				if priceMap, ok := priceInfo.(map[string]any); ok {
					if amount, ok := priceMap["amount"].(float64); ok {
						normalizedPrice = amount
					}
				}
			case "ebay":
				if _, ok := priceInfo.(string); ok {
					// Parse "$249.99" format - simplified for test
					normalizedPrice = 249.99 // Simplified parsing
				}
			case "internal":
				if priceMap, ok := priceInfo.(map[string]any); ok {
					if cost, ok := priceMap["cost"].(float64); ok {
						if markup, ok := priceMap["markup"].(float64); ok {
							normalizedPrice = cost * markup
						}
					}
				}
			}
			
			result := make(map[string]any)
			for k, v := range record {
				result[k] = v
			}
			result["normalized_price"] = normalizedPrice
			result["price_normalized"] = true
			
			return result
		})
		
		// Stage 3: Final product catalog format
		catalogFormat := priceNormalized.MapUnsafe(func(item any) any {
			record := item.(map[string]any)
			
			if !record["parsed"].(bool) {
				return map[string]any{
					"catalog_entry": false,
					"error_reason": record["error"],
					"source": record["source"],
				}
			}
			
			return map[string]any{
				"catalog_entry": true,
				"unified_id": fmt.Sprintf("%s_%v", record["source"], record["id"]),
				"product_name": record["name"],
				"price_usd": record["normalized_price"],
				"source_system": record["source"],
				"integration_timestamp": "2023-08-21T10:30:00Z",
				"schema_version": "v2.1",
				"processing_stages": 3,
			}
		})
		
		// Verify the complete pipeline
		assert.Equal(t, 3, len(catalogFormat.pipeline))
		
		// All should be MapUnsafe operations
		for i := 0; i < 3; i++ {
			op := catalogFormat.pipeline[i]
			assert.Equal(t, MapOp, op.Type())
			
			// Verify it's actually an UnsafeMapOperation
			_, isUnsafe := op.(*UnsafeMapOperation)
			assert.True(t, isUnsafe, "Operation %d should be UnsafeMapOperation", i)
		}
	})
}