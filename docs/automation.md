# Automation Surfaces

`go-arch-guard` exposes small, explicit surfaces aimed at AI agents, bots, and CI automation.

For the simplest integration path, `rules.RunAll(...)` executes the recommended
built-in rule bundle and returns one merged violation slice.

## Preset `architecture_test.go` Templates

Use `scaffold.ArchitectureTest(...)` when you want a ready-to-copy test file for a built-in preset.

```go
import "github.com/NamhaeSusan/go-arch-guard/scaffold"

src, err := scaffold.ArchitectureTest(
    scaffold.PresetHexagonal,
    scaffold.ArchitectureTestOptions{PackageName: "myapp_test"},
)
```

`PackageName` must be a valid Go package identifier.

Available presets:

- `scaffold.PresetDDD`
- `scaffold.PresetCleanArch`
- `scaffold.PresetLayered`
- `scaffold.PresetHexagonal`
- `scaffold.PresetModularMonolith`

The generated file includes `domain isolation`, `layer direction`, `naming`, `structure`, and `blast radius` checks.

## Machine-readable JSON Reports

Use `report.BuildJSONReport(...)`, `report.MarshalJSONReport(...)`, or `report.WriteJSONReport(...)` when the violations need to be consumed by another tool.

```go
import "github.com/NamhaeSusan/go-arch-guard/report"

data, err := report.MarshalJSONReport(violations)
if err != nil {
    return err
}
fmt.Println(string(data))
```

The JSON report includes:

- schema marker (`go-arch-guard.report.v1`)
- summary counts (`total`, `errors`, `warnings`, `files`)
- sorted rule ids
- each violation with string severity (`error` / `warning`)
