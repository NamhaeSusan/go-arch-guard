# 인증 & 인가 패턴

> 이 파일의 예시 코드는 `errors.md`의 `NewErrorResponse` 헬퍼를 사용한다.

## 인증 방식: Bearer Token

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

- 토큰 전달은 항상 `Authorization` 헤더. query param (`?token=...`) 금지
- 토큰 종류 (JWT, opaque 등)는 프로젝트에 따라 결정하되, 클라이언트 인터페이스는 동일

## Auth 미들웨어 구조

```go
func Auth() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractBearerToken(c)
        if token == "" {
            NewErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "인증 토큰이 필요합니다.")
            c.Abort()
            return
        }

        claims, err := verifyToken(token)
        if err != nil {
            NewErrorResponse(c, http.StatusUnauthorized, "TOKEN_INVALID", "유효하지 않은 토큰입니다.")
            c.Abort()
            return
        }

        // context에 사용자 정보 세팅
        c.Set("userId", claims.UserID)
        c.Set("userRole", claims.Role)
        c.Next()
    }
}

func extractBearerToken(c *gin.Context) string {
    header := c.GetHeader("Authorization")
    parts := strings.SplitN(header, " ", 2)
    if len(parts) != 2 || parts[0] != "Bearer" {
        return ""
    }
    return parts[1]
}
```

## Context에서 사용자 정보 꺼내기

타입 안전한 헬퍼 함수를 통해서만 접근:

```go
func UserIDFrom(c *gin.Context) string {
    v, _ := c.Get("userId")
    id, _ := v.(string)
    return id
}

func UserRoleFrom(c *gin.Context) string {
    v, _ := c.Get("userRole")
    role, _ := v.(string)
    return role
}
```

> `c.Get("userId")` 직접 호출을 핸들러에서 하지 않는다.
> 키 문자열 오타를 방지하고, 타입 단언을 한 곳에서 관리.

## 401 vs 403 판단 기준

| 상황 | 코드 | 에러 코드 |
|------|------|-----------|
| 토큰 없음 | 401 | `UNAUTHORIZED` |
| 토큰 만료 | 401 | `TOKEN_EXPIRED` |
| 토큰 변조/무효 | 401 | `TOKEN_INVALID` |
| 토큰 유효하지만 권한 없음 | 403 | `FORBIDDEN` |
| 다른 사용자의 리소스 접근 | 403 | `FORBIDDEN` |

> 401은 "누구인지 모르겠다", 403은 "누구인지는 알지만 안 된다".

## 인가 미들웨어 (역할 기반)

```go
func RequireRole(roles ...string) gin.HandlerFunc {
    allowed := make(map[string]bool, len(roles))
    for _, r := range roles {
        allowed[r] = true
    }
    return func(c *gin.Context) {
        role := UserRoleFrom(c)
        if !allowed[role] {
            NewErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "접근 권한이 없습니다.")
            c.Abort()
            return
        }
        c.Next()
    }
}
```

```go
// 사용 예
admin := v1.Group("/admin", middleware.RequireRole("ADMIN"))
{
    admin.GET("/users", handler.AdminListUsers)
    admin.DELETE("/users/:user_id", handler.AdminDeleteUser)
}
```

## 리소스 소유자 확인

역할과 별개로, 자기 리소스만 접근 가능한 경우:

```go
func GetUser(c *gin.Context) {
    var req GetUserRequest
    if err := c.ShouldBindUri(&req); err != nil {
        NewErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
        return
    }

    // 본인 리소스인지 확인
    if UserIDFrom(c) != req.UserID && UserRoleFrom(c) != "ADMIN" {
        NewErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "접근 권한이 없습니다.")
        return
    }

    // ...
}
```

> 소유자 확인은 핸들러 또는 app 레이어에서. 미들웨어로 일반화하기 어렵다.

## 공개/보호 라우트 구성

```go
func SetupRouter() *gin.Engine {
    r := gin.New()
    r.Use(middleware.Recovery(), middleware.RequestID(), middleware.Logger())

    // 운영 — 인증 없음
    r.GET("/healthz", handler.Healthz)
    r.GET("/readyz", handler.Readyz)

    // 공개 API — 인증 없음
    public := r.Group("/v1")
    {
        public.POST("/auth/login", handler.Login)
        public.POST("/auth/register", handler.Register)
        public.POST("/auth/refresh", handler.RefreshToken)
    }

    // 보호 API — 인증 필수
    protected := r.Group("/v1", middleware.Auth())
    {
        protected.GET("/users/me", handler.GetMe)
        protected.PUT("/users/me", handler.UpdateMe)
        // ...
    }

    // 관리자 API — 인증 + 역할
    admin := r.Group("/v1/admin", middleware.Auth(), middleware.RequireRole("ADMIN"))
    {
        admin.GET("/users", handler.AdminListUsers)
    }

    return r
}
```
