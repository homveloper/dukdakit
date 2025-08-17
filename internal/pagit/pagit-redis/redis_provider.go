package pagitredis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// RedisListProvider provides offset-based pagination for Redis Lists
// Ideal for game scenarios like: leaderboards, recent activities, item collections
type RedisListProvider[T any] struct {
	client redis.Cmdable
	key    string
}

// NewRedisListProvider creates a Redis list provider for offset-based pagination
func NewRedisListProvider[T any](client redis.Cmdable, key string) *RedisListProvider[T] {
	return &RedisListProvider[T]{
		client: client,
		key:    key,
	}
}

// GetData implements pagit.DataProvider interface for Redis Lists
func (p *RedisListProvider[T]) GetData(ctx context.Context, offset, limit int) ([]T, error) {
	// Redis LRANGE is 0-indexed: LRANGE key start stop
	start := int64(offset)
	stop := int64(offset + limit - 1)

	values, err := p.client.LRange(ctx, p.key, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get list range: %w", err)
	}

	result := make([]T, 0, len(values))
	for _, value := range values {
		var item T
		if err := json.Unmarshal([]byte(value), &item); err != nil {
			return nil, fmt.Errorf("failed to unmarshal item: %w", err)
		}
		result = append(result, item)
	}

	return result, nil
}

// GetTotalCount implements pagit.CountProvider interface for Redis Lists
func (p *RedisListProvider[T]) GetTotalCount(ctx context.Context) (int64, error) {
	count, err := p.client.LLen(ctx, p.key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get list length: %w", err)
	}
	return count, nil
}

// RedisSortedSetProvider provides cursor-based pagination for Redis Sorted Sets
// Ideal for game scenarios like: leaderboards, time-ordered events, score rankings
type RedisSortedSetProvider[T any] struct {
	client    redis.Cmdable
	key       string
	extractor func(T) float64 // Extract score for cursor
}

// NewRedisSortedSetProvider creates a Redis sorted set provider for cursor-based pagination
func NewRedisSortedSetProvider[T any](
	client redis.Cmdable,
	key string,
	extractor func(T) float64,
) *RedisSortedSetProvider[T] {
	return &RedisSortedSetProvider[T]{
		client:    client,
		key:       key,
		extractor: extractor,
	}
}

// GetDataAfter implements pagit.CursorDataProvider interface
func (p *RedisSortedSetProvider[T]) GetDataAfter(
	ctx context.Context,
	cursor *float64,
	limit int,
) ([]T, error) {
	var min string
	if cursor == nil {
		min = "-inf"
	} else {
		min = fmt.Sprintf("(%f", *cursor) // Exclusive range: score > cursor
	}

	return p.getRangeByScore(ctx, min, "+inf", limit)
}

// GetDataBefore implements pagit.CursorDataProvider interface
func (p *RedisSortedSetProvider[T]) GetDataBefore(
	ctx context.Context,
	cursor *float64,
	limit int,
) ([]T, error) {
	var max string
	if cursor == nil {
		// Start from the highest score and go backward (reverse order from highest)
		max = "+inf"
	} else {
		max = fmt.Sprintf("(%f", *cursor) // Exclusive range: score < cursor
	}

	// For "before" we need to get in reverse order and then reverse the slice
	values, err := p.client.ZRevRangeByScore(ctx, p.key, &redis.ZRangeBy{
		Min:   "-inf",
		Max:   max,
		Count: int64(limit),
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get sorted set range: %w", err)
	}

	result := make([]T, 0, len(values))
	for i := len(values) - 1; i >= 0; i-- { // Reverse to maintain proper order
		var item T
		if err := json.Unmarshal([]byte(values[i]), &item); err != nil {
			return nil, fmt.Errorf("failed to unmarshal item: %w", err)
		}
		result = append(result, item)
	}

	return result, nil
}

// HasDataAfter implements pagit.CursorCheckProvider interface
func (p *RedisSortedSetProvider[T]) HasDataAfter(
	ctx context.Context,
	cursor float64,
) (bool, error) {
	min := fmt.Sprintf("(%f", cursor)
	count, err := p.client.ZCount(ctx, p.key, min, "+inf").Result()
	if err != nil {
		return false, fmt.Errorf("failed to count sorted set members: %w", err)
	}
	return count > 0, nil
}

// HasDataBefore implements pagit.CursorCheckProvider interface
func (p *RedisSortedSetProvider[T]) HasDataBefore(
	ctx context.Context,
	cursor float64,
) (bool, error) {
	max := fmt.Sprintf("(%f", cursor)
	count, err := p.client.ZCount(ctx, p.key, "-inf", max).Result()
	if err != nil {
		return false, fmt.Errorf("failed to count sorted set members: %w", err)
	}
	return count > 0, nil
}

// getRangeByScore helper method for getting data by score range
func (p *RedisSortedSetProvider[T]) getRangeByScore(
	ctx context.Context,
	min, max string,
	limit int,
) ([]T, error) {
	values, err := p.client.ZRangeByScore(ctx, p.key, &redis.ZRangeBy{
		Min:   min,
		Max:   max,
		Count: int64(limit),
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get sorted set range: %w", err)
	}

	result := make([]T, 0, len(values))
	for _, value := range values {
		var item T
		if err := json.Unmarshal([]byte(value), &item); err != nil {
			return nil, fmt.Errorf("failed to unmarshal item: %w", err)
		}
		result = append(result, item)
	}

	return result, nil
}

// RedisHashProvider provides offset-based pagination for Redis Hash fields
// Ideal for game scenarios like: player inventories, game settings, user profiles
type RedisHashProvider[T any] struct {
	client redis.Cmdable
	key    string
}

// NewRedisHashProvider creates a Redis hash provider for offset-based pagination
func NewRedisHashProvider[T any](client redis.Cmdable, key string) *RedisHashProvider[T] {
	return &RedisHashProvider[T]{
		client: client,
		key:    key,
	}
}

// GetData implements pagit.DataProvider interface for Redis Hashes
func (p *RedisHashProvider[T]) GetData(ctx context.Context, offset, limit int) ([]T, error) {
	// Get all hash keys first (Redis doesn't support direct offset/limit on hashes)
	fields, err := p.client.HKeys(ctx, p.key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get hash keys: %w", err)
	}

	// Apply offset/limit to field names
	if offset >= len(fields) {
		return []T{}, nil
	}

	endIndex := offset + limit
	if endIndex > len(fields) {
		endIndex = len(fields)
	}

	selectedFields := fields[offset:endIndex]
	if len(selectedFields) == 0 {
		return []T{}, nil
	}

	// Get values for selected fields
	values, err := p.client.HMGet(ctx, p.key, selectedFields...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get hash values: %w", err)
	}

	result := make([]T, 0, len(values))
	for _, value := range values {
		if value == nil {
			continue // Skip nil values
		}

		var item T
		if err := json.Unmarshal([]byte(value.(string)), &item); err != nil {
			return nil, fmt.Errorf("failed to unmarshal item: %w", err)
		}
		result = append(result, item)
	}

	return result, nil
}

// GetTotalCount implements pagit.CountProvider interface for Redis Hashes
func (p *RedisHashProvider[T]) GetTotalCount(ctx context.Context) (int64, error) {
	count, err := p.client.HLen(ctx, p.key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get hash length: %w", err)
	}
	return count, nil
}

// Helper functions for adding data to Redis structures

// AddToList adds an item to a Redis list (useful for testing and data setup)
func AddToList[T any](ctx context.Context, client redis.Cmdable, key string, items ...T) error {
	for _, item := range items {
		data, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("failed to marshal item: %w", err)
		}

		if err := client.RPush(ctx, key, string(data)).Err(); err != nil {
			return fmt.Errorf("failed to push to list: %w", err)
		}
	}
	return nil
}

// AddToSortedSet adds an item to a Redis sorted set with a score
func AddToSortedSet[T any](ctx context.Context, client redis.Cmdable, key string, score float64, item T) error {
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	if err := client.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: string(data),
	}).Err(); err != nil {
		return fmt.Errorf("failed to add to sorted set: %w", err)
	}
	return nil
}

// AddToHash adds an item to a Redis hash with a field name
func AddToHash[T any](ctx context.Context, client redis.Cmdable, key, field string, item T) error {
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	if err := client.HSet(ctx, key, field, string(data)).Err(); err != nil {
		return fmt.Errorf("failed to set hash field: %w", err)
	}
	return nil
}
