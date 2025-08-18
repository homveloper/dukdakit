package frienditpostgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/homveloper/dukdakit/friendit"
)

// ============================================================================
// PostgreSQL Adapter Example - 사용자가 참고할 수 있는 PostgreSQL 구현 예제
// ============================================================================

// PostgresAdapter provides PostgreSQL implementation of repositories
// 이것은 예제 구현입니다. 사용자가 자신의 요구사항에 맞게 수정할 수 있습니다.
type PostgresAdapter struct {
	db *sql.DB
}

// NewPostgresAdapter creates a new PostgreSQL adapter
func NewPostgresAdapter(connectionString string) (*PostgresAdapter, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	adapter := &PostgresAdapter{db: db}
	
	// Initialize tables
	if err := adapter.initTables(); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return adapter, nil
}

// Close closes the database connection
func (pa *PostgresAdapter) Close() error {
	return pa.db.Close()
}

// initTables creates necessary tables if they don't exist
func (pa *PostgresAdapter) initTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS friendit_users (
			id VARCHAR(255) PRIMARY KEY,
			status VARCHAR(50),
			last_seen TIMESTAMP,
			username VARCHAR(255),
			display_name VARCHAR(255),
			metadata JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS friendit_friendships (
			id VARCHAR(255) PRIMARY KEY,
			user1_id VARCHAR(255) NOT NULL,
			user2_id VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL,
			metadata JSONB,
			source VARCHAR(100),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user1_id) REFERENCES friendit_users(id) ON DELETE CASCADE,
			FOREIGN KEY (user2_id) REFERENCES friendit_users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS friendit_friend_requests (
			id VARCHAR(255) PRIMARY KEY,
			sender_id VARCHAR(255) NOT NULL,
			receiver_id VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL,
			message TEXT,
			metadata JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (sender_id) REFERENCES friendit_users(id) ON DELETE CASCADE,
			FOREIGN KEY (receiver_id) REFERENCES friendit_users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS friendit_block_relations (
			id VARCHAR(255) PRIMARY KEY,
			blocker_id VARCHAR(255) NOT NULL,
			blocked_id VARCHAR(255) NOT NULL,
			reason TEXT,
			metadata JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (blocker_id) REFERENCES friendit_users(id) ON DELETE CASCADE,
			FOREIGN KEY (blocked_id) REFERENCES friendit_users(id) ON DELETE CASCADE
		)`,
	}

	for _, query := range queries {
		if _, err := pa.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %s, error: %w", query, err)
		}
	}

	// Create indexes for better performance
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_users_status ON friendit_users(status)`,
		`CREATE INDEX IF NOT EXISTS idx_friendships_user1 ON friendit_friendships(user1_id)`,
		`CREATE INDEX IF NOT EXISTS idx_friendships_user2 ON friendit_friendships(user2_id)`,
		`CREATE INDEX IF NOT EXISTS idx_friend_requests_sender ON friendit_friend_requests(sender_id)`,
		`CREATE INDEX IF NOT EXISTS idx_friend_requests_receiver ON friendit_friend_requests(receiver_id)`,
		`CREATE INDEX IF NOT EXISTS idx_block_relations_blocker ON friendit_block_relations(blocker_id)`,
		`CREATE INDEX IF NOT EXISTS idx_block_relations_blocked ON friendit_block_relations(blocked_id)`,
	}

	for _, index := range indexes {
		if _, err := pa.db.Exec(index); err != nil {
			return fmt.Errorf("failed to create index: %s, error: %w", index, err)
		}
	}

	return nil
}

// ============================================================================
// User Repository Implementation
// ============================================================================

// PostgresUserRepository implements UserRepository for PostgreSQL with BasicUser
type PostgresUserRepository struct {
	db *sql.DB
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func (pa *PostgresAdapter) NewPostgresUserRepository() *PostgresUserRepository {
	return &PostgresUserRepository{db: pa.db}
}

// Create implements UserRepository.Create for BasicUser
func (r *PostgresUserRepository) Create(ctx context.Context, user friendit.BasicUser) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `INSERT INTO friendit_users (id, status, last_seen, username, display_name, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`
	
	_, err := r.db.ExecContext(ctx, query, 
		user.ID, user.Status, user.LastSeen, user.Username, 
		user.DisplayName, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	return nil
}

// GetByID implements UserRepository.GetByID for BasicUser
func (r *PostgresUserRepository) GetByID(ctx context.Context, id friendit.UserID) (friendit.BasicUser, error) {
	var user friendit.BasicUser
	
	query := `SELECT id, status, last_seen, username, display_name, created_at, updated_at 
			  FROM friendit_users WHERE id = $1`
	
	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(&user.ID, &user.Status, &user.LastSeen, &user.Username, 
		&user.DisplayName, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, fmt.Errorf("user not found: %s", id)
		}
		return user, fmt.Errorf("failed to get user: %w", err)
	}
	
	return user, nil
}

// Update implements UserRepository.Update for BasicUser
func (r *PostgresUserRepository) Update(ctx context.Context, user friendit.BasicUser) error {
	user.UpdatedAt = time.Now()
	
	query := `UPDATE friendit_users SET status = $2, last_seen = $3, username = $4, 
			  display_name = $5, updated_at = $6 WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, 
		user.ID, user.Status, user.LastSeen, user.Username, 
		user.DisplayName, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("user not found: %s", user.ID)
	}
	
	return nil
}

// Delete implements UserRepository.Delete for BasicUser
func (r *PostgresUserRepository) Delete(ctx context.Context, id friendit.UserID) error {
	query := `DELETE FROM friendit_users WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("user not found: %s", id)
	}
	
	return nil
}

// GetByIDs implements UserRepository.GetByIDs for BasicUser
func (r *PostgresUserRepository) GetByIDs(ctx context.Context, ids []friendit.UserID) ([]friendit.BasicUser, error) {
	if len(ids) == 0 {
		return []friendit.BasicUser{}, nil
	}
	
	// PostgreSQL의 ANY 연산자 사용
	query := `SELECT id, status, last_seen, username, display_name, created_at, updated_at 
			  FROM friendit_users WHERE id = ANY($1)`
	
	// UserID 슬라이스를 string 슬라이스로 변환
	stringIDs := make([]string, len(ids))
	for i, id := range ids {
		stringIDs[i] = string(id)
	}
	
	rows, err := r.db.QueryContext(ctx, query, stringIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()
	
	var users []friendit.BasicUser
	for rows.Next() {
		var user friendit.BasicUser
		err := rows.Scan(&user.ID, &user.Status, &user.LastSeen, &user.Username, 
			&user.DisplayName, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	
	return users, nil
}

// FindByStatus implements UserRepository.FindByStatus for BasicUser
func (r *PostgresUserRepository) FindByStatus(ctx context.Context, status string) ([]friendit.BasicUser, error) {
	query := `SELECT id, status, last_seen, username, display_name, created_at, updated_at 
			  FROM friendit_users WHERE status = $1`
	
	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to query users by status: %w", err)
	}
	defer rows.Close()
	
	var users []friendit.BasicUser
	for rows.Next() {
		var user friendit.BasicUser
		err := rows.Scan(&user.ID, &user.Status, &user.LastSeen, &user.Username, 
			&user.DisplayName, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	
	return users, nil
}

// Search implements UserRepository.Search for BasicUser
func (r *PostgresUserRepository) Search(ctx context.Context, query string, limit int) ([]friendit.BasicUser, error) {
	sqlQuery := `SELECT id, status, last_seen, username, display_name, created_at, updated_at 
				 FROM friendit_users 
				 WHERE username ILIKE $1 OR display_name ILIKE $1 
				 LIMIT $2`
	
	searchPattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, sqlQuery, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	defer rows.Close()
	
	var users []friendit.BasicUser
	for rows.Next() {
		var user friendit.BasicUser
		err := rows.Scan(&user.ID, &user.Status, &user.LastSeen, &user.Username, 
			&user.DisplayName, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	
	return users, nil
}

// UpdateStatus implements UserRepository.UpdateStatus for BasicUser
func (r *PostgresUserRepository) UpdateStatus(ctx context.Context, id friendit.UserID, status string) error {
	query := `UPDATE friendit_users SET status = $2, last_seen = $3, updated_at = $4 WHERE id = $1`
	
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, id, status, now, now)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("user not found: %s", id)
	}
	
	return nil
}

// GetOnlineUsers implements UserRepository.GetOnlineUsers for BasicUser
func (r *PostgresUserRepository) GetOnlineUsers(ctx context.Context) ([]friendit.BasicUser, error) {
	return r.FindByStatus(ctx, "online")
}