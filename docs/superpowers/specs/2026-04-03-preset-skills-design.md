# Preset-Specific Skills

**Date:** 2026-04-03
**Status:** Approved

## Summary

Add two new skills (`go-arch-guard-consumer`, `go-arch-guard-batch`) to the
go-arch-guard plugin, each providing preset-specific scaffolding guidance.
Add skill tests that verify a project scaffolded per each skill passes all rules.

## Skills

### go-arch-guard-consumer

**Trigger:** user mentions consumer worker, kafka worker, rabbitmq worker, SQS worker,
message consumer, or similar in a Go project context.

**Content:**
1. Decision Flow — check `internal/` and `architecture_test.go` existence
2. Quick Init:
   - `mkdir -p internal/worker internal/service internal/store internal/model internal/pkg`
   - `scaffold.ArchitectureTest(scaffold.PresetConsumerWorker, ...)`
   - Verify with `go test -run TestArchitecture -v`
3. Layout reference — flat structure diagram
4. Direction table — worker→service→store→model
5. TypePattern guide — `worker/worker_xxx.go` → `XxxWorker.Process()` with example
6. cmd/worker/main.go pattern — consumer setup, worker dispatch
7. Options reference — WithExclude, WithSeverity

### go-arch-guard-batch

**Trigger:** user mentions batch job, cron job, scheduler, batch processing,
or similar in a Go project context.

**Content:** Same structure as consumer skill, adapted for batch:
1. Decision Flow
2. Quick Init with `internal/job` instead of `internal/worker`
3. Layout reference
4. Direction table — job→service→store→model
5. TypePattern guide — `job/job_xxx.go` → `XxxJob.Run()` with example
6. cmd/batch/main.go pattern — flag parsing, job registry, exit codes, dry-run
7. Options reference

## Skill Tests

Add to `skill_test.go`:

### TestSkill_ConsumerWorkerSetup

Creates a valid ConsumerWorker project following the skill guide:
- cmd/worker/main.go
- internal/worker/worker_order.go (OrderWorker + Process)
- internal/service/order.go
- internal/store/order.go
- internal/model/order.go
- internal/pkg/consumer/consumer.go

Runs `rules.RunAll` with `ConsumerWorker()` model → 0 violations.

### TestSkill_BatchSetup

Creates a valid Batch project following the skill guide:
- cmd/batch/main.go
- internal/job/job_expire.go (ExpireJob + Run)
- internal/service/file.go
- internal/store/file.go
- internal/model/file.go
- internal/pkg/batchutil/util.go

Runs `rules.RunAll` with `Batch()` model → 0 violations.

## Scope

| File | Action |
|---|---|
| `plugins/go-arch-guard/skills/go-arch-guard-consumer/SKILL.md` | Create |
| `plugins/go-arch-guard/skills/go-arch-guard-batch/SKILL.md` | Create |
| `plugins/go-arch-guard/.claude-plugin/plugin.json` | Bump 0.0.8 → 0.0.9 |
| `skill_test.go` | Add 2 tests |
