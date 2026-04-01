package scaffold

import (
	"fmt"
	"go/format"
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
)

// ArchitectureTestOptions controls generated architecture_test.go output.
type ArchitectureTestOptions struct {
	PackageName string
}

// ArchitectureTest returns a ready-to-copy architecture_test.go source file
// for the selected preset.
func ArchitectureTest(preset Preset, opts ArchitectureTestOptions) (string, error) {
	packageName := strings.TrimSpace(opts.PackageName)
	if packageName == "" {
		return "", fmt.Errorf("package name is required")
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
	writeCheckRun(&b, "domain isolation", fmt.Sprintf("rules.CheckDomainIsolation(pkgs, \"\", \"\"%s)", callOptionSuffix(modelFunc)))
	writeCheckRun(&b, "layer direction", fmt.Sprintf("rules.CheckLayerDirection(pkgs, \"\", \"\"%s)", callOptionSuffix(modelFunc)))
	writeCheckRun(&b, "naming", fmt.Sprintf("rules.CheckNaming(pkgs%s)", namingOptionSuffix(modelFunc)))
	writeCheckRun(&b, "structure", fmt.Sprintf("rules.CheckStructure(\".\"%s)", callOptionSuffix(modelFunc)))
	writeCheckRun(&b, "blast radius", fmt.Sprintf("rules.AnalyzeBlastRadius(pkgs, \"\", \"\"%s)", callOptionSuffix(modelFunc)))
	b.WriteString("}\n")
	return b.String()
}

func writeCheckRun(b *strings.Builder, name, call string) {
	fmt.Fprintf(b, "\tt.Run(%q, func(t *testing.T) {\n", name)
	fmt.Fprintf(b, "\t\treport.AssertNoViolations(t, %s)\n", call)
	b.WriteString("\t})\n")
}

func callOptionSuffix(modelFunc string) string {
	if modelFunc == "" {
		return ""
	}
	return ", opts..."
}

func namingOptionSuffix(modelFunc string) string {
	if modelFunc == "" {
		return ""
	}
	return ", opts..."
}
