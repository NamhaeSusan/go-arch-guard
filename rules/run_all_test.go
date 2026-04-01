package rules_test

import (
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
	want = append(want, rules.AnalyzeBlastRadius(pkgs, "github.com/kimtaeyun/testproject-dc", "../testdata/valid", opts...)...)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("RunAll() mismatch\n got: %#v\nwant: %#v", got, want)
	}
}
