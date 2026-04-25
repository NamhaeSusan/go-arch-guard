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
		name             string
		preset           scaffold.Preset
		architectureFunc string
		rulesFunc        string
	}{
		{
			name:             "ddd",
			preset:           scaffold.PresetDDD,
			architectureFunc: "DDD",
			rulesFunc:        "RecommendedDDD",
		},
		{
			name:             "clean arch",
			preset:           scaffold.PresetCleanArch,
			architectureFunc: "CleanArch",
			rulesFunc:        "RecommendedCleanArch",
		},
		{
			name:             "layered",
			preset:           scaffold.PresetLayered,
			architectureFunc: "Layered",
			rulesFunc:        "RecommendedLayered",
		},
		{
			name:             "hexagonal",
			preset:           scaffold.PresetHexagonal,
			architectureFunc: "Hexagonal",
			rulesFunc:        "RecommendedHexagonal",
		},
		{
			name:             "modular monolith",
			preset:           scaffold.PresetModularMonolith,
			architectureFunc: "ModularMonolith",
			rulesFunc:        "RecommendedModularMonolith",
		},
		{
			name:             "consumer worker",
			preset:           scaffold.PresetConsumerWorker,
			architectureFunc: "ConsumerWorker",
			rulesFunc:        "RecommendedConsumerWorker",
		},
		{
			name:             "batch",
			preset:           scaffold.PresetBatch,
			architectureFunc: "Batch",
			rulesFunc:        "RecommendedBatch",
		},
		{
			name:             "event pipeline",
			preset:           scaffold.PresetEventPipeline,
			architectureFunc: "EventPipeline",
			rulesFunc:        "RecommendedEventPipeline",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, err := scaffold.ArchitectureTest(tt.preset, scaffold.ArchitectureTestOptions{PackageName: "myapp_test"})
			if err != nil {
				t.Fatalf("ArchitectureTest() error = %v", err)
			}
			contains := []string{
				"package myapp_test",
				"func TestArchitecture(t *testing.T)",
				"\"github.com/NamhaeSusan/go-arch-guard/core\"",
				"\"github.com/NamhaeSusan/go-arch-guard/presets\"",
				"arch := presets." + tt.architectureFunc + "()",
				"ctx := core.NewContext(pkgs, \"\", \"\", arch, nil)",
				"rules := presets." + tt.rulesFunc + "()",
				"report.AssertNoViolations(t, core.Run(ctx, rules))",
			}
			for _, fragment := range contains {
				if !strings.Contains(src, fragment) {
					t.Fatalf("expected generated source to contain %q\n%s", fragment, src)
				}
			}
			for _, fragment := range []string{
				"\"github.com/NamhaeSusan/go-arch-guard/rules\"",
				"rules.RunAll",
				"rules.WithModel",
			} {
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
	if !strings.Contains(src, "arch := presets.ConsumerWorker()") {
		t.Error("generated source must call presets.ConsumerWorker()")
	}
	if !strings.Contains(src, "rules := presets.RecommendedConsumerWorker()") {
		t.Error("generated source must call presets.RecommendedConsumerWorker()")
	}
	if !strings.Contains(src, "package myapp_test") {
		t.Error("generated source must have correct package name")
	}
}

func TestArchitectureTestDefaultsInternalRoot(t *testing.T) {
	src, err := scaffold.ArchitectureTest(scaffold.PresetDDD, scaffold.ArchitectureTestOptions{PackageName: "myapp_test"})
	if err != nil {
		t.Fatal(err)
	}
	want := `analyzer.Load(".", "internal/...", "cmd/...")`
	if !strings.Contains(src, want) {
		t.Fatalf("default scaffold must emit %q\n%s", want, src)
	}
	// Default InternalRoot must NOT emit a redundant override line — the
	// normalization inside cloneArchitecture already handles it.
	if strings.Contains(src, "arch.Layout.InternalRoot") {
		t.Fatalf("default scaffold must NOT emit arch.Layout.InternalRoot assignment\n%s", src)
	}
}

func TestArchitectureTestCustomInternalRoot(t *testing.T) {
	src, err := scaffold.ArchitectureTest(scaffold.PresetDDD, scaffold.ArchitectureTestOptions{
		PackageName:  "myapp_test",
		InternalRoot: "packages",
	})
	if err != nil {
		t.Fatal(err)
	}
	wantLoad := `analyzer.Load(".", "packages/...", "cmd/...")`
	if !strings.Contains(src, wantLoad) {
		t.Fatalf("custom InternalRoot must emit %q\n%s", wantLoad, src)
	}
	if strings.Contains(src, `"internal/..."`) {
		t.Fatalf("custom InternalRoot must NOT emit \"internal/...\"\n%s", src)
	}
	// Critical: the rule classifier reads arch.Layout.InternalRoot, so the
	// generated source must explicitly set it on the preset arch. Without this
	// line, analyzer.Load would scan packages/ but rules would still classify
	// against internal/, leaving every layout-dependent rule effectively a
	// no-op.
	wantArch := `arch.Layout.InternalRoot = "packages"`
	if !strings.Contains(src, wantArch) {
		t.Fatalf("custom InternalRoot must emit %q on arch\n%s", wantArch, src)
	}
	// Order: the assignment must come AFTER `arch := presets.DDD()` and
	// BEFORE `core.NewContext(...)` so the override is in effect when the
	// context clones and normalizes the architecture.
	archIdx := strings.Index(src, "arch := presets.")
	overrideIdx := strings.Index(src, wantArch)
	ctxIdx := strings.Index(src, "core.NewContext")
	if archIdx < 0 || archIdx >= overrideIdx || overrideIdx >= ctxIdx {
		t.Fatalf("InternalRoot override must sit between presets.X() and core.NewContext(): archIdx=%d overrideIdx=%d ctxIdx=%d\n%s", archIdx, overrideIdx, ctxIdx, src)
	}
}

func TestArchitectureTestRejectsInternalRootWithSeparator(t *testing.T) {
	_, err := scaffold.ArchitectureTest(scaffold.PresetDDD, scaffold.ArchitectureTestOptions{
		PackageName:  "myapp_test",
		InternalRoot: "foo/bar",
	})
	if err == nil {
		t.Fatal("expected error for InternalRoot containing slash")
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
