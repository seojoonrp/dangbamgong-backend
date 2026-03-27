# dangbamgong-backend

iOS 전용 공백 추적 소셜 앱 당밤공의 Go 백엔드 서버.

## Tech Stack

Go 1.26 / Echo v4 / MongoDB (Atlas) / APNs (push) / robfig/cron (scheduler)

## Project Structure

```
cmd/api/main.go          # 엔트리포인트 (graceful shutdown 포함)
internal/
  config/                 # 상수 (KST, DayStartHour=16)
  database/               # MongoDB 연결
  domain/                 # AppError 타입, ErrorCode 상수
  dto/                    # 요청/응답 DTO, 응답 헬퍼 (Success, SuccessEmpty, Fail)
  model/                  # MongoDB 도큐먼트 모델 (bson + json 태그)
  handler/                # HTTP 핸들러 (Echo context 처리)
  service/                # 비즈니스 로직 + 스케줄러
  repository/             # MongoDB 데이터 접근 (5초 context timeout)
  middleware/             # JWTAuth, ErrorHandler
  auth/                   # JWT(HS256, 30일), 소셜 로그인(Google/Kakao/Apple)
  push/                   # APNs 클라이언트
docs/                     # Swagger 자동생성 문서
```

## Architecture

**Handler -> Service -> Repository -> MongoDB**

- 모든 의존성은 생성자 주입 (server.go에서 조립)
- protected 라우트는 `middleware.JWTAuth()` 적용, `c.Get(middleware.ContextKeyUserID)` 로 userID 추출
- 에러는 `domain.AppError` 반환 -> `middleware.ErrorHandler`에서 JSON 변환

## Commands

```
make build       # go build -o main cmd/api/main.go
make run         # go run cmd/api/main.go
make test        # go test ./... -v
make watch       # air (live reload)
make docker-run  # MongoDB 컨테이너 실행
```

## Conventions

### 응답 형식

```go
dto.Success[T](c, http.StatusOK, data)     // {"success":true,"data":{...}}
dto.SuccessEmpty(c, http.StatusOK)         // {"success":true}
dto.Fail(c, statusCode, domain.ErrXxx)     // {"success":false,"code":"XXX"}
```

### 에러 생성

```go
domain.NewBadRequest(domain.ErrXxx, "message")
domain.NewNotFound(domain.ErrXxx, "message")
domain.NewInternal("message")
```

### Repository 패턴

- 매 쿼리마다 `ctx, cancel := context.WithTimeout(ctx, 5*time.Second)` 사용
- `mongo.ErrNoDocuments` -> nil 반환 (not found)
- BSON 필터: `bson.M{}`, 업데이트: `bson.M{"$set": ...}`

### 모델 태그

- BSON: snake_case (`bson:"field_name"`)
- JSON: camelCase (`json:"fieldName"`)

### 날짜/시간

- 하루 기준: **16:00 KST** (이전이면 전날, 이후면 당일)
- `config.CalcTargetDay(t)` 으로 대상 날짜 계산
- targetDay 포맷: `"2006-01-02"`

## Environment Variables

| 변수                                                        | 설명                          |
| ----------------------------------------------------------- | ----------------------------- |
| PORT                                                        | 서버 포트 (기본 8000)         |
| APP_ENV                                                     | 환경 (development/production) |
| DB_URI                                                      | MongoDB Atlas 연결 URI        |
| DB_NAME                                                     | DB 이름 (기본 dangbamgong)    |
| JWT_SECRET                                                  | JWT 서명 키                   |
| GOOGLE_WEB_CLIENT_ID                                        | Google OAuth 클라이언트 ID    |
| GOOGLE_IOS_CLIENT_ID                                        | Google iOS 클라이언트 ID      |
| KAKAO_ADMIN_KEY                                             | Kakao REST API 키             |
| APPLE_KEY_PATH / APPLE_KEY_ID / APPLE_TEAM_ID / APPLE_TOPIC | Apple 인증 설정               |

## Others

### User preferences

- 당밤공은 개발자의 golang 백엔드 실력 향상을 목표로 한 사이드 프로젝트임
- 모든 것을 구현하지 말고, 배울 점이 있는 구현 부분은 TODO 주석으로 남기기
