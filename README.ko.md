# go-arch-guard

[![CI](https://github.com/NamhaeSusan/go-arch-guard/actions/workflows/ci.yml/badge.svg)](https://github.com/NamhaeSusan/go-arch-guard/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/NamhaeSusan/go-arch-guard/branch/main/graph/badge.svg)](https://codecov.io/gh/NamhaeSusan/go-arch-guard)
[![Go Report Card](https://goreportcard.com/badge/github.com/NamhaeSusan/go-arch-guard)](https://goreportcard.com/report/github.com/NamhaeSusan/go-arch-guard)

`go test`로 Go 프로젝트의 아키텍처 가드레일을 적용하는 도구이며, 특히 AI 코딩 에이전트와 빠르게 움직이는 팀에 맞춰 설계되었습니다.

격리, 레이어 방향, 구조, 네이밍, 블래스트 반경 규칙을 정의하고, 프로젝트 형태가 벗어나면 일반 테스트에서 실패시킵니다. **DDD**, **Clean Architecture**, **Layered**, **Hexagonal**, **Modular Monolith**, **Consumer/Worker**, **Batch**, **Event-Driven Pipeline** 프리셋을 기본 제공하며, 완전한 커스텀 아키텍처 모델도 지원합니다. 별도 CLI나 설정 포맷 없이, Go 테스트만으로 동작합니다.

AI 에이전트 친화적인 기본 surface:

- `scaffold.ArchitectureTest(...)` --- 바로 붙여 넣을 수 있는 `architecture_test.go` 생성
- `rules.RunAll(...)` --- 권장 rule 묶음을 한 번에 실행
- `report.MarshalJSONReport(...)` --- 봇과 자동 수정 루프가 읽기 쉬운 JSON 출력

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

## 프리셋

| 프리셋 | 타입 | 서브레이어 | 방향 |
|--------|------|-----------|------|
| `DDD()` | Domain | handler, app, core/model, core/repo, core/svc, event, infra | handler->app->core/\*, infra->core/repo+core/model+event |
| `CleanArch()` | Domain | handler, usecase, entity, gateway, infra | handler->usecase->entity+gateway, infra->gateway+entity |
| `Layered()` | Domain | handler, service, repository, model | handler->service->repository+model |
| `Hexagonal()` | Domain | handler, usecase, port, domain, adapter | handler->usecase->port+domain, adapter->port+domain |
| `ModularMonolith()` | Domain | api, application, core, infrastructure | api->application->core, infrastructure->core |
| `ConsumerWorker()` | Flat | worker, service, store, model | worker→service+model, service→store+model, store→model |
| `Batch()` | Flat | job, service, store, model | job→service+model, service→store+model, store→model |
| `EventPipeline()` | Flat | command, aggregate, event, projection, eventstore, readstore, model | command→aggregate+eventstore+model, aggregate→event+model, projection→event+readstore+model |

Domain 프리셋은 `internal/domain/{name}/{layer}/` 레이아웃을 사용합니다.
Flat 프리셋은 `internal/{layer}/` 레이아웃을 사용합니다 (domain 디렉토리 없음).

전체 레이아웃 다이어그램과 방향 테이블은 [프리셋 상세](docs/presets.md)를 참고하세요.

### 커스텀 모델 옵션

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
| `WithAppDir("app")` | 컴포지션 루트 최상위 디렉토리 (예: `internal/app/`). 이 패키지들은 무제한 import 가능. 비워두면 비활성화. |
| `WithServerDir("server")` | 트랜스포트 레이어 최상위 디렉토리 (예: `internal/server/`). 하위 디렉토리는 모두 트랜스포트로 분류. app+pkg만 import 가능. 비워두면 비활성화. |
| `WithRequireAlias(bool)` | 도메인 루트에 alias 파일 필수 여부 |
| `WithAliasFileName("alias.go")` | alias 파일명 |
| `WithRequireModel(bool)` | 도메인에 모델 디렉토리 필수 여부 |
| `WithModelPath("core/model")` | 도메인 모델 디렉토리 경로 |
| `WithDTOAllowedLayers([]string{...})` | DTO 허용 서브레이어 |
| `WithBannedPkgNames([]string{...})` | internal/ 하위 금지 패키지명 |
| `WithLegacyPkgNames([]string{...})` | 마이그레이션 경고 패키지명 |
| `WithLayerDirNames(map[string]bool{...})` | 네이밍 체크 시 "레이어" 디렉토리 이름 |
| `WithInterfacePatternExclude(map[string]bool{...})` | 인터페이스 패턴 검사 제외 레이어 |
| `WithPortLayers([]string{...})` | 포트 레이어(순수 인터페이스 정의)로 분류할 서브레이어. 비어있지 않으면 authoritative (정확히 일치해야 함). |
| `WithContractLayers([]string{...})` | 계약 레이어(포트 + svc류)로 분류할 서브레이어. `ContractLayers ⊇ PortLayers`; 헬퍼는 두 목록을 union하여 판정한다. |

`NewModel`은 `DDD()` 기본값에서 시작하므로, 커스텀 모델은 별도 지정이 없으면 `PortLayers=["core/repo"]`, `ContractLayers=["core/repo","core/svc"]`를 상속한다. basename fallback(`repo`/`gateway`/`svc`)을 다시 쓰려면 **두 목록을 모두** 명시적으로 비워야 한다: `WithPortLayers(nil), WithContractLayers(nil)`. 한쪽만 비우면 authoritative exact-match 경로가 유지된다 (자세한 내용은 `NewModel` godoc 참조).

## 격리 규칙

`rules.CheckDomainIsolation(pkgs, module, root, opts...)`

도메인 간 누수를 차단합니다. 격리 없이는 도메인 A의 변경이 도메인 B를 조용히 깨뜨릴 수 있으며,
이것이 DDD 프로젝트에서 가장 흔한 의도치 않은 결합 원인입니다.

### `isolation.cross-domain`

도메인은 다른 도메인을 직접 import할 수 없습니다.

```go
// internal/domain/order/app/service.go
package app

import _ "myapp/internal/domain/user/app"  // 위반
```

```go
// 크로스 도메인 조율에는 orchestration 사용
package orchestration

import (
    "myapp/internal/domain/order"
    "myapp/internal/domain/user"
)
```

### `isolation.cmd-deep-import`

`cmd/`는 도메인 루트 패키지(alias)만 import할 수 있고, 하위 패키지는 안 됩니다.

```go
// cmd/server/main.go
import _ "myapp/internal/domain/order/app"  // 너무 깊음

import _ "myapp/internal/domain/order"  // 도메인 루트만 허용
```

### `isolation.orchestration-deep-import`

오케스트레이션은 도메인 루트만 import하여 결합 표면을 최소화해야 합니다.

```go
// internal/orchestration/checkout.go
import _ "myapp/internal/domain/order/app"  // 너무 깊음

import _ "myapp/internal/domain/order"  // 도메인 루트만 허용
```

### `isolation.pkg-imports-domain`

공유 `pkg/`는 어떤 도메인도 import할 수 없습니다 --- 도메인에 무관해야 합니다.

```go
// internal/pkg/logger/logger.go
import _ "myapp/internal/domain/order"  // 위반: pkg가 도메인에 의존
```

### `isolation.pkg-imports-orchestration`

공유 `pkg/`는 오케스트레이션을 import할 수 없습니다.

### `isolation.domain-imports-orchestration`

도메인은 오케스트레이션을 import할 수 없습니다 --- 오케스트레이션이 도메인을 조율하지, 그 반대가 아닙니다.

### `isolation.stray-imports-orchestration`

`cmd/`와 오케스트레이션 자체만 오케스트레이션에 의존할 수 있습니다.

### `isolation.stray-imports-domain`

도메인이 아닌 내부 패키지(orchestration/cmd/pkg/app/transport 제외)는 도메인을 import할 수 없습니다.

### `isolation.transport-imports-domain`

트랜스포트 패키지(`internal/server/<proto>/`)는 도메인 하위 패키지를 직접 import할 수 없습니다.
컴포지션 루트(`internal/app/`)를 통해야 합니다.

### `isolation.transport-imports-orchestration`

트랜스포트 패키지는 오케스트레이션을 직접 import할 수 없습니다.

### `isolation.transport-imports-unclassified`

트랜스포트 패키지는 분류되지 않은 내부 패키지(예: `internal/config`, `internal/bootstrap`)를 import할 수 없습니다.
트랜스포트가 의존하는 모든 것은 `internal/app/` (컴포지션 루트) 또는 `internal/pkg/`를 거쳐야 합니다.

**Import 매트릭스 (DDD + app/server):**

| from | 도메인 루트 | 도메인 하위 | 오케스트레이션 | 공유 패키지 | app | 트랜스포트 |
|------|:-:|:-:|:-:|:-:|:-:|:-:|
| **같은 도메인** | O | O | X | O | X | X |
| **다른 도메인** | X | X | X | O | X | X |
| **오케스트레이션** | O | X | O | O | X | X |
| **cmd** | O | X | O | O | X | X |
| **공유 패키지** | X | X | X | O | X | X |
| **app (컴포지션 루트)** | O | O | O | O | O | X |
| **트랜스포트** | X | X | X | O | O | O |

> **Flat 레이아웃 프리셋** (ConsumerWorker, Batch, EventPipeline): 격리할 도메인이 없으므로
> 격리 규칙이 완전히 스킵됩니다.

## 레이어 방향 규칙

`rules.CheckLayerDirection(pkgs, module, root, opts...)`

레이어 간 역방향 의존성을 차단합니다. 방향 강제 없이는 내부 레이어(model, entity)가
외부 레이어의 import를 점진적으로 축적하여, 추출이나 독립 테스트가 불가능해집니다.

### `layer.direction`

import는 프리셋의 방향 매트릭스에 정의된 허용 방향을 따라야 합니다.

```go
// DDD 프리셋: core/svc는 core/model만 import 가능
package svc // internal/domain/order/core/svc/

import _ "myapp/internal/domain/order/app"  // 역방향

import _ "myapp/internal/domain/order/core/model"  // 허용
```

### `layer.inner-imports-pkg`

`PkgRestricted`로 표시된 내부 레이어는 공유 `pkg/`를 import할 수 없습니다.
핵심 도메인 로직을 인프라 관심사로부터 자유롭게 유지합니다.

```go
// DDD: core/model은 PkgRestricted
package model // internal/domain/order/core/model/

import _ "myapp/internal/pkg/logger"  // model은 자족적이어야 함
```

### `layer.unknown-sublayer`

도메인 하위에 인식된 서브레이어 이름과 일치하지 않는 디렉토리를 감지합니다.

```
internal/domain/order/utils/   "utils"는 인식된 서브레이어가 아님
```

> **Flat 레이아웃 프리셋**: 도메인 내부가 아닌 `internal/` 최상위에서 레이어를 검사합니다.

## 구조 규칙

`rules.CheckStructure(root, opts...)`

바이브 코딩 중 구조적 드리프트를 방지하는 파일시스템 레이아웃 규칙을 강제합니다.

### `structure.internal-top-level`

`internal/` 최상위에는 허용된 디렉토리만 존재할 수 있습니다.

```
// DDD: domain/, orchestration/, pkg/만 허용
internal/
  domain/          허용
  orchestration/   허용
  pkg/             허용
  config/          허용 목록에 없음
```

### `structure.banned-package`

쓰레기통이 되기 쉬운 모호한 패키지명을 차단합니다.

기본 금지 목록: `util`, `common`, `misc`, `helper`, `shared`, `services`

```
internal/domain/order/app/util/   "util"은 금지됨
```

### `structure.legacy-package`

마이그레이션이 필요한 패키지명을 경고합니다: `router`, `bootstrap`

### `structure.misplaced-layer`

레이어 디렉토리(`app`, `handler`, `infra`)는 도메인 슬라이스 안에만 있어야 하며,
internal/ 최상위에 떠 있으면 안 됩니다.

### `structure.middleware-placement`

`middleware/`는 `internal/pkg/middleware/`에 있어야 하며, 도메인에 흩어지면 안 됩니다.

### `structure.domain-alias-exists` (DDD만)

각 도메인 루트는 공개 API surface로 `alias.go` 파일을 정의해야 합니다.

### `structure.domain-alias-package`

alias 파일의 패키지 이름은 디렉토리 이름과 일치해야 합니다.

### `structure.domain-alias-exclusive`

도메인 루트 디렉토리는 `alias.go`만 포함할 수 있습니다 --- 나머지 코드는 서브레이어에 넣어야 합니다.

### `structure.domain-alias-no-interface`

alias 파일은 인터페이스를 직접 정의할 수 없습니다 --- 크로스 도메인 계약이 누수됩니다.

### `structure.domain-alias-contract-reexport`

alias 파일은 계약 서브레이어(repo/svc)의 타입을 re-export할 수 없습니다 --- 크로스 도메인 숨은 의존성이 생깁니다.

### `structure.domain-model-required` (DDD만)

각 도메인에 `core/model/` 디렉토리와 하나 이상의 Go 파일이 있어야 합니다.

### `structure.dto-placement`

DTO 파일(`dto.go`, `*_dto.go`)은 허용된 레이어(handler, app)에만 존재할 수 있습니다.

## 네이밍 규칙

`rules.CheckNaming(pkgs, opts...)`

코드베이스를 일관되고 grep 친화적으로 유지하는 Go 네이밍 규칙을 강제합니다.

### `naming.no-stutter`

exported 타입은 패키지 이름을 반복하면 안 됩니다.

```go
package repo

type RepoOrder struct{}  // 더듬거림: repo.RepoOrder
type Order struct{}      // 깔끔: repo.Order
```

### `naming.no-impl-suffix`

exported 타입은 `Impl`로 끝나면 안 됩니다. unexported 타입을 대신 사용하세요.

```go
type OrderServiceImpl struct{}  // Impl 접미사
type orderService struct{}      // unexported
```

### `naming.snake-case-file`

모든 Go 파일명은 snake_case여야 합니다.

```
OrderService.go   위반
order_service.go  올바름
```

### `structure.repo-file-interface`

`repo/` (또는 `core/repo/`) 파일은 파일명과 일치하는 인터페이스를 포함해야 합니다.

```go
// repo/order.go는 다음을 정의해야:
type Order interface { ... }  // 파일명과 일치
```

### `structure.repo-file-extra-interface`

`repo/` 파일에는 인터페이스가 정확히 1개만 있어야 합니다. 추가 인터페이스는 별도 파일로 분리하세요.

```go
// repo/review.go
type Review interface { Find() }   // 올바름
type Helper interface { Assist() } // 위반: helper.go로 이동
```

### `interface.too-many-methods`

repo 인터페이스의 메서드 수가 `WithMaxRepoInterfaceMethods`로 설정한 상한을 초과하면 위반입니다. 기본 비활성.

```go
rules.CheckNaming(pkgs, rules.WithMaxRepoInterfaceMethods(10))
```

```go
// repo/review.go
type Review interface {
    // 메서드 11개 --- 위반 (max 10)
}
```

### `naming.no-layer-suffix`

파일명은 레이어 이름을 불필요하게 반복하면 안 됩니다.

```
// service/ 디렉토리 안에서:
order_service.go  "_service" 접미사 불필요
order.go          올바름
```

### `structure.interface-placement` (DDD만)

Repository 포트 인터페이스(이름이 `Repository` 또는 `Repo`로 끝나는 것)는
`core/repo/`에만 정의해야 합니다. Consumer-defined interface(사용처에서
작은 인터페이스를 선언하는 Go 관례)는 `handler/`, `app/`, `svc/` 등
사용처 어디든 허용됩니다.

`type X = otherdomain.Repo` 처럼 다른 도메인의 repository 인터페이스를
재노출하는 alias도 함께 감지합니다 — 이런 cross-domain 경계 코드는
`orchestration/`에 두어야 합니다.

### `testing.no-handmock`

테스트 파일은 hand-rolled mock/fake/stub struct를 정의하면 안 됩니다.
mockery 등 생성 도구를 대신 사용하세요.

### `naming.type-pattern-mismatch` (flat 프리셋)

TypePattern 접두사와 일치하는 파일은 대응하는 타입을 정의해야 합니다.

```go
// worker/worker_order.go는 다음을 정의해야:
type OrderWorker struct{}  // 기대됨

type SomethingElse struct{}  // OrderWorker가 기대됨
```

### `naming.type-pattern-missing-method` (flat 프리셋)

TypePattern으로 매칭된 타입은 필수 메서드를 가져야 합니다.

```go
type OrderWorker struct{}
// Process 메서드 누락 --- 위반

func (w *OrderWorker) Process(ctx context.Context) error { ... }  // 올바름
```

## 인터페이스 패턴 규칙

`rules.CheckInterfacePattern(pkgs, opts...)`

Go 인터페이스 모범 사례를 강제합니다: 비공개 구현체, `New()` 전용 생성자,
인터페이스 반환 타입, 패키지당 단일 인터페이스.

### `interface.exported-impl`

exported struct는 인터페이스를 구현하면 안 됩니다 --- 구현 타입을 unexported로 만들어
소비자가 concrete 타입에 의존하지 않도록 합니다.

```go
type RepositoryImpl struct{ db *sql.DB }  // exported struct가 interface 구현
type repository struct{ db *sql.DB }      // unexported --- 올바름
```

### `interface.constructor-name`

생성자는 `New`여야 하며, `NewXxx` 변형은 불허합니다. 모든 패키지에서 일관된
팩토리 패턴을 강제합니다.

```go
func NewRepository(db *sql.DB) Repository  // NewXxx 불허
func New(db *sql.DB) Repository            // 올바름
```

### `interface.constructor-returns-interface`

`New()`는 concrete 타입이 아닌 인터페이스를 반환해야 합니다. 호출자가
구현이 아닌 계약에 의존하도록 보장합니다.

```go
func New(db *sql.DB) *repository  // concrete 타입 반환
func New(db *sql.DB) Repository   // 인터페이스 반환 --- 올바름
```

### `interface.single-per-package`

패키지당 exported 인터페이스는 최대 1개 (Warning). 하나의 패키지에 여러 인터페이스가 있으면
보통 패키지의 책임이 너무 많다는 신호입니다.

프리셋별 제외 레이어(진입점, model, event, pkg)는 `InterfacePatternExclude`로 제어합니다.

### `interface.cross-domain-anonymous`

도메인 외부 *그리고 orchestration 외부*에서 선언된 anonymous interface가 method signature에 다른 도메인 타입을 참조하는 경우를 감지합니다. 기본 severity: **Error**.

이 룰은 **cross-domain 추상화는 orchestration 패키지가 소유한다**는 컨벤션을 강제합니다. cmd/ (또는 internal/pkg/) 같은 wiring 코드가 도메인 타입에 대해 inline anonymous interface를 선언하면, 통제되지 않은 *두 번째* cross-domain 표면을 만드는 셈입니다 — 그 어댑터/추상화는 `internal/orchestration/`에 있어야 합니다.

```go
// flagged: cmd/가 도메인 타입을 추상화하는 inline interface 선언
package main

import "example.com/p/internal/domain/user"

type adapter struct {
    repo interface {                                          // ← cmd/의 cross-domain anonymous
        GetByID(ctx context.Context, id string) (*user.User, error)
    }
}
```

```go
// flagged 안 됨: 같은 모양이지만 orchestration 안 (cross-domain 조정의 지정된 장소)
package orchestration

import "example.com/p/internal/domain/user"

type userInfoAdapter struct {
    repo interface {                                          // ← anonymous지만 orchestration은 exempt
        GetByID(ctx context.Context, id string) (*user.User, error)
    }
}
```

flag된 위치를 고치는 방법은 **어댑터를 orchestration 패키지로 이동**하고, wiring 코드는 자체 interface를 선언하는 대신 orchestration 생성자를 호출하는 것입니다.

스킵:
- 테스트 파일 (`_test.go`)
- 빈 interface (`interface{}`) 및 메서드 선언 없는 interface
- Embedded interface (`interface { io.Reader }`)
- 같은 도메인 내 참조
- `internal/<OrchestrationDir>/` 안의 모든 패키지 — orchestration은 cross-domain 조정의 지정된 레이어
- `DomainDir`이 없는 플랫 레이아웃 모델 (ConsumerWorker, Batch, EventPipeline)

### `interface.container-only`

패키지 안에서 선언된 named interface가 **struct field 타입으로만** 사용되고 함수 파라미터나 반환 타입으로는 한 번도 사용되지 않는 경우를 감지합니다. 기본 severity: **Warning**.

이는 vibe-coding 잡음 패턴으로, interface를 추상화가 아니라 *값 컨테이너*로 쓰는 신호입니다. 주로 wiring 레이어에서 어떤 값을 들고 있어야 하는데 concrete 타입이 노출되지 않아서 (예: `alias.go`가 생성자만 re-export하고 타입은 안 함), 필드 타입을 부여하기 위해 local interface를 임시로 만든 경우에 발생합니다.

```go
// flagged: container-only — 파라미터나 반환에 한 번도 안 쓰임
type userRepo interface {
    GetByID(id string) string
}

type holder struct {
    r userRepo  // 유일한 사용처
}
```

```go
// flagged 안 됨: 정상 consumer-defined interface
type userRepo interface {
    GetByID(id string) string
}

func newHolder(r userRepo) *holder {  // 파라미터로 사용 → 진짜 추상화
    return &holder{r: r}
}
```

스킵:
- 테스트 파일 (`_test.go`) — mock/fake fixture가 같은 모양을 갖기 때문
- 타입 alias (`type Foo = pkg.Foo`)
- struct의 embedded field (anonymous embedding)
- 어디에서도 안 쓰이는 interface (다른 smell 카테고리)

이 룰은 **fix를 강제하지 않습니다**. 그저 smell을 짚을 뿐. 일반적인 두 가지 해결 방법:
1. concrete 타입을 `alias.go`에서 re-export해서 필드가 직접 들 수 있게 한다.
2. 값을 두 함수 사이의 struct field가 아니라 한 함수 내부의 local 변수로 다시 짠다.

`WithSeverity(Error)`로 hard rule로 승격할 수 있습니다.

## 블래스트 반경

`rules.AnalyzeBlastRadius(pkgs, module, root, opts...)`

IQR 기반 통계적 이상치로 비정상 커플링 패키지를 탐지합니다. 기본 severity: Warning. 내부 패키지 5개 미만이면 스킵.

| 규칙 | 의미 |
|------|------|
| `blast.high-coupling` | transitive dependents가 통계적 이상치 |

| 메트릭 | 정의 |
|--------|------|
| Ca (Afferent Coupling) | 이 패키지를 import하는 패키지 수 |
| Ce (Efferent Coupling) | 이 패키지가 import하는 패키지 수 |
| Instability | Ce / (Ca + Ce) |
| Transitive Dependents | BFS로 추적한 전체 역방향 도달 가능 집합 |

## 트랜잭션 경계

### `CheckTxBoundary` (옵트인)

트랜잭션을 **시작**할 수 있는 위치를 제한하고, 트랜잭션 타입이 허용 레이어 밖의
함수 시그니처로 **누설**되는 것을 막습니다. 완전 옵트인 — 설정 전엔 아무것도
하지 않으며 `RunAll`에 자동 포함됩니다 (설정 전까지 no-op).

```go
violations := rules.CheckTxBoundary(pkgs, module, root,
    rules.WithTxBoundary(rules.TxBoundaryConfig{
        StartSymbols: []string{
            "database/sql.(*DB).BeginTx",
            "database/sql.(*DB).Begin",
        },
        Types:         []string{"database/sql.Tx"},
        AllowedLayers: []string{"app"}, // 비어 있으면 기본값
    }),
)
```

발생 가능한 규칙 ID: `tx.start-outside-allowed-layer`, `tx.type-in-signature`.

## 세터 패턴

### `CheckNoSetters`

포인터 리시버를 가진 내보내기 세터 메서드(`Set*`, 매개변수 1개 이상)를 검출하여
명시적인 생성자 파라미터 사용을 유도합니다.

**권장 수정**: 의존성을 생성자의 명시적 파라미터로 추가 (`NewService(..., dep)`).
`With*` 옵션은 정말로 선택적이고 여러 조합이 필요한 경우에만 사용. 설정류
옵션에도 setter는 대체로 맞지 않음.

- 플루언트 빌더(리시버 타입을 반환하는 메서드)는 제외됩니다.
- 테스트 파일과 `testdata/` 또는 `mocks/` 하위 패키지는 자동 제외됩니다.
- 기본 severity: Warning. 엄격하게 적용하려면 `WithSeverity(rules.Error)`를 사용하세요.

```go
// 기본: Warning severity
report.AssertNoViolations(t, rules.CheckNoSetters(pkgs))

// 엄격: Error severity
report.AssertNoViolations(t, rules.CheckNoSetters(pkgs, rules.WithSeverity(rules.Error)))
```

발생 가능한 규칙 ID: `setter.forbidden`.

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
| `rules.CheckInterfacePattern(pkgs, opts...)` | 인터페이스 패턴 모범 사례 |
| `rules.CheckTxBoundary(pkgs, module, root, opts...)` | 트랜잭션 경계 강제 (옵트인) |
| `rules.CheckNoSetters(pkgs, opts...)` | 내보내기 세터 검출 (기본 Warning) |
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
| `rules.WithMaxRepoInterfaceMethods(10)` | repo 인터페이스 메서드 수 제한 |
| `rules.WithTxBoundary(cfg)` | 트랜잭션 경계 검사 설정 |

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

## 외부 Import 위생 --- 이 라이브러리가 아닌 AI 도구 지침으로 강제

`go-arch-guard`는 **프로젝트 내부** import만 검사합니다. 외부 의존성 위생은 AI 도구 지침과 코드 리뷰로 강제하세요.

**AI 도구의 시스템 프롬프트에 복사:**

```text
# 외부 Import 제약 (go-arch-guard는 이를 강제하지 않음)

- core/model, core/repo, core/svc, event --- stdlib만, 서드파티 금지
- handler --- HTTP/gRPC 프레임워크 허용, 영속성 라이브러리 금지
- infra --- 영속성/메시징 라이브러리 허용, HTTP 프레임워크 금지
- app --- 일반적으로 자유, 인프라 라이브러리 직접 import 지양
```

## 라이선스

MIT
