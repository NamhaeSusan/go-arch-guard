---
name: go-arch-guard
description: Go 서버 프로젝트에 go-arch-guard 아키텍처 가드레일 설정. 의존성/구조/네이밍 규칙 적용.
---

# go-arch-guard Setup Guide

Go 서버 프로젝트에 아키텍처 가드레일을 `go test`로 적용하는 방법.

## When to Use

- 새 Go 서버 프로젝트 초기 설정 시
- 기존 프로젝트에 아키텍처 검증 추가 시
- `architecture_test.go` 작성/수정 요청 시

---

## 1. Install

### Go library

```bash
go get github.com/NamhaeSusan/go-arch-guard
```

### Claude Code plugin

```text
/plugin marketplace add NamhaeSusan/go-arch-guard
/plugin install go-arch-guard@go-arch-guard-marketplace
```

---

## 2. Target Structure

```text
cmd/
  api/
    main.go

internal/
  domain/
    {domain}/
      alias.go          # public surface — type aliases only
      app/              # application service
      core/
        model/          # domain model (required)
        repo/           # repository interface
        svc/            # domain service
      event/            # domain events
      handler/http/     # HTTP handler
      infra/persistence # infrastructure
  orchestration/        # cross-domain coordination
  pkg/                  # shared utilities
```

**Key constraints:**
- `internal/` 아래에는 `domain/`, `orchestration/`, `pkg/`만 허용
- 각 domain root에는 `alias.go` 필수, 다른 non-test .go 파일 금지
- 각 domain에 `core/model/` 디렉토리에 최소 1개 .go 파일 필수

---

## 3. architecture_test.go Template

프로젝트 루트에 생성:

```go
package myproject_test

import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/report"
	"github.com/NamhaeSusan/go-arch-guard/rules"
)

func TestArchitecture(t *testing.T) {
	pkgs, err := analyzer.Load(".", "internal/...", "cmd/...")
	if err != nil {
		t.Log(err) // partial load OK
	}
	if len(pkgs) == 0 {
		t.Fatalf("no packages loaded: %v", err)
	}

	// module과 root에 ""를 넘기면 자동 추출
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

---

## 4. Import Rules Quick Reference

### Domain Isolation (who can import whom)

| from → to | domain root | domain sub-pkg | orchestration | pkg |
|-----------|:-----------:|:--------------:|:-------------:|:---:|
| **same domain** | ✅ | ✅ | ❌ | ✅ |
| **other domain** | ❌ | ❌ | ❌ | ✅ |
| **orchestration** | ✅ | ❌ | ✅ | ✅ |
| **cmd** | ✅ | ❌ | ✅ | ✅ |
| **pkg** | ❌ | ❌ | ❌ | ✅ |

### Layer Direction (intra-domain)

```text
handler → app → core/model, core/repo, core/svc, event
core → core/model
core/repo, core/svc, event → core/model
infra → core/repo, core/model, event
```

- `core/*`, `event`는 `internal/pkg` import 금지

---

## 5. Options

### Migration — 특정 경로 제외

```go
rules.CheckDomainIsolation(pkgs, "", "",
	rules.WithExclude("internal/legacy/..."),
)
```

### Warning Mode — 실패 없이 로그만

```go
rules.CheckDomainIsolation(pkgs, "", "",
	rules.WithSeverity(rules.Warning),
)
```

---

## 6. Banned Patterns

| Category | Banned |
|----------|--------|
| Package names | `util`, `common`, `misc`, `helper`, `shared`, `services` |
| Legacy dirs | `router`, `bootstrap`, misplaced `app`/`handler`/`infra` under `internal/` |
| Naming | type stuttering (`order.OrderService`), `Impl` suffix, non-snake_case files, hand-rolled mocks in `_test.go` |
| Placement | `middleware/` must be at `internal/pkg/middleware/` |
| DTOs | `dto.go` in `core/`, `event/`, `infra/` (allowed in `handler/`, `app/`) |

---

## 7. Cross-Domain Interface 금지 원칙 (CRITICAL)

도메인 내에서 interface는 **`core/repo/`에만** 정의할 수 있으며, 이는 **infra가 구현하는 persistence 추상화 전용**입니다.

### 위반 시 올바른 수정 방향

cross-domain 데이터가 필요한 endpoint는 **도메인 handler가 아닌 `orchestration/handler/http/`로 이동**해야 합니다. interface를 다른 패키지로 옮기는 것은 해결이 아닙니다.

### 금지되는 우회 패턴

| 우회 시도 | 잡는 규칙 |
|----------|----------|
| handler/app/core/svc에서 interface 정의 | `naming.domain-interface-repo-only` |
| app에서 core/repo interface를 type alias | `naming.domain-interface-repo-only` |
| alias.go에서 interface re-export | `structure.domain-alias-no-interface` |
| alias.go에서 core/repo, core/svc type alias | `structure.domain-alias-no-interface` |

### core/repo에 cross-domain interface를 넣지 말 것

`core/repo/`는 persistence interface 전용입니다. `GetUserByID`, `SetUserActive` 같은 다른 도메인 조작 메서드는 repo에 넣지 마세요. 이런 기능이 필요하면 해당 endpoint를 `orchestration/handler/http/`로 이동하세요.

---

## Blast Radius

`AnalyzeBlastRadius`는 내부 패키지 간 의존 그래프를 분석하여 coupling이 비정상적으로 높은 패키지를 Warning으로 보고한다.

- 설정 불필요 — IQR 기반 통계적 이상치 자동 탐지
- 기본 severity: Warning (테스트를 실패시키지 않음)
- 내부 패키지 5개 미만이면 분석 스킵

| Rule | 의미 |
|------|------|
| `blast-radius.high-coupling` | transitive dependents가 통계적 이상치인 패키지 |

---

## 8. Common Setup Scenarios

### New project from scratch

1. Create target structure directories
2. Add `alias.go` + `core/model/*.go` per domain
3. Copy architecture_test.go template
4. `go test -run TestArchitecture -v`

### Existing project with violations

1. Copy template, run to see all violations
2. Add `WithExclude` for paths being migrated
3. Fix violations incrementally, remove excludes as fixed
4. Optionally use `WithSeverity(rules.Warning)` for non-critical rules during transition
