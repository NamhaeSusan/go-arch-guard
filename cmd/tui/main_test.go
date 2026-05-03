package main

import (
	"sort"
	"strings"
	"testing"
)

func TestPresetNamesAreSorted(t *testing.T) {
	got := strings.Split(presetNames(), ", ")
	if !sort.StringsAreSorted(got) {
		t.Fatalf("presetNames() = %q, want sorted names", presetNames())
	}
}
