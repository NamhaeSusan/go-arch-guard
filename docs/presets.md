# Preset Details

Detailed layout diagrams and direction tables for each built-in preset.

For custom models, see [Model Concepts](model-concepts.md).

## DDD Layout

```text
internal/
├── domain/
│   └── order/
│       ├── alias.go              # public surface (required)
│       ├── handler/http/         # inbound adapters
│       ├── app/                  # application service
│       ├── core/
│       │   ├── model/            # domain model (required)
│       │   ├── repo/             # repository interface
│       │   └── svc/              # domain service interface
│       ├── event/                # domain events
│       └── infra/persistence/    # outbound adapters
├── orchestration/                # cross-domain coordination
└── pkg/                          # shared utilities
```

DDD layer direction:

| from | allowed to import |
|------|-------------------|
| `handler` | `app` |
| `app` | `core/model`, `core/repo`, `core/svc`, `event` |
| `core` | `core/model` |
| `core/model` | nothing |
| `core/repo` | `core/model` |
| `core/svc` | `core/model` |
| `event` | `core/model` |
| `infra` | `core/repo`, `core/model`, `event` |

## Clean Architecture Layout

```text
internal/
├── domain/
│   └── product/
│       ├── handler/              # interface adapters (controllers)
│       ├── usecase/              # application business rules
│       ├── entity/               # enterprise business rules
│       ├── gateway/              # data access interfaces
│       └── infra/                # frameworks & drivers
├── orchestration/
└── pkg/
```

Clean Architecture layer direction:

| from | allowed to import |
|------|-------------------|
| `handler` | `usecase` |
| `usecase` | `entity`, `gateway` |
| `entity` | nothing |
| `gateway` | `entity` |
| `infra` | `gateway`, `entity` |

## Layered (Spring-style) Layout

```text
internal/
├── domain/
│   └── order/
│       ├── handler/              # HTTP/gRPC handlers
│       ├── service/              # business logic
│       ├── repository/           # data access
│       └── model/                # domain models
├── orchestration/
└── pkg/
```

Layered direction:

| from | allowed to import |
|------|-------------------|
| `handler` | `service` |
| `service` | `repository`, `model` |
| `repository` | `model` |
| `model` | nothing |

## Hexagonal (Ports & Adapters) Layout

```text
internal/
├── domain/
│   └── order/
│       ├── handler/              # driving adapters (HTTP, gRPC)
│       ├── usecase/              # application logic
│       ├── port/                 # interfaces (inbound + outbound)
│       ├── domain/               # entities, value objects
│       └── adapter/              # driven adapters (DB, messaging)
├── orchestration/
└── pkg/
```

Hexagonal direction:

| from | allowed to import |
|------|-------------------|
| `handler` | `usecase` |
| `usecase` | `port`, `domain` |
| `port` | `domain` |
| `domain` | nothing |
| `adapter` | `port`, `domain` |

## Modular Monolith Layout

```text
internal/
├── domain/
│   └── order/
│       ├── api/                  # module public interface
│       ├── application/          # use cases
│       ├── core/                 # entities, value objects
│       └── infrastructure/       # DB, external services
├── orchestration/
└── pkg/
```

Modular Monolith direction:

| from | allowed to import |
|------|-------------------|
| `api` | `application` |
| `application` | `core` |
| `core` | nothing |
| `infrastructure` | `core` |

## Consumer/Worker Layout (Flat)

Unlike domain-centric presets, the Consumer/Worker preset uses a **flat layout** ---
layers live directly under `internal/` with no `domain/` directory.

```text
internal/
├── worker/            # worker_order.go, worker_payment.go
├── service/           # business logic
├── store/             # persistence (DB, external APIs)
├── model/             # data structures
└── pkg/               # shared infra (consumer lib, logging)
    └── consumer/
```

Consumer/Worker direction:

| from | allowed to import |
|------|-------------------|
| `worker` | `service`, `model` |
| `service` | `store`, `model` |
| `store` | `model` |
| `model` | nothing |

All layers may import `pkg/` except `model` (restricted).

**Type pattern enforcement:** Files matching `worker_*.go` in `worker/` must define
a corresponding exported type with a `Process` method:
- `worker_order.go` -> must define `OrderWorker` with `Process` method
- `worker_payment.go` -> must define `PaymentWorker` with `Process` method

Domain isolation rules are not applicable and are skipped entirely.

## Batch Layout (Flat)

The Batch preset uses the same flat layout as Consumer/Worker, with `job/` as the
entry-point layer for cron/scheduler-triggered batch processing.

```text
internal/
├── job/               # job_expire_files.go, job_cleanup_trash.go
├── service/           # business logic
├── store/             # persistence (DB, external APIs)
├── model/             # data structures
└── pkg/               # shared infra (batchutil, logging)
```

Batch direction:

| from | allowed to import |
|------|-------------------|
| `job` | `service`, `model` |
| `service` | `store`, `model` |
| `store` | `model` |
| `model` | nothing |

All layers may import `pkg/` except `model` (restricted).

**Type pattern enforcement:** Files matching `job_*.go` in `job/` must define
a corresponding exported type with a `Run` method:
- `job_expire_files.go` -> must define `ExpireFilesJob` with `Run` method
- `job_cleanup_trash.go` -> must define `CleanupTrashJob` with `Run` method

Domain isolation rules are not applicable and are skipped entirely.

## Event-Driven Pipeline Layout (Flat)

The Event-Driven Pipeline preset uses a flat layout for event-sourcing / CQRS
projects, with dedicated directories for commands, aggregates, events,
projections, and stores.

```text
internal/
├── command/          # command handlers (command_create_order.go)
├── aggregate/        # aggregate roots (aggregate_order.go)
├── event/            # domain events
├── projection/       # read-model projectors
├── eventstore/       # event persistence
├── readstore/        # read-model persistence
├── model/            # shared value objects / DTOs
└── pkg/              # shared infra (eventbus, logging)
```

Event-Driven Pipeline direction:

| from | allowed to import |
|------|-------------------|
| `command` | `aggregate`, `eventstore`, `model` |
| `aggregate` | `event`, `model` |
| `event` | `model` |
| `projection` | `event`, `readstore`, `model` |
| `eventstore` | `event`, `model` |
| `readstore` | `model` |
| `model` | nothing |

All layers may import `pkg/` except `event` and `model` (restricted).

**Type pattern enforcement:** Files matching `command_*.go` in `command/` must
define a corresponding exported type with an `Execute` method:
- `command_create_order.go` -> must define `CreateOrderCommand` with `Execute` method

Files matching `aggregate_*.go` in `aggregate/` must define a corresponding
exported type with an `Apply` method:
- `aggregate_order.go` -> must define `OrderAggregate` with `Apply` method

Domain isolation rules are not applicable and are skipped entirely.
