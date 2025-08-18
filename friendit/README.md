# Friendit - 친구 관리 서비스

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue.svg)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Friendit은 게임 서버와 소셜 애플리케이션을 위한 고성능 친구 관리 서비스입니다. 최대한의 커스터마이징과 자유로운 관리를 제공합니다.

## 🚀 Quick Start

```go
package main

import (
    "context"
    "github.com/homveloper/friendit"
)

func main() {
    // 기본 친구 요청
    friendit.Request().
        From("user123").
        To("user456").
        WithMessage("Let's play together!").
        Send()

    // 고급 필터링
    friends := friendit.Filter().
        User("user123").
        Online().
        Game("wow", "lol").
        SortBy(friendit.LastActive.Desc()).
        Get()
}
```

## ✨ 주요 기능

### 🎯 최대 커스터마이징
- **Fluent API**: 직관적이고 유연한 체이닝 인터페이스
- **이벤트 시스템**: 커스텀 훅과 미들웨어 지원
- **플러그인 아키텍처**: 확장 가능한 컴포넌트 시스템
- **정책 엔진**: 비즈니스 룰 커스터마이징

### 🏗️ Clean Architecture
- **도메인 계층**: 순수 비즈니스 로직
- **인프라 계층**: 플러그인 방식 저장소 (MongoDB, Redis, PostgreSQL)
- **서비스 계층**: 복잡한 유스케이스 처리

### 🔧 프로덕션 준비
- **다중 저장소 지원**: MongoDB, Redis, PostgreSQL, 인메모리
- **실시간 업데이트**: WebSocket/SSE 지원
- **성능 최적화**: 캐싱, 배치 처리, 페이지네이션
- **모니터링**: 메트릭, 로깅, 추적

## 📦 설치

```bash
go get github.com/homveloper/friendit
```

## 🏛️ 아키텍처

```
friendit/
├── domain/          # 도메인 엔터티 및 비즈니스 로직
├── repository/      # 저장소 인터페이스
├── service/         # 서비스 레이어
├── adapters/        # 인프라 어댑터 (MongoDB, Redis 등)
├── api/             # 공개 API 인터페이스
└── examples/        # 사용 예제
```

## 📖 문서

- [설계 문서](./DESIGN.md)
- [API 참조](./docs/api.md)
- [예제 모음](./examples/)

## 🤝 기여

이 프로젝트는 MIT 라이선스 하에 배포됩니다. 기여를 환영합니다!

## 📄 라이선스

MIT License - 자세한 내용은 [LICENSE](LICENSE) 파일을 참조하세요.