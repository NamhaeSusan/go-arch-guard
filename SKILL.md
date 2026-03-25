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

```bash
go get github.com/NamhaeSusan/go-arch-guard
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

## 7. Common Setup Scenarios

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
