package structural

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/core"
)

const (
	ruleDTOPlacement = "structural.dto-placement"
	dtoPlacement     = "structural.dto-placement"
)

var (
	defaultDTOFilenames        = []string{"dto.go"}
	defaultDTOFilenameSuffixes = []string{"_dto.go"}
)

// DTOPlacement flags DTO files (default: "dto.go" or any "*_dto.go" except
// "*_dto_test.go") found in domain sublayers outside Structure.DTOAllowedLayers.
// Override the filename rules with WithDTOFilenames / WithDTOFilenameSuffixes.
//
// The rule is a no-op when Layout.DomainDir is empty (flat layouts have no
// domain sublayer concept). It does NOT emit meta.rule-disabled-by-config in
// that case because flat-layout users typically don't bundle this rule —
// presets that target flat layouts already exclude DTOPlacement.
type DTOPlacement struct {
	severity         core.Severity
	dtoFilenames     []string
	dtoFilenameSufxs []string
}

func NewDTOPlacement(opts ...Option) *DTOPlacement {
	cfg := newConfig(opts, core.Error)
	filenames := cfg.dtoFilenames
	if filenames == nil {
		filenames = defaultDTOFilenames
	}
	suffixes := cfg.dtoFilenameSuffixes
	if suffixes == nil {
		suffixes = defaultDTOFilenameSuffixes
	}
	return &DTOPlacement{severity: cfg.severity, dtoFilenames: filenames, dtoFilenameSufxs: suffixes}
}

func (r *DTOPlacement) Spec() core.RuleSpec {
	return withSeverity(core.RuleSpec{
		ID:              ruleDTOPlacement,
		Description:     "DTO files must live in DTOAllowedLayers",
		DefaultSeverity: r.severity,
	}, r.severity)
}

func (r *DTOPlacement) Check(ctx *core.Context) []core.Violation {
	if ctx == nil {
		return nil
	}
	arch := ctx.Arch()
	if !hasInternalDir(ctx.Root(), arch.Layout.InternalRoot) {
		return []core.Violation{metaLayoutNotSupported(ruleDTOPlacement)}
	}
	if arch.Layout.DomainDir == "" {
		return nil
	}
	internalDir := filepath.Join(ctx.Root(), arch.Layout.InternalRoot)
	domainDir := filepath.Join(internalDir, filepath.FromSlash(arch.Layout.DomainDir))
	if _, err := os.Stat(domainDir); err != nil {
		return nil
	}
	var violations []core.Violation
	_ = filepath.WalkDir(domainDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") {
			return nil
		}
		if !r.isDTOFilename(name) {
			return nil
		}
		rel, relErr := filepath.Rel(filepath.Dir(internalDir), path)
		if relErr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if ctx.IsExcluded(rel) || isDTOAllowedSublayer(arch, rel) {
			return nil
		}
		violations = append(violations, violation(r.severity, dtoPlacement, rel,
			`"`+name+`" found in forbidden layer`,
			"DTOs belong in "+strings.Join(arch.Structure.DTOAllowedLayers, ", ")))
		return nil
	})
	return violations
}

func (r *DTOPlacement) isDTOFilename(name string) bool {
	if strings.HasSuffix(name, "_test.go") {
		return false
	}
	if slices.Contains(r.dtoFilenames, name) {
		return true
	}
	for _, suffix := range r.dtoFilenameSufxs {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

// isDTOAllowedSublayer reports whether the file at rel sits in one of the
// DTOAllowedLayers. The check tries every nested depth from the sublayer
// start position so DTOAllowedLayers entries like "core/repo" match nested
// directories — "core" is tested first, then "core/repo", and so on. The
// last segment of parts is the filename, so it is excluded from candidates.
func isDTOAllowedSublayer(arch core.Architecture, rel string) bool {
	domainDepth := len(strings.Split(arch.Layout.DomainDir, "/"))
	parts := strings.Split(rel, "/")
	sublayerStart := 1 + domainDepth + 1
	if len(parts) <= sublayerStart {
		return false
	}
	for depth := 1; sublayerStart+depth < len(parts); depth++ {
		candidate := strings.Join(parts[sublayerStart:sublayerStart+depth], "/")
		if slices.Contains(arch.Structure.DTOAllowedLayers, candidate) {
			return true
		}
	}
	return false
}

var _ core.Rule = (*DTOPlacement)(nil)
