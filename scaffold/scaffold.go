package scaffold

import (
	"fmt"
	"go/format"
	"go/token"
	"strings"
)

// Preset identifies a built-in architecture model template.
type Preset string

const (
	PresetDDD             Preset = "ddd"
	PresetCleanArch       Preset = "clean-arch"
	PresetLayered         Preset = "layered"
	PresetHexagonal       Preset = "hexagonal"
	PresetModularMonolith Preset = "modular-monolith"
	PresetConsumerWorker  Preset = "consumer-worker"
	PresetBatch           Preset = "batch"
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

	modelFunc, err := presetModelFunc(preset)
	if err != nil {
		return "", err
	}

	src := renderArchitectureTest(packageName, modelFunc)
	formatted, err := format.Source([]byte(src))
	if err != nil {
		return "", fmt.Errorf("format generated template: %w", err)
	}
	return string(formatted), nil
}

func presetModelFunc(preset Preset) (string, error) {
	switch preset {
	case PresetDDD:
		return "", nil
	case PresetCleanArch:
		return "CleanArch", nil
	case PresetLayered:
		return "Layered", nil
	case PresetHexagonal:
		return "Hexagonal", nil
	case PresetModularMonolith:
		return "ModularMonolith", nil
	case PresetConsumerWorker:
		return "ConsumerWorker", nil
	case PresetBatch:
		return "Batch", nil
	default:
		return "", fmt.Errorf("unknown preset %q", preset)
	}
}

func renderArchitectureTest(packageName, modelFunc string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "package %s\n\n", packageName)
	b.WriteString(`import (
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
`)
	if modelFunc != "" {
		fmt.Fprintf(&b, "\n\tm := rules.%s()\n\topts := []rules.Option{rules.WithModel(m)}\n", modelFunc)
	}
	b.WriteString("\n")
	if modelFunc == "" {
		b.WriteString("\treport.AssertNoViolations(t, rules.RunAll(pkgs, \"\", \"\"))\n")
	} else {
		b.WriteString("\treport.AssertNoViolations(t, rules.RunAll(pkgs, \"\", \"\", opts...))\n")
	}
	b.WriteString("}\n")
	return b.String()
}
