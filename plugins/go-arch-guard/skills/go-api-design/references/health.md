# Health Check & 운영 엔드포인트

## 엔드포인트 목록

| 경로 | 용도 | K8s Probe |
|------|------|-----------|
| `GET /healthz` | 프로세스 살아있는지 | livenessProbe |
| `GET /readyz` | 트래픽 받을 준비 됐는지 | readinessProbe |

> 버전 prefix 없음. `/v1/healthz` 아님. 운영 엔드포인트는 API 버저닝 대상이 아니다.
> Auth 미들웨어 밖에 둔다 (인증 없이 접근 가능해야 probe가 동작).

## 응답 구조

```go
type HealthResponse struct {
    Status string        `json:"status"`           // "ok" | "degraded" | "unhealthy"
    Checks []HealthCheck `json:"checks,omitempty"` // readyz에서만
}

type HealthCheck struct {
    Name    string `json:"name"`    // "database", "redis", "kafka"
    Status  string `json:"status"`  // "ok" | "unhealthy"
    Message string `json:"message,omitempty"`
}
```

### /healthz — 단순 liveness

의존성 체크 없이 즉시 응답. 프로세스가 살아있으면 ok:

```json
// 200 OK
{ "status": "ok" }
```

```go
func Healthz(c *gin.Context) {
    c.JSON(http.StatusOK, HealthResponse{Status: "ok"})
}
```

### /readyz — readiness (의존성 포함)

DB, Redis, Kafka 등 핵심 의존성을 체크:

```json
// 200 OK — 모두 정상
{
  "status": "ok",
  "checks": [
    { "name": "database", "status": "ok" },
    { "name": "redis", "status": "ok" }
  ]
}

// 503 Service Unavailable — 하나라도 실패
{
  "status": "unhealthy",
  "checks": [
    { "name": "database", "status": "ok" },
    { "name": "redis", "status": "unhealthy", "message": "connection refused" }
  ]
}
```

```go
func Readyz(c *gin.Context) {
    ctx := c.Request.Context()
    checks := []HealthCheck{
        checkDatabase(ctx),
        checkRedis(ctx),
    }

    status := "ok"
    httpStatus := http.StatusOK
    for _, ch := range checks {
        if ch.Status != "ok" {
            status = "unhealthy"
            httpStatus = http.StatusServiceUnavailable
            break
        }
    }

    c.JSON(httpStatus, HealthResponse{Status: status, Checks: checks})
}

func checkDatabase(ctx context.Context) HealthCheck {
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()
    if err := db.PingContext(ctx); err != nil {
        return HealthCheck{Name: "database", Status: "unhealthy", Message: err.Error()}
    }
    return HealthCheck{Name: "database", Status: "ok"}
}
```

## 라우터 등록

```go
func SetupRouter() *gin.Engine {
    r := gin.New()

    // 운영 엔드포인트 — auth 밖에
    r.GET("/healthz", handler.Healthz)
    r.GET("/readyz", handler.Readyz)

    // API 엔드포인트 — auth 안에
    v1 := r.Group("/v1", middleware.Auth())
    // ...
}
```

## 주의사항

- `/healthz`에 DB 체크 넣지 않는다. DB가 죽으면 pod가 재시작되는데, DB 문제는 pod 재시작으로 해결 안 됨
- `/readyz` 체크에 timeout 필수 (2초 이내). 느리면 probe 자체가 타임아웃
- 민감한 정보 (connection string, credential) 노출 금지
