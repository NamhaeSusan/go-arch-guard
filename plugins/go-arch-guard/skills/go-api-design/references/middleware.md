# 미들웨어 순서 & 공통 헤더

> 이 파일의 예시 코드는 `errors.md`의 `NewErrorResponse` 헬퍼를 사용한다.

## 미들웨어 체인 순서

바깥에서 안쪽 순서. 순서를 바꾸지 않는다:

```
Recovery → RequestID → Logger → CORS → RateLimit → Auth → Handler
```

```go
func SetupMiddleware(r *gin.Engine) {
    r.Use(
        gin.Recovery(),              // 1. panic recovery (항상 최외곽)
        middleware.RequestID(),      // 2. X-Request-ID 생성/전파
        middleware.Logger(),         // 3. 요청/응답 로깅
        middleware.CORS(),           // 4. CORS (preflight 빠르게 반환)
        middleware.RateLimit(),      // 5. rate limit (인증 전에 차단)
    )

    // 6. Auth는 그룹 단위로 적용
    authorized := r.Group("/v1", middleware.Auth())
    {
        // protected routes
    }

    // public routes (auth 없이)
    public := r.Group("/v1")
    {
        public.POST("/auth/login", handler.Login)
    }
}
```

> Recovery가 최외곽이어야 어떤 미들웨어에서 panic이 나도 잡힌다.
> RateLimit은 Auth 앞에 둬서, 인증 로직 실행 전에 과부하를 차단한다.

## X-Request-ID

모든 요청에 고유 ID를 부여. 클라이언트가 보내면 전파, 없으면 생성:

```go
func RequestID() gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.GetHeader("X-Request-ID")
        if id == "" {
            id = uuid.NewString()
        }
        c.Set("requestID", id)
        c.Header("X-Request-ID", id)
        c.Next()
    }
}
```

- 로그에 항상 포함: `logger.With("requestId", id)`
- Kafka 이벤트 발행 시 header `correlation-id`로 전파
- 하위 서비스 호출 시 `X-Request-ID` 헤더로 전파

## Rate Limit 응답 헤더

429 응답 시 아래 헤더를 포함한다:

```
X-RateLimit-Limit: 100        # 윈도우당 최대 요청 수
X-RateLimit-Remaining: 0      # 남은 요청 수
X-RateLimit-Reset: 1717200000 # 리셋 시각 (Unix timestamp)
Retry-After: 30               # 재시도까지 대기 초
```

```go
func RateLimitExceeded(c *gin.Context, limit int, reset time.Time) {
    c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
    c.Header("X-RateLimit-Remaining", "0")
    c.Header("X-RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))
    c.Header("Retry-After", strconv.Itoa(int(time.Until(reset).Seconds())))
    NewErrorResponse(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "요청 한도를 초과했습니다.")
}
```

## CORS 기본 설정

```go
func CORS() gin.HandlerFunc {
    return cors.New(cors.Config{
        AllowOrigins:     []string{"https://example.com"},  // 프로덕션은 명시적으로
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
        ExposeHeaders:    []string{"X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    })
}
```

> `AllowOrigins: []string{"*"}`는 개발 환경에서만. 프로덕션은 명시적 도메인만 허용.

## 공통 응답 헤더 정리

| 헤더 | 값 | 설정 위치 |
|------|----|-----------|
| `X-Request-ID` | UUID | RequestID 미들웨어 |
| `Content-Type` | `application/json` | 프레임워크 자동 |
| `X-RateLimit-*` | 숫자 | RateLimit 미들웨어 |
| `Retry-After` | 초 | 429 응답 시 |
| `Deprecation` | `true` | deprecated 엔드포인트 |
| `Sunset` | RFC 7231 날짜 | deprecated 엔드포인트 |
