# go-arch-guard

`go test`로 Go 프로젝트의 아키텍처 가드레일을 적용합니다.

격리, 레이어 방향, 구조, 네이밍 규칙을 정의하고, 프로젝트 형태가 벗어나면 일반 테스트에서 실패시킵니다. 별도 CLI나 설정 포맷 없이, Go 테스트만으로 동작합니다.

## 독선적 기본값

이 라이브러리에 포함된 규칙, 경로(`internal/domain/`, `internal/orchestration/`, `internal/pkg/`), 서브레이어 이름, 레이어 방향 매트릭스는 **NamhaeSusan의 컨벤션**을 반영합니다. 범용 Go 모범 사례가 아닙니다.

자체 프로젝트에 도입하려면 현재 규칙셋을 **참조 구현**으로 취급하고, 팀 아키텍처에 맞게 조정(또는 재작성)하세요.

## 왜 필요한가

아키텍처는 보통 깊은 이론적 위반이 아니라 몇 가지 큰 실수를 통해 무너집니다:

- 크로스 도메인 import
- 숨겨진 컴포지션 루트
- 패키지 배치 드리프트
- 의도한 프로젝트 형태를 깨는 네이밍

`go-arch-guard`는 정적 분석으로 이런 큰 실수를 조기에 잡습니다. Go 패키지 내부의 모든 의미론적 뉘앙스를 모델링하지 않으며, Go가 자체적으로 거부하는 것(예: import cycle)은 주요 대상이 아닙니다.

## 설치

```bash
go get github.com/NamhaeSusan/go-arch-guard
```

## 빠른 시작

프로젝트 루트에 `architecture_test.go`를 생성합니다:

```go
package myproject_test

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/report"
	"github.com/NamhaeSusan/go-arch-guard/rules"
)

func TestArchitecture(t *testing.T) {
	root := "."
	module := "github.com/yourmodule"

	pkgs, err := analyzer.Load(root, "internal/...", "cmd/...")
	if err != nil {
		// 일부 패키지만 실패하면 에러와 함께 유효한 패키지를 반환합니다.
		// t.Log로 분석을 계속합니다.
		t.Log(err)
	}
	if len(pkgs) == 0 {
		t.Fatalf("no packages loaded: %v", err)
	}

	t.Run("domain isolation", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckDomainIsolation(pkgs, module, root))
	})
	t.Run("layer direction", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckLayerDirection(pkgs, module, root))
	})
	t.Run("naming", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckNaming(pkgs))
	})
	t.Run("structure", func(t *testing.T) {
		report.AssertNoViolations(t, rules.CheckStructure(root))
	})
}
```

실행:

```bash
go test -run TestArchitecture -v
```

위반이 있을 때 출력 예시:

```text
=== RUN   TestArchitecture/domain_isolation
    [ERROR] violation: domain "order" must not import domain "user" (file: internal/domain/order/app/service.go:5, rule: isolation.cross-domain, fix: use orchestration/ for cross-domain orchestration or move shared types to pkg/)
    found 1 architecture violation(s)
--- FAIL: TestArchitecture/domain_isolation
```

### 간소화된 사용법

`module`과 `root`에 빈 문자열을 전달하면 로드된 패키지에서 자동 추출합니다:

```go
t.Run("domain isolation", func(t *testing.T) {
	report.AssertNoViolations(t, rules.CheckDomainIsolation(pkgs, "", ""))
})
t.Run("layer direction", func(t *testing.T) {
	report.AssertNoViolations(t, rules.CheckLayerDirection(pkgs, "", ""))
})
```

모듈을 결정할 수 없는 경우(예: 모듈 메타데이터 없이 패키지를 로드한 경우) `meta.no-matching-packages` 경고가 발생합니다.

## 대상 아키텍처

`go-arch-guard`는 도메인 중심 수직 슬라이스 레이아웃을 가정합니다.

`internal/` 최상위에는 `domain/`, `orchestration/`, `pkg/`만 허용됩니다. 추가 최상위 지원 패키지는 거부됩니다.

```text
cmd/
`-- api/
    |-- main.go
    |-- wire.go
    `-- routes.go

internal/
|-- domain/
|   |-- order/
|   |   |-- alias.go
|   |   |-- app/
|   |   |   `-- service.go
|   |   |-- core/
|   |   |   |-- model/
|   |   |   |   `-- order.go
|   |   |   |-- repo/
|   |   |   |   `-- repository.go
|   |   |   `-- svc/
|   |   |       `-- order.go
|   |   |-- event/
|   |   |   `-- events.go
|   |   |-- handler/
|   |   |   `-- http/
|   |   |       `-- handler.go
|   |   `-- infra/
|   |       `-- persistence/
|   |           `-- store.go
|   `-- user/
|       `-- ...
|-- orchestration/
|   |-- handler/
|   |   `-- http/
|   |       `-- handler.go
|   `-- create_order.go
`-- pkg/
    |-- middleware/
    `-- transport/http/
```

### 도메인 루트

각 도메인 루트 패키지는 해당 도메인의 공개 import surface입니다. 루트에는 반드시 `alias.go`를 정의해야 하며, 추가 비테스트 Go 파일을 포함할 수 없습니다.

예시:

```go
// internal/domain/order/alias.go
package order

import (
	"mymodule/internal/domain/order/app"
	orderhttp "mymodule/internal/domain/order/handler/http"
)

type Service = app.Service
type Handler = orderhttp.Handler
```

외부 코드는 `internal/domain/order`를 import해야 하며, `internal/domain/order/app`이나 `internal/domain/order/handler/http`를 직접 import해서는 안 됩니다.

### 도메인 레이어

도메인 내에서 모델링되는 서브레이어:

- `handler`
- `app`
- `core`
- `core/model`
- `core/repo`
- `core/svc`
- `event`
- `infra`

알려지지 않은 도메인 서브레이어는 거부됩니다.

### 오케스트레이션

`internal/orchestration`은 크로스 도메인 조율 레이어입니다.

- 도메인 루트만 import할 수 있으며, 도메인 하위 패키지는 불가합니다.
- 도메인 import 시 반드시 도메인 루트 패키지(`alias.go`)를 통해야 하며, `app/`, `handler/` 등 하위 패키지는 불가합니다.
- 필요 시 `internal/pkg/...`의 공유 헬퍼를 import할 수 있습니다.
- 기존 비도메인 내부 패키지도 import할 수 있습니다.
- 외부에서는 보호된 레이어입니다: `cmd/...`와 `internal/orchestration/...`만 오케스트레이션에 의존할 수 있으며, 도메인, `pkg`, 기타 내부 패키지는 불가합니다.

### 공유 패키지

`internal/pkg`는 공유 유틸리티용입니다.

- `pkg`는 도메인을 import할 수 없습니다.
- `pkg`는 오케스트레이션을 import할 수 없습니다.
- 내부 도메인 레이어(`core`, `core/model`, `core/repo`, `core/svc`, `event`)는 `internal/pkg/...`를 import할 수 없습니다.

## 규칙

### 도메인 격리

`rules.CheckDomainIsolation(pkgs, module, root, opts...)`

목적:

- 크로스 도메인 import 차단
- 도메인 루트 패키지를 통한 외부 접근 강제
- 오케스트레이션과 `pkg`가 숨겨진 의존성 지름길이 되는 것 방지

| 규칙 | 의미 |
|------|------|
| `isolation.cross-domain` | 도메인 A는 도메인 B를 import할 수 없음 |
| `isolation.cmd-deep-import` | `cmd/`는 도메인 루트만 import 가능, 하위 패키지 불가 |
| `isolation.orchestration-deep-import` | `orchestration/`은 도메인 루트만 import 가능, 하위 패키지 불가 |
| `isolation.pkg-imports-domain` | `pkg/`는 도메인을 import할 수 없음 |
| `isolation.pkg-imports-orchestration` | `pkg/`는 오케스트레이션을 import할 수 없음 |
| `isolation.domain-imports-orchestration` | 도메인은 오케스트레이션을 import할 수 없음 |
| `isolation.internal-imports-orchestration` | cmd/orchestration 외 패키지는 오케스트레이션을 import할 수 없음 |
| `isolation.internal-imports-domain` | 미등록 내부 패키지는 도메인을 import할 수 없음 |

Import 매트릭스:

| 출발 | 도착 | 허용? |
|------|------|:-----:|
| 같은 도메인 | 같은 도메인 | O |
| 모든 곳 | `internal/pkg/...` | O |
| `orchestration/...` | 도메인 루트 | O |
| `orchestration/...` | 도메인 하위 패키지 | X |
| `orchestration/...` | `internal/pkg/...` | O |
| `orchestration/...` | 기타 비도메인 내부 패키지 | O |
| `cmd/...` | `internal/orchestration/...` | O |
| `cmd/...` | 도메인 루트 | O |
| `cmd/...` | 도메인 하위 패키지 | X |
| 도메인 | `internal/orchestration/...` | X |
| `internal/pkg/...` | 모든 도메인 | X |
| `internal/pkg/...` | `internal/orchestration/...` | X |
| 기타 내부 패키지 | 모든 도메인 | X |
| 기타 내부 패키지 | `internal/orchestration/...` | X |
| 도메인 A | 도메인 B | X |

### 레이어 방향

`rules.CheckLayerDirection(pkgs, module, root, opts...)`

목적:

- 도메인 내 허용된 의존 방향 강제
- 알려지지 않은 도메인 서브레이어 거부
- 내부 레이어에서 `internal/pkg/...` 사용 금지

| 규칙 | 의미 |
|------|------|
| `layer.direction` | 허용된 레이어 방향을 위반하는 import |
| `layer.inner-imports-pkg` | 내부 레이어(`core/*`, `event`)가 `internal/pkg/`를 import |
| `layer.unknown-sublayer` | 도메인에서 알 수 없는 서브레이어 발견 |

도메인 내 허용 import:

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

참고:

- 같은 서브레이어 간 import는 허용
- 도메인 루트 패키지는 `CheckLayerDirection`에서 검사하지 않음
- `core`, `core/model`, `core/repo`, `core/svc`, `event`는 `internal/pkg/...`를 import할 수 없음

예시:

```go
// OK: infra가 core/repo를 import (허용)
// internal/domain/order/infra/persistence/store.go
import "mymodule/internal/domain/order/core/repo"

// BAD: core/svc가 core/repo를 import (허용 목록에 없음)
// internal/domain/order/core/svc/order.go
import "mymodule/internal/domain/order/core/repo" // layer.direction

// BAD: handler가 infra를 직접 import (허용 목록에 없음)
// internal/domain/order/handler/http/handler.go
import "mymodule/internal/domain/order/infra/persistence" // layer.direction

// BAD: core/model이 internal/pkg를 import (내부 레이어 제한)
// internal/domain/order/core/model/order.go
import "mymodule/internal/pkg/clock" // layer.inner-imports-pkg
```

### 구조

`rules.CheckStructure(root, opts...)`

| 규칙 | 의미 |
|------|------|
| `structure.internal-top-level` | `internal/` 바로 아래에는 `domain`, `orchestration`, `pkg`만 허용 |
| `structure.banned-package` | `util`, `common`, `misc`, `helper`, `shared`, `services`는 `internal/` 어디서든 금지 |
| `structure.legacy-package` | `router`, `bootstrap`은 마이그레이션이 필요한 레거시 패키지 |
| `structure.misplaced-layer` | 도메인 슬라이스나 orchestration handler 외부의 `app`/`handler`/`infra` 디렉토리 |
| `structure.middleware-placement` | `middleware/`는 `internal/pkg/middleware/`에만 허용 |
| `structure.domain-root-alias-required` | 각 도메인 루트에 `alias.go` 필수 |
| `structure.domain-root-alias-package` | `alias.go`의 패키지 이름은 도메인 디렉토리와 일치해야 함 |
| `structure.domain-root-alias-only` | 도메인 루트에는 비테스트 Go 파일로 `alias.go`만 허용 |
| `structure.domain-alias-no-interface` | `alias.go`에서 interface re-export 금지 — 크로스 도메인 의존 의심 |
| `structure.domain-model-required` | 각 도메인에 `core/model/`에 최소 1개의 직접 비테스트 Go 파일 필수 |
| `structure.dto-placement` | `dto.go` 또는 `*_dto.go`는 내부 도메인 레이어(`core/`, `event/`)나 `infra/`에 불가; `handler/`와 `app/`에서만 허용 |

### 네이밍

`rules.CheckNaming(pkgs, opts...)`

이 규칙 세트는 경계 규칙보다 의도적으로 더 독선적입니다.

| 규칙 | 의미 |
|------|------|
| `naming.no-stutter` | exported 타입이 패키지 이름을 반복 |
| `naming.no-impl-suffix` | exported 타입이 `Impl`로 끝남 |
| `naming.snake-case-file` | 파일 이름이 snake_case가 아님 |
| `naming.repo-file-interface` | `repo/` 아래 파일이 매칭되는 interface를 정의하지 않음 |
| `naming.no-layer-suffix` | 파일 이름이 레이어 이름을 불필요하게 반복 |
| `naming.domain-interface-repo-only` | `core/repo/` 외부에서 도메인 interface 정의, 또는 `core/repo` interface를 type alias로 re-export |
| `naming.no-handmock` | 테스트 파일에서 mock/fake/stub 접두사 struct에 메서드를 직접 정의 |

## 옵션

### 심각도

기본 심각도는 `Error`입니다. 테스트를 실패시키지 않고 로그만 남기려면 `Warning`을 사용합니다:

```go
violations := rules.CheckDomainIsolation(
	pkgs, module, root,
	rules.WithSeverity(rules.Warning),
)
report.AssertNoViolations(t, violations) // 통과 — Error만 실패
```

### 경로 제외

마이그레이션 중 하위 트리를 건너뜁니다:

```go
rules.CheckDomainIsolation(
	pkgs, module, root,
	rules.WithExclude("internal/legacy/..."),
	rules.WithExclude("internal/domain/payment/core/model/..."),
)
```

- 패턴은 슬래시를 사용하는 프로젝트 상대 경로
- `...`는 해당 루트와 모든 하위를 매칭
- 모듈 수식 경로(`github.com/yourmodule/...`)는 사용하지 마세요
- 모든 체크 함수에 동일한 형식 적용

## 진단

| 규칙 | 의미 |
|------|------|
| `meta.no-matching-packages` | `module` 인수가 로드된 패키지와 일치하지 않음 — 보통 설정 오류 |

## TUI 뷰어

프로젝트의 패키지 구조와 의존성을 인터랙티브 터미널 UI로 시각화합니다.

```bash
# go-arch-guard를 의존성으로 쓰는 프로젝트에서
go run github.com/NamhaeSusan/go-arch-guard/cmd/tui .

# 또는 go-arch-guard 저장소에서 직접 실행
go run ./cmd/tui /path/to/your/project
```

의존성이 없다면 먼저 추가:

```bash
go get github.com/NamhaeSusan/go-arch-guard@latest
```

기능:
- 건강 상태 기반 트리 색상 — 초록 (정상), 노랑 `⚠` (경고), 빨강 `✗` (에러)
- 패키지 선택 시 imports, 역방향 의존성, 커플링 메트릭(Ca, Ce, Instability, Transitive Dependents) 표시
- 그룹 노드 선택 시 (예: `domain/`) 하위 전체 위반 요약: 에러 우선, 경고 후순
- 위반 상세 정보: 심각도, 규칙 ID, 메시지, 수정 가이드
- 검색/필터 — `/` 키로 패키지명 필터링, `Esc`로 초기화
- 키보드: `↑↓` 탐색, `Enter` 열기/닫기, `/` 검색, `Tab` 디테일 패널 스크롤, `q` 종료

## API 레퍼런스

| 함수 | 설명 |
|------|------|
| `analyzer.Load(dir, patterns...)` | 분석용 Go 패키지 로드 |
| `rules.CheckDomainIsolation(pkgs, module, root, opts...)` | 크로스 도메인 및 오케스트레이션 경계 검사 (`""`로 자동 추출) |
| `rules.CheckLayerDirection(pkgs, module, root, opts...)` | 도메인 내 의존 방향 검사 (`""`로 자동 추출) |
| `rules.CheckNaming(pkgs, opts...)` | 네이밍 컨벤션 검사 |
| `rules.CheckStructure(root, opts...)` | 파일시스템 구조 검사 |
| `report.AssertNoViolations(t, violations)` | `Error` 위반 시 테스트 실패 |
| `rules.WithSeverity(rules.Warning)` | 위반을 경고로 다운그레이드 |
| `rules.WithExclude("internal/path/...")` | 프로젝트 상대 하위 트리 또는 파일 건너뛰기 |

## 외부 Import 위생 — 이 라이브러리가 아닌 AI 도구 지침으로 강제

`go-arch-guard`는 **프로젝트 내부** import만 검사합니다. 어떤 레이어가 어떤 외부 패키지를 사용할 수 있는지(예: `core/model`이 `gorm.io/gorm`을 import)는 제한하지 않으며 **앞으로도 하지 않을 것입니다**.

이는 의도적입니다. `WithBannedImport("core/...", "gorm.io/...")`같은 규칙은 단순해 보이지만 금세 허용 목록 유지보수 부담이 그 가치를 초과합니다. 외부 의존성 위생은 AI 도구 지침과 코드 리뷰로 강제하는 것이 더 낫습니다.

**아래 제약사항을 AI 도구의 시스템 프롬프트나 프로젝트 규칙 파일(예: `CLAUDE.md`, `AGENTS.md`, `.cursorrules`)에 복사하세요:**

```text
# 외부 Import 제약 (go-arch-guard는 이를 강제하지 않음 — 직접 해야 함)

- core/model, core/repo, core/svc, event — stdlib만, 서드파티 import 금지
- handler — HTTP/gRPC 프레임워크 허용, 영속성 라이브러리 금지
- infra — 영속성/메시징 라이브러리 허용, HTTP 프레임워크 import 금지
- app — 일반적으로 자유, 단 인프라 라이브러리 직접 import 지양
```

이것이 의도된 강제 메커니즘입니다. `go-arch-guard`에 외부 import 규칙 추가를 요청하는 이슈나 PR을 열지 마세요.

## 라이선스

MIT
