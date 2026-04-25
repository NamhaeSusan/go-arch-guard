---
name: go-arch-guard-batch
description: Use when scaffolding or maintaining a cron/scheduler batch job Go project with go-arch-guard. Handles flat-layout Batch preset setup and type pattern guidance.
user_invocable: true
---

# go-arch-guard — Batch

cron, 스케줄러(Airflow 등), 수동 실행 기반 배치 프로젝트에 아키텍처 가드레일을 적용한다.
**플랫 레이아웃** — `internal/domain/` 없이 레이어가 `internal/` 바로 아래에 위치.

## Decision Flow

1. `internal/` 존재 여부 확인
2. `architecture_test.go` 존재 여부 확인

**새 프로젝트** → **Quick Init** 실행
**기존 프로젝트** → **Layout Reference** 참조하여 수정

---

## Quick Init

### Step 1: 스캐폴딩

```bash
go get github.com/NamhaeSusan/go-arch-guard
mkdir -p internal/job internal/service internal/store internal/model internal/pkg
```

### Step 2: architecture_test.go 생성

파일이 이미 존재하면 덮어쓰지 않고 유저에게 확인한다.
패키지명은 **유효한 Go 패키지 식별자**여야 한다.

```go
import "github.com/NamhaeSusan/go-arch-guard/scaffold"

src, err := scaffold.ArchitectureTest(
    scaffold.PresetBatch,
    scaffold.ArchitectureTestOptions{PackageName: "myapp_test"},
)
if err != nil {
    return err
}
```

생성된 결과를 `architecture_test.go`에 저장한다. 비표준 패키지 루트(`packages/`, `src/`)를 쓰면 `ArchitectureTestOptions{..., InternalRoot: "packages"}`로 같이 지정한다.

### Step 3: 검증

```bash
go test -run TestArchitecture -v
```

### Step 4: 안내

> 스캐폴딩 완료. `internal/job/` 아래에 첫 번째 배치 job을 추가하세요.
> 파일명은 `job_xxx.go`, 타입은 `XxxJob`, `Run` 메서드 필수.

---

## Layout Reference

```text
cmd/batch/main.go
internal/
├── job/               # job_expire_files.go, job_cleanup_trash.go
├── service/           # 비즈니스 로직
├── store/             # 영속성 (DB, 외부 API)
├── model/             # 데이터 구조체
└── pkg/               # 공유 인프라
    └── batchutil/     # 배치 공통 (progress reporter 등)
```

### 레이어 방향

| from | 허용된 import |
|------|--------------|
| `job` | `service`, `model` |
| `service` | `store`, `model` |
| `store` | `model` |
| `model` | 없음 |

모든 레이어는 `pkg/`를 import 가능 (`model` 제외).

---

## TypePattern

`job/` 내 `job_*.go` 파일은 대응하는 타입과 메서드를 반드시 정의해야 한다.

| 파일명 | 필수 타입 | 필수 메서드 |
|--------|----------|-----------|
| `job_expire_files.go` | `ExpireFilesJob` | `Run` |
| `job_cleanup_trash.go` | `CleanupTrashJob` | `Run` |

`job/` 내 prefix가 없는 파일 (예: `job.go`, `params.go`)은 검사 대상 아님.

### 예시

```go
// job/job_expire_files.go
package job

import "context"

type ExpireFilesJob struct {
    service *service.FileService
}

func (j *ExpireFilesJob) Run(ctx context.Context) error {
    cursor := ""
    for {
        files, next, err := j.service.FetchExpired(ctx, cursor, 100)
        if err != nil {
            return err
        }
        if len(files) == 0 {
            break
        }
        for _, f := range files {
            _ = j.service.Expire(ctx, f)
        }
        cursor = next
    }
    return nil
}
```

### 공통 타입 (선택사항)

`job/job.go`에 공통 인터페이스나 타입을 정의할 수 있다. 강제는 아님.

```go
// job/job.go (optional)
package job

type Params struct {
    StartDate  string
    EndDate    string
    BatchSize  int
    DryRun     bool
    ResumeFrom string
}

type Result struct {
    Total     int
    Processed int
    Skipped   int
    Failed    int
}
```

---

## cmd/batch/main.go 패턴

```go
package main

import (
    "context"
    "flag"
    "log"
    "os"
)

func main() {
    jobName := flag.String("job", "", "job name to run")
    dryRun := flag.Bool("dry-run", false, "dry run mode")
    flag.Parse()

    cfg := config.Load()

    registry := map[string]interface{ Run(context.Context) error }{
        "expire_files":  newExpireFilesJob(cfg),
        "cleanup_trash": newCleanupTrashJob(cfg),
    }

    j, ok := registry[*jobName]
    if !ok {
        log.Fatalf("unknown job: %s", *jobName)
    }

    if err := j.Run(context.Background()); err != nil {
        log.Fatalf("job failed: %v", err)
    }

    log.Println("done")
}
```

**실행:** `./batch --job expire_files --dry-run`

---

## Options

```go
ctx := core.NewContext(pkgs, "", "", presets.Batch(), []string{"internal/legacy/..."})
core.Run(ctx, presets.RecommendedBatch(),
    core.WithSeverityOverride("layer.direction", core.Warning))
```

---

## 적용되는 규칙

| 카테고리 | 규칙 |
|---------|------|
| 레이어 방향 | `layer.direction`, `layer.inner-imports-pkg` |
| 구조 | `structure.internal-top-level`, `structure.banned-package`, `structure.legacy-package`, `structure.middleware-placement` |
| 네이밍 | `naming.no-stutter`, `naming.no-impl-suffix`, `naming.snake-case-file`, `naming.no-layer-suffix`, `testing.no-handmock` |
| 타입 패턴 | `naming.type-pattern-mismatch`, `naming.type-pattern-missing-method` |
| 인터페이스 패턴 | `interface.exported-impl`, `interface.constructor-name`, `interface.constructor-returns-interface`, `interface.single-per-package` |
| 커플링 | `blast.high-coupling` |

**미적용:** `isolation.*` (도메인 격리) — 플랫 레이아웃에는 도메인 개념 없음.
