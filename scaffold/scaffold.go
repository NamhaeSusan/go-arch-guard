package scaffold

import (
	"fmt"
	"go/format"
	"go/token"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/presets"
)

// Preset identifies a built-in architecture model template.
type Preset = presets.Preset

const (
	PresetDDD             = presets.PresetDDD
	PresetCleanArch       = presets.PresetCleanArch
	PresetLayered         = presets.PresetLayered
	PresetHexagonal       = presets.PresetHexagonal
	PresetModularMonolith = presets.PresetModularMonolith
	PresetConsumerWorker  = presets.PresetConsumerWorker
	PresetBatch           = presets.PresetBatch
	PresetEventPipeline   = presets.PresetEventPipeline
)

// ArchitectureTestOptions controls generated architecture_test.go output.
type ArchitectureTestOptions struct {
	PackageName string
}

// ArchitectureTest returns a ready-to-copy architecture_test.go source file
// for the selected preset. PackageName must be a valid Go package identifier
// such as "myapp_test".
func ArchitectureTest(preset Preset, opts ArchitectureTestOptions) (string, error) {
	packageName := strings.TrimSpace(opts.PackageName)
	if packageName == "" {
		return "", fmt.Errorf("package name is required")
	}
	if !token.IsIdentifier(packageName) {
		return "", fmt.Errorf("package name must be a valid Go identifier: %q", packageName)
	}

	funcs, err := presetFunctions(preset)
	if err != nil {
		return "", err
	}

	src := renderArchitectureTest(packageName, funcs)
	formatted, err := format.Source([]byte(src))
	if err != nil {
		return "", fmt.Errorf("format generated template: %w", err)
	}
	return string(formatted), nil
}

type presetFuncs struct {
	architecture string
	rules        string
}

func presetFunctions(preset Preset) (presetFuncs, error) {
	switch preset {
	case PresetDDD:
		return presetFuncs{architecture: "DDD", rules: "RecommendedDDD"}, nil
	case PresetCleanArch:
		return presetFuncs{architecture: "CleanArch", rules: "RecommendedCleanArch"}, nil
	case PresetLayered:
		return presetFuncs{architecture: "Layered", rules: "RecommendedLayered"}, nil
	case PresetHexagonal:
		return presetFuncs{architecture: "Hexagonal", rules: "RecommendedHexagonal"}, nil
	case PresetModularMonolith:
		return presetFuncs{architecture: "ModularMonolith", rules: "RecommendedModularMonolith"}, nil
	case PresetConsumerWorker:
		return presetFuncs{architecture: "ConsumerWorker", rules: "RecommendedConsumerWorker"}, nil
	case PresetBatch:
		return presetFuncs{architecture: "Batch", rules: "RecommendedBatch"}, nil
	case PresetEventPipeline:
		return presetFuncs{architecture: "EventPipeline", rules: "RecommendedEventPipeline"}, nil
	default:
		return presetFuncs{}, fmt.Errorf("unknown preset %q", preset)
	}
}

func renderArchitectureTest(packageName string, funcs presetFuncs) string {
	var b strings.Builder
	fmt.Fprintf(&b, "package %s\n\n", packageName)
	b.WriteString(`import (
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/presets"
	"github.com/NamhaeSusan/go-arch-guard/report"
)

func TestArchitecture(t *testing.T) {
	pkgs, err := analyzer.Load(".", "internal/...", "cmd/...")
	if err != nil {
		t.Log(err)
	}
	if len(pkgs) == 0 {
		t.Fatalf("no packages loaded: %v", err)
	}
`)
	fmt.Fprintf(&b, "\n\tarch := presets.%s()\n", funcs.architecture)
	b.WriteString("\tctx := core.NewContext(pkgs, \"\", \"\", arch, nil)\n")
	fmt.Fprintf(&b, "\trules := presets.%s()\n\n", funcs.rules)
	b.WriteString("\treport.AssertNoViolations(t, core.Run(ctx, rules))\n")
	b.WriteString("}\n")
	return b.String()
}
