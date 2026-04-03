package rules_test

import (
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/rules"
)

func violationRuleCount(vs []rules.Violation) map[string]int {
	m := make(map[string]int)
	for _, v := range vs {
		m[v.Rule]++
	}
	return m
}

func violationRules(vs []rules.Violation) []string {
	seen := make(map[string]bool)
	var result []string
	for _, v := range vs {
		if !seen[v.Rule] {
			seen[v.Rule] = true
			result = append(result, v.Rule)
		}
	}
	sort.Strings(result)
	return result
}

func TestRunAll_DefaultModelMatchesManualComposition(t *testing.T) {
	pkgs, err := analyzer.Load("../testdata/valid", "internal/...", "cmd/...")
	if err != nil {
		t.Fatalf("load packages: %v", err)
	}

	got := rules.RunAll(pkgs, "github.com/kimtaeyun/testproject-dc", "../testdata/valid")

	var want []rules.Violation
	want = append(want, rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc", "../testdata/valid")...)
	want = append(want, rules.CheckLayerDirection(pkgs, "github.com/kimtaeyun/testproject-dc", "../testdata/valid")...)
	want = append(want, rules.CheckNaming(pkgs)...)
	want = append(want, rules.CheckStructure("../testdata/valid")...)
	want = append(want, rules.CheckTypePatterns(pkgs)...)
	want = append(want, rules.CheckInterfacePattern(pkgs)...)
	want = append(want, rules.AnalyzeBlastRadius(pkgs, "github.com/kimtaeyun/testproject-dc", "../testdata/valid")...)

	if !reflect.DeepEqual(violationRuleCount(got), violationRuleCount(want)) {
		t.Fatalf("RunAll() rule set mismatch\n got: %v\nwant: %v", violationRules(got), violationRules(want))
	}
}

func TestRunAll_WithModelMatchesManualComposition(t *testing.T) {
	pkgs, err := analyzer.Load("../testdata/valid", "internal/...", "cmd/...")
	if err != nil {
		t.Fatalf("load packages: %v", err)
	}

	opts := []rules.Option{rules.WithModel(rules.Hexagonal())}
	got := rules.RunAll(pkgs, "github.com/kimtaeyun/testproject-dc", "../testdata/valid", opts...)

	var want []rules.Violation
	want = append(want, rules.CheckDomainIsolation(pkgs, "github.com/kimtaeyun/testproject-dc", "../testdata/valid", opts...)...)
	want = append(want, rules.CheckLayerDirection(pkgs, "github.com/kimtaeyun/testproject-dc", "../testdata/valid", opts...)...)
	want = append(want, rules.CheckNaming(pkgs, opts...)...)
	want = append(want, rules.CheckStructure("../testdata/valid", opts...)...)
	want = append(want, rules.CheckTypePatterns(pkgs, opts...)...)
	want = append(want, rules.CheckInterfacePattern(pkgs, opts...)...)
	want = append(want, rules.AnalyzeBlastRadius(pkgs, "github.com/kimtaeyun/testproject-dc", "../testdata/valid", opts...)...)

	if !reflect.DeepEqual(violationRuleCount(got), violationRuleCount(want)) {
		t.Fatalf("RunAll(WithModel) rule set mismatch\n got: %v\nwant: %v", violationRules(got), violationRules(want))
	}
}

func TestRunAll_EmptyModuleAndRootAutoExtract(t *testing.T) {
	pkgs, err := analyzer.Load("../testdata/valid", "internal/...", "cmd/...")
	if err != nil {
		t.Fatalf("load packages: %v", err)
	}

	got := rules.RunAll(pkgs, "", "")
	want := rules.RunAll(pkgs, "github.com/kimtaeyun/testproject-dc", "../testdata/valid")

	if !reflect.DeepEqual(violationRuleCount(got), violationRuleCount(want)) {
		t.Fatalf("RunAll() auto-extract rule set mismatch\n got: %v\nwant: %v", violationRules(got), violationRules(want))
	}
}

func TestRunAll_ConsumerWorker_IncludesTypePatterns(t *testing.T) {
	root := t.TempDir()
	module := "example.com/runall-cw"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")
	// worker_order.go without OrderWorker → naming.type-pattern-mismatch
	writeTestFile(t, filepath.Join(root, "internal", "worker", "worker_order.go"),
		"package worker\n\ntype BadName struct{}\n")
	writeTestFile(t, filepath.Join(root, "internal", "model", "order.go"),
		"package model\n")

	pkgs := loadTestPackages(t, root)
	violations := rules.RunAll(pkgs, module, root, rules.WithModel(rules.ConsumerWorker()))
	found := false
	for _, v := range violations {
		if v.Rule == "naming.type-pattern-mismatch" {
			found = true
		}
	}
	if !found {
		t.Error("RunAll should include naming.type-pattern-mismatch from CheckTypePatterns")
	}
}

func TestRunAll_IncludesInterfacePattern(t *testing.T) {
	root := t.TempDir()
	module := "example.com/runall-ip"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")
	writeTestFile(t, filepath.Join(root, "internal", "store", "store.go"),
		`package store

type Store interface {
	Find() error
}

type StoreImpl struct{}
func (s *StoreImpl) Find() error { return nil }
`)

	pkgs := loadTestPackages(t, root)
	violations := rules.RunAll(pkgs, module, root, rules.WithModel(rules.ConsumerWorker()))
	found := false
	for _, v := range violations {
		if v.Rule == "interface.exported-impl" {
			found = true
		}
	}
	if !found {
		t.Error("RunAll should include interface.exported-impl from CheckInterfacePattern")
	}
}
