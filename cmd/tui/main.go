package main

import (
	"fmt"
	"os"

	"github.com/NamhaeSusan/go-arch-guard/analyzer"
	"github.com/NamhaeSusan/go-arch-guard/tui"
)

func main() {
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	pkgs, err := analyzer.Load(dir, "internal/...", "cmd/...")
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: %v\n", err)
	}
	if len(pkgs) == 0 {
		fmt.Fprintln(os.Stderr, "no packages found")
		os.Exit(1)
	}

	module := ""
	for _, pkg := range pkgs {
		if pkg.Module != nil && pkg.Module.Path != "" {
			module = pkg.Module.Path
			break
		}
	}

	if err := tui.Run(pkgs, module); err != nil {
		fmt.Fprintf(os.Stderr, "tui error: %v\n", err)
		os.Exit(1)
	}
}
