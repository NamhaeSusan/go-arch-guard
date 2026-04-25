---
name: go-arch-guard-event-pipeline
description: Use when scaffolding or maintaining an event-sourcing / CQRS Go project with go-arch-guard. Handles flat-layout EventPipeline preset setup with command, aggregate, event, projection, eventstore, readstore layers.
user_invocable: true
---

# go-arch-guard — Event-Driven Pipeline

이벤트 소싱 / CQRS 프로젝트에 아키텍처 가드레일을 적용한다.
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
mkdir -p internal/command internal/aggregate internal/event \
         internal/projection internal/eventstore internal/readstore \
         internal/model internal/pkg
```

### Step 2: architecture_test.go 생성

```go
import "github.com/NamhaeSusan/go-arch-guard/scaffold"

src, err := scaffold.ArchitectureTest(
    scaffold.PresetEventPipeline,
    scaffold.ArchitectureTestOptions{PackageName: "myapp_test"},
)
if err != nil {
    return err
}
```

비표준 패키지 루트(`packages/`, `src/`)를 쓰면 `ArchitectureTestOptions{..., InternalRoot: "packages"}`로 같이 지정한다.

### Step 3: 검증

```bash
go test -run TestArchitecture -v
```

### Step 4: 안내

> 스캐폴딩 완료.
> - `internal/command/`에 커맨드 추가: `command_xxx.go` → `XxxCommand.Execute()`
> - `internal/aggregate/`에 애그리거트 추가: `aggregate_xxx.go` → `XxxAggregate.Apply()`

---

## Layout Reference

```text
cmd/pipeline/main.go
internal/
├── command/       # command_create_order.go → CreateOrderCommand.Execute()
├── aggregate/     # aggregate_order.go → OrderAggregate.Apply()
├── event/         # 이벤트 정의 (OrderCreated, OrderShipped 등)
├── projection/    # 읽기 모델 구축
├── eventstore/    # 이벤트 스토어 (append-only)
├── readstore/     # 읽기 스토어 (projection용 CRUD)
├── model/         # 공유 타입
└── pkg/           # 공유 인프라 (eventbus 등)
```

### 레이어 방향

| from | 허용된 import |
|------|--------------|
| `command` | `aggregate`, `eventstore`, `model` |
| `aggregate` | `event`, `model` |
| `event` | `model` |
| `projection` | `event`, `readstore`, `model` |
| `eventstore` | `event`, `model` |
| `readstore` | `model` |
| `model` | 없음 |

`model`, `event`를 제외한 모든 레이어는 `pkg/`를 import 가능.

---

## TypePattern

### command/

`command/` 내 `command_*.go` 파일은 대응하는 타입과 메서드를 정의해야 한다.

| 파일명 | 필수 타입 | 필수 메서드 |
|--------|----------|-----------|
| `command_create_order.go` | `CreateOrderCommand` | `Execute` |
| `command_ship_order.go` | `ShipOrderCommand` | `Execute` |

### aggregate/

`aggregate/` 내 `aggregate_*.go` 파일은 대응하는 타입과 메서드를 정의해야 한다.

| 파일명 | 필수 타입 | 필수 메서드 |
|--------|----------|-----------|
| `aggregate_order.go` | `OrderAggregate` | `Apply` |
| `aggregate_user.go` | `UserAggregate` | `Apply` |

prefix가 없는 파일 (예: `helper.go`, `types.go`)은 검사 대상 아님.

### 예시

```go
// command/command_create_order.go
package command

import "context"

type CreateOrderCommand struct {
    agg   *aggregate.OrderAggregate
    store eventstore.Store
}

func (c *CreateOrderCommand) Execute(ctx context.Context) error {
    evt := c.agg.Apply(ctx)
    return c.store.Save(ctx, evt)
}
```

```go
// aggregate/aggregate_order.go
package aggregate

import "context"

type OrderAggregate struct{}

func (a *OrderAggregate) Apply(ctx context.Context) error {
    // 비즈니스 규칙 검증 + 이벤트 생성
    return nil
}
```

---

## 데이터 흐름

```
Command → Aggregate → Event
      ↘ EventStore ←──┘
                  ↓
            Projection → ReadStore
```

1. **Command** 수신 → **Aggregate**에 위임 + **EventStore**에 저장
2. **Aggregate**가 비즈니스 규칙 검증 → **Event** 생성 (저장소 의존 없음)
3. **Projection**이 **Event**를 구독 → 읽기 모델 구축 → **ReadStore**에 저장

---

## Options

```go
ctx := core.NewContext(pkgs, "", "", presets.EventPipeline(), []string{"internal/legacy/..."})
core.Run(ctx, presets.RecommendedEventPipeline(),
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
