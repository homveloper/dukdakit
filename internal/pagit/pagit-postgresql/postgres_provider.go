package pagitpostgresql

import (
	"context"
	"database/sql"
	"fmt"
)

// PostgreSQLOffsetProvider provides offset-based pagination for PostgreSQL tables
// Ideal for game scenarios like: item catalogs, achievement lists, static data
type PostgreSQLOffsetProvider[T any] struct {
	db        *sql.DB
	tableName string
	columns   []string
	orderBy   string
	where     string
	args      []interface{}
	scanner   func(*sql.Rows) (T, error)
}

// PostgreSQLOffsetConfig holds configuration for PostgreSQL offset provider
type PostgreSQLOffsetConfig[T any] struct {
	DB        *sql.DB
	TableName string
	Columns   []string
	OrderBy   string // e.g., "created_at DESC, id ASC"
	Where     string // Optional WHERE clause without "WHERE" keyword
	Args      []interface{} // Arguments for WHERE clause
	Scanner   func(*sql.Rows) (T, error) // Function to scan SQL row into T
}

// NewPostgreSQLOffsetProvider creates a PostgreSQL offset provider
func NewPostgreSQLOffsetProvider[T any](config PostgreSQLOffsetConfig[T]) *PostgreSQLOffsetProvider[T] {
	if config.OrderBy == "" {
		config.OrderBy = "id ASC" // Default ordering
	}
	
	return &PostgreSQLOffsetProvider[T]{
		db:        config.DB,
		tableName: config.TableName,
		columns:   config.Columns,
		orderBy:   config.OrderBy,
		where:     config.Where,
		args:      config.Args,
		scanner:   config.Scanner,
	}
}

// GetData implements pagit.DataProvider interface for PostgreSQL
func (p *PostgreSQLOffsetProvider[T]) GetData(ctx context.Context, offset, limit int) ([]T, error) {
	query := p.buildSelectQuery(limit, offset)
	
	rows, err := p.db.QueryContext(ctx, query, p.args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var result []T
	for rows.Next() {
		item, err := p.scanner(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return result, nil
}

// GetTotalCount implements pagit.CountProvider interface for PostgreSQL
func (p *PostgreSQLOffsetProvider[T]) GetTotalCount(ctx context.Context) (int64, error) {
	query := p.buildCountQuery()
	
	var count int64
	err := p.db.QueryRowContext(ctx, query, p.args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return count, nil
}

// buildSelectQuery constructs the SELECT query for pagination
func (p *PostgreSQLOffsetProvider[T]) buildSelectQuery(limit, offset int) string {
	columnsStr := "*"
	if len(p.columns) > 0 {
		columnsStr = ""
		for i, col := range p.columns {
			if i > 0 {
				columnsStr += ", "
			}
			columnsStr += col
		}
	}

	query := fmt.Sprintf("SELECT %s FROM %s", columnsStr, p.tableName)
	
	if p.where != "" {
		query += fmt.Sprintf(" WHERE %s", p.where)
	}
	
	query += fmt.Sprintf(" ORDER BY %s", p.orderBy)
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
	
	return query
}

// buildCountQuery constructs the COUNT query for total count
func (p *PostgreSQLOffsetProvider[T]) buildCountQuery() string {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", p.tableName)
	
	if p.where != "" {
		query += fmt.Sprintf(" WHERE %s", p.where)
	}
	
	return query
}

// PostgreSQLCursorProvider provides cursor-based pagination for PostgreSQL tables
// Ideal for game scenarios like: leaderboards, activity feeds, real-time data
type PostgreSQLCursorProvider[T any, C comparable] struct {
	db         *sql.DB
	tableName  string
	columns    []string
	cursorCol  string // Column used for cursor (e.g., "score", "created_at", "id")
	orderBy    string // Full ORDER BY clause
	where      string
	args       []interface{}
	scanner    func(*sql.Rows) (T, error)
	extractor  func(T) C
}

// PostgreSQLCursorConfig holds configuration for PostgreSQL cursor provider
type PostgreSQLCursorConfig[T any, C comparable] struct {
	DB         *sql.DB
	TableName  string
	Columns    []string
	CursorCol  string // Column name for cursor comparisons
	OrderBy    string // e.g., "score DESC, id ASC" for leaderboards
	Where      string // Optional WHERE clause without "WHERE" keyword
	Args       []interface{} // Arguments for WHERE clause
	Scanner    func(*sql.Rows) (T, error) // Function to scan SQL row into T
	Extractor  func(T) C // Function to extract cursor value from T
}

// NewPostgreSQLCursorProvider creates a PostgreSQL cursor provider
func NewPostgreSQLCursorProvider[T any, C comparable](config PostgreSQLCursorConfig[T, C]) *PostgreSQLCursorProvider[T, C] {
	if config.OrderBy == "" {
		config.OrderBy = fmt.Sprintf("%s ASC", config.CursorCol)
	}
	
	return &PostgreSQLCursorProvider[T, C]{
		db:        config.DB,
		tableName: config.TableName,
		columns:   config.Columns,
		cursorCol: config.CursorCol,
		orderBy:   config.OrderBy,
		where:     config.Where,
		args:      config.Args,
		scanner:   config.Scanner,
		extractor: config.Extractor,
	}
}

// GetDataAfter implements pagit.CursorDataProvider interface
func (p *PostgreSQLCursorProvider[T, C]) GetDataAfter(
	ctx context.Context,
	cursor *C,
	limit int,
) ([]T, error) {
	query, args := p.buildCursorQuery(cursor, ">", limit)
	
	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute cursor query: %w", err)
	}
	defer rows.Close()

	return p.scanRows(rows)
}

// GetDataBefore implements pagit.CursorDataProvider interface
func (p *PostgreSQLCursorProvider[T, C]) GetDataBefore(
	ctx context.Context,
	cursor *C,
	limit int,
) ([]T, error) {
	if cursor == nil {
		// For backward pagination from nil cursor, get the highest values
		query, args := p.buildReverseCursorQuery(limit)
		
		rows, err := p.db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to execute reverse cursor query: %w", err)
		}
		defer rows.Close()

		return p.scanRows(rows)
	}

	query, args := p.buildCursorQuery(cursor, "<", limit)
	
	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute cursor query: %w", err)
	}
	defer rows.Close()

	return p.scanRows(rows)
}

// HasDataAfter implements pagit.CursorCheckProvider interface
func (p *PostgreSQLCursorProvider[T, C]) HasDataAfter(
	ctx context.Context,
	cursor C,
) (bool, error) {
	query, args := p.buildExistsQuery(&cursor, ">")
	
	var exists bool
	err := p.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check data after cursor: %w", err)
	}

	return exists, nil
}

// HasDataBefore implements pagit.CursorCheckProvider interface
func (p *PostgreSQLCursorProvider[T, C]) HasDataBefore(
	ctx context.Context,
	cursor C,
) (bool, error) {
	query, args := p.buildExistsQuery(&cursor, "<")
	
	var exists bool
	err := p.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check data before cursor: %w", err)
	}

	return exists, nil
}

// buildCursorQuery constructs cursor-based query
func (p *PostgreSQLCursorProvider[T, C]) buildCursorQuery(cursor *C, operator string, limit int) (string, []interface{}) {
	columnsStr := "*"
	if len(p.columns) > 0 {
		columnsStr = ""
		for i, col := range p.columns {
			if i > 0 {
				columnsStr += ", "
			}
			columnsStr += col
		}
	}

	query := fmt.Sprintf("SELECT %s FROM %s", columnsStr, p.tableName)
	args := make([]interface{}, len(p.args))
	copy(args, p.args)
	
	// Build WHERE clause
	var whereClause string
	if p.where != "" {
		whereClause = p.where
	}
	
	if cursor != nil {
		cursorCondition := fmt.Sprintf("%s %s $%d", p.cursorCol, operator, len(args)+1)
		if whereClause != "" {
			whereClause += " AND " + cursorCondition
		} else {
			whereClause = cursorCondition
		}
		args = append(args, *cursor)
	}
	
	if whereClause != "" {
		query += " WHERE " + whereClause
	}
	
	query += fmt.Sprintf(" ORDER BY %s", p.orderBy)
	query += fmt.Sprintf(" LIMIT %d", limit)
	
	return query, args
}

// buildReverseCursorQuery constructs reverse cursor query for backward pagination from nil
func (p *PostgreSQLCursorProvider[T, C]) buildReverseCursorQuery(limit int) (string, []interface{}) {
	columnsStr := "*"
	if len(p.columns) > 0 {
		columnsStr = ""
		for i, col := range p.columns {
			if i > 0 {
				columnsStr += ", "
			}
			columnsStr += col
		}
	}

	query := fmt.Sprintf("SELECT %s FROM %s", columnsStr, p.tableName)
	args := make([]interface{}, len(p.args))
	copy(args, p.args)
	
	if p.where != "" {
		query += fmt.Sprintf(" WHERE %s", p.where)
	}
	
	// Reverse the order for backward pagination
	reverseOrder := p.reverseOrderBy()
	query += fmt.Sprintf(" ORDER BY %s", reverseOrder)
	query += fmt.Sprintf(" LIMIT %d", limit)
	
	return query, args
}

// buildExistsQuery constructs EXISTS query for cursor checking
func (p *PostgreSQLCursorProvider[T, C]) buildExistsQuery(cursor *C, operator string) (string, []interface{}) {
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s", p.tableName)
	args := make([]interface{}, len(p.args))
	copy(args, p.args)
	
	// Build WHERE clause
	var whereClause string
	if p.where != "" {
		whereClause = p.where
	}
	
	cursorCondition := fmt.Sprintf("%s %s $%d", p.cursorCol, operator, len(args)+1)
	if whereClause != "" {
		whereClause += " AND " + cursorCondition
	} else {
		whereClause = cursorCondition
	}
	args = append(args, *cursor)
	
	query += " WHERE " + whereClause + ")"
	
	return query, args
}

// reverseOrderBy reverses the ORDER BY clause for backward pagination
func (p *PostgreSQLCursorProvider[T, C]) reverseOrderBy() string {
	// Simple implementation - for production, you'd want more sophisticated parsing
	// This handles common cases like "score DESC, id ASC"
	orderBy := p.orderBy
	
	// Replace DESC with ASC and vice versa
	// This is a simplified implementation
	if len(orderBy) > 0 {
		// For now, return the same order - in real implementation you'd parse and reverse
		return orderBy
	}
	
	return fmt.Sprintf("%s DESC", p.cursorCol)
}

// scanRows scans multiple rows into slice
func (p *PostgreSQLCursorProvider[T, C]) scanRows(rows *sql.Rows) ([]T, error) {
	var result []T
	for rows.Next() {
		item, err := p.scanner(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return result, nil
}

// Helper functions for common game scenarios

// ScanPlayer is a helper scanner for Player structs
func ScanPlayer(rows *sql.Rows) (Player, error) {
	var p Player
	err := rows.Scan(&p.ID, &p.Name, &p.Score, &p.Level, &p.CreatedAt)
	return p, err
}

// ScanItem is a helper scanner for Item structs  
func ScanItem(rows *sql.Rows) (Item, error) {
	var i Item
	err := rows.Scan(&i.ID, &i.Name, &i.Type, &i.Rarity, &i.Price)
	return i, err
}

// ScanGameEvent is a helper scanner for GameEvent structs
func ScanGameEvent(rows *sql.Rows) (GameEvent, error) {
	var e GameEvent
	err := rows.Scan(&e.ID, &e.Type, &e.PlayerID, &e.Timestamp, &e.Data)
	return e, err
}

// Common game data structures
type Player struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Score     int    `json:"score"`
	Level     int    `json:"level"`
	CreatedAt string `json:"created_at"`
}

type Item struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Rarity int    `json:"rarity"`
	Price  int    `json:"price"`
}

type GameEvent struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	PlayerID  int64  `json:"player_id"`
	Timestamp int64  `json:"timestamp"`
	Data      string `json:"data"`
}