---
name: go-api-design
description: >
  Go로 REST API를 설계하거나 구현할 때 일관된 컨벤션을 강제하는 가이드. 바이브코딩 시 
  URL/리소스 네이밍, HTTP 상태코드 & 에러 응답, 버저닝, 페이징/필터링/정렬, 요청/응답 
  필드명 등이 중구난방이 되지 않도록 표준을 제시한다. Go API 코드를 작성하거나 리뷰하거나 
  설계를 논의할 때 반드시 이 skill을 참조한다. gin, echo, chi, net/http 모두 해당.
---

# Go REST API 설계 표준 가이드

> Google API Design Guide + Microsoft REST API Guidelines 기반.  
> 코드 생성 시 이 가이드의 컨벤션을 **강제**한다. 이유 없이 벗어나지 않는다.

세부 규칙은 references/ 안에 있다. 아래 섹션별로 필요한 파일을 로드해서 따른다:

| 주제 | 파일 |
|------|------|
| URL & 리소스 네이밍 | `references/naming.md` |
| HTTP 상태코드 & 에러 응답 | `references/errors.md` |
| 요청/응답 구조 & 필드명 | `references/payload.md` |
| 페이징 / 필터링 / 정렬 | `references/pagination.md` |
| 버저닝 | `references/versioning.md` |
| Kafka 메시지 컨벤션 | `references/kafka.md` |
| 미들웨어 순서 & 공통 헤더 | `references/middleware.md` |
| Health Check & 운영 | `references/health.md` |
| 인증 & 인가 패턴 | `references/auth.md` |
| gRPC 컨벤션 | `references/grpc.md` |

---

## 빠른 체크리스트

API 코드를 작성하기 전에 아래를 확인한다:

- [ ] URL은 복수 명사, snake_case, 동사 금지
- [ ] 에러 응답은 통일된 `ErrorResponse` 구조체 사용
- [ ] 상태코드는 의미에 맞게 (201 vs 200, 422 vs 400 등)
- [ ] 응답 필드명은 `camelCase` (JSON)
- [ ] 목록 응답은 항상 페이징 메타 포함
- [ ] 버전은 URL prefix (`/v1/`)
- [ ] 필터/정렬은 query param, 리소스 식별자는 path param
- [ ] Kafka 이벤트는 Event envelope 구조 사용
- [ ] 토픽명은 `{domain}.{entity}.{action}` 형식
- [ ] consumer는 event.ID 기반 멱등성 보장
- [ ] 미들웨어 순서: Recovery → RequestID → Logger → CORS → RateLimit → Auth
- [ ] `/healthz` (liveness), `/readyz` (readiness) 구현
- [ ] 인증은 Bearer token, context 헬퍼로 사용자 정보 접근
- [ ] gRPC: 표준 메서드명, Enum 0번은 UNSPECIFIED, errdetails 사용

---

## 작업 흐름

1. **새 엔드포인트 설계** → `naming.md` → `payload.md` → `errors.md` 순서로 확인
2. **목록 API 구현** → `pagination.md` 필수 확인
3. **에러 처리 구현** → `errors.md` 에서 상태코드 + 구조체 확인
4. **Kafka 이벤트 설계** → `kafka.md` 에서 envelope + 토픽 네이밍 확인
5. **gRPC 서비스 설계** → `grpc.md` 에서 메서드명 + 에러 코드 확인
6. **미들웨어 구성** → `middleware.md` 에서 체인 순서 확인
7. **인증/인가 구현** → `auth.md` 에서 패턴 확인
8. **기존 코드 리뷰** → ��크리스트 기준으로 지적
