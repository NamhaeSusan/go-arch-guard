package goarchguard_test

import (
	"fmt"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/presets"
)

func Example() {
	pkgs, err := analyzer.Load("testdata/valid", "internal/...", "cmd/...")
	if err != nil {
		fmt.Println(err)
		return
	}

	arch := presets.DDD()
	ctx := core.NewContext(pkgs, "github.com/kimtaeyun/testproject-dc", "testdata/valid", arch, nil)
	rs := presets.RecommendedDDD()
	violations := core.Run(ctx, rs)

	fmt.Printf("violations: %d\n", len(violations))
	// Output: violations: 0
}
