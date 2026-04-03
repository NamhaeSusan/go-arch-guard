# gRPC 컨벤션

## Proto 파일 구조

```
proto/
├── {domain}/
│   └── v1/
│       ├── {service}.proto        # 서비스 정의
│       ├── {service}_messages.proto  # 메시지만 (선택)
│       └── common.proto           # 도메인 공통 타입
└── common/
    └── v1/
        ├── pagination.proto
        └── error_details.proto
```

```protobuf
syntax = "proto3";
package myapp.order.v1;

option go_package = "github.com/myorg/myapp/gen/order/v1;orderv1";
```

> `go_package`는 항상 gen/ 아래 생성. 수동 코드와 섞지 않는다.

## 서비스 네이밍

`{Resource}Service` 형식. CRUD는 표준 메서드명 사용:

```protobuf
service OrderService {
    rpc CreateOrder(CreateOrderRequest)   returns (Order);
    rpc GetOrder(GetOrderRequest)         returns (Order);
    rpc ListOrders(ListOrdersRequest)     returns (ListOrdersResponse);
    rpc UpdateOrder(UpdateOrderRequest)   returns (Order);
    rpc DeleteOrder(DeleteOrderRequest)   returns (google.protobuf.Empty);

    // Custom method — 동사로 시작
    rpc CancelOrder(CancelOrderRequest)   returns (Order);
    rpc ShipOrder(ShipOrderRequest)       returns (Order);
}
```

| 패턴 | 메서드명 | 요청 | 응답 |
|------|----------|------|------|
| 생성 | `Create{Resource}` | `Create{Resource}Request` | `{Resource}` |
| 조회 | `Get{Resource}` | `Get{Resource}Request` | `{Resource}` |
| 목록 | `List{Resource}s` | `List{Resource}sRequest` | `List{Resource}sResponse` |
| 수정 | `Update{Resource}` | `Update{Resource}Request` | `{Resource}` |
| 삭제 | `Delete{Resource}` | `Delete{Resource}Request` | `Empty` |
| 커스텀 | `{Verb}{Resource}` | `{Verb}{Resource}Request` | `{Resource}` |

## 메시지 네이밍

```protobuf
// 리소스 메시지 — 접미사 없음
message Order {
    string id = 1;
    string user_id = 2;
    OrderStatus status = 3;
    int64 total_amount = 4;
    string currency = 5;
    google.protobuf.Timestamp created_at = 6;
    google.protobuf.Timestamp updated_at = 7;
}

// 요청 — {Method}{Resource}Request
message CreateOrderRequest {
    string user_id = 1;
    repeated OrderItem items = 2;
}

message GetOrderRequest {
    string id = 1;
}

// 목록 응답 — List{Resource}sResponse
message ListOrdersResponse {
    repeated Order orders = 1;
    int32 total_count = 2;
    string next_page_token = 3;
}
```

## 필드 규칙

- **snake_case** (protobuf 표준, Go에서 자동 PascalCase 변환)
- ID 필드: `id`, 외래키: `{resource}_id`
- 시간: `google.protobuf.Timestamp`, 필드명 `_at` 접미사
- Boolean: `is_` / `has_` 접두사
- Enum: 0번은 항상 `UNSPECIFIED`

```protobuf
enum OrderStatus {
    ORDER_STATUS_UNSPECIFIED = 0;  // 항상 0번 = UNSPECIFIED
    ORDER_STATUS_PENDING = 1;
    ORDER_STATUS_IN_PROGRESS = 2;
    ORDER_STATUS_COMPLETED = 3;
    ORDER_STATUS_CANCELLED = 4;
}
```

> Enum 값은 `{TYPE}_{VALUE}` 형식. 전역 네임스페이스 충돌 방지.

## 페이지네이션

Google AIP-158 기반 token 방식:

```protobuf
message ListOrdersRequest {
    int32 page_size = 1;       // 최대 100, 기본 20
    string page_token = 2;     // 다음 페이지 토큰
    string filter = 3;         // 필터 표현식 (선택)
    string order_by = 4;       // 정렬 (e.g. "created_at desc")
}

message ListOrdersResponse {
    repeated Order orders = 1;
    string next_page_token = 2;  // 빈 문자열이면 마지막 페이지
    int32 total_count = 3;       // 전체 건수 (비용 높으면 생략 가능)
}
```

## 에러 처리

gRPC status code 사용. 커스텀 에러 디테일은 `errdetails`로:

| 상황 | gRPC Code | REST 대응 |
|------|-----------|-----------|
| 요청 형식 오류 | `InvalidArgument` | 400 |
| 인증 없음 | `Unauthenticated` | 401 |
| 권한 없음 | `PermissionDenied` | 403 |
| 리소스 없음 | `NotFound` | 404 |
| 중복 | `AlreadyExists` | 409 |
| 비즈니스 규칙 위반 | `FailedPrecondition` | 422 |
| rate limit | `ResourceExhausted` | 429 |
| 내부 오류 | `Internal` | 500 |

```go
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/genproto/googleapis/rpc/errdetails"
)

// 단순 에러
func notFound(resource, id string) error {
    return status.Errorf(codes.NotFound, "%s %q not found", resource, id)
}

// 필드 유효성 에러 (디테일 포함)
func validationError(violations []*errdetails.BadRequest_FieldViolation) error {
    st := status.New(codes.InvalidArgument, "validation failed")
    detail := &errdetails.BadRequest{FieldViolations: violations}
    st, _ = st.WithDetails(detail)
    return st.Err()
}

// 사용 예
return validationError([]*errdetails.BadRequest_FieldViolation{
    {Field: "email", Description: "invalid email format"},
    {Field: "name", Description: "name is required"},
})
```

## Interceptor 체인 (미들웨어 대응)

REST 미들웨어 순서와 동일한 개념 (gRPC는 CORS가 불필요하므로 생략):

```go
func NewGRPCServer() *grpc.Server {
    return grpc.NewServer(
        grpc.ChainUnaryInterceptor(
            recovery.UnaryServerInterceptor(),     // 1. panic recovery
            requestid.UnaryServerInterceptor(),    // 2. request ID (metadata)
            logging.UnaryServerInterceptor(),      // 3. 로깅
            ratelimit.UnaryServerInterceptor(),    // 4. rate limit
            auth.UnaryServerInterceptor(),         // 5. 인증
        ),
        grpc.ChainStreamInterceptor(
            recovery.StreamServerInterceptor(),
            requestid.StreamServerInterceptor(),
            logging.StreamServerInterceptor(),
            ratelimit.StreamServerInterceptor(),
            auth.StreamServerInterceptor(),
        ),
    )
}
```

## Metadata (gRPC 헤더)

REST 헤더와 동일한 메타데이터를 metadata로 전달:

| Metadata Key | 값 | 용도 |
|--------------|----|------|
| `x-request-id` | UUID | 분산 추적 |
| `authorization` | `Bearer {token}` | 인증 |

```go
// 클라이언트에서 전송
md := metadata.Pairs(
    "x-request-id", uuid.NewString(),
    "authorization", "Bearer "+token,
)
ctx := metadata.NewOutgoingContext(ctx, md)
resp, err := client.GetOrder(ctx, req)

// 서버에서 수신
md, ok := metadata.FromIncomingContext(ctx)
requestID := md.Get("x-request-id")
```

## Health Check (gRPC)

`grpc.health.v1.Health` 표준 서비스 사용:

```go
import "google.golang.org/grpc/health"
import healthpb "google.golang.org/grpc/health/grpc_health_v1"

healthServer := health.NewServer()
healthpb.RegisterHealthServer(grpcServer, healthServer)

// 서비스별 상태 설정
healthServer.SetServingStatus("myapp.order.v1.OrderService", healthpb.HealthCheckResponse_SERVING)
```

> 커스텀 health 엔드포인트 만들지 않는다. 표준 프로토콜 사용.

## Proto 버저닝

URL 버저닝과 동일한 원칙. package에 버전 포함:

```protobuf
package myapp.order.v1;  // v1
package myapp.order.v2;  // breaking change 시
```

하위 호환 변경 (필드 추가)은 같은 package에서. Breaking change만 v2로.

> 삭제된 필드 번호는 `reserved`로 보호:

```protobuf
message Order {
    reserved 6;              // 삭제된 필드 번호 재사용 방지
    reserved "old_field";    // 삭제된 필드 이름 재사용 방지
}
```

## buf lint 설정

Proto 파일 품질은 `buf lint`로 강제한다. 프로젝트 루트에 `buf.yaml`:

```yaml
version: v2
lint:
  use:
    - STANDARD           # buf 표준 규칙 전체
  except:
    - PACKAGE_NO_IMPORT_CYCLE  # 필요 시 제외
  disallow_comment_ignores: true  # // buf:lint:ignore 금지 (예외 없이 준수)
breaking:
  use:
    - FILE               # 하위 호환성 검증
```

### 주요 lint 규칙과 이 가이드의 매핑

| buf 규칙 | 이 가이드에서 해당 | 효과 |
|----------|--------------------|------|
| `ENUM_ZERO_VALUE_SUFFIX` | Enum 0번 = UNSPECIFIED | `_UNSPECIFIED` suffix 강제 |
| `ENUM_VALUE_PREFIX` | Enum `{TYPE}_{VALUE}` | 타입명 prefix 강제 |
| `SERVICE_SUFFIX` | `{Resource}Service` | Service suffix 강제 |
| `RPC_REQUEST_STANDARD_NAME` | `{Method}Request` | 요청 메시지 네이밍 |
| `RPC_RESPONSE_STANDARD_NAME` | `{Method}Response` | 응답 메시지 네이밍 |
| `FIELD_LOWER_SNAKE_CASE` | 필드 snake_case | 필드명 형식 |
| `PACKAGE_VERSION_SUFFIX` | `v1`, `v2` 버저닝 | 패키지 버전 suffix |
| `RPC_REQUEST_RESPONSE_UNIQUE` | 메시지 재사용 금지 | 각 RPC별 전용 메시지 |

### CI에서 실행

```yaml
# GitHub Actions
- uses: bufbuild/buf-setup-action@v1
- uses: bufbuild/buf-lint-action@v1

# 또는 직접
- run: buf lint
- run: buf breaking --against '.git#branch=main'
```

### 로컬 개발

```bash
# 설치
brew install bufbuild/buf/buf   # macOS
go install github.com/bufbuild/buf/cmd/buf@latest  # Go

# 사용
buf lint                        # lint
buf breaking --against '.git#branch=main'  # breaking change 검증
buf generate                    # 코드 생성
```

> `protoc` 직접 사용 대신 `buf generate`를 권장한다. `buf.gen.yaml`로 코드 생성을 선언적으로 관리.
```
