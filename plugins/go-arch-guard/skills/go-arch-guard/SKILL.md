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
**기존 프로젝트** → **Architecture Reference** 참조하여 수정/리팩토링

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

비표준 패키지 루트(`packages/`, `src/` 등)를 쓰는 프로젝트는 `InternalRoot`도 함께 지정한다 — 그래야 생성된 `analyzer.Load(".", "<root>/...", "cmd/...")`가 실제 레이아웃과 매칭된다:

```go
src, err := scaffold.ArchitectureTest(
    scaffold.PresetDDD,
    scaffold.ArchitectureTestOptions{
        PackageName:  "myapp_test",
        InternalRoot: "packages",
    },
)
```

생성된 결과를 `architecture_test.go`에 저장한다. 생성 템플릿은 내부에서
`core.NewContext(...)`, `presets.{Preset}()`, `presets.Recommended{Preset}()`,
`core.Run(...)`, `report.AssertNoViolations(...)`를 사용한다.

수동 작성 시에도 같은 흐름을 따른다: `analyzer.Load`로 패키지를 로드하고,
프리셋 아키텍처로 `core.NewContext`를 만든 뒤, 권장 ruleset을 `core.Run`에 넘긴다.
세부 rule을 개별로 다뤄야 할 때만 `core.NewRuleSet(...)`으로 직접 조합한다.
rule이 panic을 내면 `core.Run`은 `meta.rule-panic` Error violation으로 변환하고
나머지 rule 실행을 계속한다.

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
> 레이어 구조는 아래 **Architecture Reference**를 참고하세요.

---

## Architecture Reference

### DDD (기본값)

```text
internal/
├── domain/{domain}/
│   ├── alias.go           # public surface (필수)
│   ├── handler/           # 인바운드 어댑터
│   ├── app/               # 애플리케이션 서비스
│   ├── core/model/        # 도메인 모델 (필수)
│   ├── core/repo/         # 레포지토리 인터페이스
│   ├── core/svc/          # 도메인 서비스 인터페이스
│   ├── event/             # 도메인 이벤트
│   └── infra/             # 아웃바운드 어댑터
├── app/                   # 컴포지션 루트 (DI 와이어링)
├── server/
│   └── http/              # 트랜스포트 레이어 (http, grpc, …)
├── orchestration/         # 크로스 도메인 조정
└── pkg/                   # 공유 유틸리티
```

- `alias.go` 필수, domain root에 다른 .go 금지
- `core/model/`에 최소 1개 .go 필수
- `core/*`, `event`는 `internal/pkg` import 금지
- interface는 `core/repo/`에만 정의
- `internal/app/` (컴포지션 루트): 무제한 import 가능
- `internal/server/<proto>/` (트랜스포트): `internal/app/`, `internal/pkg/`, 형제 트랜스포트만 import 가능. 위반 규칙: `isolation.transport-imports-domain`, `isolation.transport-imports-orchestration`, `isolation.transport-imports-unclassified`

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
arch := core.Architecture{
    Layers: core.LayerModel{
        Sublayers: []string{"api", "logic", "data"},
        Direction: map[string][]string{
            "api": {"logic"}, "logic": {"data"}, "data": {},
        },
        InternalTopLevel: map[string]bool{
            "module": true,
            "pkg":    true,
        },
    },
    Layout: core.LayoutModel{
        DomainDir: "module",
        SharedDir: "pkg",
    },
    Naming: core.NamingPolicy{
        BannedPkgNames: []string{"util", "common", "misc", "helper", "shared", "services"},
        LegacyPkgNames: []string{"router", "bootstrap"},
        AliasFileName:  "alias.go",
    },
    Structure: core.StructurePolicy{
        RequireAlias: false,
        RequireModel: false,
    },
}
```

전체 필드: `LayerModel.Sublayers`, `LayerModel.Direction`, `LayerModel.PkgRestricted`,
`LayerModel.InternalTopLevel`, `LayerModel.LayerDirNames`, `LayerModel.PortLayers`,
`LayerModel.ContractLayers`, `LayoutModel.InternalRoot` (default `"internal"`,
`"packages"`/`"src"` 등 비표준 패키지 루트 지원), `LayoutModel.DomainDir`,
`LayoutModel.OrchestrationDir`, `LayoutModel.SharedDir`, `LayoutModel.AppDir`,
`LayoutModel.ServerDir`, `NamingPolicy.BannedPkgNames`, `NamingPolicy.LegacyPkgNames`,
`NamingPolicy.AliasFileName`, `StructurePolicy.RequireAlias`,
`StructurePolicy.RequireModel`, `StructurePolicy.ModelPath`,
`StructurePolicy.DTOAllowedLayers`, `StructurePolicy.TypePatterns`,
`StructurePolicy.InterfacePatternExclude`

검증은 `arch.Validate()` 또는 `core.Validate(arch)`로 수행한다. 빈 `InternalRoot`는 `cloneArchitecture` 시점에 `"internal"`로 정규화되므로 룰에서는 항상 비어있지 않은 값을 읽는다.

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
ctx := core.NewContext(pkgs, "", "", presets.DDD(), []string{"internal/legacy/..."})
core.Run(ctx, presets.RecommendedDDD(),
    core.WithSeverityOverride("isolation.cross-domain", core.Warning))
```

Exclude 패턴은 정규화 후 매칭됨: `/internal/foo`, `internal/foo`, `internal/foo/`, `./internal/foo` 모두 동일.

### Meta Violations

런타임이 환경 이슈를 알리는 `meta.*` violation. `(Rule, Message)` pair로 dedup.

| ID | Severity | 의미 |
|---|---|---|
| `meta.no-matching-packages` | Warning | 모듈 path가 로드된 패키지와 매칭 안 됨 |
| `meta.layout-not-supported` | Warning | 레이아웃 의존 룰이 `<root>/<InternalRoot>/`를 못 찾음 |
| `meta.rule-panic` | Error | 룰 panic 발생; 다른 룰은 계속 실행 |
| `meta.unknown-violation-id` | per rule | 룰이 미선언 violation ID emit |

### Tx Boundary (opt-in)

트랜잭션 시작 위치와 tx 타입 누설을 차단합니다. 프로젝트마다 DB/SDK 조합이 다르므로
완전 옵트인입니다. `StartSymbols`, `Types` 둘 다 비어 있으면 해당 검사는 건너뜁니다.

```go
ruleset := presets.RecommendedDDD().With(tx.New(tx.Config{
    StartSymbols:  []string{"database/sql.(*DB).BeginTx"},
    Types:         []string{"database/sql.Tx"},
    AllowedLayers: []string{"app"}, // default when empty
}))
```

위반 규칙 ID: `tx.start-outside-allowed-layer`, `tx.type-in-signature`.

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

Repository 포트 interface(이름이 `Repository`/`Repo`로 끝나는 것)는 **`core/repo/`에만** 정의 가능. Consumer-defined interface(사용처에서 선언하는 Go 관례)는 `handler/`, `app/`, `svc/` 어디든 허용. cross-domain 데이터 필요 시 → `orchestration/handler/`로 이동.

| 우회 시도 | 잡는 규칙 |
|----------|----------|
| handler/app/svc에서 `*Repository`/`*Repo` interface 정의 | `structure.interface-placement` |
| handler/app/svc에서 `type X = otherdomain.Repo` alias | `structure.interface-placement` |
| alias.go에서 interface 직접 정의 | `structure.domain-alias-no-interface` |
| alias.go에서 repo/svc 타입 re-export | `structure.domain-alias-contract-reexport` |

---

## Vibe-coding Smell (Warning만)

빌드를 깨지 않고 개발자에게 알리는 **Warning 카테고리** 룰. 강제(error)가 아니라 *고지*다.

| 룰 | 잡는 패턴 |
|----|----------|
| `interface.container-only` | 패키지에서 선언된 named interface가 struct field 타입으로만 쓰이고 함수 파라미터/반환에 한 번도 안 쓰임. wiring 레이어가 값을 들기 위해 만든 임시 컨테이너 interface 패턴을 잡는다. `interfaces.WithSeverity(core.Error)`로 hard rule 승격 가능. |
| `setter.forbidden` | 포인터 리시버를 가진 내보내기 세터 메서드(`Set*`, 매개변수 1개 이상)를 검출. **권장 수정**: 의존성을 생성자의 명시적 파라미터로 추가 (`NewService(..., dep)`). `With*` 옵션은 정말로 선택적이고 여러 조합이 필요한 경우에만 사용 — 설정류 옵션에도 setter는 대체로 맞지 않음. 플루언트 빌더(리시버 타입 반환 메서드), 테스트 파일, `testdata/`·`mocks/` 하위 패키지는 자동 제외. `types.NewNoSetter()`. |

## Cross-Domain Anonymous Interface (Hard rule, Error)

| 룰 | 잡는 패턴 |
|----|----------|
| `interface.cross-domain-anonymous` | 도메인 외부 *그리고 orchestration 외부*에서 선언된 anonymous interface가 method signature에 다른 도메인 타입을 참조하면 위반. cmd/ 또는 internal/pkg/ 같은 wiring 코드가 도메인 타입에 대해 inline ad-hoc 추상화를 선언하는 패턴을 잡는다. **Severity: Error** — cross-domain 추상화는 orchestration 패키지가 소유한다는 컨벤션을 강제. **fix: 어댑터를 `internal/orchestration/`으로 이동하고 wiring 코드는 orchestration 생성자를 호출**. orchestration 패키지(서브패키지 포함)는 by-design exempt. |

---

## Existing Project with Violations

1. `architecture_test.go` 생성 후 실행 → 전체 위반 확인
2. `core.NewContext(..., exclude)`로 마이그레이션 중 경로 제외
3. 점진적으로 위반 수정, exclude 제거
4. 전환기에는 `core.WithSeverityOverride(..., core.Warning)` 사용
