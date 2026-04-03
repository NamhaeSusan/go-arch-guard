# URL & 리소스 네이밍

## 기본 원칙

- **명사 복수형** 사용, 동사 금지
- **snake_case** (하이픈 X, camelCase X)
- **소문자만** 사용
- **계층은 최대 3단계**까지 (`/resources/{id}/sub-resources/{id}`)

```
✅ GET /v1/users
✅ GET /v1/users/{user_id}/orders
✅ POST /v1/files/{file_id}/versions

❌ GET /v1/getUsers
❌ GET /v1/user
❌ GET /v1/users/{user_id}/orders/{order_id}/items/{item_id}/details
```

## Path Parameter 네이밍

리소스명 + `_id` 접미사:

```
/v1/users/{user_id}
/v1/orders/{order_id}
/v1/files/{file_id}
```

Go 구조체 예시 (gin):
```go
type GetUserRequest struct {
    UserID string `uri:"user_id" binding:"required,uuid"`
}
```

## 동사가 필요한 경우 (Custom Method)

순수 CRUD로 표현 안 되는 액션은 `:action` suffix 사용 (Google AIP-136):

```
POST /v1/files/{file_id}:move
POST /v1/users/{user_id}:deactivate
POST /v1/orders/{order_id}:cancel
```

> ⚠️ `/cancel-order`, `/moveFile` 같은 형태는 사용하지 않는다.

## 관계 리소스

소유 관계가 명확할 때만 중첩. 그 외엔 필터로 처리:

```
✅ GET /v1/users/{user_id}/addresses       # user에 완전히 종속
✅ GET /v1/orders?user_id={user_id}        # order는 독립 리소스

❌ GET /v1/users/{user_id}/orders/{order_id}/items  # 3단계 이상 금지
✅ GET /v1/order_items?order_id={order_id}          # 대신 이렇게
```

## 싱글턴 리소스

단 하나만 존재하는 경우 단수 + id 없음:

```
GET  /v1/users/{user_id}/profile
PUT  /v1/users/{user_id}/profile
```
