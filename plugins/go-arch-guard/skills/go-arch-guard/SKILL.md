---
name: go-arch-guard
description: Go 서버 프로젝트에 go-arch-guard 아키텍처 가드레일 설정. DDD / Clean Architecture 프리셋 또는 커스텀 모델 지원.
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

## 2. Choose Architecture Model

### Option A: DDD (기본값)

```text
internal/
├── domain/
│   └── {domain}/
│       ├── alias.go              # public surface (필수)
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

**레이어 방향:**

```text
handler → app → core/model, core/repo, core/svc, event
core, core/repo, core/svc, event → core/model
infra → core/repo, core/model, event
```

**DDD 제약:**
- 각 domain root에 `alias.go` 필수, 다른 non-test .go 파일 금지
- 각 domain에 `core/model/` 디렉토리에 최소 1개 .go 파일 필수
- `core/*`, `event`는 `internal/pkg` import 금지
- interface는 `core/repo/`에만 정의 가능

### Option B: Clean Architecture

```text
internal/
├── domain/
│   └── {domain}/
│       ├── handler/              # 인터페이스 어댑터 (컨트롤러)
│       ├── usecase/              # 애플리케이션 비즈니스 규칙
│       ├── entity/               # 엔터프라이즈 비즈니스 규칙
│       ├── gateway/              # 데이터 접근 인터페이스
│       └── infra/                # 프레임워크 & 드라이버
├── orchestration/
└── pkg/
```

**레이어 방향:**

```text
handler → usecase → entity, gateway
gateway → entity
infra → gateway, entity
entity → (nothing)
```

**CleanArch 제약:**
- `alias.go` 불필요
- `entity`는 `internal/pkg` import 금지
- DTO는 `handler/`, `usecase/`에서만 허용

### Option C: Custom

DDD 기본값에서 시작하여 오버라이드:

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
    rules.WithRequireAlias(false),
    rules.WithRequireModel(false),
)
```

---

## 3. architecture_test.go Template

### DDD (기본값 — WithModel 불필요)

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

### Clean Architecture

```go
func TestArchitecture(t *testing.T) {
    pkgs, err := analyzer.Load(".", "internal/...", "cmd/...")
    if err != nil {
        t.Log(err)
    }
    if len(pkgs) == 0 {
        t.Fatalf("no packages loaded: %v", err)
    }

    m := rules.CleanArch()
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

---

## 4. Domain Isolation (공통 — 모든 모델 동일)

| from → to | domain root | domain sub-pkg | orchestration | shared pkg |
|-----------|:-:|:-:|:-:|:-:|
| **same domain** | O | O | X | O |
| **other domain** | X | X | X | O |
| **orchestration** | O | X | O | O |
| **cmd** | O | X | O | O |
| **shared pkg** | X | X | X | O |

---

## 5. Options

### Migration — 특정 경로 제외

```go
rules.WithExclude("internal/legacy/...")
```

### Warning Mode — 실패 없이 로그만

```go
rules.WithSeverity(rules.Warning)
```

### Model Options (커스텀 모델용)

`WithSublayers`, `WithDirection`, `WithDomainDir`, `WithSharedDir`, `WithOrchestrationDir`, `WithRequireAlias`, `WithAliasFileName`, `WithRequireModel`, `WithModelPath`, `WithPkgRestricted`, `WithDTOAllowedLayers`, `WithBannedPkgNames`, `WithLegacyPkgNames`, `WithLayerDirNames`

---

## 6. Banned Patterns (공통)

| Category | Banned |
|----------|--------|
| Package names | `util`, `common`, `misc`, `helper`, `shared`, `services` |
| Legacy dirs | `router`, `bootstrap`, misplaced `app`/`handler`/`infra` |
| Naming | type stuttering (`order.OrderService`), `Impl` suffix, non-snake_case files, hand-rolled mocks |
| Placement | `middleware/` must be at `internal/pkg/middleware/` |
| DTOs | DDD: `handler/`, `app/`에서만 허용. CleanArch: `handler/`, `usecase/`에서만 허용 |

---

## 7. DDD: Cross-Domain Interface 금지 원칙

> DDD 모델에만 적용. CleanArch에서는 이 규칙 비활성.

도메인 내에서 interface는 **`core/repo/`에만** 정의 가능.

### 위반 시 올바른 수정 방향

cross-domain 데이터가 필요한 endpoint → **`orchestration/handler/http/`로 이동**. interface를 다른 패키지로 옮기는 것은 해결이 아님.

| 우회 시도 | 잡는 규칙 |
|----------|----------|
| handler/app/core/svc에서 interface 정의 | `naming.domain-interface-repo-only` |
| alias.go에서 interface re-export | `structure.domain-alias-no-interface` |
| alias.go에서 core/repo, core/svc type alias | `structure.domain-alias-no-interface` |

---

## 8. Blast Radius (공통)

IQR 기반 통계적 이상치로 비정상 커플링 패키지를 Warning으로 탐지. 설정 불필요. 내부 패키지 5개 미만이면 스킵.

---

## 9. Common Setup Scenarios

### New project from scratch

**DDD:**
1. 디렉토리 구조 생성
2. 각 domain에 `alias.go` + `core/model/*.go` 추가
3. `architecture_test.go` 템플릿 복사
4. `go test -run TestArchitecture -v`

**CleanArch:**
1. 디렉토리 구조 생성 (handler/usecase/entity/gateway/infra)
2. `architecture_test.go` 템플릿 복사 (`rules.CleanArch()` + `WithModel` 사용)
3. `go test -run TestArchitecture -v`

### Existing project with violations

1. 템플릿 복사 후 실행하여 전체 위반 확인
2. 마이그레이션 중인 경로에 `WithExclude` 추가
3. 점진적으로 위반 수정, exclude 제거
4. 전환기에는 `WithSeverity(rules.Warning)` 사용 가능
