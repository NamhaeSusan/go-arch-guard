package rules_test

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/rules"
)

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
	want = append(want, rules.AnalyzeBlastRadius(pkgs, "github.com/kimtaeyun/testproject-dc", "../testdata/valid")...)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("RunAll() mismatch\n got: %#v\nwant: %#v", got, want)
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
	want = append(want, rules.AnalyzeBlastRadius(pkgs, "github.com/kimtaeyun/testproject-dc", "../testdata/valid", opts...)...)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("RunAll() mismatch\n got: %#v\nwant: %#v", got, want)
	}
}

func TestRunAll_EmptyModuleAndRootAutoExtract(t *testing.T) {
	pkgs, err := analyzer.Load("../testdata/valid", "internal/...", "cmd/...")
	if err != nil {
		t.Fatalf("load packages: %v", err)
	}

	got := rules.RunAll(pkgs, "", "")
	want := rules.RunAll(pkgs, "github.com/kimtaeyun/testproject-dc", "../testdata/valid")

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("RunAll() auto-extract must match explicit values\n got: %#v\nwant: %#v", got, want)
	}
}

func TestRunAll_ConsumerWorker_IncludesTypePatterns(t *testing.T) {
	root := t.TempDir()
	module := "example.com/runall-cw"

	writeTestFile(t, filepath.Join(root, "go.mod"), "module "+module+"\n\ngo 1.21\n")
	// worker_order.go without OrderWorker → naming.worker-type-mismatch
	writeTestFile(t, filepath.Join(root, "internal", "worker", "worker_order.go"),
		"package worker\n\ntype BadName struct{}\n")
	writeTestFile(t, filepath.Join(root, "internal", "model", "order.go"),
		"package model\n")

	pkgs := loadTestPackages(t, root)
	violations := rules.RunAll(pkgs, module, root, rules.WithModel(rules.ConsumerWorker()))
	found := false
	for _, v := range violations {
		if v.Rule == "naming.worker-type-mismatch" {
			found = true
		}
	}
	if !found {
		t.Error("RunAll should include naming.worker-type-mismatch from CheckTypePatterns")
	}
}
