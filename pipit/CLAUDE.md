# CLAUDE.md

이 파일은 Claude Code가 Pipit 프로젝트에서 작업할 때 따라야 할 지침을 제공합니다.

## Project Overview

**Pipit**은 Go를 위한 고성능 함수형 프로그래밍 파이프라인 라이브러리입니다. 지연 평가(lazy evaluation), 메소드 체이닝, 제네릭 타입 안전성을 제공하며, Go의 idiomatic patterns를 따릅니다.

### Project Name Origin
"Pipit" = Pipe + it (파이프라인 + 접미사)  
데이터를 Unix 파이프처럼 연결하여 처리한다는 의미를 담고 있습니다.

## 작업 관리

### TODO 트래킹
작업 진행 상황과 TODO 트래킹은 **PIPIT_IMPLEMENTATION_WORKFLOW.md** 문서를 확인하세요.
- 각 Phase별 진행률 추적
- 상세한 TODO 체크리스트
- 의존성 매핑 및 시간 추정

## 핵심 설계 원칙

### 1. 지연 평가 (Lazy Evaluation)
- **필수 원칙**: 모든 성공한 함수형 프레임워크가 채택
- 중간 연산은 실행되지 않고 파이프라인에만 추가
- 터미널 연산 호출 시점에 실제 계산 수행
- 메모리 효율성과 조기 종료(early termination) 지원

### 2. Fluent API (메소드 체이닝)
- **표준 패턴**: 메소드 체이닝으로 가독성 향상
- 각 중간 연산은 새로운 Query 인스턴스 반환
- 불변성(immutability) 원칙 준수

### 3. 타입 안전성
- **중요**: 컴파일 타임 에러 검출
- Go 1.21+ 제네릭 적극 활용
- 타입 변환 시에도 안전성 보장

### 4. 성능과 사용성의 균형
- **추상화 비용 최소화**: Go의 단순함 유지
- 런타임 오버헤드 최소화
- 네이티브 루프 대비 15% 이내 성능 목표

## Go 언어 특성 활용

### Context 통합

### 에러 처리 명시화

### 제네릭 활용

### Goroutine 안전성
- 모든 공유 상태는 mutex로 보호
- Iterator는 thread-safe 구현
- Context를 통한 취소 지원

## 아키텍처 지침

### Go-First Design
- **Go의 단순함과 명확성 유지**
- 과도한 추상화 지양
- 관용적 Go 코드 작성

### Zero-Dependency
- **표준 라이브러리만 사용**
- 외부 의존성 최소화
- 빌드 복잡성 제거

### Performance-Conscious
- **지연 평가로 메모리 효율성 확보**
- 가비지 생성 최소화
- 컴파일 타임 최적화 활용

## 코딩 컨벤션

### 주석

- 모든 주석은 영어로 작성하며, godoc 포맷을 지원해야 합니다.

### 네이밍
- **함수형 표준 용어 사용**: Map, Filter, Reduce (Where, Select 대신)
- Go 관례 준수: PascalCase for public, camelCase for private
- 의미 있는 이름 선택

### 테스트
- testify framework 필수 사용
- 함수명에 기능명 포함: `TestQuery_FilterChaining`
- Context cancellation 테스트 필수
- 벤치마크 테스트 포함

## 중요 참고사항

- **메모리 풀링**: 대용량 데이터 처리 시 고려
- **벤치마킹**: 네이티브 루프와 성능 비교 필수
- **문서화**: Godoc 형식으로 상세한 예제 포함
- **호환성**: Go 1.21+ 제네릭 기능 활용

이 지침을 따라 일관성 있고 고품질의 Pipit 라이브러리를 구축하세요.