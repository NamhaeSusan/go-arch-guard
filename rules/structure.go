package rules

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func CheckStructure(projectRoot string, opts ...Option) []Violation {
	cfg := NewConfig(opts...)
	m := cfg.model()
	var violations []Violation

	internalDir := filepath.Join(projectRoot, "internal")
	if _, err := os.Stat(internalDir); err != nil {
		return nil
	}

	violations = append(violations, checkInternalTopLevelPackages(internalDir, m, cfg)...)
	violations = append(violations, checkPackageNames(internalDir, m, cfg)...)
	violations = append(violations, checkMiddlewarePlacement(internalDir, m, cfg)...)

	// Domain-specific checks only apply when DomainDir is set.
	if m.DomainDir != "" {
		domainDir := filepath.Join(internalDir, m.DomainDir)
		if m.RequireAlias {
			violations = append(violations, checkDomainRootAliasRequired(domainDir, m, cfg)...)
			violations = append(violations, checkDomainRootAliasPackage(domainDir, m, cfg)...)
			violations = append(violations, checkDomainRootAliasOnly(domainDir, m, cfg)...)
			violations = append(violations, checkDomainAliasNoInterface(domainDir, m, cfg)...)
		}
		if m.RequireModel {
			violations = append(violations, checkDomainModelRequired(domainDir, m, cfg)...)
		}
		violations = append(violations, checkDTOPlacement(internalDir, m, cfg)...)
	}

	return violations
}

func checkInternalTopLevelPackages(internalDir string, m Model, cfg Config) []Violation {
	var violations []Violation
	entries, err := os.ReadDir(internalDir)
	if err != nil {
		return nil
	}

	var allowedNames []string
	for name := range m.InternalTopLevel {
		allowedNames = append(allowedNames, "internal/"+name+"/")
	}

	for _, entry := range entries {
		relPath := filepath.ToSlash(filepath.Join("internal", entry.Name()))

		if entry.IsDir() {
			if cfg.IsExcluded(relPath + "/") {
				continue
			}
			if m.InternalTopLevel[entry.Name()] {
				continue
			}
			violations = append(violations, Violation{
				File:     relPath + "/",
				Rule:     "structure.internal-top-level",
				Message:  `internal/ top-level package "` + entry.Name() + `" is not allowed`,
				Fix:      fmt.Sprintf("use only %v at the internal/ top level", allowedNames),
				Severity: cfg.Sev,
			})
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		if cfg.IsExcluded(relPath) {
			continue
		}
		violations = append(violations, Violation{
			File:     relPath,
			Rule:     "structure.internal-top-level",
			Message:  `internal/ top-level Go file "` + entry.Name() + `" is not allowed`,
			Fix:      fmt.Sprintf("move code under %v", allowedNames),
			Severity: cfg.Sev,
		})
	}
	return violations
}

func checkDomainRootAliasRequired(domainDir string, m Model, cfg Config) []Violation {
	var violations []Violation
	entries, err := os.ReadDir(domainDir)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		relPath := filepath.ToSlash(filepath.Join("internal", m.DomainDir, e.Name()))
		if cfg.IsExcluded(relPath + "/") {
			continue
		}
		aliasPath := filepath.Join(domainDir, e.Name(), m.AliasFileName)
		if _, err := os.Stat(aliasPath); err == nil {
			continue
		}
		violations = append(violations, Violation{
			File:     relPath + "/",
			Rule:     "structure.domain-root-alias-required",
			Message:  `domain root "` + e.Name() + `" must define ` + m.AliasFileName,
			Fix:      "add " + m.AliasFileName + " as the single public surface file for the domain root package",
			Severity: cfg.Sev,
		})
	}
	return violations
}

func checkDomainRootAliasPackage(domainDir string, m Model, cfg Config) []Violation {
	var violations []Violation
	entries, err := os.ReadDir(domainDir)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		relPath := filepath.ToSlash(filepath.Join("internal", m.DomainDir, e.Name()))
		if cfg.IsExcluded(relPath + "/") {
			continue
		}
		aliasPath := filepath.Join(domainDir, e.Name(), m.AliasFileName)
		if _, err := os.Stat(aliasPath); err != nil {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), aliasPath, nil, parser.PackageClauseOnly)
		if err != nil {
			continue
		}
		if file.Name.Name == e.Name() {
			continue
		}
		violations = append(violations, Violation{
			File:     relPath + "/" + m.AliasFileName,
			Rule:     "structure.domain-root-alias-package",
			Message:  m.AliasFileName + ` package name must match domain root "` + e.Name() + `"`,
			Fix:      `set "package ` + e.Name() + `" in ` + m.AliasFileName,
			Severity: cfg.Sev,
		})
	}
	return violations
}

func checkDomainRootAliasOnly(domainDir string, m Model, cfg Config) []Violation {
	var violations []Violation
	entries, err := os.ReadDir(domainDir)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		rootDir := filepath.Join(domainDir, e.Name())
		rootEntries, err := os.ReadDir(rootDir)
		if err != nil {
			continue
		}
		for _, rootEntry := range rootEntries {
			if rootEntry.IsDir() {
				continue
			}
			name := rootEntry.Name()
			if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") || name == m.AliasFileName {
				continue
			}
			relPath := filepath.ToSlash(filepath.Join("internal", m.DomainDir, e.Name(), name))
			if cfg.IsExcluded(relPath) {
				continue
			}
			violations = append(violations, Violation{
				File:     relPath,
				Rule:     "structure.domain-root-alias-only",
				Message:  `domain root "` + e.Name() + `" must expose its public API from ` + m.AliasFileName + ` only`,
				Fix:      `move "` + name + `" into a sub-package or merge the public API into ` + m.AliasFileName,
				Severity: cfg.Sev,
			})
		}
	}
	return violations
}

func checkDomainAliasNoInterface(domainDir string, m Model, cfg Config) []Violation {
	var violations []Violation
	entries, err := os.ReadDir(domainDir)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		relPath := filepath.ToSlash(filepath.Join("internal", m.DomainDir, e.Name()))
		if cfg.IsExcluded(relPath + "/") {
			continue
		}
		aliasPath := filepath.Join(domainDir, e.Name(), m.AliasFileName)
		if _, err := os.Stat(aliasPath); err != nil {
			continue
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, aliasPath, nil, 0)
		if err != nil {
			continue
		}
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if _, isIface := ts.Type.(*ast.InterfaceType); isIface {
					violations = append(violations, Violation{
						File:     relPath + "/" + m.AliasFileName,
						Line:     fset.Position(ts.Name.Pos()).Line,
						Rule:     "structure.domain-alias-no-interface",
						Message:  m.AliasFileName + ` re-exports interface "` + ts.Name.Name + `" — suspected cross-domain dependency; use ` + m.OrchestrationDir + `/ instead`,
						Fix:      "move cross-domain coordination to " + m.OrchestrationDir + "/handler/ or " + m.OrchestrationDir + "/",
						Severity: cfg.Sev,
					})
				}
				if ts.Assign != 0 {
					if sel, ok := ts.Type.(*ast.SelectorExpr); ok {
						if ident, ok := sel.X.(*ast.Ident); ok {
							impPath := resolveIdentImportPath(file, ident.Name)
							if src := matchContractSublayer(m, impPath); src != "" {
								violations = append(violations, Violation{
									File:     relPath + "/" + m.AliasFileName,
									Line:     fset.Position(ts.Name.Pos()).Line,
									Rule:     "structure.domain-alias-contract-reexport",
									Message:  m.AliasFileName + ` re-exports "` + ts.Name.Name + `" from ` + src + ` — suspected cross-domain dependency; use ` + m.OrchestrationDir + `/ instead`,
									Fix:      "move cross-domain coordination to " + m.OrchestrationDir + "/handler/ or " + m.OrchestrationDir + "/",
									Severity: cfg.Sev,
								})
							}
						}
					}
				}
			}
		}
	}
	return violations
}

func checkPackageNames(internalDir string, m Model, cfg Config) []Violation {
	var violations []Violation
	_ = filepath.WalkDir(internalDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(filepath.Dir(internalDir), path)
		if relErr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if rel == "internal" {
			return nil
		}
		if cfg.IsExcluded(rel + "/") {
			return filepath.SkipDir
		}
		if !hasNonTestGoFiles(path) {
			return nil
		}

		name := d.Name()
		for _, banned := range m.BannedPkgNames {
			if name == banned {
				violations = append(violations, Violation{
					File:     rel + "/",
					Rule:     "structure.banned-package",
					Message:  `package "` + name + `" is banned`,
					Fix:      fmt.Sprintf("move to specific domain or %s/", m.SharedDir),
					Severity: cfg.Sev,
				})
			}
		}

		for _, legacy := range m.LegacyPkgNames {
			if name == legacy {
				violations = append(violations, Violation{
					File:     rel + "/",
					Rule:     "structure.legacy-package",
					Message:  `legacy package "` + name + `" should be migrated`,
					Fix:      fmt.Sprintf("move app-specific wiring to cmd/ and shared helpers to internal/%s/", m.SharedDir),
					Severity: cfg.Sev,
				})
			}
		}

		if isMisplacedLayerDirWith(m, rel, name) {
			violations = append(violations, Violation{
				File:     rel + "/",
				Rule:     "structure.misplaced-layer",
				Message:  `layer package "` + name + `" is misplaced`,
				Fix:      fmt.Sprintf("place app/handler/infra only in domain slices or %s handler", m.OrchestrationDir),
				Severity: cfg.Sev,
			})
		}
		return nil
	})
	return violations
}

func checkMiddlewarePlacement(internalDir string, m Model, cfg Config) []Violation {
	allowedPath := filepath.ToSlash(filepath.Join("internal", m.SharedDir, "middleware"))
	var violations []Violation
	_ = filepath.WalkDir(internalDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		if d.Name() != "middleware" {
			return nil
		}
		if !hasNonTestGoFiles(path) {
			return nil
		}
		rel, _ := filepath.Rel(filepath.Dir(internalDir), path)
		rel = filepath.ToSlash(rel)
		if cfg.IsExcluded(rel + "/") {
			return nil
		}
		if rel == allowedPath {
			return nil
		}
		violations = append(violations, Violation{
			File:     rel + "/",
			Rule:     "structure.middleware-placement",
			Message:  `middleware found at "` + rel + `"`,
			Fix:      "move middleware to " + allowedPath + "/",
			Severity: cfg.Sev,
		})
		return nil
	})
	return violations
}

func checkDomainModelRequired(domainDir string, m Model, cfg Config) []Violation {
	var violations []Violation
	entries, err := os.ReadDir(domainDir)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		relPath := filepath.ToSlash(filepath.Join("internal", m.DomainDir, e.Name()))
		if cfg.IsExcluded(relPath + "/") {
			continue
		}
		modelDir := filepath.Join(domainDir, e.Name(), filepath.FromSlash(m.ModelPath))
		if hasNonTestGoFiles(modelDir) {
			continue
		}
		violations = append(violations, Violation{
			File:     relPath + "/",
			Rule:     "structure.domain-model-required",
			Message:  `domain "` + e.Name() + `" missing a direct non-test Go file in ` + m.ModelPath + `/`,
			Fix:      "add at least one non-test Go file directly under " + m.ModelPath + "/",
			Severity: cfg.Sev,
		})
	}
	return violations
}

func checkDTOPlacement(internalDir string, m Model, cfg Config) []Violation {
	var violations []Violation
	domainDir := filepath.Join(internalDir, m.DomainDir)
	if _, err := os.Stat(domainDir); err != nil {
		return nil
	}
	_ = filepath.WalkDir(domainDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		name := d.Name()
		if !strings.HasSuffix(name, ".go") {
			return nil
		}
		if name == "dto.go" || (strings.HasSuffix(name, "_dto.go") && !strings.HasSuffix(name, "_test.go")) {
			rel, _ := filepath.Rel(filepath.Dir(internalDir), path)
			rel = filepath.ToSlash(rel)
			if cfg.IsExcluded(rel) {
				return nil
			}
			if isDTOAllowedSublayerWith(m, rel) {
				return nil
			}
			violations = append(violations, Violation{
				File:     rel,
				Rule:     "structure.dto-placement",
				Message:  `"` + name + `" found in forbidden layer`,
				Fix:      fmt.Sprintf("DTOs belong in %v", m.DTOAllowedLayers),
				Severity: cfg.Sev,
			})
		}
		return nil
	})
	return violations
}

func isDTOAllowedSublayerWith(m Model, relPath string) bool {
	// relPath = "internal/<domainDir>/<domainName>/<sublayer>/..."
	domainDirDepth := len(strings.Split(m.DomainDir, "/"))
	parts := strings.Split(relPath, "/")
	// sublayer sits at: "internal"(1) + domainDir segments + domainName(1)
	sublayerIdx := 1 + domainDirDepth + 1
	if len(parts) <= sublayerIdx {
		return false
	}
	return slices.Contains(m.DTOAllowedLayers, parts[sublayerIdx])
}

func isMisplacedLayerDirWith(m Model, rel, name string) bool {
	switch name {
	case "app", "infra":
		return !matchesDomainLayerWith(m, rel, name)
	case "handler":
		return !matchesDomainLayerWith(m, rel, name) && rel != filepath.ToSlash(filepath.Join("internal", m.OrchestrationDir, "handler"))
	default:
		return false
	}
}

func matchesDomainLayerWith(m Model, rel, name string) bool {
	parts := strings.Split(rel, "/")
	return len(parts) == 4 && parts[0] == "internal" && parts[1] == m.DomainDir && parts[2] != "" && parts[3] == name
}

func hasNonTestGoFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") {
			return true
		}
	}
	return false
}
