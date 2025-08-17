package pagitredis

import (
	"context"
	"fmt"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/homveloper/dukdakit"
	"github.com/homveloper/dukdakit/internal/pagit"
)

// Test data structures for game scenarios
type Player struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Score int    `json:"score"`
	Level int    `json:"level"`
}

type Item struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Rarity int    `json:"rarity"`
}

type GameEvent struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	PlayerID  int64  `json:"player_id"`
	Timestamp int64  `json:"timestamp"`
	Data      string `json:"data"`
}

func setupRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return mr, client
}

func TestRedisListProvider_OffsetPagination(t *testing.T) {
	mr, client := setupRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	key := "test:players"

	// Setup test data - player leaderboard
	players := []Player{
		{ID: 1, Name: "Player1", Score: 1000, Level: 10},
		{ID: 2, Name: "Player2", Score: 950, Level: 9},
		{ID: 3, Name: "Player3", Score: 900, Level: 8},
		{ID: 4, Name: "Player4", Score: 850, Level: 7},
		{ID: 5, Name: "Player5", Score: 800, Level: 6},
	}

	err := AddToList(ctx, client, key, players...)
	require.NoError(t, err)

	provider := NewRedisListProvider[Player](client, key)

	t.Run("FirstPage", func(t *testing.T) {
		config := pagit.OffsetConfig{
			Page:     1,
			PageSize: 2,
		}

		result, err := dukdakit.PaginateOffset(ctx, provider, config)
		require.NoError(t, err)

		offsetResult, ok := result.(pagit.OffsetResult[Player])
		require.True(t, ok)

		assert.Len(t, offsetResult.Data, 2)
		assert.Equal(t, int64(1), offsetResult.Data[0].ID)
		assert.Equal(t, int64(2), offsetResult.Data[1].ID)
		assert.Equal(t, int64(5), offsetResult.TotalCount)
		assert.Equal(t, 1, offsetResult.Page)
		assert.Equal(t, 3, offsetResult.TotalPages)
		assert.True(t, offsetResult.HasNext)
		assert.False(t, offsetResult.HasPrev)
	})

	t.Run("SecondPage", func(t *testing.T) {
		config := pagit.OffsetConfig{
			Page:     2,
			PageSize: 2,
		}

		result, err := dukdakit.PaginateOffset(ctx, provider, config)
		require.NoError(t, err)

		offsetResult, ok := result.(pagit.OffsetResult[Player])
		require.True(t, ok)

		assert.Len(t, offsetResult.Data, 2)
		assert.Equal(t, int64(3), offsetResult.Data[0].ID)
		assert.Equal(t, int64(4), offsetResult.Data[1].ID)
		assert.True(t, offsetResult.HasNext)
		assert.True(t, offsetResult.HasPrev)
	})

	t.Run("LastPage", func(t *testing.T) {
		config := pagit.OffsetConfig{
			Page:     3,
			PageSize: 2,
		}

		result, err := dukdakit.PaginateOffset(ctx, provider, config)
		require.NoError(t, err)

		offsetResult, ok := result.(pagit.OffsetResult[Player])
		require.True(t, ok)

		assert.Len(t, offsetResult.Data, 1)
		assert.Equal(t, int64(5), offsetResult.Data[0].ID)
		assert.False(t, offsetResult.HasNext)
		assert.True(t, offsetResult.HasPrev)
	})

	t.Run("EmptyPage", func(t *testing.T) {
		config := pagit.OffsetConfig{
			Page:     10,
			PageSize: 2,
		}

		result, err := dukdakit.PaginateOffset(ctx, provider, config)
		require.NoError(t, err)

		offsetResult, ok := result.(pagit.OffsetResult[Player])
		require.True(t, ok)

		assert.Empty(t, offsetResult.Data)
		assert.False(t, offsetResult.HasNext)
		assert.True(t, offsetResult.HasPrev)
	})
}

func TestRedisSortedSetProvider_CursorPagination(t *testing.T) {
	mr, client := setupRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	key := "test:leaderboard"

	// Setup test data - score-based leaderboard
	players := []Player{
		{ID: 1, Name: "TopPlayer", Score: 1000, Level: 10},
		{ID: 2, Name: "SecondPlayer", Score: 950, Level: 9},
		{ID: 3, Name: "ThirdPlayer", Score: 900, Level: 8},
		{ID: 4, Name: "FourthPlayer", Score: 850, Level: 7},
		{ID: 5, Name: "FifthPlayer", Score: 800, Level: 6},
	}

	// Add players to sorted set with scores
	for _, player := range players {
		err := AddToSortedSet(ctx, client, key, float64(player.Score), player)
		require.NoError(t, err)
	}

	extractor := func(p Player) float64 { return float64(p.Score) }
	provider := NewRedisSortedSetProvider(client, key, extractor)

	t.Run("FromStart", func(t *testing.T) {
		config := pagit.CursorConfig[float64]{
			PageSize:  2,
			Cursor:    nil, // Start from beginning
			Direction: pagit.CursorForward,
		}

		result, err := dukdakit.PaginateCursor(ctx, provider, config, extractor)
		require.NoError(t, err)

		assert.Len(t, result.Data, 2)
		assert.Equal(t, int64(5), result.Data[0].ID) // Lowest score first (800)
		assert.Equal(t, int64(4), result.Data[1].ID) // Next lowest (850)
		assert.True(t, result.HasNext)
		assert.False(t, result.HasPrev)
		assert.NotNil(t, result.NextCursor)
		assert.Equal(t, float64(850), *result.NextCursor)
	})

	t.Run("ForwardFromCursor", func(t *testing.T) {
		cursor := float64(850)
		config := pagit.CursorConfig[float64]{
			PageSize:  2,
			Cursor:    &cursor,
			Direction: pagit.CursorForward,
		}

		result, err := dukdakit.PaginateCursor(ctx, provider, config, extractor)
		require.NoError(t, err)

		assert.Len(t, result.Data, 2)
		assert.Equal(t, int64(3), result.Data[0].ID) // Score 900
		assert.Equal(t, int64(2), result.Data[1].ID) // Score 950
		assert.True(t, result.HasNext)
		assert.True(t, result.HasPrev)
	})

	t.Run("BackwardFromCursor", func(t *testing.T) {
		cursor := float64(900)
		config := pagit.CursorConfig[float64]{
			PageSize:  2,
			Cursor:    &cursor,
			Direction: pagit.CursorBackward,
		}

		result, err := dukdakit.PaginateCursor(ctx, provider, config, extractor)
		require.NoError(t, err)

		assert.Len(t, result.Data, 2)
		assert.Equal(t, int64(5), result.Data[0].ID) // Score 800
		assert.Equal(t, int64(4), result.Data[1].ID) // Score 850
		assert.True(t, result.HasNext)
		assert.False(t, result.HasPrev) // No data before score 800
	})

	t.Run("HasDataAfter", func(t *testing.T) {
		hasData, err := provider.HasDataAfter(ctx, 900)
		require.NoError(t, err)
		assert.True(t, hasData) // Should have 950 and 1000

		hasData, err = provider.HasDataAfter(ctx, 1000)
		require.NoError(t, err)
		assert.False(t, hasData) // No data after highest score
	})

	t.Run("HasDataBefore", func(t *testing.T) {
		hasData, err := provider.HasDataBefore(ctx, 900)
		require.NoError(t, err)
		assert.True(t, hasData) // Should have 800 and 850

		hasData, err = provider.HasDataBefore(ctx, 800)
		require.NoError(t, err)
		assert.False(t, hasData) // No data before lowest score
	})
}

func TestRedisHashProvider_OffsetPagination(t *testing.T) {
	mr, client := setupRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	key := "test:inventory"

	// Setup test data - player inventory
	items := []Item{
		{ID: 1, Name: "Iron Sword", Type: "weapon", Rarity: 1},
		{ID: 2, Name: "Magic Shield", Type: "armor", Rarity: 2},
		{ID: 3, Name: "Health Potion", Type: "consumable", Rarity: 1},
		{ID: 4, Name: "Steel Armor", Type: "armor", Rarity: 3},
		{ID: 5, Name: "Fire Scroll", Type: "consumable", Rarity: 2},
	}

	// Add items to hash
	for _, item := range items {
		field := fmt.Sprintf("item_%d", item.ID)
		err := AddToHash(ctx, client, key, field, item)
		require.NoError(t, err)
	}

	provider := NewRedisHashProvider[Item](client, key)

	t.Run("FirstPage", func(t *testing.T) {
		config := pagit.OffsetConfig{
			Page:     1,
			PageSize: 2,
		}

		result, err := dukdakit.PaginateOffset(ctx, provider, config)
		require.NoError(t, err)

		offsetResult, ok := result.(pagit.OffsetResult[Item])
		require.True(t, ok)

		assert.Len(t, offsetResult.Data, 2)
		assert.Equal(t, int64(5), offsetResult.TotalCount)
		assert.Equal(t, 1, offsetResult.Page)
		assert.Equal(t, 3, offsetResult.TotalPages)
		assert.True(t, offsetResult.HasNext)
		assert.False(t, offsetResult.HasPrev)
	})

	t.Run("GetTotalCount", func(t *testing.T) {
		totalCount, err := provider.GetTotalCount(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(5), totalCount)
	})
}

func TestRedisProvider_ErrorHandling(t *testing.T) {
	mr, client := setupRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()

	t.Run("NonExistentKey", func(t *testing.T) {
		provider := NewRedisListProvider[Player](client, "nonexistent:key")

		config := pagit.OffsetConfig{
			Page:     1,
			PageSize: 10,
		}

		result, err := dukdakit.PaginateOffset(ctx, provider, config)
		require.NoError(t, err)

		offsetResult, ok := result.(pagit.OffsetResult[Player])
		require.True(t, ok)

		assert.Empty(t, offsetResult.Data)
		assert.Equal(t, int64(0), offsetResult.TotalCount)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		key := "test:invalid"

		// Add invalid JSON data
		err := client.RPush(ctx, key, "invalid json data").Err()
		require.NoError(t, err)

		provider := NewRedisListProvider[Player](client, key)

		data, err := provider.GetData(ctx, 0, 10)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal item")
		assert.Nil(t, data)
	})
}

func TestRedisProvider_GameScenarios(t *testing.T) {
	mr, client := setupRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()

	t.Run("PlayerLeaderboard", func(t *testing.T) {
		key := "leaderboard:global"

		// Add top players
		topPlayers := []Player{
			{ID: 1, Name: "Legend", Score: 9999, Level: 100},
			{ID: 2, Name: "Master", Score: 8888, Level: 95},
			{ID: 3, Name: "Expert", Score: 7777, Level: 90},
			{ID: 4, Name: "Pro", Score: 6666, Level: 85},
			{ID: 5, Name: "Advanced", Score: 5555, Level: 80},
		}

		for _, player := range topPlayers {
			err := AddToSortedSet(ctx, client, key, float64(player.Score), player)
			require.NoError(t, err)
		}

		provider := NewRedisSortedSetProvider(client, key, func(p Player) float64 { return float64(p.Score) })

		// Get top 3 players (backward from highest - gets highest scores in ascending order)
		config := pagit.CursorConfig[float64]{
			PageSize:  3,
			Cursor:    nil,
			Direction: pagit.CursorBackward, // Start from highest scores and go backward
		}

		result, err := dukdakit.PaginateCursor(ctx, provider, config, func(p Player) float64 { return float64(p.Score) })
		require.NoError(t, err)

		assert.Len(t, result.Data, 3)
		assert.Equal(t, "Expert", result.Data[0].Name) // Score 7777 (highest 3rd)
		assert.Equal(t, "Master", result.Data[1].Name) // Score 8888 (highest 2nd)
		assert.Equal(t, "Legend", result.Data[2].Name) // Score 9999 (highest 1st)
	})

	t.Run("RecentGameEvents", func(t *testing.T) {
		key := "events:recent"

		events := []GameEvent{
			{ID: 1, Type: "login", PlayerID: 101, Timestamp: 1000, Data: "player logged in"},
			{ID: 2, Type: "achievement", PlayerID: 102, Timestamp: 2000, Data: "completed quest"},
			{ID: 3, Type: "purchase", PlayerID: 103, Timestamp: 3000, Data: "bought item"},
			{ID: 4, Type: "logout", PlayerID: 101, Timestamp: 4000, Data: "player logged out"},
		}

		err := AddToList(ctx, client, key, events...)
		require.NoError(t, err)

		provider := NewRedisListProvider[GameEvent](client, key)

		config := pagit.OffsetConfig{
			Page:     1,
			PageSize: 2,
		}

		result, err := dukdakit.PaginateOffset(ctx, provider, config)
		require.NoError(t, err)

		offsetResult, ok := result.(pagit.OffsetResult[GameEvent])
		require.True(t, ok)

		assert.Len(t, offsetResult.Data, 2)
		assert.Equal(t, "login", offsetResult.Data[0].Type)
		assert.Equal(t, "achievement", offsetResult.Data[1].Type)
		assert.Equal(t, int64(4), offsetResult.TotalCount)
	})
}
