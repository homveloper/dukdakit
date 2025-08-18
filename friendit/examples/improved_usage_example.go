package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/homveloper/dukdakit/friendit"
	memory "github.com/homveloper/dukdakit/friendit/adapters/memory"
)

// ============================================================================
// 개선된 Friendit 사용 예제 - 통합된 동시성 안전 서비스
// ============================================================================

func main() {
	ctx := context.Background()

	// 동시성 안전 메모리 기반 저장소 생성
	userRepo := memory.NewConcurrentMemoryUserRepository()

	// 기본 팩토리 생성 (사용자가 필요에 따라 커스터마이징 가능)

	// 서비스 설정
	config := &friendit.ServiceConfig{
		MaxFriends:        500,
		AllowSelfRequests: false,
		RequireMessage:    false,
	}

	// 단일 사용자 서비스 (동시성 보장 내장)
	userService := friendit.NewUserService(userRepo, config)

	// 예제: 동시성 안전 사용자 상태 업데이트
	fmt.Println("=== 동시성 안전 사용자 관리 ===")

	// 1. 사용자 생성
	user := &friendit.BasicUser{
		ID:       "user1",
		Status:   "online",
		Metadata: make(map[string]any),
	}
	err := userRepo.Create(ctx, user)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}
	fmt.Printf("Created user: %s\n", user.ID)

	// 2. 동시성 안전 상태 업데이트
	err = userService.UpdateUserStatus(ctx, "user1", "away")
	if err != nil {
		log.Fatalf("Failed to update user status: %v", err)
	}
	fmt.Println("Updated user status to 'away'")

	// 3. 사용자 조회
	retrievedUser, err := userService.GetUser(ctx, "user1")
	if err != nil {
		log.Fatalf("Failed to get user: %v", err)
	}
	fmt.Printf("Retrieved user: %s (Status: %s)\n", retrievedUser.ID, retrievedUser.Status)

	// 4. 동시 업데이트 테스트 (Race condition 방지 확인)
	fmt.Println("\n=== 동시성 테스트 ===")

	// 동시에 상태 업데이트 시도
	go func() {
		err := userService.UpdateUserStatus(ctx, "user1", "busy")
		if err != nil {
			fmt.Printf("Goroutine 1 failed: %v\n", err)
		} else {
			fmt.Println("Goroutine 1: Status updated to 'busy'")
		}
	}()

	go func() {
		err := userService.UpdateUserStatus(ctx, "user1", "offline")
		if err != nil {
			fmt.Printf("Goroutine 2 failed: %v\n", err)
		} else {
			fmt.Println("Goroutine 2: Status updated to 'offline'")
		}
	}()

	// 잠깐 대기
	time.Sleep(100 * time.Millisecond)

	// 최종 상태 확인
	finalUser, err := userService.GetUser(ctx, "user1")
	if err != nil {
		log.Fatalf("Failed to get final user: %v", err)
	}
	fmt.Printf("Final user status: %s\n", finalUser.Status)

	fmt.Println("\n=== 개선된 동시성 안전 Friendit 예제 완료! ===")
	fmt.Println("✅ 별도 ConcurrentService 없이 기존 서비스에서 동시성 보장")
	fmt.Println("✅ 원자적 연산으로 Race condition 방지")
	fmt.Println("✅ 단일 API로 간편한 사용")
}

// UserService 인터페이스의 GetUser 메서드 - 이미 구현되어 있어야 함
// 만약 없다면 repository에서 직접 조회
func getUserSafely(service *friendit.BasicUserService[*friendit.BasicUser], ctx context.Context, userID friendit.UserID) (*friendit.BasicUser, error) {
	// 실제로는 service.GetUser()가 있어야 하지만, 없다면 repository 직접 사용
	// return service.repo.GetByID(ctx, userID)
	return nil, fmt.Errorf("GetUser method needs to be implemented")
}
