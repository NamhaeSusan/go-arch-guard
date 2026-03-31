---
name: go-arch-guard
description: Use when user explicitly mentions go-arch-guard, architecture guardrails, or architecture_test.go in a Go project. Handles initial scaffolding and preset configuration.
user_invocable: true
---

# go-arch-guard

Go 프로젝트에 아키텍처 가드레일을 `go test`로 적용.

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
| DDD | `DDD()` (기본값) | handler→app→core/{model,repo,svc},event; infra→core/repo |
| Clean Architecture | `CleanArch()` | handler→usecase→entity+gateway; infra→gateway |
| Layered | `Layered()` | handler→service→repository+model; repository→model |
| Hexagonal | `Hexagonal()` | handler→usecase→port+domain; adapter→port+domain |
| Modular Monolith | `ModularMonolith()` | api→application→core; infrastructure→core |

### Step 2: 스캐폴딩

```bash
go get github.com/NamhaeSusan/go-arch-guard
mkdir -p internal/domain internal/orchestration internal/pkg
```

### Step 3: architecture_test.go 생성

파일이 이미 존재하면 덮어쓰지 않고 유저에게 확인한다.
패키지명은 go.mod module의 마지막 세그먼트. 예: `module github.com/user/myapp` → `package myapp_test`

#### DDD (WithModel 불필요)

```go
package {project}_test

import (
    "testing"

    "github.com/NamhaeSusan/go-arch-guard/analyzer"
    "github.com/NamhaeSusan/go-arch-guard/report"
    "github.com/NamhaeSusan/go-arch-guard/rules"
)

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

#### 그 외 프리셋 (CleanArch / Layered / Hexagonal / ModularMonolith)

DDD 템플릿에서 아래 2줄만 추가하고, 각 체크 함수에 `opts...` 전달:

```go
m := rules.{Preset}()  // CleanArch, Layered, Hexagonal, or ModularMonolith
opts := []rules.Option{rules.WithModel(m)}
```

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

전체 옵션: `WithSublayers`, `WithDirection`, `WithPkgRestricted`, `WithDomainDir`, `WithOrchestrationDir`, `WithSharedDir`, `WithRequireAlias`, `WithAliasFileName`, `WithRequireModel`, `WithModelPath`, `WithDTOAllowedLayers`, `WithBannedPkgNames`, `WithLegacyPkgNames`, `WithLayerDirNames`

---

## Domain Isolation (모든 모델 공통)

| from → to | domain root | domain sub-pkg | orchestration | shared pkg |
|-----------|:-:|:-:|:-:|:-:|
| **same domain** | O | O | X | O |
| **other domain** | X | X | X | O |
| **orchestration** | O | X | O | O |
| **cmd** | O | X | O | O |
| **shared pkg** | X | X | X | O |

---

## Options

```go
rules.WithExclude("internal/legacy/...")  // 마이그레이션 중 경로 제외
rules.WithSeverity(rules.Warning)         // 실패 없이 로그만
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
| handler/app/svc에서 interface 정의 | `naming.domain-interface-repo-only` |
| alias.go에서 interface re-export | `structure.domain-alias-no-interface` |

---

## Existing Project with Violations

1. `architecture_test.go` 생성 후 실행 → 전체 위반 확인
2. `WithExclude`로 마이그레이션 중 경로 제외
3. 점진적으로 위반 수정, exclude 제거
4. 전환기에는 `WithSeverity(rules.Warning)` 사용
