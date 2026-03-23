package rules

import (
	"os"
	"path/filepath"
	"strings"
)

var bannedPackageNames = []string{"util", "common", "misc", "helper", "shared"}

var legacyPackageNames = []string{"handler", "app", "infra"}

func CheckStructure(projectRoot string, opts ...Option) []Violation {
	cfg := NewConfig(opts...)
	var violations []Violation

	internalDir := filepath.Join(projectRoot, "internal")
	if _, err := os.Stat(internalDir); err != nil {
		return nil
	}

	violations = append(violations, checkBannedPackages(internalDir, cfg)...)
	violations = append(violations, checkLegacyPackages(internalDir, cfg)...)
	violations = append(violations, checkMiddlewarePlacement(internalDir, cfg)...)

	domainDir := filepath.Join(internalDir, "domain")
	violations = append(violations, checkDomainRootAliasOnly(domainDir, cfg)...)
	violations = append(violations, checkDomainModelRequired(domainDir, cfg)...)
	violations = append(violations, checkDTOPlacement(internalDir, cfg)...)

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
			relPath := filepath.Join("internal", "domain", e.Name(), name)
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

func checkBannedPackages(internalDir string, cfg Config) []Violation {
	var violations []Violation
	entries, err := os.ReadDir(internalDir)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		relPath := filepath.Join("internal", e.Name())
		if cfg.IsExcluded(relPath + "/") {
			continue
		}
		for _, banned := range bannedPackageNames {
			if e.Name() == banned {
				violations = append(violations, Violation{
					File:     relPath + "/",
					Rule:     "structure.banned-package",
					Message:  `package "` + e.Name() + `" is banned`,
					Fix:      "move to specific domain or pkg/",
					Severity: cfg.Sev,
				})
			}
		}
	}
	return violations
}

func checkLegacyPackages(internalDir string, cfg Config) []Violation {
	var violations []Violation
	entries, err := os.ReadDir(internalDir)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		relPath := filepath.Join("internal", e.Name())
		if cfg.IsExcluded(relPath + "/") {
			continue
		}
		for _, legacy := range legacyPackageNames {
			if e.Name() == legacy {
				violations = append(violations, Violation{
					File:     relPath + "/",
					Rule:     "structure.legacy-package",
					Message:  `legacy package "` + e.Name() + `" should be migrated`,
					Fix:      "move handlers to domain/*/handler/, middleware to pkg/, router to internal/router/",
					Severity: cfg.Sev,
				})
			}
		}
	}
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
		if cfg.IsExcluded(rel + "/") {
			return nil
		}
		// middleware/ under pkg/ is OK
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
		relPath := filepath.Join("internal", "domain", e.Name())
		if cfg.IsExcluded(relPath + "/") {
			continue
		}
		modelFile := filepath.Join(domainDir, e.Name(), "model.go")
		modelDir := filepath.Join(domainDir, e.Name(), "core", "model")
		_, fileErr := os.Stat(modelFile)
		_, dirErr := os.Stat(modelDir)
		if fileErr != nil && dirErr != nil {
			violations = append(violations, Violation{
				File:     relPath + "/",
				Rule:     "structure.domain-model-required",
				Message:  `domain "` + e.Name() + `" missing required file "model.go"`,
				Fix:      "create model.go with domain entities",
				Severity: cfg.Sev,
			})
		}
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
