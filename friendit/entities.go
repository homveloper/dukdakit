package friendit

import (
	"time"
)

// 기본 식별자 타입들
type UserID string
type RequestID string
type FriendshipID string
type BlockID string

// ============================================================================
// 사용자 정의 가능한 인터페이스 - 서비스 동작에 필요한 최소 필드만 정의
// ============================================================================

// UserEntity represents the minimum required fields for a user
// 사용자는 이 인터페이스를 구현하는 자신만의 User 구조체를 정의할 수 있습니다
type UserEntity interface {
	GetID() UserID
	GetStatus() string // online, offline, away, busy 등 자유롭게 정의
	GetLastSeen() *time.Time
	SetStatus(status string)
	SetLastSeen(t time.Time)
}

// FriendshipEntity represents the minimum required fields for a friendship
// 친구 관계 엔터티 - 사용자가 추가 필드를 자유롭게 확장 가능
type FriendshipEntity interface {
	GetID() FriendshipID
	GetUser1ID() UserID
	GetUser2ID() UserID
	GetStatus() string // active, pending, blocked 등 자유롭게 정의
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	SetUser1ID(userID UserID)
	SetUser2ID(userID UserID)
	SetStatus(status string)
	SetCreatedAt(t time.Time)
	SetUpdatedAt(t time.Time)
	SetSource(source string)
	SetMetadata(metadata map[string]any)
}

// FriendRequestEntity represents the minimum required fields for a friend request
// 친구 요청 엔터티 - 사용자가 추가 필드와 로직을 자유롭게 확장 가능
type FriendRequestEntity interface {
	GetID() RequestID
	GetSenderID() UserID
	GetReceiverID() UserID
	GetStatus() string // pending, accepted, rejected 등 자유롭게 정의
	GetCreatedAt() time.Time
	GetExpiresAt() *time.Time
	GetUpdatedAt() time.Time
	SetSenderID(senderID UserID)
	SetReceiverID(receiverID UserID)
	SetStatus(status string)
	SetCreatedAt(t time.Time)
	SetExpiresAt(t time.Time)
	SetUpdatedAt(t time.Time)
	SetMessage(message string)
	SetPriority(priority string)
	SetMetadata(metadata map[string]any)
	IsExpired() bool
}

// BlockRelationEntity represents the minimum required fields for a block relation
// 차단 관계 엔터티 - 사용자가 추가 정보를 자유롭게 확장 가능
type BlockRelationEntity interface {
	GetID() BlockID
	GetBlockerID() UserID
	GetBlockedID() UserID
	GetStatus() string
	GetCreatedAt() time.Time
	GetExpiresAt() *time.Time
	SetBlockerID(blockerID UserID)
	SetBlockedID(blockedID UserID)
	SetStatus(status string)
	SetCreatedAt(t time.Time)
	SetExpiresAt(t time.Time)
	SetReason(reason string)
	SetMetadata(metadata map[string]any)
}

// ============================================================================
// 기본 구현 예제 - 사용자가 참고할 수 있는 구조체들
// ============================================================================

// BasicUser - 기본적인 User 구현 예제 (사용자가 필요에 따라 필드 추가/수정 가능)
type BasicUser struct {
	ID       UserID     `json:"id" bson:"_id"`
	Status   string     `json:"status" bson:"status"`
	LastSeen *time.Time `json:"last_seen" bson:"last_seen"`
	// 사용자가 추가할 수 있는 선택적 필드들 예시:
	Username    string         `json:"username,omitempty" bson:"username,omitempty"`
	DisplayName string         `json:"display_name,omitempty" bson:"display_name,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty" bson:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt   time.Time      `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}

// BasicFriendship - 기본적인 Friendship 구현 예제
type BasicFriendship struct {
	ID        FriendshipID `json:"id" bson:"_id"`
	User1ID   UserID       `json:"user1_id" bson:"user1_id"`
	User2ID   UserID       `json:"user2_id" bson:"user2_id"`
	Status    string       `json:"status" bson:"status"`
	CreatedAt time.Time    `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" bson:"updated_at"`
	// 사용자가 추가할 수 있는 선택적 필드들:
	Metadata map[string]any `json:"metadata,omitempty" bson:"metadata,omitempty"`
	Source   string         `json:"source,omitempty" bson:"source,omitempty"` // 친구 추가 경로
}

// BasicFriendRequest - 기본적인 FriendRequest 구현 예제
type BasicFriendRequest struct {
	ID         RequestID  `json:"id" bson:"_id"`
	SenderID   UserID     `json:"sender_id" bson:"sender_id"`
	ReceiverID UserID     `json:"receiver_id" bson:"receiver_id"`
	Status     string     `json:"status" bson:"status"`
	CreatedAt  time.Time  `json:"created_at" bson:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at" bson:"expires_at"`
	UpdatedAt  time.Time  `json:"updated_at" bson:"updated_at"`
	// 사용자가 추가할 수 있는 선택적 필드들:
	Message  string         `json:"message,omitempty" bson:"message,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

// BasicBlockRelation - 기본적인 BlockRelation 구현 예제
type BasicBlockRelation struct {
	ID        BlockID    `json:"id" bson:"_id"`
	BlockerID UserID     `json:"blocker_id" bson:"blocker_id"`
	BlockedID UserID     `json:"blocked_id" bson:"blocked_id"`
	Status    string     `json:"status" bson:"status"`
	CreatedAt time.Time  `json:"created_at" bson:"created_at"`
	ExpiresAt *time.Time `json:"expires_at" bson:"expires_at"`
	// 사용자가 추가할 수 있는 선택적 필드들:
	Reason   string         `json:"reason,omitempty" bson:"reason,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

// ============================================================================
// 기본 구현체들의 인터페이스 메서드 구현
// ============================================================================

// BasicUser가 UserEntity 인터페이스를 구현
func (u *BasicUser) GetID() UserID           { return u.ID }
func (u *BasicUser) GetStatus() string       { return u.Status }
func (u *BasicUser) GetLastSeen() *time.Time { return u.LastSeen }
func (u *BasicUser) SetStatus(status string) {
	u.Status = status
	u.UpdatedAt = time.Now()
}
func (u *BasicUser) SetLastSeen(t time.Time) {
	u.LastSeen = &t
	u.UpdatedAt = time.Now()
}

// BasicFriendship이 FriendshipEntity 인터페이스를 구현
func (f *BasicFriendship) GetID() FriendshipID     { return f.ID }
func (f *BasicFriendship) GetUser1ID() UserID      { return f.User1ID }
func (f *BasicFriendship) GetUser2ID() UserID      { return f.User2ID }
func (f *BasicFriendship) GetStatus() string       { return f.Status }
func (f *BasicFriendship) GetCreatedAt() time.Time { return f.CreatedAt }
func (f *BasicFriendship) GetUpdatedAt() time.Time { return f.UpdatedAt }
func (f *BasicFriendship) SetUser1ID(userID UserID) { f.User1ID = userID }
func (f *BasicFriendship) SetUser2ID(userID UserID) { f.User2ID = userID }
func (f *BasicFriendship) SetStatus(status string) {
	f.Status = status
	f.UpdatedAt = time.Now()
}
func (f *BasicFriendship) SetCreatedAt(t time.Time) { f.CreatedAt = t }
func (f *BasicFriendship) SetUpdatedAt(t time.Time) { f.UpdatedAt = t }
func (f *BasicFriendship) SetSource(source string) { f.Source = source }
func (f *BasicFriendship) SetMetadata(metadata map[string]any) { f.Metadata = metadata }

// BasicFriendRequest가 FriendRequestEntity 인터페이스를 구현
func (fr *BasicFriendRequest) GetID() RequestID         { return fr.ID }
func (fr *BasicFriendRequest) GetSenderID() UserID      { return fr.SenderID }
func (fr *BasicFriendRequest) GetReceiverID() UserID    { return fr.ReceiverID }
func (fr *BasicFriendRequest) GetStatus() string        { return fr.Status }
func (fr *BasicFriendRequest) GetCreatedAt() time.Time  { return fr.CreatedAt }
func (fr *BasicFriendRequest) GetExpiresAt() *time.Time { return fr.ExpiresAt }
func (fr *BasicFriendRequest) GetUpdatedAt() time.Time  { return fr.UpdatedAt }
func (fr *BasicFriendRequest) SetSenderID(senderID UserID)    { fr.SenderID = senderID }
func (fr *BasicFriendRequest) SetReceiverID(receiverID UserID) { fr.ReceiverID = receiverID }
func (fr *BasicFriendRequest) SetStatus(status string) {
	fr.Status = status
	fr.UpdatedAt = time.Now()
}
func (fr *BasicFriendRequest) SetCreatedAt(t time.Time)  { fr.CreatedAt = t }
func (fr *BasicFriendRequest) SetExpiresAt(t time.Time)  { fr.ExpiresAt = &t }
func (fr *BasicFriendRequest) SetUpdatedAt(t time.Time) { fr.UpdatedAt = t }
func (fr *BasicFriendRequest) SetMessage(message string) { fr.Message = message }
func (fr *BasicFriendRequest) SetPriority(priority string) { 
	// Priority could be stored in Metadata or as a separate field
	if fr.Metadata == nil {
		fr.Metadata = make(map[string]any)
	}
	fr.Metadata["priority"] = priority
}
func (fr *BasicFriendRequest) SetMetadata(metadata map[string]any) { fr.Metadata = metadata }
func (fr *BasicFriendRequest) IsExpired() bool {
	if fr.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*fr.ExpiresAt)
}

// BasicBlockRelation이 BlockRelationEntity 인터페이스를 구현
func (br *BasicBlockRelation) GetID() BlockID           { return br.ID }
func (br *BasicBlockRelation) GetBlockerID() UserID     { return br.BlockerID }
func (br *BasicBlockRelation) GetBlockedID() UserID     { return br.BlockedID }
func (br *BasicBlockRelation) GetStatus() string        { return br.Status }
func (br *BasicBlockRelation) GetCreatedAt() time.Time  { return br.CreatedAt }
func (br *BasicBlockRelation) GetExpiresAt() *time.Time { return br.ExpiresAt }
func (br *BasicBlockRelation) SetBlockerID(blockerID UserID) { br.BlockerID = blockerID }
func (br *BasicBlockRelation) SetBlockedID(blockedID UserID) { br.BlockedID = blockedID }
func (br *BasicBlockRelation) SetStatus(status string)       { br.Status = status }
func (br *BasicBlockRelation) SetCreatedAt(t time.Time)      { br.CreatedAt = t }
func (br *BasicBlockRelation) SetExpiresAt(t time.Time)      { br.ExpiresAt = &t }
func (br *BasicBlockRelation) SetReason(reason string)       { br.Reason = reason }
func (br *BasicBlockRelation) SetMetadata(metadata map[string]any) { br.Metadata = metadata }

// ============================================================================
// 유틸리티 메서드들 (모든 구현체에서 사용 가능)
// ============================================================================

// GetOtherUserID returns the other user's ID in the friendship
func GetOtherUserID(friendship FriendshipEntity, userID UserID) UserID {
	if friendship.GetUser1ID() == userID {
		return friendship.GetUser2ID()
	}
	return friendship.GetUser1ID()
}

// InvolvesFriendship returns true if the friendship involves the specified user
func InvolvesFriendship(friendship FriendshipEntity, userID UserID) bool {
	return friendship.GetUser1ID() == userID || friendship.GetUser2ID() == userID
}

// InvolvesBlock returns true if the block relation involves the specified user
func InvolvesBlock(block BlockRelationEntity, userID UserID) bool {
	return block.GetBlockerID() == userID || block.GetBlockedID() == userID
}

// IsBlockedBy returns true if userID is blocked by blockerID
func IsBlockedBy(block BlockRelationEntity, userID UserID, blockerID UserID) bool {
	return block.GetBlockedID() == userID && block.GetBlockerID() == blockerID
}

// CanBeAccepted returns true if the request can be accepted
func CanBeAccepted(request FriendRequestEntity) bool {
	return request.GetStatus() == "pending" && !request.IsExpired()
}

// CanBeRejected returns true if the request can be rejected
func CanBeRejected(request FriendRequestEntity) bool {
	return request.GetStatus() == "pending"
}
