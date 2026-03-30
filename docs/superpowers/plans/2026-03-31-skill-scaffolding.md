# Skill Scaffolding Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Quick Init scaffolding flow to the go-arch-guard SKILL.md so that selecting a preset auto-generates directories and `architecture_test.go`.

**Architecture:** Single file modification — restructure SKILL.md to add a decision flow at the top that detects project state and routes to either Quick Init (new project) or existing guide (existing project). The Quick Init section contains step-by-step instructions for the Claude Code agent to execute.

**Tech Stack:** Markdown (Claude Code skill format)

---

### Task 1: Add Quick Init flow to SKILL.md

**Files:**
- Modify: `plugins/go-arch-guard/skills/go-arch-guard/SKILL.md`

- [ ] **Step 1: Update description frontmatter**

Change line 3 from:
```
description: Go 서버 프로젝트에 go-arch-guard 아키텍처 가드레일 설정. DDD / Clean Architecture / Layered / Hexagonal / Modular Monolith 프리셋 또는 커스텀 모델 지원.
```
to:
```
description: Go 서버 프로젝트에 go-arch-guard 아키텍처 가드레일 설정 및 초기 스캐폴딩. DDD / Clean Architecture / Layered / Hexagonal / Modular Monolith 프리셋 또는 커스텀 모델 지원.
```

- [ ] **Step 2: Update "When to Use" section**

Replace lines 10-14:
```markdown
## When to Use

- 새 Go 서버 프로젝트 초기 설정 시
- 기존 프로젝트에 아키텍처 검증 추가 시
- `architecture_test.go` 작성/수정 요청 시
```

with:
```markdown
## When to Use

- 새 Go 서버 프로젝트 초기 스캐폴딩 시
- 기존 프로젝트에 아키텍처 검증 추가 시
- `architecture_test.go` 작성/수정 요청 시

## Decision Flow

프로젝트 상태를 확인하고 적절한 흐름으로 분기한다.

1. `go.mod` 존재 여부 확인
2. `internal/` 디렉토리 존재 여부 확인
3. `architecture_test.go` 존재 여부 확인

**새 프로젝트** (`internal/` 없음 AND `architecture_test.go` 없음):
→ **Quick Init** (아래) 실행

**기존 프로젝트** (`internal/` 있음 OR `architecture_test.go` 있음):
→ **섹션 2. Choose Architecture Model** 부터 참조
```

- [ ] **Step 3: Add Quick Init section after "Decision Flow" and before "## 1. Install"**

Insert the following section:

```markdown
---

## Quick Init

새 프로젝트에 go-arch-guard를 처음 설정할 때 사용한다.
프리셋을 선택하면 기본 디렉토리 구조와 `architecture_test.go`를 생성한다.

### Step 1: 프리셋 선택

유저에게 아래 프리셋 중 하나를 선택하도록 질문한다:

| 프리셋 | 설명 | 적합한 프로젝트 |
|--------|------|----------------|
| **DDD** | handler→app→core/model,repo,svc→event, infra | 도메인 모델 중심, alias.go로 캡슐화 |
| **Clean Architecture** | handler→usecase→entity+gateway, infra→gateway | Uncle Bob 스타일, entity 중심 |
| **Layered** | handler→service→repository+model | Spring 스타일, 가장 단순한 레이어 구조 |
| **Hexagonal** | handler→usecase→port+domain, adapter→port | 포트 & 어댑터, 인터페이스 분리 중시 |
| **Modular Monolith** | api→application→domain, infrastructure→domain | 모듈 단위 격리, MSA 전환 준비 |

### Step 2: 의존성 설치

```bash
go get github.com/NamhaeSusan/go-arch-guard
```

### Step 3: 디렉토리 생성

아래 명령을 실행한다:

```bash
mkdir -p internal/domain internal/orchestration internal/pkg
```

### Step 4: architecture_test.go 생성

선택한 프리셋에 따라 프로젝트 루트에 `architecture_test.go`를 생성한다.
**주의:** 파일이 이미 존재하면 덮어쓰지 않고 유저에게 확인한다.

패키지명은 프로젝트의 go.mod module 경로에서 마지막 세그먼트를 사용한다.
예: `module github.com/user/myapp` → `package myapp_test`

#### DDD (기본값)

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

#### Clean Architecture / Layered / Hexagonal / Modular Monolith

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

    m := rules.{Preset}()  // CleanArch, Layered, Hexagonal, or ModularMonolith
    opts := []rules.Option{rules.WithModel(m)}

    t.Run("domain isolation", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckDomainIsolation(pkgs, "", "", opts...))
    })
    t.Run("layer direction", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckLayerDirection(pkgs, "", "", opts...))
    })
    t.Run("naming", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckNaming(pkgs, opts...))
    })
    t.Run("structure", func(t *testing.T) {
        report.AssertNoViolations(t, rules.CheckStructure(".", opts...))
    })
    t.Run("blast radius", func(t *testing.T) {
        report.AssertNoViolations(t, rules.AnalyzeBlastRadius(pkgs, "", "", opts...))
    })
}
```

`{project}`는 go.mod의 module 마지막 세그먼트, `{Preset}`은 선택한 프리셋 함수명으로 치환한다.

### Step 5: 검증

```bash
go test -run TestArchitecture -v
```

도메인이 아직 없으므로 모든 체크가 통과해야 한다. 이후 도메인을 추가하면서 가드레일이 적용된다.

### Step 6: 다음 단계 안내

유저에게 안내한다:

> 스캐폴딩이 완료됐습니다. 이제 `internal/domain/` 아래에 첫 번째 도메인을 추가하세요.
> 선택한 프리셋의 레이어 구조는 **섹션 2**를 참고하세요.
```

- [ ] **Step 4: Verify the SKILL.md structure**

Read the modified file and verify:
- Decision Flow is between "When to Use" and "1. Install"
- Quick Init is between Decision Flow and "1. Install"
- All existing sections (1-9) are preserved
- No duplicate content between Quick Init templates and section 3 templates

- [ ] **Step 5: Run tests to verify nothing broke**

```bash
cd /Users/kimtaeyun/go-arch-guard && go test ./... -count=1
```

Expected: All PASS (SKILL.md changes don't affect Go tests)

- [ ] **Step 6: Run the skill_test.go specifically**

```bash
cd /Users/kimtaeyun/go-arch-guard && go test . -run TestSkill -v -count=1
```

Check if there's a skill content validation test. If it exists and fails, fix.

- [ ] **Step 7: Commit**

```bash
git add plugins/go-arch-guard/skills/go-arch-guard/SKILL.md
git commit -m "feat: add Quick Init scaffolding flow to go-arch-guard skill"
```
