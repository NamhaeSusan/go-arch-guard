package goarchguard_test

import (
	"fmt"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/rules"
)

func Example() {
	pkgs, err := analyzer.Load("testdata/valid", "internal/...", "cmd/...")
	if err != nil {
		fmt.Println(err)
		return
	}

	violations := rules.RunAll(pkgs, "github.com/kimtaeyun/testproject-dc", "testdata/valid")

	fmt.Printf("violations: %d\n", len(violations))
	// Output: violations: 0
}
