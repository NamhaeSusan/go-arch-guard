---
name: go-arch-guard-consumer
description: Use when scaffolding or maintaining a Kafka/RabbitMQ/SQS consumer worker Go project with go-arch-guard. Handles flat-layout ConsumerWorker preset setup and type pattern guidance.
user_invocable: true
---

# go-arch-guard — Consumer/Worker

Kafka, RabbitMQ, SQS 등 메시지 컨슈머 프로젝트에 아키텍처 가드레일을 적용한다.
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
mkdir -p internal/worker internal/service internal/store internal/model internal/pkg
```

### Step 2: architecture_test.go 생성

파일이 이미 존재하면 덮어쓰지 않고 유저에게 확인한다.
패키지명은 **유효한 Go 패키지 식별자**여야 한다.

```go
import "github.com/NamhaeSusan/go-arch-guard/scaffold"

src, err := scaffold.ArchitectureTest(
    scaffold.PresetConsumerWorker,
    scaffold.ArchitectureTestOptions{PackageName: "myapp_test"},
)
if err != nil {
    return err
}
```

생성된 결과를 `architecture_test.go`에 저장한다.

### Step 3: 검증

```bash
go test -run TestArchitecture -v
```

### Step 4: 안내

> 스캐폴딩 완료. `internal/worker/` 아래에 첫 번째 worker를 추가하세요.
> 파일명은 `worker_xxx.go`, 타입은 `XxxWorker`, `Process` 메서드 필수.

---

## Layout Reference

```text
cmd/worker/main.go
internal/
├── worker/            # worker_order.go, worker_payment.go
├── service/           # 비즈니스 로직
├── store/             # 영속성 (DB, 외부 API)
├── model/             # 데이터 구조체
└── pkg/               # 공유 인프라
    └── consumer/      # 메시지 수신 공통 로직
```

### 레이어 방향

| from | 허용된 import |
|------|--------------|
| `worker` | `service`, `model` |
| `service` | `store`, `model` |
| `store` | `model` |
| `model` | 없음 |

모든 레이어는 `pkg/`를 import 가능 (`model` 제외).

---

## TypePattern

`worker/` 내 `worker_*.go` 파일은 대응하는 타입과 메서드를 반드시 정의해야 한다.

| 파일명 | 필수 타입 | 필수 메서드 |
|--------|----------|-----------|
| `worker_order.go` | `OrderWorker` | `Process` |
| `worker_payment.go` | `PaymentWorker` | `Process` |

`worker/` 내 prefix가 없는 파일 (예: `helper.go`)은 검사 대상 아님.

### 예시

```go
// worker/worker_order.go
package worker

import "context"

type OrderWorker struct {
    service *service.OrderService
}

func (w *OrderWorker) Process(ctx context.Context) error {
    // 메시지 수신 → service 호출
    return w.service.HandleOrderEvent(ctx, event)
}
```

---

## cmd/worker/main.go 패턴

```go
package main

import (
    "context"
    "log"
    "os/signal"
)

func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
    defer stop()

    cfg := config.Load()
    consumer := pkg.NewConsumer(cfg.BrokerAddrs, cfg.GroupID)

    // worker별 토픽 구독
    orderWorker := &worker.OrderWorker{...}
    consumer.Subscribe("order-events", orderWorker.Process)

    if err := consumer.Run(ctx); err != nil {
        log.Fatal(err)
    }
}
```

---

## Options

```go
rules.WithExclude("internal/legacy/...")  // 마이그레이션 중 경로 제외
rules.WithSeverity(rules.Warning)         // 실패 없이 로그만
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
