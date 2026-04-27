package structural_test

import (
	"path/filepath"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/rules/structural"
)

func TestBannedPackage(t *testing.T) {
	t.Run("valid fixture has no banned package violations", func(t *testing.T) {
		violations := runRule(t, "../../testdata/valid", structural.NewBannedPackage())
		assertNoRulePrefix(t, violations, "structure.")
	})

	t.Run("detects invalid fixture banned and legacy packages", func(t *testing.T) {
		violations := runRule(t, "../../testdata/invalid", structural.NewBannedPackage())

		assertViolation(t, violations, "structure.banned-package", "internal/platform/common/")
		assertViolation(t, violations, "structure.legacy-package", "internal/router/")
	})

	t.Run("defaults to warning", func(t *testing.T) {
		violations := runRule(t, "../../testdata/invalid", structural.NewBannedPackage())
		for _, v := range violations {
			if v.Rule == "structure.banned-package" && v.DefaultSeverity != core.Warning {
				t.Fatalf("DefaultSeverity = %v, want Warning", v.DefaultSeverity)
			}
		}
	})

	// Path-aware exemption: when the configured SharedDir name happens to
	// match an entry in BannedPkgNames (e.g. SharedDir="shared"), the rule
	// must exempt internal/<SharedDir>/ at the top level only. Nested
	// directories with the same name are still flagged so the dumping-ground
	// anti-pattern cannot return inside domains.
	t.Run("top-level SharedDir exempt even when its name is banned", func(t *testing.T) {
		root := t.TempDir()
		writeTestFile(t, filepath.Join(root, "internal/shared/foo.go"), "package shared\n")
		writeTestFile(t, filepath.Join(root, "internal/domain/order/shared/bar.go"), "package shared\n")
		writeTestFile(t, filepath.Join(root, "internal/domain/order/util/baz.go"), "package util\n")

		arch := dddArch()
		arch.Layout.SharedDir = "shared"

		ctx := core.NewContext(nil, "github.com/example/app", root, arch, nil)
		violations := core.Run(ctx, core.NewRuleSet(structural.NewBannedPackage()))

		for _, v := range violations {
			if v.Rule == "structure.banned-package" && v.File == "internal/shared/" {
				t.Fatalf("top-level SharedDir must be exempt, got: %+v", v)
			}
		}
		assertViolation(t, violations, "structure.banned-package", "internal/domain/order/shared/")
		assertViolation(t, violations, "structure.banned-package", "internal/domain/order/util/")
	})

	t.Run("when SharedDir is not banned the rule behaves as before", func(t *testing.T) {
		root := t.TempDir()
		writeTestFile(t, filepath.Join(root, "internal/pkg/foo.go"), "package pkg\n")
		writeTestFile(t, filepath.Join(root, "internal/domain/order/shared/bar.go"), "package shared\n")

		arch := dddArch() // SharedDir="pkg" (not in banned list)

		ctx := core.NewContext(nil, "github.com/example/app", root, arch, nil)
		violations := core.Run(ctx, core.NewRuleSet(structural.NewBannedPackage()))

		for _, v := range violations {
			if v.Rule == "structure.banned-package" && v.File == "internal/pkg/" {
				t.Fatalf("internal/pkg/ must not be flagged, got: %+v", v)
			}
		}
		assertViolation(t, violations, "structure.banned-package", "internal/domain/order/shared/")
	})
}
