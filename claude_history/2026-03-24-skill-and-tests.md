# go-arch-guard Skill + Skill Tests

## Task
Go 일반 서버용 go-arch-guard skill.md 작성 및 skill 패턴 검증 테스트 추가.

## Changes

### ~/.claude/skills/go-arch-guard/SKILL.md (new)
- go-arch-guard를 Go 서버 프로젝트에 설정하는 방법을 Claude가 참조하는 skill
- install, target structure, architecture_test.go template, import rules quick reference, options, banned patterns, common setup scenarios 포함

### skill_test.go (new)
- `TestSkill_NewProjectSetup` — skill template대로 새 프로젝트 구조 생성 시 4개 rule 모두 PASS 검증
- `TestSkill_AutoExtractModuleRoot` — `""` 자동 추출 검증
- `TestSkill_ExcludeOption` — WithExclude 마이그레이션 시나리오 검증
- `TestSkill_WarningMode` — WithSeverity(Warning) 모드 검증
- `TestSkill_BannedPatterns` — 금지 패키지명 탐지 검증
- `TestSkill_CrossDomainViolation` — cross-domain import 탐지 검증
- `TestSkill_LayerDirectionViolation` — 레이어 방향 위반 탐지 검증

## Verification
- `go test -run TestSkill -v`: 7/7 PASS
- `go test ./...`: all packages PASS
- `make lint`: 0 issues
