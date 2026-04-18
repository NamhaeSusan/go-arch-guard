package rules_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"golang.org/x/tools/go/packages"
)

var (
	validOnce sync.Once
	validPkgs []*packages.Package
	validErr  error

	invalidOnce sync.Once
	invalidPkgs []*packages.Package
	invalidErr  error

	blastOnce sync.Once
	blastPkgs []*packages.Package
	blastErr  error
)

func loadValid(t *testing.T) []*packages.Package {
	t.Helper()
	validOnce.Do(func() {
		validPkgs, validErr = analyzer.Load("../testdata/valid", "internal/...", "cmd/...")
	})
	if validErr != nil {
		t.Fatal(validErr)
	}
	return validPkgs
}

func loadInvalid(t *testing.T) []*packages.Package {
	t.Helper()
	invalidOnce.Do(func() {
		invalidPkgs, invalidErr = analyzer.Load("../testdata/invalid", "internal/...", "cmd/...")
	})
	if invalidErr != nil {
		t.Fatal(invalidErr)
	}
	return invalidPkgs
}

func loadBlast(t *testing.T) []*packages.Package {
	t.Helper()
	blastOnce.Do(func() {
		blastPkgs, blastErr = analyzer.Load("../testdata/blast", "internal/...")
	})
	if blastErr != nil {
		t.Fatal(blastErr)
	}
	return blastPkgs
}

var (
	txboundaryOnce sync.Once
	txboundaryPkgs []*packages.Package
	txboundaryErr  error
)

func loadTxBoundary(t *testing.T) []*packages.Package {
	t.Helper()
	txboundaryOnce.Do(func() {
		txboundaryPkgs, txboundaryErr = analyzer.Load("../testdata/txboundary", "internal/...", "cmd/...")
	})
	if txboundaryErr != nil {
		t.Fatal(txboundaryErr)
	}
	return txboundaryPkgs
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
