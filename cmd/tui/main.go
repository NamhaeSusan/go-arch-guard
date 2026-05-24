package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/presets"
	"github.com/NamhaeSusan/go-arch-guard/tui"
)

type presetEntry struct {
	arch    func() core.Architecture
	ruleset func() core.RuleSet
}

var presetTable = map[string]presetEntry{
	"ddd":              {presets.DDD, presets.RecommendedDDD},
	"cleanarch":        {presets.CleanArch, presets.RecommendedCleanArch},
	"layered":          {presets.Layered, presets.RecommendedLayered},
	"hexagonal":        {presets.Hexagonal, presets.RecommendedHexagonal},
	"modular-monolith": {presets.ModularMonolith, presets.RecommendedModularMonolith},
	"consumer-worker":  {presets.ConsumerWorker, presets.RecommendedConsumerWorker},
	"batch":            {presets.Batch, presets.RecommendedBatch},
	"event-pipeline":   {presets.EventPipeline, presets.RecommendedEventPipeline},
}

func presetNames() string {
	names := make([]string, 0, len(presetTable))
	for k := range presetTable {
		names = append(names, k)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

func main() {
	preset := flag.String("preset", "ddd", "architecture preset: "+presetNames())
	internalRoot := flag.String("internal-root", "", `package root to analyze, e.g. "internal", "packages", or "src"`)
	flag.Parse()

	dir := "."
	if flag.NArg() > 0 {
		dir = flag.Arg(0)
	}

	entry, ok := presetTable[*preset]
	if !ok {
		fmt.Fprintf(os.Stderr, "error: unknown preset %q (one of: %s)\n", *preset, presetNames())
		os.Exit(2)
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	arch := entry.arch()
	applyInternalRoot(&arch, *internalRoot)
	pkgs, err := analyzer.Load(dir, loadPatterns(arch)...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: %v\n", err)
	}
	if len(pkgs) == 0 {
		fmt.Fprintln(os.Stderr, "no packages found")
		os.Exit(1)
	}

	module := ""
	for _, pkg := range pkgs {
		if pkg.Module != nil && pkg.Module.Path != "" {
			module = pkg.Module.Path
			break
		}
	}
	if module == "" {
		fmt.Fprintln(os.Stderr, "error: cannot determine module path; ensure go.mod is present")
		os.Exit(1)
	}

	if err := tui.Run(pkgs, module, absDir, arch, entry.ruleset()); err != nil {
		fmt.Fprintf(os.Stderr, "tui error: %v\n", err)
		os.Exit(1)
	}
}

func applyInternalRoot(arch *core.Architecture, internalRoot string) {
	internalRoot = strings.Trim(strings.TrimSpace(internalRoot), "/")
	if internalRoot != "" {
		arch.Layout.InternalRoot = internalRoot
	}
}

func loadPatterns(arch core.Architecture) []string {
	internalRoot := strings.Trim(strings.TrimSpace(arch.Layout.InternalRoot), "/")
	if internalRoot == "" {
		internalRoot = "internal"
	}
	return []string{internalRoot + "/...", "cmd/..."}
}
