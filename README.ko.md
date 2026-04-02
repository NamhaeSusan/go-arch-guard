# go-arch-guard

[![CI](https://github.com/NamhaeSusan/go-arch-guard/actions/workflows/ci.yml/badge.svg)](https://github.com/NamhaeSusan/go-arch-guard/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/NamhaeSusan/go-arch-guard/branch/main/graph/badge.svg)](https://codecov.io/gh/NamhaeSusan/go-arch-guard)
[![Go Report Card](https://goreportcard.com/badge/github.com/NamhaeSusan/go-arch-guard)](https://goreportcard.com/report/github.com/NamhaeSusan/go-arch-guard)

`go test`로 Go 프로젝트의 아키텍처 가드레일을 적용하는 도구이며, 특히 AI 코딩 에이전트와 빠르게 움직이는 팀에 맞춰 설계되었습니다.

격리, 레이어 방향, 구조, 네이밍, 블래스트 반경 규칙을 정의하고, 프로젝트 형태가 벗어나면 일반 테스트에서 실패시킵니다. **DDD**, **Clean Architecture**, **Layered**, **Hexagonal**, **Modular Monolith**, **Consumer/Worker**, **Batch**, **Event-Driven Pipeline** 프리셋을 기본 제공하며, 완전한 커스텀 아키텍처 모델도 지원합니다. 별도 CLI나 설정 포맷 없이, Go 테스트만으로 동작합니다.

AI 에이전트 친화적인 기본 surface:

- `scaffold.ArchitectureTest(...)` — 바로 붙여 넣을 수 있는 `architecture_test.go` 생성
- `rules.RunAll(...)` — 권장 rule 묶음을 한 번에 실행
- `report.MarshalJSONReport(...)` — 봇과 자동 수정 루프가 읽기 쉬운 JSON 출력

## 왜 필요한가

아키텍처는 보통 깊은 이론적 위반이 아니라 몇 가지 큰 실수를 통해 무너집니다:

- 크로스 도메인 import
- 숨겨진 컴포지션 루트
- 패키지 배치 드리프트
- 의도한 프로젝트 형태를 깨는 네이밍

`go-arch-guard`는 정적 분석으로 이런 큰 실수를 조기에 잡습니다. AI 에이전트가 손쉽게 스캐폴딩하고 유지할 수 있을 만큼 단순하게 설계하면서도, 사람이 경계를 검토하기에 충분한 가드레일을 제공하는 데 초점을 둡니다.

## 설치

```bash
go get github.com/NamhaeSusan/go-arch-guard
```

## 빠른 시작

### 프리셋 템플릿 생성

AI 에이전트나 스캐폴딩 도구라면 아래 예제를 손으로 복사하기보다
바로 `architecture_test.go` 템플릿을 생성할 수 있습니다:

```go
import "github.com/NamhaeSusan/go-arch-guard/scaffold"

src, err := scaffold.ArchitectureTest(
    scaffold.PresetHexagonal,
    scaffold.ArchitectureTestOptions{PackageName: "myapp_test"},
)
```

`PackageName`은 유효한 Go 패키지 식별자여야 합니다. 하이픈이 있는
module basename을 그대로 쓰면 안 됩니다.

사용 가능한 프리셋: `PresetDDD`, `PresetCleanArch`, `PresetLayered`,
`PresetHexagonal`, `PresetModularMonolith`, `PresetConsumerWorker`, `PresetBatch`,
`PresetEventPipeline`.

### 권장 shortcut

각 rule을 직접 append 하지 않고 권장 기본 묶음을 한 번에 실행하려면:

```go
violations := rules.RunAll(pkgs, "", "")
report.AssertNoViolations(t, violations)
```

기본값이 아닌 모델이나 severity/exclude 옵션이 필요할 때만 `opts...`를 넘기면 됩니다.

### 개별 rule 제어 (DDD 예시)

각 체크를 세밀하게 제어하려면 직접 조합합니다:

```go
func TestArchitecture(t *testing.T) {
    pkgs, err := analyzer.Load(".", "internal/...", "cmd/...")
    if err != nil {
        t.Log(err)
    }
    if len(pkgs) == 0 {
        t.Fatalf("no packages loaded: %v", err)
    }

    t.Run("domain isolation", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckDomainIsolation(pkgs, "", ""))
    })
    t.Run("layer direction", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckLayerDirection(pkgs, "", ""))
    })
    t.Run("naming", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckNaming(pkgs))
    })
    t.Run("structure", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckStructure("."))
    })
    t.Run("blast radius", func(t *testing.T) {
        report.AssertNoViolations(t, rules.AnalyzeBlastRadius(pkgs, "", ""))
    })
}
```

다른 프리셋을 쓸 때는 모델 함수만 교체하고 `opts...`를 전달합니다:

```go
m := rules.CleanArch() // 또는 Layered(), Hexagonal(), ModularMonolith(), ConsumerWorker(), Batch(), EventPipeline()
opts := []rules.Option{rules.WithModel(m)}

rules.CheckDomainIsolation(pkgs, "", "", opts...)
rules.CheckLayerDirection(pkgs, "", "", opts...)
// ... 모든 Check* 함수에 동일하게 opts... 전달
```

### 커스텀 모델

```go
m := rules.NewModel(
    rules.WithDomainDir("module"),
    rules.WithSharedDir("lib"),
    rules.WithSublayers([]string{"api", "logic", "data"}),
    rules.WithDirection(map[string][]string{
        "api":   {"logic"},
        "logic": {"data"},
        "data":  {},
    }),
)
opts := []rules.Option{rules.WithModel(m)}
```

실행:

```bash
go test -run TestArchitecture -v
```

`module`과 `root`에 빈 문자열을 전달하면 로드된 패키지에서 자동 추출합니다.

## 아키텍처 모델

### 프리셋

| 프리셋 | 서브레이어 | 방향 | Alias 필수 | 모델 필수 |
|--------|-----------|------|:-:|:-:|
| `DDD()` | handler, app, core/model, core/repo, core/svc, event, infra | handler→app→core/\*, infra→core/repo | O | O |
| `CleanArch()` | handler, usecase, entity, gateway, infra | handler→usecase→entity+gateway, infra→gateway | X | X |
| `Layered()` | handler, service, repository, model | handler→service→repository+model, repository→model | X | X |
| `Hexagonal()` | handler, usecase, port, domain, adapter | handler→usecase→port+domain, adapter→port+domain | X | X |
| `ModularMonolith()` | api, application, core, infrastructure | api→application→core, infrastructure→core | X | X |
| `ConsumerWorker()` | worker, service, store, model | worker→service→store→model | X | X |
| `Batch()` | job, service, store, model | job→service→store→model | X | X |
| `EventPipeline()` | command, aggregate, event, projection, eventstore, readstore, model | command→aggregate+eventstore, aggregate→event, projection→event/readstore | X | X |

### DDD 레이아웃

```text
internal/
├── domain/
│   └── order/
│       ├── alias.go              # 공개 surface (필수)
│       ├── handler/http/         # 인바운드 어댑터
│       ├── app/                  # 애플리케이션 서비스
│       ├── core/
│       │   ├── model/            # 도메인 모델 (필수)
│       │   ├── repo/             # 레포지토리 인터페이스
│       │   └── svc/              # 도메인 서비스 인터페이스
│       ├── event/                # 도메인 이벤트
│       └── infra/persistence/    # 아웃바운드 어댑터
├── orchestration/                # 크로스 도메인 조율
└── pkg/                          # 공유 유틸리티
```

DDD 레이어 방향:

| 출발 | import 가능 대상 |
|------|-----------------|
| `handler` | `app` |
| `app` | `core/model`, `core/repo`, `core/svc`, `event` |
| `core` | `core/model` |
| `core/model` | 없음 |
| `core/repo` | `core/model` |
| `core/svc` | `core/model` |
| `event` | `core/model` |
| `infra` | `core/repo`, `core/model`, `event` |

### Clean Architecture 레이아웃

```text
internal/
├── domain/
│   └── product/
│       ├── handler/              # 인터페이스 어댑터 (컨트롤러)
│       ├── usecase/              # 애플리케이션 비즈니스 규칙
│       ├── entity/               # 엔터프라이즈 비즈니스 규칙
│       ├── gateway/              # 데이터 접근 인터페이스
│       └── infra/                # 프레임워크 & 드라이버
├── orchestration/
└── pkg/
```

Clean Architecture 레이어 방향:

| 출발 | import 가능 대상 |
|------|-----------------|
| `handler` | `usecase` |
| `usecase` | `entity`, `gateway` |
| `entity` | 없음 |
| `gateway` | `entity` |
| `infra` | `gateway`, `entity` |

### Layered (Spring 스타일) 레이아웃

```text
internal/
├── domain/
│   └── order/
│       ├── handler/              # HTTP/gRPC 핸들러
│       ├── service/              # 비즈니스 로직
│       ├── repository/           # 데이터 접근
│       └── model/                # 도메인 모델
├── orchestration/
└── pkg/
```

Layered 레이어 방향:

| 출발 | import 가능 대상 |
|------|-----------------|
| `handler` | `service` |
| `service` | `repository`, `model` |
| `repository` | `model` |
| `model` | 없음 |

### Hexagonal (포트 & 어댑터) 레이아웃

```text
internal/
├── domain/
│   └── order/
│       ├── handler/              # 드라이빙 어댑터 (HTTP, gRPC)
│       ├── usecase/              # 애플리케이션 로직
│       ├── port/                 # 인터페이스 (인바운드 + 아웃바운드)
│       ├── domain/               # 엔티티, 값 객체
│       └── adapter/              # 드리븐 어댑터 (DB, 메시징)
├── orchestration/
└── pkg/
```

Hexagonal 레이어 방향:

| 출발 | import 가능 대상 |
|------|-----------------|
| `handler` | `usecase` |
| `usecase` | `port`, `domain` |
| `port` | `domain` |
| `domain` | 없음 |
| `adapter` | `port`, `domain` |

### Modular Monolith 레이아웃

```text
internal/
├── domain/
│   └── order/
│       ├── api/                  # 모듈 공개 인터페이스
│       ├── application/          # 유즈케이스
│       ├── core/                 # 엔티티, 값 객체
│       └── infrastructure/       # DB, 외부 서비스
├── orchestration/
└── pkg/
```

Modular Monolith 레이어 방향:

| 출발 | import 가능 대상 |
|------|-----------------|
| `api` | `application` |
| `application` | `core` |
| `core` | 없음 |
| `infrastructure` | `core` |

### Consumer/Worker 레이아웃 (플랫)

도메인 중심 프리셋과 달리, Consumer/Worker 프리셋은 **플랫 레이아웃**을 사용합니다 —
레이어가 `domain/` 디렉토리 없이 `internal/` 바로 아래에 위치합니다.

```text
internal/
├── worker/            # worker_order.go, worker_payment.go
├── service/           # 비즈니스 로직
├── store/             # 영속성 (DB, 외부 API)
├── model/             # 데이터 구조체
└── pkg/               # 공유 인프라 (consumer 라이브러리, 로깅)
    └── consumer/
```

Consumer/Worker 방향:

| from | 허용된 import |
|------|--------------|
| `worker` | `service`, `model` |
| `service` | `store`, `model` |
| `store` | `model` |
| `model` | 없음 |

`model`을 제외한 모든 레이어는 `pkg/`를 import 할 수 있습니다.

**타입 패턴 강제:** `worker/` 내 `worker_*.go` 파일은 대응하는 exported 타입과
`Process` 메서드를 반드시 정의해야 합니다:
- `worker_order.go` → `OrderWorker` 타입 + `Process` 메서드 필수
- `worker_payment.go` → `PaymentWorker` 타입 + `Process` 메서드 필수

도메인 격리 규칙은 적용되지 않습니다.

### Batch 레이아웃 (플랫)

Batch 프리셋은 Consumer/Worker와 동일한 플랫 레이아웃을 사용하며,
cron/스케줄러 기반 배치 처리의 진입점 레이어로 `job/`을 사용합니다.

```text
internal/
├── job/               # job_expire_files.go, job_cleanup_trash.go
├── service/           # 비즈니스 로직
├── store/             # 영속성 (DB, 외부 API)
├── model/             # 데이터 구조체
└── pkg/               # 공유 인프라 (batchutil, 로깅)
```

Batch 방향:

| from | 허용된 import |
|------|--------------|
| `job` | `service`, `model` |
| `service` | `store`, `model` |
| `store` | `model` |
| `model` | 없음 |

`model`을 제외한 모든 레이어는 `pkg/`를 import 할 수 있습니다.

**타입 패턴 강제:** `job/` 내 `job_*.go` 파일은 대응하는 exported 타입과
`Run` 메서드를 반드시 정의해야 합니다:
- `job_expire_files.go` → `ExpireFilesJob` 타입 + `Run` 메서드 필수
- `job_cleanup_trash.go` → `CleanupTrashJob` 타입 + `Run` 메서드 필수

도메인 격리 규칙은 적용되지 않습니다.

### Event-Driven Pipeline 레이아웃 (플랫)

Event-Driven Pipeline 프리셋은 이벤트 소싱 / CQRS 프로젝트를 위한 플랫 레이아웃을 사용하며,
커맨드, 애그리거트, 이벤트, 프로젝션, 스토어를 위한 전용 디렉토리를 제공합니다.

```text
internal/
├── command/          # 커맨드 핸들러 (command_create_order.go)
├── aggregate/        # 애그리거트 루트 (aggregate_order.go)
├── event/            # 도메인 이벤트
├── projection/       # 읽기 모델 프로젝터
├── eventstore/       # 이벤트 영속성
├── readstore/        # 읽기 모델 영속성
├── model/            # 공유 값 객체 / DTO
└── pkg/              # 공유 인프라 (eventbus, 로깅)
```

Event-Driven Pipeline 방향:

| from | 허용된 import |
|------|--------------|
| `command` | `aggregate`, `eventstore`, `model` |
| `aggregate` | `event`, `model` |
| `event` | `model` |
| `projection` | `event`, `readstore`, `model` |
| `eventstore` | `event`, `model` |
| `readstore` | `model` |
| `model` | 없음 |

`event`와 `model`을 제외한 모든 레이어는 `pkg/`를 import 할 수 있습니다.

**타입 패턴 강제:** `command/` 내 `command_*.go` 파일은 대응하는 exported 타입과
`Execute` 메서드를 반드시 정의해야 합니다:
- `command_create_order.go` → `CreateOrderCommand` 타입 + `Execute` 메서드 필수

`aggregate/` 내 `aggregate_*.go` 파일은 대응하는 exported 타입과
`Apply` 메서드를 반드시 정의해야 합니다:
- `aggregate_order.go` → `OrderAggregate` 타입 + `Apply` 메서드 필수

도메인 격리 규칙은 적용되지 않습니다.

### 커스텀 모델

DDD 기본값에서 시작하여 필요한 부분만 오버라이드:

```go
m := rules.NewModel(
    rules.WithDomainDir("module"),          // internal/module/
    rules.WithOrchestrationDir("workflow"), // internal/workflow/
    rules.WithSharedDir("lib"),             // internal/lib/
    rules.WithSublayers([]string{"api", "logic", "data"}),
    rules.WithDirection(map[string][]string{
        "api":   {"logic"},
        "logic": {"data"},
        "data":  {},
    }),
    rules.WithRequireAlias(false),
    rules.WithRequireModel(false),
)
```

전체 모델 옵션:

| 옵션 | 설명 |
|------|------|
| `WithSublayers([]string{...})` | 인식할 서브레이어 이름 |
| `WithDirection(map[string][]string{...})` | 허용 import 방향 매트릭스 |
| `WithPkgRestricted(map[string]bool{...})` | 공유 패키지 import 금지 서브레이어 |
| `WithDomainDir("domain")` | 도메인 최상위 디렉토리명 |
| `WithOrchestrationDir("orchestration")` | 오케스트레이션 최상위 디렉토리명 |
| `WithSharedDir("pkg")` | 공유 패키지 최상위 디렉토리명 |
| `WithRequireAlias(bool)` | 도메인 루트에 alias 파일 필수 여부 |
| `WithAliasFileName("alias.go")` | alias 파일명 |
| `WithRequireModel(bool)` | 도메인에 모델 디렉토리 필수 여부 |
| `WithModelPath("core/model")` | 도메인 모델 디렉토리 경로 |
| `WithDTOAllowedLayers([]string{...})` | DTO 허용 서브레이어 |
| `WithBannedPkgNames([]string{...})` | internal/ 하위 금지 패키지명 |
| `WithLegacyPkgNames([]string{...})` | 마이그레이션 경고 패키지명 |
| `WithLayerDirNames(map[string]bool{...})` | 네이밍 체크 시 "레이어" 디렉토리 이름 |

## 격리 규칙

`rules.CheckDomainIsolation(pkgs, module, root, opts...)`

| 규칙 | 의미 |
|------|------|
| `isolation.cross-domain` | 도메인 A는 도메인 B를 import할 수 없음 |
| `isolation.cmd-deep-import` | `cmd/`는 도메인 루트만 import 가능 |
| `isolation.orchestration-deep-import` | 오케스트레이션은 도메인 루트만 import 가능 |
| `isolation.pkg-imports-domain` | 공유 패키지는 도메인을 import할 수 없음 |
| `isolation.pkg-imports-orchestration` | 공유 패키지는 오케스트레이션을 import할 수 없음 |
| `isolation.domain-imports-orchestration` | 도메인은 오케스트레이션을 import할 수 없음 |
| `isolation.internal-imports-orchestration` | cmd/orchestration 외 패키지는 오케스트레이션을 import할 수 없음 |
| `isolation.internal-imports-domain` | 미등록 내부 패키지는 도메인을 import할 수 없음 |

Import 매트릭스:

| from → to | 도메인 루트 | 도메인 하위 | 오케스트레이션 | 공유 패키지 |
|-----------|:-:|:-:|:-:|:-:|
| **같은 도메인** | O | O | X | O |
| **다른 도메인** | X | X | X | O |
| **오케스트레이션** | O | X | O | O |
| **cmd** | O | X | O | O |
| **공유 패키지** | X | X | X | O |

## 레이어 방향 규칙

`rules.CheckLayerDirection(pkgs, module, root, opts...)`

| 규칙 | 의미 |
|------|------|
| `layer.direction` | 허용된 레이어 방향 위반 import |
| `layer.inner-imports-pkg` | 내부 레이어가 공유 패키지를 import (`PkgRestricted` 제어) |
| `layer.unknown-sublayer` | 도메인에서 알 수 없는 서브레이어 |

방향 매트릭스는 아키텍처 모델에서 정의됩니다. `WithDirection`으로 완전 커스터마이징 가능.

## 구조 규칙

`rules.CheckStructure(root, opts...)`

| 규칙 | 의미 |
|------|------|
| `structure.internal-top-level` | `internal/` 아래 허용된 최상위 패키지만 |
| `structure.banned-package` | 금지된 패키지명 (기본: `util`, `common`, `misc`, `helper`, `shared`, `services`) |
| `structure.legacy-package` | 마이그레이션 필요 레거시 패키지 |
| `structure.misplaced-layer` | 도메인 슬라이스 외부의 `app`/`handler`/`infra` |
| `structure.middleware-placement` | `middleware/`는 공유 패키지에만 |
| `structure.domain-root-alias-required` | 도메인 루트에 alias 파일 필수 (DDD만) |
| `structure.domain-root-alias-package` | alias 파일 패키지명이 디렉토리와 일치 |
| `structure.domain-root-alias-only` | 도메인 루트에 alias 파일만 허용 |
| `structure.domain-alias-no-interface` | alias 파일에서 interface re-export 금지 |
| `structure.domain-model-required` | 도메인에 모델 디렉토리 필수 (DDD만) |
| `structure.dto-placement` | DTO 파일은 handler/app에만 |

## 네이밍 규칙

`rules.CheckNaming(pkgs, opts...)`

| 규칙 | 의미 |
|------|------|
| `naming.no-stutter` | exported 타입이 패키지 이름 반복 |
| `naming.no-impl-suffix` | exported 타입이 `Impl`로 끝남 |
| `naming.snake-case-file` | 파일명이 snake_case가 아님 |
| `naming.repo-file-interface` | repo/ 파일에 매칭 interface 없음 |
| `naming.no-layer-suffix` | 파일명이 레이어 이름 불필요 반복 |
| `naming.domain-interface-repo-only` | repo 서브레이어 외부에서 도메인 interface 정의 (DDD만) |
| `naming.no-handmock` | 테스트에서 hand-rolled mock/fake/stub 정의 |
| `naming.type-pattern-mismatch` | `worker_*.go`/`job_*.go` 파일에 매칭 타입 미정의 |
| `naming.type-pattern-missing-method` | 타입에 필수 메서드 없음 (예: `Process`, `Run`, `Execute`, `Apply`) |

## 블래스트 반경

`rules.AnalyzeBlastRadius(pkgs, module, root, opts...)`

IQR 기반 통계적 이상치로 비정상 커플링 패키지를 탐지합니다. 기본 severity: Warning. 내부 패키지 5개 미만이면 스킵.

| 규칙 | 의미 |
|------|------|
| `blast-radius.high-coupling` | transitive dependents가 통계적 이상치 |

| 메트릭 | 정의 |
|--------|------|
| Ca (Afferent Coupling) | 이 패키지를 import하는 패키지 수 |
| Ce (Efferent Coupling) | 이 패키지가 import하는 패키지 수 |
| Instability | Ce / (Ca + Ce) |
| Transitive Dependents | BFS로 추적한 전체 역방향 도달 가능 집합 |

## 옵션

```go
// 경고로 다운그레이드
rules.WithSeverity(rules.Warning)

// 마이그레이션 중 경로 제외
rules.WithExclude("internal/legacy/...")

// 아키텍처 모델 적용
rules.WithModel(rules.CleanArch())
```

## TUI 뷰어

```bash
go run github.com/NamhaeSusan/go-arch-guard/cmd/tui .
```

건강 상태 트리 색상, 커플링 메트릭, 위반 상세, 검색/필터 (`/`), 키보드 탐색 지원.

## API 레퍼런스

| 함수 | 설명 |
|------|------|
| `analyzer.Load(dir, patterns...)` | 분석용 Go 패키지 로드 |
| `rules.CheckDomainIsolation(pkgs, module, root, opts...)` | 크로스 도메인 경계 검사 |
| `rules.CheckLayerDirection(pkgs, module, root, opts...)` | 도메인 내 방향 검사 |
| `rules.CheckNaming(pkgs, opts...)` | 네이밍 검사 |
| `rules.CheckStructure(root, opts...)` | 파일시스템 구조 검사 |
| `rules.AnalyzeBlastRadius(pkgs, module, root, opts...)` | 커플링 이상치 탐지 |
| `rules.RunAll(pkgs, module, root, opts...)` | 권장 기본 rule 묶음 실행 |
| `report.AssertNoViolations(t, violations)` | Error 위반 시 테스트 실패 |
| `report.BuildJSONReport(violations)` | 기계가 읽기 쉬운 JSON 리포트 구성 |
| `report.MarshalJSONReport(violations)` | JSON 리포트 직렬화 |
| `report.WriteJSONReport(w, violations)` | JSON 리포트 쓰기 |
| `scaffold.ArchitectureTest(preset, opts)` | 프리셋별 `architecture_test.go` 템플릿 생성 |
| `rules.DDD()` | DDD 아키텍처 모델 (기본값) |
| `rules.CleanArch()` | Clean Architecture 모델 |
| `rules.Layered()` | Spring 스타일 레이어드 모델 |
| `rules.Hexagonal()` | 포트 & 어댑터 모델 |
| `rules.ModularMonolith()` | 모듈 기반 레이어드 모델 |
| `rules.ConsumerWorker()` | Consumer/Worker 플랫 레이아웃 모델 |
| `rules.Batch()` | Batch 플랫 레이아웃 모델 |
| `rules.EventPipeline()` | 이벤트 소싱 / CQRS 플랫 레이아웃 모델 |
| `rules.CheckTypePatterns(pkgs, opts...)` | AST 기반 타입 패턴 강제 |
| `rules.NewModel(opts...)` | 커스텀 모델 빌더 |
| `rules.WithModel(m)` | 커스텀 모델 적용 |
| `rules.WithSeverity(rules.Warning)` | 경고로 다운그레이드 |
| `rules.WithExclude("path/...")` | 하위 트리 건너뛰기 |

## 기계 친화적인 JSON 출력

CI, 봇, AI 수정 루프에서는 같은 위반 목록을 JSON으로 내보낼 수 있습니다:

```go
import "github.com/NamhaeSusan/go-arch-guard/report"

data, err := report.MarshalJSONReport(violations)
if err != nil {
    return err
}
fmt.Println(string(data))
```

## Claude Code 플러그인

```text
/plugin marketplace add NamhaeSusan/go-arch-guard
/plugin install go-arch-guard@go-arch-guard-marketplace
```

## 외부 Import 위생 — 이 라이브러리가 아닌 AI 도구 지침으로 강제

`go-arch-guard`는 **프로젝트 내부** import만 검사합니다. 외부 의존성 위생은 AI 도구 지침과 코드 리뷰로 강제하세요.

**AI 도구의 시스템 프롬프트에 복사:**

```text
# 외부 Import 제약 (go-arch-guard는 이를 강제하지 않음)

- core/model, core/repo, core/svc, event — stdlib만, 서드파티 금지
- handler — HTTP/gRPC 프레임워크 허용, 영속성 라이브러리 금지
- infra — 영속성/메시징 라이브러리 허용, HTTP 프레임워크 금지
- app — 일반적으로 자유, 인프라 라이브러리 직접 import 지양
```

## 라이선스

MIT
