package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"golang.org/x/tools/go/packages"
)

// LoadOptions configures the build whose production packages are analyzed.
type LoadOptions struct {
	// BuildFlags are passed to the Go build system, for example
	// []string{"-tags=integration"}. Explicit flags override inherited flags
	// of the same name. GOFLAGS and other Go environment variables are also
	// inherited normally.
	BuildFlags []string
	// DisableBuildFlagInheritance disables inheritance of -tags, -race,
	// -msan, and -asan recorded in the calling binary's build information.
	// It does not disable normal Go environment variable inheritance.
	DisableBuildFlagInheritance bool
}

// Load parses Go packages matching the given patterns under dir. Relative
// directory patterns such as "internal/..." are resolved under dir, patterns
// beginning with "./" are passed through, module-path patterns such as
// "github.com/acme/project/..." are passed through, and absolute filesystem
// patterns are passed through unchanged.
// Build tags and enabled race/memory/address sanitizer flags recorded in the
// calling binary's build information are inherited. Use LoadWithOptions to
// override flags or disable this inheritance.
// When some packages contain errors (e.g. type-check failures), they are
// skipped and the successfully loaded packages are returned alongside a
// non-nil error describing what was skipped. Callers that want partial
// analysis should check len(pkgs) rather than treating err as fatal.
func Load(dir string, patterns ...string) ([]*packages.Package, error) {
	return LoadWithOptions(dir, LoadOptions{}, patterns...)
}

// LoadWithOptions is Load with explicit build flags. It retains Load's
// partial-analysis behavior: callers using it as a CI guard must reject any
// returned error and packages with IllTyped or Errors set.
func LoadWithOptions(dir string, opts LoadOptions, patterns ...string) ([]*packages.Package, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve dir: %w", err)
	}
	if _, err := os.Stat(absDir); err != nil {
		return nil, fmt.Errorf("project root not found: %w", err)
	}

	// Prefix relative directory patterns with "./" so go/packages treats
	// them as filesystem queries against absDir. Patterns that already start
	// with "./" or that look like an absolute module query (e.g.
	// "github.com/foo/bar/...") are passed through unchanged — double-
	// prefixing would either yield "././..." or wedge a module path into a
	// filesystem lookup that go/packages rejects.
	//
	// Module-path detection looks at the first path segment only: a dot in
	// the first segment ("github.com/...") implies a module query, while
	// "internal/..." has no dot in its first segment despite the trailing
	// "..." wildcard.
	prefixed := make([]string, len(patterns))
	for i, p := range patterns {
		switch {
		case strings.HasPrefix(p, "./"):
			prefixed[i] = p
		case filepath.IsAbs(p):
			prefixed[i] = p
		case looksLikeModulePath(p):
			prefixed[i] = p
		default:
			prefixed[i] = "./" + p
		}
	}

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedImports | packages.NeedFiles |
			packages.NeedSyntax | packages.NeedModule |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps,
		Dir:        absDir,
		BuildFlags: loadBuildFlags(opts),
	}
	pkgs, err := packages.Load(cfg, prefixed...)
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}

	var result []*packages.Package
	var loadErrs []string
	for _, pkg := range pkgs {
		if depErr := firstFatalErrorInDeps(pkg); depErr != "" {
			loadErrs = append(loadErrs, depErr)
			continue
		}
		if len(pkg.Errors) == 0 {
			result = append(result, pkg)
			continue
		}
		// Tolerate pure type-check failures on the root package itself (e.g.,
		// undefined identifier) because downstream rules can still inspect the
		// AST/imports. Reject parse and list errors since those leave TypesInfo
		// unreliable.
		//
		// Seed from pkg.IllTyped (which go/packages sets to true whenever any
		// error exists, type or otherwise) and then narrow it: the loop below
		// flips it to false the first time we hit a non-TypeError. The
		// pkg.IllTyped seed is only meaningful AFTER firstFatalErrorInDeps has
		// returned empty above, because otherwise this block never runs.
		onlyTypeErrors := pkg.IllTyped
		for _, e := range pkg.Errors {
			if e.Kind != packages.TypeError {
				onlyTypeErrors = false
				break
			}
		}
		if onlyTypeErrors {
			result = append(result, pkg)
			continue
		}
		// Syntax or load errors on the root package are reported and it is skipped.
		for _, e := range pkg.Errors {
			loadErrs = append(loadErrs, e.Error())
		}
	}
	if len(loadErrs) > 0 {
		return result, fmt.Errorf("packages with errors were skipped (%d): %s",
			len(loadErrs), summarizeLoadErrs(loadErrs))
	}
	return result, nil
}

func loadBuildFlags(opts LoadOptions) []string {
	if !opts.DisableBuildFlagInheritance {
		if info, ok := debug.ReadBuildInfo(); ok {
			return mergeBuildFlags(info.Settings, opts.BuildFlags)
		}
	}
	return append([]string(nil), opts.BuildFlags...)
}

func mergeBuildFlags(settings []debug.BuildSetting, explicit []string) []string {
	overridden := make(map[string]bool)
	for _, flag := range explicit {
		name, _, _ := strings.Cut(flag, "=")
		overridden[name] = true
	}
	var flags []string
	for _, setting := range settings {
		if overridden[setting.Key] {
			continue
		}
		switch setting.Key {
		case "-tags":
			if setting.Value != "" {
				flags = append(flags, setting.Key+"="+setting.Value)
			}
		case "-race", "-msan", "-asan":
			if setting.Value == "true" {
				flags = append(flags, setting.Key+"=true")
			}
		}
	}
	return append(flags, explicit...)
}

// looksLikeModulePath reports whether p has a dot in its first path segment,
// the canonical signal of a Go module path (e.g. "github.com/x/y" or
// "example.com/foo"). A pattern like "internal/..." has no dot in its first
// segment ("internal") even though the trailing "..." wildcard contains
// dots, so it is NOT a module path and still wants the "./" prefix.
func looksLikeModulePath(p string) bool {
	firstSegment, _, _ := strings.Cut(p, "/")
	if firstSegment == "..." || firstSegment == "" {
		return false
	}
	return strings.Contains(firstSegment, ".")
}

// summarizeLoadErrs joins errors with "; " and caps the output at the first
// few entries so the final error message stays readable when many packages
// fail at once. Callers that need every detail can re-run with verbose
// tooling; the returned error here is for human-eyeballable CI logs.
const maxLoadErrSamples = 5

func summarizeLoadErrs(errs []string) string {
	if len(errs) <= maxLoadErrSamples {
		return strings.Join(errs, "; ")
	}
	head := strings.Join(errs[:maxLoadErrSamples], "; ")
	return fmt.Sprintf("%s; ... and %d more", head, len(errs)-maxLoadErrSamples)
}

// firstFatalErrorInDeps traverses the transitive import graph of pkg
// (excluding pkg itself) and returns the first fatal error found in any
// dependency. Any non-TypeError in a dep is treated as fatal:
//   - ParseError / ListError are clearly fatal.
//   - UnknownError is also treated as fatal because go/packages uses it for
//     missing-or-unreadable export data and toolchain/source skew, which
//     leave the root's TypesInfo incomplete. With NeedTypes|NeedTypesInfo|
//     NeedDeps enabled, downstream rules consume TypesInfo and would silently
//     misreport if we tolerated UnknownError here.
//
// Only TypeError in a dep is tolerated, mirroring the root-package policy:
// rules can still inspect AST/imports meaningfully when a dep has a pure
// type-check failure. An empty string means all deps are clean.
func firstFatalErrorInDeps(root *packages.Package) string {
	var found string
	packages.Visit([]*packages.Package{root}, nil, func(dep *packages.Package) {
		if found != "" || dep == root {
			return
		}
		for _, e := range dep.Errors {
			if e.Kind != packages.TypeError {
				found = fmt.Sprintf("dependency %s: %s", dep.PkgPath, e.Error())
				return
			}
		}
	})
	return found
}
