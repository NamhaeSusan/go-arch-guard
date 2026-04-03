# 페이징 / 필터링 / 정렬

## 페이징 방식

**offset 기반** (기본, 간단한 경우):

```
GET /v1/users?page=1&page_size=20
```

**cursor 기반** (대용량, 실시간 데이터):

```
GET /v1/files?limit=20&cursor=eyJpZCI6MTAwfQ==
```

> MyBox 같은 파일 스토리지는 cursor 기반 권장.

## 요청 파라미터

```
page       int    페이지 번호 (1부터, 기본값 1)
page_size  int    페이지 크기 (기본값 20, 최대 100)
limit      int    cursor 방식에서 건수
cursor     string 다음 페이지 커서 (base64 인코딩)
```

```go
type PaginationQuery struct {
    Page     int `form:"page" binding:"omitempty,min=1"`
    PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

func (q *PaginationQuery) SetDefaults() {
    if q.Page == 0 { q.Page = 1 }
    if q.PageSize == 0 { q.PageSize = 20 }
}

func (q *PaginationQuery) Offset() int {
    return (q.Page - 1) * q.PageSize
}
```

## 응답 구조

> 요청 query param은 `snake_case`, 응답 JSON 필드는 `camelCase` (`payload.md` 참고).

**offset 기반**:

```go
type PagedResponse[T any] struct {
    Items      []T        `json:"items"`
    Pagination Pagination `json:"pagination"`
}

type Pagination struct {
    Page       int  `json:"page"`
    PageSize   int  `json:"pageSize"`
    TotalItems int  `json:"totalItems"`
    TotalPages int  `json:"totalPages"`
    HasNext    bool `json:"hasNext"`
    HasPrev    bool `json:"hasPrev"`
}
```

```json
{
  "items": [...],
  "pagination": {
    "page": 2,
    "pageSize": 20,
    "totalItems": 153,
    "totalPages": 8,
    "hasNext": true,
    "hasPrev": true
  }
}
```

**cursor 기반**:

```go
type CursorPagedResponse[T any] struct {
    Items      []T    `json:"items"`
    NextCursor string `json:"nextCursor,omitempty"`
    HasMore    bool   `json:"hasMore"`
}
```

```json
{
  "items": [...],
  "nextCursor": "eyJpZCI6MTIwfQ==",
  "hasMore": true
}
```

## 필터링

query param으로, 필드명은 snake_case:

```
GET /v1/files?status=ACTIVE&created_after=2024-01-01&owner_id=usr_123
```

```go
type FileFilterQuery struct {
    Status       string    `form:"status"`
    OwnerID      string    `form:"owner_id"`
    CreatedAfter time.Time `form:"created_after" time_format:"2006-01-02"`
    CreatedBefore time.Time `form:"created_before" time_format:"2006-01-02"`
    PaginationQuery
}
```

> 날짜 필터: `created_after`, `created_before` (ISO 8601 날짜 형식)

## 정렬

```
GET /v1/files?sort=created_at&order=desc
GET /v1/files?sort=-created_at          # 마이너스 prefix = desc (Google 방식)
```

```go
type SortQuery struct {
    Sort  string `form:"sort"`   // 필드명
    Order string `form:"order"`  // "asc" | "desc"
}

// 허용된 정렬 필드만 화이트리스트 처리
var allowedSortFields = map[string]string{
    "created_at": "created_at",
    "name":       "name",
    "size":       "size",
}
```

> ⚠️ 정렬 필드는 반드시 화이트리스트로 SQL injection 방지

## 검색

전문 검색은 `q` 파라미터:

```
GET /v1/files?q=report&page=1&page_size=20
```
