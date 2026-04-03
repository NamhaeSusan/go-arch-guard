# HTTP 상태코드 & 에러 응답

## 상태코드 사용 기준

| 코드 | 언제 |
|------|------|
| 200 OK | GET, PUT, PATCH 성공 |
| 201 Created | POST로 리소스 생성 성공 |
| 204 No Content | DELETE 성공, 응답 바디 없음 |
| 400 Bad Request | 요청 형식 오류 (JSON 파싱 실패 등) |
| 401 Unauthorized | 인증 없음 (토큰 없거나 만료) |
| 403 Forbidden | 인증은 됐지만 권한 없음 |
| 404 Not Found | 리소스 없음 |
| 409 Conflict | 중복 생성 시도 (이미 존재하는 email 등) |
| 422 Unprocessable Entity | 형식은 맞지만 비즈니스 유효성 실패 |
| 429 Too Many Requests | Rate limit 초과 |
| 500 Internal Server Error | 서버 내부 오류 |

> 400 vs 422 구분: 400은 파싱/바인딩 실패, 422는 비즈니스 규칙 위반.

## 통일된 에러 응답 구조체

**모든** 에러는 아래 구조로 응답한다:

```go
type ErrorResponse struct {
    Code    string `json:"code"`              // 머신리더블 에러 코드 (대문자 SNAKE_CASE)
    Message string `json:"message"`           // 사람이 읽는 메시지
    Details []ErrorDetail `json:"details,omitempty"` // 필드 단위 에러 (422에서 주로)
}

type ErrorDetail struct {
    Field   string `json:"field"`    // 에러 발생 필드명 (camelCase)
    Message string `json:"message"`  // 필드별 메시지
}
```

응답 예시:

```json
// 422 - 유효성 실패
{
  "code": "VALIDATION_FAILED",
  "message": "입력값이 올바르지 않습니다.",
  "details": [
    { "field": "email", "message": "이메일 형식이 아닙니다." },
    { "field": "name", "message": "이름은 2자 이상이어야 합니다." }
  ]
}

// 404 - 리소스 없음
{
  "code": "USER_NOT_FOUND",
  "message": "사용자를 찾을 수 없습니다."
}

// 409 - 중복
{
  "code": "EMAIL_ALREADY_EXISTS",
  "message": "이미 사용 중인 이메일입니다."
}
```

## 에러 코드 네이밍

`대문자_SNAKE_CASE`, 리소스명 포함:

```
USER_NOT_FOUND
FILE_ALREADY_EXISTS
VALIDATION_FAILED
UNAUTHORIZED
FORBIDDEN
INTERNAL_ERROR
```

## Go 구현 패턴 (gin 예시)

```go
// 에러 헬퍼
func NewErrorResponse(c *gin.Context, status int, code, message string, details ...ErrorDetail) {
    resp := ErrorResponse{Code: code, Message: message}
    if len(details) > 0 {
        resp.Details = details
    }
    c.JSON(status, resp)
}

// 사용 예
func GetUser(c *gin.Context) {
    var req GetUserRequest
    if err := c.ShouldBindUri(&req); err != nil {
        NewErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
        return
    }

    user, err := userService.Get(c, req.UserID)
    if errors.Is(err, ErrNotFound) {
        NewErrorResponse(c, http.StatusNotFound, "USER_NOT_FOUND", "사용자를 찾을 수 없습니다.")
        return
    }
    // ...
}
```

## 성공 응답

성공 응답은 래퍼 없이 리소스 직접 반환 (단건):

```json
// GET /v1/users/{user_id}
{
  "id": "usr_123",
  "name": "홍길동",
  "email": "hong@example.com",
  "createdAt": "2024-01-01T00:00:00Z"
}
```

목록은 `pagination.md` 참고.
