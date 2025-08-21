# 함수형 프로그래밍 라이브러리 네이밍 후보군
## DukDakit 생태계를 위한 이름 제안서

*작성일: 2025년 8월*

---

## 🎯 네이밍 전략 및 기준

### DukDakit 네이밍 컨벤션 분석
기존 DukDakit 모듈들의 패턴:
- **Timex** - Time + x (시간 관련 유틸리티)
- **Pagit** - Page + it (페이지네이션)
- **Retry** - 직관적 영어 단어

### 네이밍 기준
1. **기억하기 쉬움**: 개발자가 쉽게 기억할 수 있는 이름
2. **의미 명확성**: 함수형/쿼리 기능임을 직관적으로 알 수 있음
3. **DukDakit 일관성**: 기존 네이밍 패턴과 조화
4. **발음 용이성**: 한국어/영어 모두에서 자연스러운 발음
5. **도메인 검색 가능**: 구글링 시 관련 정보를 찾기 쉬움
6. **Go 컨벤션**: Go 언어 커뮤니티의 네이밍 관례 준수

---

## 🏆 1등급 후보군 (강력 추천)

### 1. **Flowit** (플로우잇)
```go
dukdakit.Flowit.From(data).Where(pred).Select(mapper).ToSlice()
```

**어원**: Flow + it (데이터 흐름 + DukDakit 접미사)
**의미**: 데이터의 흐름을 처리한다는 직관적 의미
**장점**:
- ✅ 함수형 파이프라인의 "흐름" 개념을 잘 표현
- ✅ DukDakit 네이밍 패턴(~it) 완벽 부합
- ✅ 발음하기 쉬움 (플로우-잇)
- ✅ 기존 라이브러리와 충돌 가능성 낮음
- ✅ 게임 개발에서 친숙한 "플로우" 용어

**사용 예시**:
```go
// 플레이어 랭킹 계산
topPlayers := dukdakit.Flowit.From(players).
    Where(func(p Player) bool { return p.IsActive }).
    OrderByDesc(func(p Player) int { return p.Score }).
    Take(10).
    ToSlice()
```

### 2. **Chainx** (체인엑스)
```go
dukdakit.Chainx.From(data).Where(pred).Select(mapper).ToSlice()
```

**어원**: Chain + x (메소드 체이닝 + DukDakit x 접미사)
**의미**: 메소드 체이닝을 통한 연속적 데이터 처리
**장점**:
- ✅ 메소드 체이닝의 핵심 개념 표현
- ✅ Timex와 일관된 x 접미사 사용
- ✅ 간결하고 기술적 느낌
- ✅ 체인 패턴은 프로그래밍에서 잘 알려진 개념

**사용 예시**:
```go
// 이벤트 로그 분석
errorLogs := dukdakit.Chainx.From(logs).
    Filter(func(log LogEntry) bool { return log.Level == "ERROR" }).
    Map(func(log LogEntry) string { return log.Message }).
    ToSlice()
```

### 3. **Streamx** (스트림엑스)
```go
dukdakit.Streamx.From(data).Where(pred).Select(mapper).ToSlice()
```

**어원**: Stream + x (데이터 스트림 + DukDakit x 접미사)
**의미**: 스트림 방식의 데이터 처리
**장점**:
- ✅ Java Streams, C# LINQ와 유사한 개념으로 익숙함
- ✅ 스트림 처리는 게임 서버에서 중요한 개념
- ✅ x 접미사로 DukDakit 일관성 유지
- ✅ 함수형 프로그래밍 커뮤니티에서 널리 사용되는 용어

**잠재적 단점**:
- ⚠️ Java Stream과 개념적 유사성으로 차별화 부족 가능성

---

## 🥈 2등급 후보군 (우수)

### 4. **Linqx** (링큐엑스)
```go
dukdakit.Linqx.From(data).Where(pred).Select(mapper).ToSlice()
```

**어원**: LINQ + x (Language Integrated Query + x 접미사)
**의미**: Go 버전의 LINQ
**장점**:
- ✅ C# LINQ와 직접적 연관성으로 개념 이해 쉬움
- ✅ 함수형 쿼리의 대명사인 LINQ 브랜딩 활용
**단점**:
- ❌ Microsoft 상표와의 잠재적 법적 이슈
- ❌ 독창성 부족

### 5. **Pipit** (피핏)
```go
dukdakit.Pipit.From(data).Where(pred).Select(mapper).ToSlice()
```

**어원**: Pipe + it (파이프라인 + it 접미사)
**의미**: 데이터 파이프라인 처리
**장점**:
- ✅ Unix 파이프의 개념과 연결
- ✅ it 접미사로 DukDakit 패턴 유지
- ✅ 귀여운 새 이름과도 연결 (Pipit = 종달새과)
**단점**:
- ⚠️ 파이프 개념이 모든 개발자에게 직관적이지 않을 수 있음

### 6. **Queryx** (쿼리엑스)
```go
dukdakit.Queryx.From(data).Where(pred).Select(mapper).ToSlice()
```

**어원**: Query + x (쿼리 + x 접미사)  
**의미**: 쿼리 기반 데이터 처리
**장점**:
- ✅ 쿼리는 직관적이고 이해하기 쉬운 개념
- ✅ x 접미사로 Timex와 일관성
**단점**:
- ⚠️ 너무 직접적이어서 창의성 부족

---

## 🥉 3등급 후보군 (괜찮음)

### 7. **Funkit** (펀킷)
```go
dukdakit.Funkit.From(data).Where(pred).Select(mapper).ToSlice()
```

**어원**: Functional + kit (함수형 + 키트)
**의미**: 함수형 프로그래밍 도구 키트
**장점**:
- ✅ DukDakit의 "kit" 테마와 완벽한 조화
- ✅ 함수형 프로그래밍임을 명확히 표현
**단점**:
- ⚠️ "Fun"이 재미있다는 뜻으로 오해될 수 있음

### 8. **Mapit** (맵잇)
```go
dukdakit.Mapit.From(data).Where(pred).Select(mapper).ToSlice()
```

**어원**: Map + it (맵 함수 + it 접미사)
**의미**: 매핑 기반 데이터 변환
**장점**:
- ✅ 함수형 프로그래밍의 핵심인 Map 강조
- ✅ it 접미사로 네이밍 일관성
**단점**:
- ❌ Map 하나의 기능만 강조해서 제한적 느낌
- ❌ 지도(Map)와 혼동 가능성

### 9. **Transformx** (트랜스폼엑스)
```go
dukdakit.Transformx.From(data).Where(pred).Select(mapper).ToSlice()
```

**어원**: Transform + x (변환 + x 접미사)
**의미**: 데이터 변환 처리
**장점**:
- ✅ 데이터 변환의 본질을 정확히 표현
- ✅ x 접미사로 네이밍 일관성
**단점**:
- ❌ 너무 길어서 타이핑하기 불편
- ❌ 게임 개발에서 Transform(좌표변환)과 혼동 가능

---

## 🎨 창의적 후보군

### 10. **Fluxit** (플럭싯)
```go
dukdakit.Fluxit.From(data).Where(pred).Select(mapper).ToSlice()
```

**어원**: Flux + it (흐름 + it 접미사)
**의미**: 데이터 플럭스 처리
**장점**:
- ✅ Reactive Programming의 Flux 개념 차용
- ✅ 독창적이면서도 기술적
**단점**:
- ⚠️ Flux는 일부 개발자에게만 친숙한 개념

### 11. **Rangit** (랭잇)  
```go
dukdakit.Rangit.From(data).Where(pred).Select(mapper).ToSlice()
```

**어원**: Range + it (범위 + it 접미사) 
**의미**: 범위 기반 데이터 처리
**장점**:
- ✅ Range는 함수형에서 중요한 개념
- ✅ it 접미사로 네이밍 일관성
**단점**:
- ⚠️ Range만으로는 전체 기능을 표현하기 부족

### 12. **Selectx** (셀렉트엑스)
```go
dukdakit.Selectx.From(data).Where(pred).Select(mapper).ToSlice()
```

**어원**: Select + x (선택 + x 접미사)
**의미**: 선택적 데이터 처리  
**장점**:
- ✅ SQL/LINQ의 SELECT와 직접적 연관
- ✅ x 접미사로 네이밍 일관성
**단점**:
- ❌ Select 하나의 기능만 강조

---

## 🌟 한국어 영감 후보군

### 13. **Yeollit** (열릿)
```go
dukdakit.Yeollit.From(data).Where(pred).Select(mapper).ToSlice()
```

**어원**: 열다(Open) + it (열어서 처리한다 + it 접미사)
**의미**: 데이터를 열어서 처리한다
**장점**:
- ✅ DukDakit의 한국어 네이밍 철학과 부합
- ✅ 독창적이고 기억하기 쉬움
**단점**:
- ❌ 영어권 개발자에게 발음/이해 어려움

### 14. **Ppulit** (뿔릿)
```go
dukdakit.Ppulit.From(data).Where(pred).Select(mapper).ToSlice()  
```

**어원**: 뽑다(Pull) + it (뽑아서 처리한다 + it 접미사)
**의미**: 데이터를 뽑아내어 처리한다
**장점**:
- ✅ DukDakit의 뚝딱(DukDak)과 유사한 한국어 의성어
- ✅ Pull의 개념이 데이터 처리와 잘 맞음
**단점**:
- ❌ 발음이 어색할 수 있음

---

## 📊 종합 평가 및 추천

### 최종 추천 순위

| 순위 | 이름 | 점수 | 강점 | 약점 |
|------|------|------|-------|-------|
| 🥇 | **Flowit** | 95/100 | 직관적, 일관성, 발음 용이 | 없음 |
| 🥈 | **Chainx** | 90/100 | 기술적 정확성, 간결함 | 체인 개념이 일부에게 생소 |
| 🥉 | **Streamx** | 85/100 | 널리 알려진 개념, 전문적 | 차별화 부족 |
| 4 | Pipit | 80/100 | 독창적, 귀여움 | 파이프 개념 이해 필요 |
| 5 | Queryx | 75/100 | 명확한 의미 | 창의성 부족 |

### 🏆 최종 추천: **Flowit**

**추천 이유**:
1. **직관성**: "Flow"는 데이터의 흐름을 완벽히 표현
2. **일관성**: DukDakit의 "~it" 패턴과 완벽 일치  
3. **사용성**: 발음하기 쉽고 기억하기 쉬움
4. **확장성**: 향후 다양한 데이터 플로우 기능 추가 가능
5. **브랜딩**: 게임 개발에서 친숙한 "플로우" 개념

### 사용 예시 비교

```go
// 현재 querit
dukdakit.Querit.From(players).Where(isActive).Select(getName).ToSlice()

// 추천 Flowit  
dukdakit.Flowit.From(players).Where(isActive).Select(getName).ToSlice()

// 기존 DukDakit 스타일과 비교
dukdakit.Timex.Elapsed(baseTime, now)
dukdakit.Pagit.NewCursor(data, 10)  
dukdakit.Flowit.From(data).Take(10).ToSlice() // 자연스러운 조화
```

### 대안 선택 가이드

**기술적 정확성을 원한다면**: `Chainx`
**친숙함을 원한다면**: `Streamx`  
**독창성을 원한다면**: `Pipit`
**한국적 정체성을 원한다면**: `Yeollit`

---

## 🎨 브랜딩 아이덴티티

### Flowit 로고 컨셉
```
┌─────────────────┐
│  ╭─→ Filter ──→ │
│  ├─→ Map ────→  │  Flowit
│  ╰─→ Reduce ─→  │  데이터가 흘러가는 모습
└─────────────────┘
```

### 태그라인 제안
- "Let Data Flow" (데이터를 흐르게 하라)
- "Flow Your Way" (당신만의 방식으로 흘려보내라)
- "Smooth Data Flow" (부드러운 데이터 플로우)

---

*이 네이밍 제안서는 DukDakit 생태계의 일관성과 함수형 프로그래밍의 본질을 모두 고려하여 작성되었습니다.*