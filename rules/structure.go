package rules

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

var bannedPackageNames = []string{"util", "common", "misc", "helper", "shared"}

var legacyPackageNames = []string{"router", "bootstrap"}

func CheckStructure(projectRoot string, opts ...Option) []Violation {
	cfg := NewConfig(opts...)
	var violations []Violation

	internalDir := filepath.Join(projectRoot, "internal")
	if _, err := os.Stat(internalDir); err != nil {
		return nil
	}

	violations = append(violations, checkPackageNames(internalDir, cfg)...)
	violations = append(violations, checkMiddlewarePlacement(internalDir, cfg)...)

	domainDir := filepath.Join(internalDir, "domain")
	violations = append(violations, checkDomainRootAliasRequired(domainDir, cfg)...)
	violations = append(violations, checkDomainRootAliasPackage(domainDir, cfg)...)
	violations = append(violations, checkDomainRootAliasOnly(domainDir, cfg)...)
	violations = append(violations, checkDomainModelRequired(domainDir, cfg)...)
	violations = append(violations, checkDTOPlacement(internalDir, cfg)...)

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

func checkPackageNames(internalDir string, cfg Config) []Violation {
	var violations []Violation
	_ = filepath.Walk(internalDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
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

		name := info.Name()
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
	_ = filepath.Walk(internalDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}
		if info.Name() != "middleware" {
			return nil
		}
		rel, _ := filepath.Rel(filepath.Dir(internalDir), path)
		rel = filepath.ToSlash(rel)
		if cfg.IsExcluded(rel + "/") {
			return nil
		}
		if strings.Contains(rel, "pkg/") {
			return nil
		}
		violations = append(violations, Violation{
			File:     rel + "/",
			Rule:     "structure.middleware-placement",
			Message:  `middleware found at "` + rel + `"`,
			Fix:      "move middleware to pkg/middleware/",
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
			Message:  `domain "` + e.Name() + `" missing non-empty core/model/`,
			Fix:      "add at least one non-test Go file under core/model/",
			Severity: cfg.Sev,
		})
	}
	return violations
}

func checkDTOPlacement(internalDir string, cfg Config) []Violation {
	var violations []Violation
	for _, forbidden := range []string{"domain", "infra"} {
		dir := filepath.Join(internalDir, forbidden)
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			name := info.Name()
			if !strings.HasSuffix(name, ".go") {
				return nil
			}
			if name == "dto.go" || (strings.HasSuffix(name, "_dto.go") && !strings.HasSuffix(name, "_test.go")) {
				rel, _ := filepath.Rel(filepath.Dir(internalDir), path)
				rel = filepath.ToSlash(rel)
				if cfg.IsExcluded(rel) {
					return nil
				}
				violations = append(violations, Violation{
					File:     rel,
					Rule:     "structure.dto-placement",
					Message:  `"` + name + `" found in ` + forbidden + "/",
					Fix:      "DTOs belong in handler/ or app/",
					Severity: cfg.Sev,
				})
			}
			return nil
		})
	}
	return violations
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
