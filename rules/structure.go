package rules

import (
	"os"
	"path/filepath"
	"strings"
)

var bannedPackageNames = []string{"util", "common", "misc", "helper", "shared"}

func CheckStructure(projectRoot string, opts ...Option) []Violation {
	cfg := NewConfig(opts...)
	var violations []Violation

	internalDir := filepath.Join(projectRoot, "internal")
	if _, err := os.Stat(internalDir); err != nil {
		return nil
	}

	violations = append(violations, checkBannedPackages(internalDir, cfg)...)

	domainDir := filepath.Join(internalDir, "domain")
	violations = append(violations, checkDomainModelRequired(domainDir, cfg)...)
	violations = append(violations, checkDTOPlacement(internalDir, cfg)...)

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
		modelPath := filepath.Join(domainDir, e.Name(), "model.go")
		if _, err := os.Stat(modelPath); err != nil {
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
