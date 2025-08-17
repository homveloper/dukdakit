package pagitpostgresql

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/homveloper/dukdakit"
	"github.com/homveloper/dukdakit/internal/pagit"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create tables for testing
	createTables := []string{
		`CREATE TABLE players (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			score INTEGER NOT NULL,
			level INTEGER NOT NULL,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE items (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			rarity INTEGER NOT NULL,
			price INTEGER NOT NULL
		)`,
		`CREATE TABLE game_events (
			id INTEGER PRIMARY KEY,
			type TEXT NOT NULL,
			player_id INTEGER NOT NULL,
			timestamp INTEGER NOT NULL,
			data TEXT NOT NULL
		)`,
	}

	for _, query := range createTables {
		_, err := db.Exec(query)
		require.NoError(t, err)
	}

	return db
}

func seedPlayers(t *testing.T, db *sql.DB) {
	players := []Player{
		{ID: 1, Name: "TopPlayer", Score: 1000, Level: 10, CreatedAt: "2024-01-01"},
		{ID: 2, Name: "SecondPlayer", Score: 950, Level: 9, CreatedAt: "2024-01-02"},
		{ID: 3, Name: "ThirdPlayer", Score: 900, Level: 8, CreatedAt: "2024-01-03"},
		{ID: 4, Name: "FourthPlayer", Score: 850, Level: 7, CreatedAt: "2024-01-04"},
		{ID: 5, Name: "FifthPlayer", Score: 800, Level: 6, CreatedAt: "2024-01-05"},
	}

	for _, p := range players {
		_, err := db.Exec(
			"INSERT INTO players (id, name, score, level, created_at) VALUES (?, ?, ?, ?, ?)",
			p.ID, p.Name, p.Score, p.Level, p.CreatedAt,
		)
		require.NoError(t, err)
	}
}

func seedItems(t *testing.T, db *sql.DB) {
	items := []Item{
		{ID: 1, Name: "Iron Sword", Type: "weapon", Rarity: 1, Price: 100},
		{ID: 2, Name: "Magic Shield", Type: "armor", Rarity: 2, Price: 200},
		{ID: 3, Name: "Health Potion", Type: "consumable", Rarity: 1, Price: 50},
		{ID: 4, Name: "Steel Armor", Type: "armor", Rarity: 3, Price: 500},
		{ID: 5, Name: "Fire Scroll", Type: "consumable", Rarity: 2, Price: 150},
	}

	for _, i := range items {
		_, err := db.Exec(
			"INSERT INTO items (id, name, type, rarity, price) VALUES (?, ?, ?, ?, ?)",
			i.ID, i.Name, i.Type, i.Rarity, i.Price,
		)
		require.NoError(t, err)
	}
}

func seedGameEvents(t *testing.T, db *sql.DB) {
	events := []GameEvent{
		{ID: 1, Type: "login", PlayerID: 101, Timestamp: 1000, Data: "player logged in"},
		{ID: 2, Type: "achievement", PlayerID: 102, Timestamp: 2000, Data: "completed quest"},
		{ID: 3, Type: "purchase", PlayerID: 103, Timestamp: 3000, Data: "bought item"},
		{ID: 4, Type: "logout", PlayerID: 101, Timestamp: 4000, Data: "player logged out"},
		{ID: 5, Type: "battle", PlayerID: 104, Timestamp: 5000, Data: "won battle"},
	}

	for _, e := range events {
		_, err := db.Exec(
			"INSERT INTO game_events (id, type, player_id, timestamp, data) VALUES (?, ?, ?, ?, ?)",
			e.ID, e.Type, e.PlayerID, e.Timestamp, e.Data,
		)
		require.NoError(t, err)
	}
}

func TestPostgreSQLOffsetProvider_BasicPagination(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	seedPlayers(t, db)

	ctx := context.Background()

	config := PostgreSQLOffsetConfig[Player]{
		DB:        db,
		TableName: "players",
		Columns:   []string{"id", "name", "score", "level", "created_at"},
		OrderBy:   "score DESC",
		Scanner:   ScanPlayer,
	}

	provider := NewPostgreSQLOffsetProvider(config)

	t.Run("FirstPage", func(t *testing.T) {
		pageConfig := pagit.OffsetConfig{
			Page:     1,
			PageSize: 2,
		}

		result, err := dukdakit.PaginateOffset(ctx, provider, pageConfig)
		require.NoError(t, err)

		offsetResult, ok := result.(pagit.OffsetResult[Player])
		require.True(t, ok)

		assert.Len(t, offsetResult.Data, 2)
		assert.Equal(t, "TopPlayer", offsetResult.Data[0].Name)     // Highest score first
		assert.Equal(t, "SecondPlayer", offsetResult.Data[1].Name) // Second highest
		assert.Equal(t, int64(5), offsetResult.TotalCount)
		assert.Equal(t, 1, offsetResult.Page)
		assert.Equal(t, 3, offsetResult.TotalPages)
		assert.True(t, offsetResult.HasNext)
		assert.False(t, offsetResult.HasPrev)
	})

	t.Run("SecondPage", func(t *testing.T) {
		pageConfig := pagit.OffsetConfig{
			Page:     2,
			PageSize: 2,
		}

		result, err := dukdakit.PaginateOffset(ctx, provider, pageConfig)
		require.NoError(t, err)

		offsetResult, ok := result.(pagit.OffsetResult[Player])
		require.True(t, ok)

		assert.Len(t, offsetResult.Data, 2)
		assert.Equal(t, "ThirdPlayer", offsetResult.Data[0].Name)
		assert.Equal(t, "FourthPlayer", offsetResult.Data[1].Name)
		assert.True(t, offsetResult.HasNext)
		assert.True(t, offsetResult.HasPrev)
	})

	t.Run("WithWhereClause", func(t *testing.T) {
		configWithWhere := PostgreSQLOffsetConfig[Player]{
			DB:        db,
			TableName: "players",
			Columns:   []string{"id", "name", "score", "level", "created_at"},
			OrderBy:   "score DESC",
			Where:     "level >= $1",
			Args:      []interface{}{8},
			Scanner:   ScanPlayer,
		}

		providerWithWhere := NewPostgreSQLOffsetProvider(configWithWhere)

		pageConfig := pagit.OffsetConfig{
			Page:     1,
			PageSize: 10,
		}

		result, err := dukdakit.PaginateOffset(ctx, providerWithWhere, pageConfig)
		require.NoError(t, err)

		offsetResult, ok := result.(pagit.OffsetResult[Player])
		require.True(t, ok)

		assert.Len(t, offsetResult.Data, 3) // Only players with level >= 8
		assert.Equal(t, int64(3), offsetResult.TotalCount)
	})
}

func TestPostgreSQLCursorProvider_CursorPagination(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	seedPlayers(t, db)

	ctx := context.Background()

	config := PostgreSQLCursorConfig[Player, int]{
		DB:        db,
		TableName: "players",
		Columns:   []string{"id", "name", "score", "level", "created_at"},
		CursorCol: "score",
		OrderBy:   "score ASC", // Ascending order for cursor testing
		Scanner:   ScanPlayer,
		Extractor: func(p Player) int { return p.Score },
	}

	provider := NewPostgreSQLCursorProvider(config)

	t.Run("FromStart", func(t *testing.T) {
		cursorConfig := pagit.CursorConfig[int]{
			PageSize:  2,
			Cursor:    nil, // Start from beginning
			Direction: pagit.CursorForward,
		}

		result, err := dukdakit.PaginateCursor(ctx, provider, cursorConfig, func(p Player) int { return p.Score })
		require.NoError(t, err)

		assert.Len(t, result.Data, 2)
		assert.Equal(t, "FifthPlayer", result.Data[0].Name)  // Lowest score first (800)
		assert.Equal(t, "FourthPlayer", result.Data[1].Name) // Next lowest (850)
		assert.True(t, result.HasNext)
		assert.False(t, result.HasPrev)
		assert.NotNil(t, result.NextCursor)
		assert.Equal(t, 850, *result.NextCursor)
	})

	t.Run("ForwardFromCursor", func(t *testing.T) {
		cursor := 850
		cursorConfig := pagit.CursorConfig[int]{
			PageSize:  2,
			Cursor:    &cursor,
			Direction: pagit.CursorForward,
		}

		result, err := dukdakit.PaginateCursor(ctx, provider, cursorConfig, func(p Player) int { return p.Score })
		require.NoError(t, err)

		assert.Len(t, result.Data, 2)
		assert.Equal(t, "ThirdPlayer", result.Data[0].Name)  // Score 900
		assert.Equal(t, "SecondPlayer", result.Data[1].Name) // Score 950
		assert.True(t, result.HasNext)
		assert.True(t, result.HasPrev)
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

func TestPostgreSQLProvider_GameScenarios(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	seedPlayers(t, db)
	seedItems(t, db)
	seedGameEvents(t, db)

	ctx := context.Background()

	t.Run("PlayerLeaderboard", func(t *testing.T) {
		config := PostgreSQLCursorConfig[Player, int]{
			DB:        db,
			TableName: "players",
			Columns:   []string{"id", "name", "score", "level", "created_at"},
			CursorCol: "score",
			OrderBy:   "score DESC, id ASC", // Highest scores first
			Scanner:   ScanPlayer,
			Extractor: func(p Player) int { return p.Score },
		}

		provider := NewPostgreSQLCursorProvider(config)

		// Get top 3 players
		cursorConfig := pagit.CursorConfig[int]{
			PageSize:  3,
			Cursor:    nil,
			Direction: pagit.CursorForward,
		}

		result, err := dukdakit.PaginateCursor(ctx, provider, cursorConfig, func(p Player) int { return p.Score })
		require.NoError(t, err)

		assert.Len(t, result.Data, 3)
		assert.Equal(t, "TopPlayer", result.Data[0].Name)    // Score 1000
		assert.Equal(t, "SecondPlayer", result.Data[1].Name) // Score 950
		assert.Equal(t, "ThirdPlayer", result.Data[2].Name)  // Score 900
	})

	t.Run("ItemCatalog", func(t *testing.T) {
		config := PostgreSQLOffsetConfig[Item]{
			DB:        db,
			TableName: "items",
			Columns:   []string{"id", "name", "type", "rarity", "price"},
			OrderBy:   "rarity DESC, price ASC", // Rare items first, then by price
			Scanner:   ScanItem,
		}

		provider := NewPostgreSQLOffsetProvider(config)

		pageConfig := pagit.OffsetConfig{
			Page:     1,
			PageSize: 3,
		}

		result, err := dukdakit.PaginateOffset(ctx, provider, pageConfig)
		require.NoError(t, err)

		offsetResult, ok := result.(pagit.OffsetResult[Item])
		require.True(t, ok)

		assert.Len(t, offsetResult.Data, 3)
		assert.Equal(t, "Steel Armor", offsetResult.Data[0].Name) // Rarity 3
		assert.Equal(t, int64(5), offsetResult.TotalCount)
	})

	t.Run("WeaponsByRarity", func(t *testing.T) {
		config := PostgreSQLOffsetConfig[Item]{
			DB:        db,
			TableName: "items",
			Columns:   []string{"id", "name", "type", "rarity", "price"},
			OrderBy:   "rarity DESC",
			Where:     "type = $1 AND rarity >= $2",
			Args:      []interface{}{"weapon", 1},
			Scanner:   ScanItem,
		}

		provider := NewPostgreSQLOffsetProvider(config)

		pageConfig := pagit.OffsetConfig{
			Page:     1,
			PageSize: 10,
		}

		result, err := dukdakit.PaginateOffset(ctx, provider, pageConfig)
		require.NoError(t, err)

		offsetResult, ok := result.(pagit.OffsetResult[Item])
		require.True(t, ok)

		assert.Len(t, offsetResult.Data, 1) // Only weapons
		assert.Equal(t, "Iron Sword", offsetResult.Data[0].Name)
		assert.Equal(t, int64(1), offsetResult.TotalCount)
	})

	t.Run("RecentGameEvents", func(t *testing.T) {
		config := PostgreSQLOffsetConfig[GameEvent]{
			DB:        db,
			TableName: "game_events",
			Columns:   []string{"id", "type", "player_id", "timestamp", "data"},
			OrderBy:   "timestamp DESC", // Most recent first
			Scanner:   ScanGameEvent,
		}

		provider := NewPostgreSQLOffsetProvider(config)

		pageConfig := pagit.OffsetConfig{
			Page:     1,
			PageSize: 3,
		}

		result, err := dukdakit.PaginateOffset(ctx, provider, pageConfig)
		require.NoError(t, err)

		offsetResult, ok := result.(pagit.OffsetResult[GameEvent])
		require.True(t, ok)

		assert.Len(t, offsetResult.Data, 3)
		assert.Equal(t, "battle", offsetResult.Data[0].Type)      // Most recent (timestamp 5000)
		assert.Equal(t, "logout", offsetResult.Data[1].Type)     // Second most recent (4000)
		assert.Equal(t, "purchase", offsetResult.Data[2].Type)   // Third most recent (3000)
		assert.Equal(t, int64(5), offsetResult.TotalCount)
	})

	t.Run("PlayerSpecificEvents", func(t *testing.T) {
		config := PostgreSQLCursorConfig[GameEvent, int64]{
			DB:        db,
			TableName: "game_events",
			Columns:   []string{"id", "type", "player_id", "timestamp", "data"},
			CursorCol: "timestamp",
			OrderBy:   "timestamp ASC",
			Where:     "player_id = $1",
			Args:      []interface{}{101},
			Scanner:   ScanGameEvent,
			Extractor: func(e GameEvent) int64 { return e.Timestamp },
		}

		provider := NewPostgreSQLCursorProvider(config)

		cursorConfig := pagit.CursorConfig[int64]{
			PageSize:  10,
			Cursor:    nil,
			Direction: pagit.CursorForward,
		}

		result, err := dukdakit.PaginateCursor(ctx, provider, cursorConfig, func(e GameEvent) int64 { return e.Timestamp })
		require.NoError(t, err)

		assert.Len(t, result.Data, 2) // Only events for player 101
		assert.Equal(t, "login", result.Data[0].Type)
		assert.Equal(t, "logout", result.Data[1].Type)
	})
}

func TestPostgreSQLProvider_ErrorHandling(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	t.Run("NonExistentTable", func(t *testing.T) {
		config := PostgreSQLOffsetConfig[Player]{
			DB:        db,
			TableName: "nonexistent_table",
			Scanner:   ScanPlayer,
		}

		provider := NewPostgreSQLOffsetProvider(config)

		pageConfig := pagit.OffsetConfig{
			Page:     1,
			PageSize: 10,
		}

		_, err := dukdakit.PaginateOffset(ctx, provider, pageConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute query")
	})

	t.Run("InvalidSQL", func(t *testing.T) {
		config := PostgreSQLOffsetConfig[Player]{
			DB:        db,
			TableName: "players",
			Where:     "invalid_column = $1", // Non-existent column
			Args:      []interface{}{"test"},
			Scanner:   ScanPlayer,
		}

		provider := NewPostgreSQLOffsetProvider(config)

		pageConfig := pagit.OffsetConfig{
			Page:     1,
			PageSize: 10,
		}

		_, err := dukdakit.PaginateOffset(ctx, provider, pageConfig)
		assert.Error(t, err)
	})
}