# Guard correctness and compatibility

The guard must distinguish a valid architecture from an analysis it could not
complete. These changes close the failure cases found in the September 2026
source audit while retaining partial package analysis for inspection tools.

## Loading and generated tests

`analyzer.Load` inherits the caller binary's recorded `-tags`, `-race`, `-msan`,
and `-asan` flags, so tagged code compiled into a test is also visible to its
nested analysis. `LoadWithOptions` accepts `LoadOptions.BuildFlags`; explicit
flags take precedence over inherited flags of the same name.
`DisableBuildFlagInheritance` opts out. Go environment settings and explicit
package patterns still determine the analyzed platform and package set.

```go
pkgs, err := analyzer.LoadWithOptions(".", analyzer.LoadOptions{
    BuildFlags: []string{"-tags=integration"},
}, "./...")
```

The lower-level loader can return partial results and can retain packages with
type errors. Generated architecture tests reject load errors, `IllTyped` or
`Errors` on packages, an empty production package set, and missing module
metadata. Existing handwritten tests should enforce the same checks. The
scaffold's `BuildFlags` field embeds explicit flags when requested. Its package
root is required; the optional `cmd` tree is analyzed when present.

Always execute architecture guards with `go test -count=1`: Go's test cache may
miss a newly added file in a package loaded indirectly by the guard.

## Rule failures and severity

- A rule emitting an undeclared violation ID produces an Error diagnostic,
  `meta.unknown-violation-id`, just like a rule panic makes incomplete analysis
  fail. Environmental meta diagnostics remain warnings unless overridden.
- Only `core.Error` and `core.Warning` are valid severities. Invalid values in
  rule specs or runtime overrides are rejected by the runner.
- Reports treat an unknown severity as an error, including violations supplied
  directly to report helpers without passing through `core.Run`.

Consumers with misspelled violation IDs or invalid severity constants may now
fail where they previously received a warning or misleading success.

## Symbol and type analysis

- Parenthesized calls and explicit generic instantiations resolve to the same
  symbol as ordinary calls. Function values remain outside direct-call analysis.
- Type aliases are resolved before transaction signature checks. Map keys and
  values, callbacks, anonymous composite types, and generic arguments are
  traversed. Named wrappers remain distinct API types; their hidden underlying
  representations are not expanded.
- `analysisutil.WalkSignatureNamedTypes` exposes the recursive traversal.
- `InspectTypeSpecs` and `ResolveIdentImportPath` accept optional `*types.Info`.
  Supplying it resolves the actual package name for versioned imports rather
  than assuming the last import-path segment is the Go identifier. Raw string
  imports and generic aliases are recognized. Untyped callers retain the
  syntax-only fallback.

## Rule behavior

- Filename checks validate every dot-separated stem segment. Generated names
  such as `service.pb.go` remain valid; `service.BadName.go` is rejected.
- Domains cannot depend on composition or transport roots under the Isolation
  rule. This is reported as `dependency.domain-imports-composition`.
- Nested roots such as `src/internal` are classified by path segments in purity
  and constructor checks, consistently with other root-aware rules.
- Interface and transaction rules honor package/file exclusions. Excluded
  declarations do not become interface-policy evidence for included files.
- Interface method caps use the complete typed method set, including embedding.
  Generic interface constructor results are accepted.
- Domain panic checks include standard `log.Panic*` functions and `log.Logger`
  Panic/Fatal methods in addition to direct panic and exit functions.
- Hand-written mock structs and receiver methods are matched across test files
  in the same package. Internal and external test packages are kept separate.

## Regression coverage

Regression cases cover inherited and explicit build flags, generated guards
against partial/invalid/empty input, dotted filenames, generic and parenthesized
calls, aliased and nested transaction types, mismatched import package names,
invalid severity and rule IDs, reverse domain dependencies, nested roots,
exclusion behavior, embedded/generic interfaces, logger exits, and split mock
files. The tests include positive controls so a blanket rejection cannot satisfy
the policy checks.
