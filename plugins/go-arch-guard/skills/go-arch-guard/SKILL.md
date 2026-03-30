---
name: go-arch-guard
description: Go 서버 프로젝트에 go-arch-guard 아키텍처 가드레일 설정 및 초기 스캐폴딩. DDD / Clean Architecture / Layered / Hexagonal / Modular Monolith 프리셋 또는 커스텀 모델 지원.
---

# go-arch-guard Setup Guide

Go 서버 프로젝트에 아키텍처 가드레일을 `go test`로 적용하는 방법.

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

### Option C: Layered (Spring 스타일)

```text
internal/
├── domain/
│   └── {domain}/
│       ├── handler/              # HTTP/gRPC 핸들러
│       ├── service/              # 비즈니스 로직
│       ├── repository/           # 데이터 접근
│       └── model/                # 도메인 모델
├── orchestration/
└── pkg/
```

**레이어 방향:**

```text
handler → service → repository, model
repository → model
model → (nothing)
```

**Layered 제약:**
- `alias.go` 불필요
- `model`은 `internal/pkg` import 금지
- DTO는 `handler/`, `service/`에서만 허용

### Option D: Hexagonal (포트 & 어댑터)

```text
internal/
├── domain/
│   └── {domain}/
│       ├── handler/              # 드라이빙 어댑터 (HTTP, gRPC)
│       ├── usecase/              # 애플리케이션 로직
│       ├── port/                 # 인터페이스 (인바운드 + 아웃바운드)
│       ├── domain/               # 엔티티, 값 객체
│       └── adapter/              # 드리븐 어댑터 (DB, 메시징)
├── orchestration/
└── pkg/
```

**레이어 방향:**

```text
handler → usecase → port, domain
adapter → port, domain
port → domain
domain → (nothing)
```

**Hexagonal 제약:**
- `alias.go` 불필요
- `domain`은 `internal/pkg` import 금지
- DTO는 `handler/`, `usecase/`에서만 허용

### Option E: Modular Monolith

```text
internal/
├── domain/
│   └── {domain}/
│       ├── api/                  # 모듈 공개 인터페이스
│       ├── application/          # 유즈케이스
│       ├── domain/               # 엔티티, 값 객체
│       └── infrastructure/       # DB, 외부 서비스
├── orchestration/
└── pkg/
```

**레이어 방향:**

```text
api → application → domain
infrastructure → domain
domain → (nothing)
```

**ModularMonolith 제약:**
- `alias.go` 불필요
- `domain`은 `internal/pkg` import 금지
- DTO는 `api/`, `application/`에서만 허용

### Option F: Custom

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

### Layered (Spring 스타일)

```go
func TestArchitecture(t *testing.T) {
    pkgs, err := analyzer.Load(".", "internal/...", "cmd/...")
    if err != nil {
        t.Log(err)
    }
    if len(pkgs) == 0 {
        t.Fatalf("no packages loaded: %v", err)
    }

    m := rules.Layered()
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

### Hexagonal (포트 & 어댑터)

```go
func TestArchitecture(t *testing.T) {
    pkgs, err := analyzer.Load(".", "internal/...", "cmd/...")
    if err != nil {
        t.Log(err)
    }
    if len(pkgs) == 0 {
        t.Fatalf("no packages loaded: %v", err)
    }

    m := rules.Hexagonal()
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

### Modular Monolith

```go
func TestArchitecture(t *testing.T) {
    pkgs, err := analyzer.Load(".", "internal/...", "cmd/...")
    if err != nil {
        t.Log(err)
    }
    if len(pkgs) == 0 {
        t.Fatalf("no packages loaded: %v", err)
    }

    m := rules.ModularMonolith()
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
| DTOs | DDD: `handler/`, `app/`에서만 허용. CleanArch: `handler/`, `usecase/`에서만 허용. Layered: `handler/`, `service/`에서만 허용. Hexagonal: `handler/`, `usecase/`에서만 허용. ModularMonolith: `api/`, `application/`에서만 허용 |

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

**Layered:**
1. 디렉토리 구조 생성 (handler/service/repository/model)
2. `architecture_test.go` 템플릿 복사 (`rules.Layered()` + `WithModel` 사용)
3. `go test -run TestArchitecture -v`

**Hexagonal:**
1. 디렉토리 구조 생성 (handler/usecase/port/domain/adapter)
2. `architecture_test.go` 템플릿 복사 (`rules.Hexagonal()` + `WithModel` 사용)
3. `go test -run TestArchitecture -v`

**ModularMonolith:**
1. 디렉토리 구조 생성 (api/application/domain/infrastructure)
2. `architecture_test.go` 템플릿 복사 (`rules.ModularMonolith()` + `WithModel` 사용)
3. `go test -run TestArchitecture -v`

### Existing project with violations

1. 템플릿 복사 후 실행하여 전체 위반 확인
2. 마이그레이션 중인 경로에 `WithExclude` 추가
3. 점진적으로 위반 수정, exclude 제거
4. 전환기에는 `WithSeverity(rules.Warning)` 사용 가능
