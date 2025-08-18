package friendit

// ============================================================================
// 도메인별 서비스 분리 완료
// ============================================================================

// 이 파일은 향후 제거될 예정입니다.
// 각 도메인별 서비스는 다음 파일들에서 정의되었습니다:
//
// - friend_request_service.go: FriendRequestService 인터페이스 및 구현체
// - friendship_service.go: FriendshipService 인터페이스 및 구현체  
// - block_service.go: BlockService 인터페이스 및 구현체
// - user_service.go: UserService 인터페이스 및 구현체
// - service_manager.go: ServiceManager 및 설정
// - options.go: 옵션 타입들 및 설정 구조체
//
// 사용법:
//   // 개별 서비스 사용
//   requestService := NewFriendRequestService(repo, config)
//   
//   // 통합 매니저 사용
//   manager := NewServiceManager(repos, options...)
//   result := manager.Request().From("user1").To("user2").Send(ctx)