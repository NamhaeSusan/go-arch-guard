package scaffold_test

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/scaffold"
)

func TestArchitectureTest(t *testing.T) {
	tests := []struct {
		name        string
		preset      scaffold.Preset
		contains    []string
		notContains []string
	}{
		{
			name:        "ddd",
			preset:      scaffold.PresetDDD,
			contains:    []string{"package myapp_test", "func TestArchitecture", "rules.RunAll(pkgs, \"\", \"\")"},
			notContains: []string{"rules.WithModel(m)"},
		},
		{
			name:     "clean arch",
			preset:   scaffold.PresetCleanArch,
			contains: []string{"m := rules.CleanArch()", "opts := []rules.Option{rules.WithModel(m)}", "rules.RunAll(pkgs, \"\", \"\", opts...)"},
		},
		{
			name:     "layered",
			preset:   scaffold.PresetLayered,
			contains: []string{"m := rules.Layered()"},
		},
		{
			name:     "hexagonal",
			preset:   scaffold.PresetHexagonal,
			contains: []string{"m := rules.Hexagonal()"},
		},
		{
			name:     "modular monolith",
			preset:   scaffold.PresetModularMonolith,
			contains: []string{"m := rules.ModularMonolith()"},
		},
		{
			name:     "consumer worker",
			preset:   scaffold.PresetConsumerWorker,
			contains: []string{"m := rules.ConsumerWorker()"},
		},
		{
			name:     "batch",
			preset:   scaffold.PresetBatch,
			contains: []string{"m := rules.Batch()"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, err := scaffold.ArchitectureTest(tt.preset, scaffold.ArchitectureTestOptions{PackageName: "myapp_test"})
			if err != nil {
				t.Fatalf("ArchitectureTest() error = %v", err)
			}
			for _, fragment := range tt.contains {
				if !strings.Contains(src, fragment) {
					t.Fatalf("expected generated source to contain %q\n%s", fragment, src)
				}
			}
			for _, fragment := range tt.notContains {
				if strings.Contains(src, fragment) {
					t.Fatalf("expected generated source not to contain %q\n%s", fragment, src)
				}
			}
			if _, err := parser.ParseFile(token.NewFileSet(), "architecture_test.go", src, parser.AllErrors); err != nil {
				t.Fatalf("generated source must parse: %v\n%s", err, src)
			}
		})
	}
}

func TestArchitectureTest_ConsumerWorker(t *testing.T) {
	src, err := scaffold.ArchitectureTest(scaffold.PresetConsumerWorker, scaffold.ArchitectureTestOptions{PackageName: "myapp_test"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(src, "rules.ConsumerWorker()") {
		t.Error("generated source must call rules.ConsumerWorker()")
	}
	if !strings.Contains(src, "package myapp_test") {
		t.Error("generated source must have correct package name")
	}
}

func TestArchitectureTestErrors(t *testing.T) {
	if _, err := scaffold.ArchitectureTest(scaffold.PresetDDD, scaffold.ArchitectureTestOptions{}); err == nil {
		t.Fatal("expected empty package name error")
	}
	if _, err := scaffold.ArchitectureTest(scaffold.PresetDDD, scaffold.ArchitectureTestOptions{PackageName: "go-arch-guard_test"}); err == nil {
		t.Fatal("expected invalid package name error")
	}
	if _, err := scaffold.ArchitectureTest(scaffold.Preset("weird"), scaffold.ArchitectureTestOptions{PackageName: "myapp_test"}); err == nil {
		t.Fatal("expected unknown preset error")
	}
}
