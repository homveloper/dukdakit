package friendit

import (
	"context"
	"fmt"
	"time"
)

// ============================================================================
// Block Service Interface and Implementation
// ============================================================================

// BlockService handles blocking operations
type BlockService[BR BlockRelationEntity] interface {
	// Block Management
	BlockUser(ctx context.Context, blockerID, blockedID UserID, options ...BlockOption) (BR, error)
	UnblockUser(ctx context.Context, blockerID, blockedID UserID) error
	IsBlocked(ctx context.Context, user1ID, user2ID UserID) (bool, error)
	
	// Query Operations
	GetBlockedUsers(ctx context.Context, userID UserID) ([]UserID, error)
	GetBlockingUsers(ctx context.Context, userID UserID) ([]UserID, error)
	GetBlockRelation(ctx context.Context, blockerID, blockedID UserID) (BR, error)
}

// ============================================================================
// Basic Implementation
// ============================================================================

// BasicBlockService implements BlockService
type BasicBlockService[BR BlockRelationEntity] struct {
	repo   BlockRelationRepository[BR]
	config *ServiceConfig
}

// NewBlockService creates a new block service
func NewBlockService[BR BlockRelationEntity](
	repo BlockRelationRepository[BR],
	config *ServiceConfig,
) BlockService[BR] {
	return &BasicBlockService[BR]{
		repo:   repo,
		config: config,
	}
}

// ============================================================================
// Service Method Implementations
// ============================================================================

// BlockUser implements BlockService.BlockUser with concurrency safety
func (s *BasicBlockService[BR]) BlockUser(ctx context.Context, blockerID, blockedID UserID, options ...BlockOption) (BR, error) {
	// Apply options
	config := &BlockConfig{}
	for _, opt := range options {
		opt(config)
	}
	
	// Validation
	if blockerID == blockedID {
		var empty BR
		return empty, fmt.Errorf("cannot block self")
	}
	
	// Create factory for atomic upsert operation
	factory := &createBlockRelationEntityFactory[BR]{
		blockerID: blockerID,
		blockedID: blockedID,
		config:    config,
		repo:      s.repo,
	}
	
	// Atomic upsert prevents race conditions and duplicate blocks
	result, err := s.repo.FindOneAndUpsert(ctx, blockerID, blockedID, factory)
	if err != nil {
		var empty BR
		return empty, fmt.Errorf("failed to block user: %w", err)
	}
	
	return result.Entity, nil
}

// UnblockUser implements BlockService.UnblockUser with concurrency safety
func (s *BasicBlockService[BR]) UnblockUser(ctx context.Context, blockerID, blockedID UserID) error {
	// Atomic delete operation prevents race conditions
	err := s.repo.DeleteByUsers(ctx, blockerID, blockedID)
	if err != nil {
		return fmt.Errorf("failed to unblock user: %w", err)
	}
	
	return nil
}

// IsBlocked implements BlockService.IsBlocked
func (s *BasicBlockService[BR]) IsBlocked(ctx context.Context, user1ID, user2ID UserID) (bool, error) {
	// Check both directions (user1 blocking user2 OR user2 blocking user1)
	_, err1 := s.repo.GetByUsers(ctx, user1ID, user2ID)
	_, err2 := s.repo.GetByUsers(ctx, user2ID, user1ID)
	
	// If either query succeeds, there's a block relationship
	return err1 == nil || err2 == nil, nil
}

// GetBlockedUsers implements BlockService.GetBlockedUsers
func (s *BasicBlockService[BR]) GetBlockedUsers(ctx context.Context, userID UserID) ([]UserID, error) {
	// Get all block relations where userID is the blocker
	blockRelations, err := s.repo.GetByBlockerID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get block relations: %w", err)
	}
	
	// Extract blocked user IDs
	blockedUserIDs := make([]UserID, len(blockRelations))
	for i, relation := range blockRelations {
		blockedUserIDs[i] = relation.GetBlockedID()
	}
	
	return blockedUserIDs, nil
}

// GetBlockingUsers implements BlockService.GetBlockingUsers
func (s *BasicBlockService[BR]) GetBlockingUsers(ctx context.Context, userID UserID) ([]UserID, error) {
	// Get all block relations where userID is the blocked user
	blockRelations, err := s.repo.GetByBlockedID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get block relations: %w", err)
	}
	
	// Extract blocker user IDs
	blockerUserIDs := make([]UserID, len(blockRelations))
	for i, relation := range blockRelations {
		blockerUserIDs[i] = relation.GetBlockerID()
	}
	
	return blockerUserIDs, nil
}

// GetBlockRelation implements BlockService.GetBlockRelation
func (s *BasicBlockService[BR]) GetBlockRelation(ctx context.Context, blockerID, blockedID UserID) (BR, error) {
	return s.repo.GetByUsers(ctx, blockerID, blockedID)
}

// ============================================================================
// Entity Factories for Atomic Operations
// ============================================================================

// createBlockRelationEntityFactory creates new block relation entities
type createBlockRelationEntityFactory[BR BlockRelationEntity] struct {
	blockerID UserID
	blockedID UserID
	config    *BlockConfig
	repo      BlockRelationRepository[BR]
}

func (f *createBlockRelationEntityFactory[BR]) CreateFn(ctx context.Context) (BR, error) {
	blockRelation := f.repo.NewEntity()
	blockRelation.SetBlockerID(f.blockerID)
	blockRelation.SetBlockedID(f.blockedID)
	blockRelation.SetStatus("active")
	blockRelation.SetCreatedAt(time.Now())
	
	if f.config.Reason != "" {
		blockRelation.SetReason(f.config.Reason)
	}
	if f.config.Duration != nil {
		blockRelation.SetExpiresAt(time.Now().Add(*f.config.Duration))
	}
	
	return blockRelation, nil
}

func (f *createBlockRelationEntityFactory[BR]) UpdateFn(ctx context.Context, existing BR) (BR, error) {
	// If block relation already exists, update reason if provided
	existing.SetStatus("active")
	
	if f.config.Reason != "" {
		existing.SetReason(f.config.Reason)
	}
	if f.config.Duration != nil {
		existing.SetExpiresAt(time.Now().Add(*f.config.Duration))
	}
	
	return existing, nil
}

// ============================================================================
// Fluent API Builder
// ============================================================================

// Block returns a block builder
func (s *BasicBlockService[BR]) Block() *BlockBuilder[BR] {
	return &BlockBuilder[BR]{service: s}
}

// BlockBuilder provides fluent interface for blocking operations
type BlockBuilder[BR BlockRelationEntity] struct {
	service   BlockService[BR]
	blockerID UserID
	blockedID UserID
	reason    string
}

// Blocker sets the blocker user ID
func (bb *BlockBuilder[BR]) Blocker(blockerID UserID) *BlockBuilder[BR] {
	bb.blockerID = blockerID
	return bb
}

// Blocked sets the blocked user ID
func (bb *BlockBuilder[BR]) Blocked(blockedID UserID) *BlockBuilder[BR] {
	bb.blockedID = blockedID
	return bb
}

// WithReason sets the blocking reason
func (bb *BlockBuilder[BR]) WithReason(reason string) *BlockBuilder[BR] {
	bb.reason = reason
	return bb
}

// Execute creates the block relation
func (bb *BlockBuilder[BR]) Execute(ctx context.Context) (BR, error) {
	// Build block options from builder state
	options := []BlockOption{}
	
	if bb.reason != "" {
		options = append(options, func(c *BlockConfig) { c.Reason = bb.reason })
	}
	
	return bb.service.BlockUser(ctx, bb.blockerID, bb.blockedID, options...)
}