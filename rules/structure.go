package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func CheckStructure(projectRoot string, opts ...Option) []Violation {
	cfg := NewConfig(opts...)
	var violations []Violation

	internalDir := filepath.Join(projectRoot, "internal")
	if _, err := os.Stat(internalDir); err != nil {
		return nil
	}

	violations = append(violations, checkInternalTopLevelPackages(internalDir, cfg)...)
	violations = append(violations, checkPackageNames(internalDir, cfg)...)
	violations = append(violations, checkMiddlewarePlacement(internalDir, cfg)...)

	domainDir := filepath.Join(internalDir, "domain")
	violations = append(violations, checkDomainRootAliasRequired(domainDir, cfg)...)
	violations = append(violations, checkDomainRootAliasPackage(domainDir, cfg)...)
	violations = append(violations, checkDomainRootAliasOnly(domainDir, cfg)...)
	violations = append(violations, checkDomainAliasNoInterface(domainDir, cfg)...)
	violations = append(violations, checkDomainModelRequired(domainDir, cfg)...)
	violations = append(violations, checkDTOPlacement(internalDir, cfg)...)

	return violations
}

func checkInternalTopLevelPackages(internalDir string, cfg Config) []Violation {
	var violations []Violation
	entries, err := os.ReadDir(internalDir)
	if err != nil {
		return nil
	}
	for _, entry := range entries {
		relPath := filepath.ToSlash(filepath.Join("internal", entry.Name()))

		if entry.IsDir() {
			if cfg.IsExcluded(relPath + "/") {
				continue
			}
			if allowedInternalTopLevel[entry.Name()] {
				continue
			}
			violations = append(violations, Violation{
				File:     relPath + "/",
				Rule:     "structure.internal-top-level",
				Message:  `internal/ top-level package "` + entry.Name() + `" is not allowed`,
				Fix:      "use only internal/domain/, internal/orchestration/, or internal/pkg/ at the internal/ top level",
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
			Fix:      "move code under internal/domain/, internal/orchestration/, or internal/pkg/",
			Severity: cfg.Sev,
		})
	}
	return violations
}

func checkDomainRootAliasRequired(domainDir string, cfg Config) []Violation {
	var violations []Violation
	entries, err := os.ReadDir(domainDir)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		relPath := filepath.ToSlash(filepath.Join("internal", "domain", e.Name()))
		if cfg.IsExcluded(relPath + "/") {
			continue
		}
		aliasPath := filepath.Join(domainDir, e.Name(), "alias.go")
		if _, err := os.Stat(aliasPath); err == nil {
			continue
		}
		violations = append(violations, Violation{
			File:     relPath + "/",
			Rule:     "structure.domain-root-alias-required",
			Message:  `domain root "` + e.Name() + `" must define alias.go`,
			Fix:      "add alias.go as the single public surface file for the domain root package",
			Severity: cfg.Sev,
		})
	}
	return violations
}

func checkDomainRootAliasPackage(domainDir string, cfg Config) []Violation {
	var violations []Violation
	entries, err := os.ReadDir(domainDir)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		relPath := filepath.ToSlash(filepath.Join("internal", "domain", e.Name()))
		if cfg.IsExcluded(relPath + "/") {
			continue
		}
		aliasPath := filepath.Join(domainDir, e.Name(), "alias.go")
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
			File:     relPath + "/alias.go",
			Rule:     "structure.domain-root-alias-package",
			Message:  `alias.go package name must match domain root "` + e.Name() + `"`,
			Fix:      `set "package ` + e.Name() + `" in alias.go`,
			Severity: cfg.Sev,
		})
	}
	return violations
}

func checkDomainRootAliasOnly(domainDir string, cfg Config) []Violation {
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
			if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") || name == "alias.go" {
				continue
			}
			relPath := filepath.ToSlash(filepath.Join("internal", "domain", e.Name(), name))
			if cfg.IsExcluded(relPath) {
				continue
			}
			violations = append(violations, Violation{
				File:     relPath,
				Rule:     "structure.domain-root-alias-only",
				Message:  `domain root "` + e.Name() + `" must expose its public API from alias.go only`,
				Fix:      `move "` + name + `" into a sub-package or merge the public API into alias.go`,
				Severity: cfg.Sev,
			})
		}
	}
	return violations
}

func checkDomainAliasNoInterface(domainDir string, cfg Config) []Violation {
	var violations []Violation
	entries, err := os.ReadDir(domainDir)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		relPath := filepath.ToSlash(filepath.Join("internal", "domain", e.Name()))
		if cfg.IsExcluded(relPath + "/") {
			continue
		}
		aliasPath := filepath.Join(domainDir, e.Name(), "alias.go")
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
				// type alias: ts.Assign != 0, check if the underlying type is interface
				// direct interface definition: ts.Type is *ast.InterfaceType
				if _, isIface := ts.Type.(*ast.InterfaceType); isIface {
					violations = append(violations, Violation{
						File:     relPath + "/alias.go",
						Line:     fset.Position(ts.Name.Pos()).Line,
						Rule:     "structure.domain-alias-no-interface",
						Message:  `alias.go re-exports interface "` + ts.Name.Name + `" — suspected cross-domain dependency; use orchestration/ instead`,
						Fix:      "move cross-domain coordination to orchestration/handler/ or orchestration/",
						Severity: cfg.Sev,
					})
				}
				// type alias to an interface from another package (e.g., type AdminOps = svc.AdminOps)
				// This is a SelectorExpr, not InterfaceType — we can't know if it's an interface
				// from AST alone. But we CAN flag all type aliases that reference core/svc/
				if ts.Assign != 0 {
					if sel, ok := ts.Type.(*ast.SelectorExpr); ok {
						if ident, ok := sel.X.(*ast.Ident); ok {
							// Check if the alias source looks like a svc/repo-adjacent import
							// We check imports to see what package the identifier refers to
							for _, imp := range file.Imports {
								impPath := strings.Trim(imp.Path.Value, `"`)
								alias := ""
								if imp.Name != nil {
									alias = imp.Name.Name
								} else {
									parts := strings.Split(impPath, "/")
									alias = parts[len(parts)-1]
								}
								isSvc := strings.Contains(impPath, "/core/svc")
								isRepo := strings.Contains(impPath, "/core/repo")
								if alias == ident.Name && (isSvc || isRepo) {
									src := "core/svc"
									if isRepo {
										src = "core/repo"
									}
									violations = append(violations, Violation{
										File:     relPath + "/alias.go",
										Line:     fset.Position(ts.Name.Pos()).Line,
										Rule:     "structure.domain-alias-no-interface",
										Message:  `alias.go re-exports "` + ts.Name.Name + `" from ` + src + ` — suspected cross-domain dependency; use orchestration/ instead`,
										Fix:      "move cross-domain coordination to orchestration/handler/ or orchestration/",
										Severity: cfg.Sev,
									})
								}
							}
						}
					}
				}
			}
		}
	}
	return violations
}

func checkPackageNames(internalDir string, cfg Config) []Violation {
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
		for _, banned := range bannedPackageNames {
			if name == banned {
				violations = append(violations, Violation{
					File:     rel + "/",
					Rule:     "structure.banned-package",
					Message:  `package "` + name + `" is banned`,
					Fix:      "move to specific domain or pkg/",
					Severity: cfg.Sev,
				})
			}
		}

		for _, legacy := range legacyPackageNames {
			if name == legacy {
				violations = append(violations, Violation{
					File:     rel + "/",
					Rule:     "structure.legacy-package",
					Message:  `legacy package "` + name + `" should be migrated`,
					Fix:      "move app-specific wiring to cmd/ and shared helpers to internal/pkg/",
					Severity: cfg.Sev,
				})
			}
		}

		if isMisplacedLayerDir(rel, name) {
			violations = append(violations, Violation{
				File:     rel + "/",
				Rule:     "structure.legacy-package",
				Message:  `legacy package "` + name + `" should be migrated`,
				Fix:      "place app/handler/infra only in domain slices or orchestration handler",
				Severity: cfg.Sev,
			})
		}
		return nil
	})
	return violations
}

func checkMiddlewarePlacement(internalDir string, cfg Config) []Violation {
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
		if rel == "internal/pkg/middleware" {
			return nil
		}
		violations = append(violations, Violation{
			File:     rel + "/",
			Rule:     "structure.middleware-placement",
			Message:  `middleware found at "` + rel + `"`,
			Fix:      "move middleware to internal/pkg/middleware/",
			Severity: cfg.Sev,
		})
		return nil
	})
	return violations
}

func checkDomainModelRequired(domainDir string, cfg Config) []Violation {
	var violations []Violation
	entries, err := os.ReadDir(domainDir)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		relPath := filepath.ToSlash(filepath.Join("internal", "domain", e.Name()))
		if cfg.IsExcluded(relPath + "/") {
			continue
		}
		modelDir := filepath.Join(domainDir, e.Name(), "core", "model")
		if hasNonTestGoFiles(modelDir) {
			continue
		}
		violations = append(violations, Violation{
			File:     relPath + "/",
			Rule:     "structure.domain-model-required",
			Message:  `domain "` + e.Name() + `" missing a direct non-test Go file in core/model/`,
			Fix:      "add at least one non-test Go file directly under core/model/",
			Severity: cfg.Sev,
		})
	}
	return violations
}

func checkDTOPlacement(internalDir string, cfg Config) []Violation {
	var violations []Violation
	// Only domain/ is walked; internal/infra/ cannot exist (blocked by structure.internal-top-level).
	domainDir := filepath.Join(internalDir, "domain")
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
			if isDTOAllowedSublayer(rel) {
				return nil
			}
			violations = append(violations, Violation{
				File:     rel,
				Rule:     "structure.dto-placement",
				Message:  `"` + name + `" found in forbidden layer`,
				Fix:      "DTOs belong in handler/ or app/",
				Severity: cfg.Sev,
			})
		}
		return nil
	})
	return violations
}

func isDTOAllowedSublayer(relPath string) bool {
	// relPath: "internal/domain/<name>/<sublayer>/..."
	parts := strings.Split(relPath, "/")
	if len(parts) < 4 {
		return false
	}
	sublayer := parts[3]
	return sublayer == "handler" || sublayer == "app"
}

func isMisplacedLayerDir(rel, name string) bool {
	switch name {
	case "app", "infra":
		return !matchesDomainLayer(rel, name)
	case "handler":
		return !matchesDomainLayer(rel, name) && rel != "internal/orchestration/handler"
	default:
		return false
	}
}

func matchesDomainLayer(rel, name string) bool {
	parts := strings.Split(rel, "/")
	return len(parts) == 4 && parts[0] == "internal" && parts[1] == "domain" && parts[2] != "" && parts[3] == name
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
