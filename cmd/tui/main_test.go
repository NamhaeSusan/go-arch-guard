package main

import (
	"os/exec"
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

func TestApplyInternalRoot(t *testing.T) {
	arch := core.Architecture{}
	applyInternalRoot(&arch, "/packages/")
	if got := arch.Layout.InternalRoot; got != "packages" {
		t.Fatalf("InternalRoot = %q, want packages", got)
	}

	applyInternalRoot(&arch, " ")
	if got := arch.Layout.InternalRoot; got != "packages" {
		t.Fatalf("blank override changed InternalRoot to %q", got)
	}
}

func TestTUICommandLoadsCustomInternalRoot(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--internal-root", "packages", "../../testdata/custom_root")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected command to fail after loading because no interactive terminal is available")
	}
	text := string(out)
	if strings.Contains(text, "no packages found") || strings.Contains(text, "pattern ./internal/...") {
		t.Fatalf("custom internal root was not used; output:\n%s", text)
	}
	if !strings.Contains(text, "tui error:") {
		t.Fatalf("expected command to reach TUI startup, got:\n%s", text)
	}
}
