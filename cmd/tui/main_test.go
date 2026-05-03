package main

import (
	"sort"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
)

func TestPresetNamesAreSorted(t *testing.T) {
	got := strings.Split(presetNames(), ", ")
	if !sort.StringsAreSorted(got) {
		t.Fatalf("presetNames() = %q, want sorted names", presetNames())
	}
}

func TestLoadPatternsRespectInternalRoot(t *testing.T) {
	arch := core.Architecture{}
	arch.Layout.InternalRoot = "packages"

	got := loadPatterns(arch)
	want := []string{"packages/...", "cmd/..."}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("loadPatterns() = %#v, want %#v", got, want)
	}
}
