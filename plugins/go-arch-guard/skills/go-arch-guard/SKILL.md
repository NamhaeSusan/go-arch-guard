---
name: go-arch-guard
description: Use when user explicitly mentions go-arch-guard, architecture guardrails, or architecture_test.go in a Go project. Handles AI-agent-friendly scaffolding, preset configuration, and automation-oriented setup.
user_invocable: true
---

# go-arch-guard

Go 프로젝트에 아키텍처 가드레일을 `go test`로 적용하고, AI 에이전트가 바로 스캐폴딩하고 유지하기 쉬운 형태로 안내한다.

## Decision Flow

프로젝트 상태를 확인하고 분기한다.

1. `internal/` 존재 여부 확인
2. `architecture_test.go` 존재 여부 확인

**새 프로젝트** (`internal/` 없음 AND `architecture_test.go` 없음) → **Quick Init** 실행
**기존 프로젝트** → **Model Reference** 참조하여 수정/리팩토링

---

## Quick Init

### Step 1: 프리셋 선택

| 프리셋 | 함수 | 레이어 방향 |
|--------|------|------------|
| DDD | `DDD()` (기본값) | handler→app→core/{model,repo,svc},event; infra→core/repo+core/model+event |
| Clean Architecture | `CleanArch()` | handler→usecase→entity+gateway; infra→gateway+entity |
| Layered | `Layered()` | handler→service→repository+model; repository→model |
| Hexagonal | `Hexagonal()` | handler→usecase→port+domain; adapter→port+domain |
| Modular Monolith | `ModularMonolith()` | api→application→core; infrastructure→core |
| Consumer/Worker | `ConsumerWorker()` | worker→service→store→model |
| Batch | `Batch()` | job→service→store→model |
| Event-Driven Pipeline | `EventPipeline()` | command→aggregate→event; projection→readstore |

### Step 2: 스캐폴딩

```bash
go get github.com/NamhaeSusan/go-arch-guard
mkdir -p internal/domain internal/orchestration internal/pkg
# Consumer/Worker 프리셋의 경우:
# mkdir -p internal/worker internal/service internal/store internal/model internal/pkg
# Batch 프리셋의 경우:
# mkdir -p internal/job internal/service internal/store internal/model internal/pkg
# Event-Driven Pipeline 프리셋의 경우:
# mkdir -p internal/command internal/aggregate internal/event internal/projection internal/eventstore internal/readstore internal/model internal/pkg
```

### Step 3: architecture_test.go 생성

파일이 이미 존재하면 덮어쓰지 않고 유저에게 확인한다.
패키지명은 **유효한 Go 패키지 식별자**여야 한다. 예: `myapp_test`
하이픈이 있는 module basename(예: `go-arch-guard`)을 그대로 쓰면 안 된다.

우선 `scaffold.ArchitectureTest(...)`로 프리셋별 템플릿을 생성하는 것을 우선한다.

```go
import "github.com/NamhaeSusan/go-arch-guard/scaffold"

src, err := scaffold.ArchitectureTest(
    scaffold.PresetDDD, // or PresetCleanArch / PresetLayered / PresetHexagonal / PresetModularMonolith / PresetConsumerWorker / PresetBatch / PresetEventPipeline
    scaffold.ArchitectureTestOptions{PackageName: "myapp_test"},
)
if err != nil {
    return err
}
```

생성된 결과를 `architecture_test.go`에 저장한다. 생성 템플릿은 내부에서
`rules.RunAll(...)`을 사용해 권장 rule 묶음을 한 번에 실행한다.

가장 간단한 통합은 `rules.RunAll(pkgs, "", "")`를 쓰는 방식이다.
기본값이 아닌 모델이나 severity/exclude 옵션이 필요할 때만 `opts...`를 넘기고,
세부 rule을 개별로 다뤄야 할 때만 `Check*` 함수를 직접 조합한다.

프리셋 매핑:

| 프리셋 | scaffold 상수 |
|--------|----------------|
| DDD | `scaffold.PresetDDD` |
| Clean Architecture | `scaffold.PresetCleanArch` |
| Layered | `scaffold.PresetLayered` |
| Hexagonal | `scaffold.PresetHexagonal` |
| Modular Monolith | `scaffold.PresetModularMonolith` |
| Consumer/Worker | `scaffold.PresetConsumerWorker` |
| Batch | `scaffold.PresetBatch` |
| Event-Driven Pipeline | `scaffold.PresetEventPipeline` |

`{Preset}`을 선택한 함수명으로 치환. `opts...`가 있는 전체 템플릿은 README.md Quick Start 참조.

### Step 4: 검증

```bash
go test -run TestArchitecture -v
```

도메인이 아직 없으므로 모든 체크 통과. 이후 도메인 추가 시 가드레일 적용.

### Step 5: 안내

> 스캐폴딩 완료. `internal/domain/` 아래에 첫 번째 도메인을 추가하세요.
> 레이어 구조는 아래 **Model Reference**를 참고하세요.

---

## Model Reference

### DDD (기본값)

```text
internal/domain/{domain}/
├── alias.go           # public surface (필수)
├── handler/           # 인바운드 어댑터
├── app/               # 애플리케이션 서비스
├── core/model/        # 도메인 모델 (필수)
├── core/repo/         # 레포지토리 인터페이스
├── core/svc/          # 도메인 서비스 인터페이스
├── event/             # 도메인 이벤트
└── infra/             # 아웃바운드 어댑터
```

- `alias.go` 필수, domain root에 다른 .go 금지
- `core/model/`에 최소 1개 .go 필수
- `core/*`, `event`는 `internal/pkg` import 금지
- interface는 `core/repo/`에만 정의

### Clean Architecture

```text
internal/domain/{domain}/
├── handler/    # 컨트롤러
├── usecase/    # 비즈니스 규칙
├── entity/     # 엔터프라이즈 규칙
├── gateway/    # 데이터 접근 인터페이스
└── infra/      # 프레임워크 & 드라이버
```

- `entity`는 `internal/pkg` import 금지
- DTO는 `handler/`, `usecase/`만 허용

### Layered (Spring 스타일)

```text
internal/domain/{domain}/
├── handler/      # HTTP/gRPC 핸들러
├── service/      # 비즈니스 로직
├── repository/   # 데이터 접근
└── model/        # 도메인 모델
```

- `model`은 `internal/pkg` import 금지
- DTO는 `handler/`, `service/`만 허용

### Hexagonal (포트 & 어댑터)

```text
internal/domain/{domain}/
├── handler/   # 드라이빙 어댑터
├── usecase/   # 애플리케이션 로직
├── port/      # 인터페이스 (인바운드+아웃바운드)
├── domain/    # 엔티티, 값 객체
└── adapter/   # 드리븐 어댑터
```

- `domain`은 `internal/pkg` import 금지
- DTO는 `handler/`, `usecase/`만 허용

### Modular Monolith

```text
internal/domain/{domain}/
├── api/              # 모듈 공개 인터페이스
├── application/      # 유즈케이스
├── core/             # 엔티티, 값 객체
└── infrastructure/   # DB, 외부 서비스
```

- `core`는 `internal/pkg` import 금지
- DTO는 `api/`, `application/`만 허용

### Consumer/Worker (플랫 레이아웃)

```text
internal/
├── worker/       # 메시지 핸들러 (worker_order.go → OrderWorker.Process())
├── service/      # 비즈니스 로직
├── store/        # 영속성
├── model/        # 데이터 구조체
└── pkg/          # 공유 인프라
```

- `model`은 `internal/pkg` import 금지
- `worker_*.go` 파일은 대응하는 타입(`XxxWorker`) + `Process` 메서드 필수
- DTO는 `worker/`, `service/`만 허용
- 도메인 격리 규칙 미적용 (플랫 레이아웃)

### Batch (플랫 레이아웃)

```text
internal/
├── job/          # 배치 작업 (job_expire_files.go → ExpireFilesJob.Run())
├── service/      # 비즈니스 로직
├── store/        # 영속성
├── model/        # 데이터 구조체
└── pkg/          # 공유 인프라
```

- `model`은 `internal/pkg` import 금지
- `job_*.go` 파일은 대응하는 타입(`XxxJob`) + `Run` 메서드 필수
- DTO는 `job/`, `service/`만 허용
- 도메인 격리 규칙 미적용 (플랫 레이아웃)

### Event-Driven Pipeline (플랫 레이아웃)

```text
internal/
├── command/       # 커맨드 핸들러 (command_create_order.go → CreateOrderCommand.Execute())
├── aggregate/     # 애그리거트 루트 (aggregate_order.go → OrderAggregate.Apply())
├── event/         # 도메인 이벤트
├── projection/    # 읽기 모델 프로젝터
├── eventstore/    # 이벤트 영속성
├── readstore/     # 읽기 모델 영속성
├── model/         # 공유 값 객체
└── pkg/           # 공유 인프라
```

- `event`, `model`은 `internal/pkg` import 금지
- `command_*.go` 파일은 대응하는 타입(`XxxCommand`) + `Execute` 메서드 필수
- `aggregate_*.go` 파일은 대응하는 타입(`XxxAggregate`) + `Apply` 메서드 필수
- DTO는 `command/`, `projection/`만 허용
- 도메인 격리 규칙 미적용 (플랫 레이아웃)

### Custom

```go
m := rules.NewModel(
    rules.WithDomainDir("module"),
    rules.WithSublayers([]string{"api", "logic", "data"}),
    rules.WithDirection(map[string][]string{
        "api": {"logic"}, "logic": {"data"}, "data": {},
    }),
    rules.WithRequireAlias(false),
    rules.WithRequireModel(false),
)
```

전체 옵션: `WithSublayers`, `WithDirection`, `WithPkgRestricted`, `WithDomainDir`, `WithOrchestrationDir`, `WithSharedDir`, `WithRequireAlias`, `WithAliasFileName`, `WithRequireModel`, `WithModelPath`, `WithDTOAllowedLayers`, `WithBannedPkgNames`, `WithLegacyPkgNames`, `WithLayerDirNames`, `WithInterfacePatternExclude`

---

## Domain Isolation (모든 모델 공통)

| from → to | domain root | domain sub-pkg | orchestration | shared pkg |
|-----------|:-:|:-:|:-:|:-:|
| **same domain** | O | O | X | O |
| **other domain** | X | X | X | O |
| **orchestration** | O | X | O | O |
| **cmd** | O | X | O | O |
| **shared pkg** | X | X | X | O |

> **참고:** Consumer/Worker, Batch 프리셋은 플랫 레이아웃이므로 도메인 격리 규칙이 적용되지 않습니다.

---

## Options

```go
rules.WithExclude("internal/legacy/...")  // 마이그레이션 중 경로 제외
rules.WithSeverity(rules.Warning)         // 실패 없이 로그만
```

## Machine-readable Output

AI 에이전트나 CI 봇이 violation을 후처리해야 하면 JSON 출력을 우선한다.

```go
data, err := report.MarshalJSONReport(violations)
if err != nil {
    return err
}
fmt.Println(string(data))
```

---

## Banned Patterns

| Category | Banned |
|----------|--------|
| Package names | `util`, `common`, `misc`, `helper`, `shared`, `services` |
| Legacy dirs | `router`, `bootstrap`, misplaced `app`/`handler`/`infra` |
| Naming | stuttering (`order.OrderService`), `Impl` suffix, non-snake_case, hand-rolled mocks |
| Placement | `middleware/`는 `internal/pkg/middleware/`에만 |

---

## DDD 전용: Cross-Domain Interface 금지

도메인 내 interface는 **`core/repo/`에만** 정의 가능. cross-domain 데이터 필요 시 → `orchestration/handler/`로 이동.

| 우회 시도 | 잡는 규칙 |
|----------|----------|
| handler/app/svc에서 interface 정의 | `structure.interface-placement` |
| alias.go에서 interface 직접 정의 | `structure.domain-alias-no-interface` |
| alias.go에서 repo/svc 타입 re-export | `structure.domain-alias-contract-reexport` |

---

## Existing Project with Violations

1. `architecture_test.go` 생성 후 실행 → 전체 위반 확인
2. `WithExclude`로 마이그레이션 중 경로 제외
3. 점진적으로 위반 수정, exclude 제거
4. 전환기에는 `WithSeverity(rules.Warning)` 사용
