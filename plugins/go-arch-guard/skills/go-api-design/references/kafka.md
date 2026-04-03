# Kafka 메시지 컨벤션

## 토픽 네이밍

`{domain}.{entity}.{action}` 형식, snake_case, 소문자:

```
order.order.created
order.order.cancelled
user.user.registered
payment.payment.completed
file.file.uploaded
```

도메인과 엔티티가 같아도 생략하지 않는다 (일관성 우선).

> dead letter: `{원본토픽}.dlq`  
> retry: `{원본토픽}.retry.{n}`

```
order.order.created.dlq
order.order.created.retry.1
order.order.created.retry.2
```

## 메시지 Envelope

모든 Kafka 메시지는 아래 envelope 구조를 따른다:

```go
type Event[T any] struct {
    ID        string    `json:"id"`        // UUID v4, 멱등성 키
    Type      string    `json:"type"`      // 이벤트 타입 (토픽명과 동일)
    Source    string    `json:"source"`    // 발행 서비스명
    Time      time.Time `json:"time"`      // RFC 3339
    DataVersion string  `json:"dataVersion"` // payload 스키마 버전 ("1", "2", ...)
    Data      T         `json:"data"`      // 실제 페이로드
}
```

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "order.order.created",
  "source": "order-service",
  "time": "2024-06-01T12:30:00Z",
  "dataVersion": "1",
  "data": {
    "orderId": "ord_123",
    "userId": "usr_456",
    "totalAmount": 50000,
    "currency": "KRW"
  }
}
```

## Partition Key 전략

| 상황 | key | 이유 |
|------|-----|------|
| 엔티티 단위 순서 보장 | entity ID (`ord_123`) | 같은 주문 이벤트는 같은 파티션 |
| 사용자 단위 순서 보장 | user ID (`usr_456`) | 같은 유저 이벤트 순서 유지 |
| 순서 불필요 | null 또는 UUID | 파티션 균등 분배 |

```go
// Go 예시 (segmentio/kafka-go)
msg := kafka.Message{
    Topic: "order.order.created",
    Key:   []byte(order.ID),   // partition key
    Value: eventJSON,
}
```

> key는 항상 string으로 직렬화. 복합 key 필요 시 `{userId}:{orderId}` 형태.

## Data (Payload) 필드 규칙

REST API payload 규칙과 동일:

- JSON 필드명: `camelCase`
- ID 외래키: camelCase `{resource}Id` — `"orderId"`, `"userId"`
- 시간 필드: RFC 3339, `At` 접미사 — `"createdAt"`, `"cancelledAt"`
- Boolean: `is`/`has` 접두사 (camelCase) — `"isRefundable"`, `"hasItems"`
- Enum: `UPPER_SNAKE_CASE` 문자열

```json
{
  "orderId": "ord_123",
  "status": "COMPLETED",
  "isRefundable": true,
  "completedAt": "2024-06-01T13:00:00Z"
}
```

## 스키마 버저닝

`dataVersion` 필드로 관리. Breaking change 정의는 REST API와 동일:

**하위 호환 (dataVersion 유지):**
- 새 필드 추가 (consumer가 무시하면 됨)
- optional 필드 추가

**Breaking (dataVersion 올림):**
- 필드 삭제 또는 이름 변경
- 필드 타입 변경
- 필수 필드 추가

버전 분기가 필요한 경우, `Data`를 `json.RawMessage`로 두고 수동 역직렬화한다:

```go
// 버전 분기용 — 제네릭 대신 RawMessage 사용
type RawEvent struct {
    ID          string          `json:"id"`
    Type        string          `json:"type"`
    Source      string          `json:"source"`
    Time        time.Time       `json:"time"`
    DataVersion string          `json:"dataVersion"`
    Data        json.RawMessage `json:"data"`
}

switch event.DataVersion {
case "1":
    var data OrderCreatedV1
    json.Unmarshal(event.Data, &data)
case "2":
    var data OrderCreatedV2
    json.Unmarshal(event.Data, &data)
}
```

> 타입이 확정된 consumer에서는 `Event[T]` 제네릭을, 버전 분기가 필요한 router에서는 `RawEvent`를 사용한다.
> 실무에서는 가급적 하위 호환만 유지하고, breaking이 불가피하면 새 토픽 (`order.order.created.v2`)도 고려.

## Consumer 멱등성

`event.ID`를 기준으로 중복 처리를 방지한다:

```go
func (h *Handler) Handle(ctx context.Context, event Event[OrderCreated]) error {
    // 1. 멱등성 체크
    if h.repo.IsProcessed(ctx, event.ID) {
        return nil // 이미 처리됨
    }

    // 2. 비즈니스 로직
    if err := h.process(ctx, event.Data); err != nil {
        return err
    }

    // 3. 처리 완료 기록
    return h.repo.MarkProcessed(ctx, event.ID)
}
```

## Dead Letter 처리

재시도 초과 시 DLQ로 전송. 원본 이벤트를 감싸서 에러 정보 추가:

```go
type DeadLetterEvent struct {
    OriginalEvent json.RawMessage `json:"originalEvent"`
    Error         string          `json:"error"`
    RetryCount    int             `json:"retryCount"`
    FailedAt      time.Time       `json:"failedAt"`
}
```

```json
{
  "originalEvent": { "id": "...", "type": "order.order.created", ... },
  "error": "database connection timeout",
  "retryCount": 3,
  "failedAt": "2024-06-01T13:05:00Z"
}
```

## Header 활용

메타데이터는 Kafka header에, 비즈니스 데이터는 payload에:

| Header | 값 | 용도 |
|--------|----|------|
| `correlation-id` | UUID | 분산 추적 |
| `content-type` | `application/json` | 직렬화 포맷 |
| `source` | 서비스명 | envelope.source와 동일 (라우팅용) |
