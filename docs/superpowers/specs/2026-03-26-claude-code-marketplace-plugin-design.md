# Design: Claude Code Marketplace Plugin Wrapper

## Purpose

Make this repository installable as a Claude Code marketplace while preserving the existing root `SKILL.md` for non-Claude workflows.

The user goal is distribution convenience, not new guardrail behavior. The implementation should therefore add the smallest possible Claude Code packaging layer around the current skill.

## Chosen Approach

Keep the existing root `SKILL.md` as the canonical skill document and add a marketplace/plugin wrapper inside the repository:

- repository root `.claude-plugin/marketplace.json`
- plugin directory `plugins/go-arch-guard/.claude-plugin/plugin.json`
- plugin skill path `plugins/go-arch-guard/skills/go-arch-guard/SKILL.md`

This keeps the current repo usable as:

1. a Go library via `go get`
2. a plain skill repository via the root `SKILL.md`
3. a Claude Code marketplace via `/plugin marketplace add <repo>` and `/plugin install ...`

## Trade-offs Considered

### 1. Convert the repo fully to Claude-only plugin layout

- Pros: single packaging model
- Cons: breaks the current root-skill workflow and reduces portability

Rejected because the user explicitly approved keeping the root skill.

### 2. Separate plugin repo

- Pros: clean separation
- Cons: duplicate maintenance, extra release/distribution overhead

Rejected because the user asked for an easy install path for "this" skill in this repository.

### 3. Wrapper inside this repo

- Pros: minimal change, easy distribution, preserves existing usage
- Cons: adds a small amount of packaging metadata

Chosen.

## File-Level Design

- `SKILL.md`
  - stays as the canonical skill content
- `.claude-plugin/marketplace.json`
  - declares this repository as a plugin marketplace
- `plugins/go-arch-guard/.claude-plugin/plugin.json`
  - declares the installable Claude Code plugin
- `plugins/go-arch-guard/skills/go-arch-guard/SKILL.md`
  - points to or mirrors the canonical skill so the plugin ships the same behavior
- `README.md`
  - documents Claude Code marketplace install/validation flow

## Validation

Because this work is packaging/documentation, not Go rule behavior, the main proof is plugin validation plus regression checks:

1. `claude plugin validate .`
2. `go test ./...`
3. `make lint`

## Non-Goals

- No new architecture rules
- No changes to analyzer/rules/report packages
- No custom hooks, agents, MCP servers, or extra plugin features beyond the marketplace wrapper
