# 요청/응답 구조 & 필드명

## JSON 필드명 규칙

- **응답/요청 JSON**: `camelCase`
- **Go 구조체 필드**: `PascalCase` (Go 표준)
- **URL path/query param**: `snake_case`

```go
type UserResponse struct {
    ID        string    `json:"id"`
    FullName  string    `json:"fullName"`
    CreatedAt time.Time `json:"createdAt"`
    IsActive  bool      `json:"isActive"`
}
```

## ID 필드

- JSON에서는 항상 `"id"` (타입 접두사 optional: `"usr_123"`)
- 외래 참조는 `"userId"`, `"fileId"` 형태

```json
{
  "id": "usr_123",
  "userId": "usr_123",
  "fileId": "file_456"
}
```

> UUID 쓰는 경우 string으로, int auto-increment는 가급적 외부 노출 금지 (보안).

## 시간 필드

- 항상 **RFC 3339 / ISO 8601** (`2024-01-01T00:00:00Z`)
- 필드명은 동사 과거형 + `At` 접미사

```json
{
  "createdAt": "2024-01-01T09:00:00+09:00",
  "updatedAt": "2024-06-01T12:30:00Z",
  "deletedAt": null
}
```

```go
// Go에서 자동으로 RFC3339 직렬화됨
CreatedAt time.Time `json:"createdAt"`
```

## Boolean 필드

`is` 또는 `has` 접두사:

```json
{
  "isActive": true,
  "isDeleted": false,
  "hasChildren": true
}
```

## Null vs 필드 생략

- 값이 없는 경우: `null` 반환 (필드 아예 생략 X)
- `omitempty`는 **목록/배열**에만 사용:

```go
type UserResponse struct {
    ID      string  `json:"id"`
    Email   string  `json:"email"`
    Phone   *string `json:"phone"`           // null 허용 필드는 포인터
    Tags    []string `json:"tags,omitempty"` // 빈 배열은 생략 가능
}
```

## 요청 구조체 분리

같은 리소스라도 Create/Update 요청 구조체를 분리:

```go
type CreateUserRequest struct {
    Name     string `json:"name" binding:"required,min=2,max=50"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}

type UpdateUserRequest struct {
    Name  *string `json:"name" binding:"omitempty,min=2,max=50"`
    Phone *string `json:"phone" binding:"omitempty,e164"`
}
// PATCH는 포인터 사용 → nil이면 변경 안 함
```

## 중첩 객체

중첩은 의미 있을 때만, depth 3 이상은 flat하게:

```json
// ✅
{
  "id": "ord_123",
  "user": {
    "id": "usr_456",
    "name": "홍길동"
  },
  "address": {
    "city": "서울",
    "street": "강남대로 123"
  }
}

// ❌ 너무 깊음
{
  "order": {
    "detail": {
      "shipping": {
        "address": { "city": "서울" }
      }
    }
  }
}
```

## 열거형 (Enum)

문자열 상수, `UPPER_SNAKE_CASE`:

```json
{
  "status": "IN_PROGRESS",
  "role": "ADMIN"
}
```

```go
type OrderStatus string

const (
    OrderStatusPending    OrderStatus = "PENDING"
    OrderStatusInProgress OrderStatus = "IN_PROGRESS"
    OrderStatusCompleted  OrderStatus = "COMPLETED"
)
```
