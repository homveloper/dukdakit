package friendit

import (
	"context"
	"fmt"
	"time"
)

// ============================================================================
// 기본 엔터티들을 위한 팩토리 구현체들
// ============================================================================

// ============================================================================
// User Entity Factories
// ============================================================================

// BasicUserFactory BasicUser를 위한 팩토리
type BasicUserFactory struct{}

func (f *BasicUserFactory) NewUser(id UserID, status string) EntityFactory[*BasicUser] {
	return &createUserFactory{
		userID: id,
		status: status,
	}
}

func (f *BasicUserFactory) UpdateUserStatus(status string) EntityFactory[*BasicUser] {
	return &updateUserStatusFactory{
		newStatus: status,
	}
}

func (f *BasicUserFactory) UpdateUserLastSeen(t time.Time) EntityFactory[*BasicUser] {
	return &updateUserLastSeenFactory{
		lastSeen: t,
	}
}

// createUserFactory 새 사용자 생성 팩토리
type createUserFactory struct {
	userID UserID
	status string
}

func (f *createUserFactory) CreateFn(ctx context.Context) (*BasicUser, error) {
	now := time.Now()
	return &BasicUser{
		ID:        f.userID,
		Status:    f.status,
		LastSeen:  &now,
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  make(map[string]any),
	}, nil
}

func (f *createUserFactory) UpdateFn(ctx context.Context, existing *BasicUser) (*BasicUser, error) {
	// 이미 존재하면 상태만 업데이트
	existing.Status = f.status
	existing.UpdatedAt = time.Now()
	return existing, nil
}

// updateUserStatusFactory 사용자 상태 업데이트 팩토리
type updateUserStatusFactory struct {
	newStatus string
}

func (f *updateUserStatusFactory) CreateFn(ctx context.Context) (*BasicUser, error) {
	return nil, fmt.Errorf("cannot create new user with status update factory")
}

func (f *updateUserStatusFactory) UpdateFn(ctx context.Context, existing *BasicUser) (*BasicUser, error) {
	existing.Status = f.newStatus
	existing.LastSeen = &[]time.Time{time.Now()}[0]
	existing.UpdatedAt = time.Now()
	return existing, nil
}

// updateUserLastSeenFactory 사용자 마지막 접속 업데이트 팩토리
type updateUserLastSeenFactory struct {
	lastSeen time.Time
}

func (f *updateUserLastSeenFactory) CreateFn(ctx context.Context) (*BasicUser, error) {
	return nil, fmt.Errorf("cannot create new user with last seen update factory")
}

func (f *updateUserLastSeenFactory) UpdateFn(ctx context.Context, existing *BasicUser) (*BasicUser, error) {
	existing.LastSeen = &f.lastSeen
	existing.UpdatedAt = time.Now()
	return existing, nil
}

// ============================================================================
// Friendship Entity Factories
// ============================================================================

// BasicFriendshipFactory BasicFriendship을 위한 팩토리
type BasicFriendshipFactory struct{}

func (f *BasicFriendshipFactory) NewFriendship(user1ID, user2ID UserID, source string) EntityFactory[*BasicFriendship] {
	return &createFriendshipFactory{
		user1ID: user1ID,
		user2ID: user2ID,
		source:  source,
	}
}

func (f *BasicFriendshipFactory) UpdateFriendshipStatus(status string) EntityFactory[*BasicFriendship] {
	return &updateFriendshipStatusFactory{
		newStatus: status,
	}
}

func (f *BasicFriendshipFactory) UpdateFriendshipMetadata(metadata map[string]any) EntityFactory[*BasicFriendship] {
	return &updateFriendshipMetadataFactory{
		metadata: metadata,
	}
}

// createFriendshipFactory 새 친구관계 생성 팩토리
type createFriendshipFactory struct {
	user1ID UserID
	user2ID UserID
	source  string
}

func (f *createFriendshipFactory) CreateFn(ctx context.Context) (*BasicFriendship, error) {
	now := time.Now()
	return &BasicFriendship{
		ID:        FriendshipID(fmt.Sprintf("%s_%s_%d", f.user1ID, f.user2ID, now.UnixNano())),
		User1ID:   f.user1ID,
		User2ID:   f.user2ID,
		Status:    "active",
		Source:    f.source,
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  make(map[string]any),
	}, nil
}

func (f *createFriendshipFactory) UpdateFn(ctx context.Context, existing *BasicFriendship) (*BasicFriendship, error) {
	// 이미 존재하면 상태만 활성화
	if existing.Status != "active" {
		existing.Status = "active"
		existing.UpdatedAt = time.Now()
	}
	return existing, nil
}

// updateFriendshipStatusFactory 친구관계 상태 업데이트 팩토리
type updateFriendshipStatusFactory struct {
	newStatus string
}

func (f *updateFriendshipStatusFactory) CreateFn(ctx context.Context) (*BasicFriendship, error) {
	return nil, fmt.Errorf("cannot create new friendship with status update factory")
}

func (f *updateFriendshipStatusFactory) UpdateFn(ctx context.Context, existing *BasicFriendship) (*BasicFriendship, error) {
	existing.Status = f.newStatus
	existing.UpdatedAt = time.Now()
	return existing, nil
}

// updateFriendshipMetadataFactory 친구관계 메타데이터 업데이트 팩토리
type updateFriendshipMetadataFactory struct {
	metadata map[string]any
}

func (f *updateFriendshipMetadataFactory) CreateFn(ctx context.Context) (*BasicFriendship, error) {
	return nil, fmt.Errorf("cannot create new friendship with metadata update factory")
}

func (f *updateFriendshipMetadataFactory) UpdateFn(ctx context.Context, existing *BasicFriendship) (*BasicFriendship, error) {
	if existing.Metadata == nil {
		existing.Metadata = make(map[string]any)
	}
	for k, v := range f.metadata {
		existing.Metadata[k] = v
	}
	existing.UpdatedAt = time.Now()
	return existing, nil
}

// ============================================================================
// FriendRequest Entity Factories
// ============================================================================

// BasicFriendRequestFactory BasicFriendRequest를 위한 팩토리
type BasicFriendRequestFactory struct{}

func (f *BasicFriendRequestFactory) NewFriendRequest(senderID, receiverID UserID, message string) EntityFactory[*BasicFriendRequest] {
	return &createFriendRequestFactory{
		senderID:   senderID,
		receiverID: receiverID,
		message:    message,
	}
}

func (f *BasicFriendRequestFactory) AcceptFriendRequest() EntityFactory[*BasicFriendRequest] {
	return &acceptFriendRequestFactory{}
}

func (f *BasicFriendRequestFactory) RejectFriendRequest(reason string) EntityFactory[*BasicFriendRequest] {
	return &rejectFriendRequestFactory{
		reason: reason,
	}
}

func (f *BasicFriendRequestFactory) ExpireFriendRequest() EntityFactory[*BasicFriendRequest] {
	return &expireFriendRequestFactory{}
}

// createFriendRequestFactory 새 친구요청 생성 팩토리
type createFriendRequestFactory struct {
	senderID   UserID
	receiverID UserID
	message    string
}

func (f *createFriendRequestFactory) CreateFn(ctx context.Context) (*BasicFriendRequest, error) {
	now := time.Now()
	expiresAt := now.Add(7 * 24 * time.Hour) // 7일 후 만료
	
	return &BasicFriendRequest{
		ID:         RequestID(fmt.Sprintf("%s_%s_%d", f.senderID, f.receiverID, now.UnixNano())),
		SenderID:   f.senderID,
		ReceiverID: f.receiverID,
		Status:     "pending",
		Message:    f.message,
		CreatedAt:  now,
		UpdatedAt:  now,
		ExpiresAt:  &expiresAt,
		Metadata:   make(map[string]any),
	}, nil
}

func (f *createFriendRequestFactory) UpdateFn(ctx context.Context, existing *BasicFriendRequest) (*BasicFriendRequest, error) {
	// 이미 존재하고 pending이면 그대로 반환, 아니면 에러
	if existing.Status == "pending" && !existing.IsExpired() {
		return existing, nil
	}
	return nil, fmt.Errorf("cannot update existing friend request: status=%s, expired=%v", existing.Status, existing.IsExpired())
}

// acceptFriendRequestFactory 친구요청 수락 팩토리
type acceptFriendRequestFactory struct{}

func (f *acceptFriendRequestFactory) CreateFn(ctx context.Context) (*BasicFriendRequest, error) {
	return nil, fmt.Errorf("cannot create new friend request with accept factory")
}

func (f *acceptFriendRequestFactory) UpdateFn(ctx context.Context, existing *BasicFriendRequest) (*BasicFriendRequest, error) {
	if existing.Status != "pending" {
		return nil, fmt.Errorf("cannot accept non-pending request: status=%s", existing.Status)
	}
	if existing.IsExpired() {
		return nil, fmt.Errorf("cannot accept expired request")
	}
	
	existing.Status = "accepted"
	existing.UpdatedAt = time.Now()
	return existing, nil
}

// rejectFriendRequestFactory 친구요청 거절 팩토리
type rejectFriendRequestFactory struct {
	reason string
}

func (f *rejectFriendRequestFactory) CreateFn(ctx context.Context) (*BasicFriendRequest, error) {
	return nil, fmt.Errorf("cannot create new friend request with reject factory")
}

func (f *rejectFriendRequestFactory) UpdateFn(ctx context.Context, existing *BasicFriendRequest) (*BasicFriendRequest, error) {
	if existing.Status != "pending" {
		return nil, fmt.Errorf("cannot reject non-pending request: status=%s", existing.Status)
	}
	
	existing.Status = "rejected"
	existing.UpdatedAt = time.Now()
	if f.reason != "" {
		if existing.Metadata == nil {
			existing.Metadata = make(map[string]any)
		}
		existing.Metadata["rejection_reason"] = f.reason
	}
	return existing, nil
}

// expireFriendRequestFactory 친구요청 만료 팩토리
type expireFriendRequestFactory struct{}

func (f *expireFriendRequestFactory) CreateFn(ctx context.Context) (*BasicFriendRequest, error) {
	return nil, fmt.Errorf("cannot create new friend request with expire factory")
}

func (f *expireFriendRequestFactory) UpdateFn(ctx context.Context, existing *BasicFriendRequest) (*BasicFriendRequest, error) {
	existing.Status = "expired"
	existing.UpdatedAt = time.Now()
	return existing, nil
}

// ============================================================================
// BlockRelation Entity Factories
// ============================================================================

// BasicBlockRelationFactory BasicBlockRelation을 위한 팩토리
type BasicBlockRelationFactory struct{}

func (f *BasicBlockRelationFactory) NewBlockRelation(blockerID, blockedID UserID, reason string) EntityFactory[*BasicBlockRelation] {
	return &createBlockRelationFactory{
		blockerID: blockerID,
		blockedID: blockedID,
		reason:    reason,
	}
}

func (f *BasicBlockRelationFactory) UpdateBlockReason(reason string) EntityFactory[*BasicBlockRelation] {
	return &updateBlockReasonFactory{
		reason: reason,
	}
}

func (f *BasicBlockRelationFactory) SetBlockExpiration(expiresAt time.Time) EntityFactory[*BasicBlockRelation] {
	return &setBlockExpirationFactory{
		expiresAt: expiresAt,
	}
}

// createBlockRelationFactory 새 차단관계 생성 팩토리
type createBlockRelationFactory struct {
	blockerID UserID
	blockedID UserID
	reason    string
}

func (f *createBlockRelationFactory) CreateFn(ctx context.Context) (*BasicBlockRelation, error) {
	now := time.Now()
	return &BasicBlockRelation{
		ID:        BlockID(fmt.Sprintf("%s_%s_%d", f.blockerID, f.blockedID, now.UnixNano())),
		BlockerID: f.blockerID,
		BlockedID: f.blockedID,
		Status:    "active",
		Reason:    f.reason,
		CreatedAt: now,
		Metadata:  make(map[string]any),
	}, nil
}

func (f *createBlockRelationFactory) UpdateFn(ctx context.Context, existing *BasicBlockRelation) (*BasicBlockRelation, error) {
	// 이미 존재하면 이유를 업데이트
	if f.reason != "" && f.reason != existing.Reason {
		existing.Reason = f.reason
	}
	existing.Status = "active"
	return existing, nil
}

// updateBlockReasonFactory 차단 이유 업데이트 팩토리
type updateBlockReasonFactory struct {
	reason string
}

func (f *updateBlockReasonFactory) CreateFn(ctx context.Context) (*BasicBlockRelation, error) {
	return nil, fmt.Errorf("cannot create new block relation with reason update factory")
}

func (f *updateBlockReasonFactory) UpdateFn(ctx context.Context, existing *BasicBlockRelation) (*BasicBlockRelation, error) {
	existing.Reason = f.reason
	return existing, nil
}

// setBlockExpirationFactory 차단 만료 설정 팩토리
type setBlockExpirationFactory struct {
	expiresAt time.Time
}

func (f *setBlockExpirationFactory) CreateFn(ctx context.Context) (*BasicBlockRelation, error) {
	return nil, fmt.Errorf("cannot create new block relation with expiration factory")
}

func (f *setBlockExpirationFactory) UpdateFn(ctx context.Context, existing *BasicBlockRelation) (*BasicBlockRelation, error) {
	existing.ExpiresAt = &f.expiresAt
	return existing, nil
}