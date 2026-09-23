package scaffold_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/scaffold"
)

// Execute the generated test in real modules. Parser-only template tests cannot
// detect a successful test that silently omits broken production packages.
func TestGeneratedGuardRejectsIncompleteAnalysis(t *testing.T) {
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	sums, err := os.ReadFile(filepath.Join(root, "go.sum"))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, broken, want    string
		testOnly              bool
		noCmd, cmdFile, empty bool
		flags, commandFlags   []string
	}{
		{name: "valid"},
		{name: "internal only", noCmd: true},
		{name: "cmd is a file", cmdFile: true, want: "cmd must be a directory"},
		{name: "empty project", empty: true, want: "no production packages loaded"},
		{name: "syntax error", broken: "package broken\nfunc broken(\n", want: "packages with errors were skipped"},
		{name: "missing import", broken: "package broken\nimport _ \"example.test/guard/missing\"\n", want: "packages with errors were skipped"},
		{name: "type error", broken: "package broken\nvar value = missingIdentifier\n", want: "type/load errors"},
		{name: "no production", testOnly: true, want: "no production packages loaded"},
		{name: "inactive tagged error", broken: "//go:build guard_audit_tag\n\npackage broken\nvar value = missingIdentifier\n"},
		{name: "explicit build tags", broken: "//go:build guard_audit_tag\n\npackage broken\nvar value = missingIdentifier\n", flags: []string{"-tags=guard_audit_tag"}, want: "type/load errors"},
		{name: "inherited test build tags", broken: "//go:build guard_audit_tag\n\npackage broken\nvar value = missingIdentifier\n", commandFlags: []string{"-tags=guard_audit_tag"}, want: "type/load errors"},
		{name: "explicit tags override inherited tags", broken: "//go:build guard_audit_tag\n\npackage broken\nvar value = missingIdentifier\n", commandFlags: []string{"-tags=guard_audit_tag"}, flags: []string{"-tags=other_tag"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			write := func(name, content string) {
				t.Helper()
				p := filepath.Join(dir, name)
				if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(p, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
			}
			write("go.mod", fmt.Sprintf("module example.test/guard\n\ngo 1.26.1\n\nrequire github.com/NamhaeSusan/go-arch-guard v0.0.0\nreplace github.com/NamhaeSusan/go-arch-guard => %q\n", root))
			write("go.sum", string(sums))
			if tc.empty {
				if err := os.MkdirAll(filepath.Join(dir, "internal"), 0o755); err != nil {
					t.Fatal(err)
				}
			} else if tc.testOnly {
				write("internal/pkg/safe/safe_test.go", "package safe\n")
			} else {
				write("internal/pkg/safe/safe.go", "package safe\n")
				if tc.cmdFile {
					write("cmd", "not a directory")
				} else if !tc.noCmd {
					write("cmd/service/main.go", "package main\nfunc main() {}\n")
				}
			}
			if tc.broken != "" {
				write("internal/pkg/broken/broken.go", tc.broken)
			}
			source, err := scaffold.ArchitectureTest(scaffold.PresetDDD, scaffold.ArchitectureTestOptions{PackageName: "guard_test", BuildFlags: tc.flags})
			if err != nil {
				t.Fatal(err)
			}
			write("architecture_test.go", source)
			args := append([]string{"test", "-mod=mod", "-count=1", "-run", "^TestArchitecture$"}, tc.commandFlags...)
			args = append(args, ".")
			cmd := exec.Command("go", args...)
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=")
			output, err := cmd.CombinedOutput()
			if tc.want == "" {
				if err != nil {
					t.Fatalf("generated guard failed: %v\n%s", err, output)
				}
				return
			}
			if err == nil {
				t.Fatalf("generated guard falsely passed:\n%s", output)
			}
			if !strings.Contains(string(output), tc.want) {
				t.Fatalf("want %q, got %v:\n%s", tc.want, err, output)
			}
		})
	}
}
