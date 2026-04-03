# API 버저닝

## 방식: URL Path Prefix

```
/v1/users
/v2/users
```

> Header 버저닝(`API-Version: 2`)이나 query param(`?version=2`)은 사용하지 않는다.  
> URL에 버전이 명시적으로 드러나는 게 디버깅/캐싱/라우팅 모두 유리하다.

## 버전 번호 규칙

- 정수만 (`v1`, `v2`, `v3`)
- 마이너/패치 버전은 URL에 노출하지 않음
- Breaking change가 있을 때만 버전 올림

## Breaking Change 정의

버전을 올려야 하는 경우:

- 필드 **삭제** 또는 **이름 변경**
- 필드 타입 변경 (string → int 등)
- 필수 요청 필드 **추가**
- URL 구조 변경
- 상태코드 의미 변경

버전 안 올려도 되는 경우 (하위 호환):

- 응답에 **새 필드 추가** (클라이언트는 무시하면 됨)
- 새 엔드포인트 추가
- 선택적 요청 파라미터 추가

## Go 라우터 구성 (gin 예시)

```go
func SetupRouter() *gin.Engine {
    r := gin.New()

    v1 := r.Group("/v1")
    {
        users := v1.Group("/users")
        {
            users.GET("", handler.ListUsers)
            users.POST("", handler.CreateUser)
            users.GET("/:user_id", handler.GetUser)
            users.PUT("/:user_id", handler.UpdateUser)
            users.DELETE("/:user_id", handler.DeleteUser)
        }
    }

    v2 := r.Group("/v2")
    {
        // breaking change가 있는 경우만
        users := v2.Group("/users")
        {
            users.GET("", handler.ListUsersV2)
        }
    }

    return r
}
```

## 버전 병행 운영

- 이전 버전은 최소 **6개월** 유지 후 Deprecation 공지
- Deprecated 응답에는 헤더 추가:

```
Deprecation: true
Sunset: Sat, 01 Jan 2026 00:00:00 GMT
Link: </v2/users>; rel="successor-version"
```

## 현실적인 전략

MVP나 내부 서비스는 처음부터 `/v1`으로 시작하되, v2 필요성이 생기기 전까지 하위 호환 변경만 유지한다.
