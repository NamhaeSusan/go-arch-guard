package scaffold_test

import (
	"fmt"
	"strings"

	"github.com/NamhaeSusan/go-arch-guard/scaffold"
)

func ExampleArchitectureTest() {
	src, err := scaffold.ArchitectureTest(scaffold.PresetHexagonal, scaffold.ArchitectureTestOptions{
		PackageName: "myapp_test",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(strings.Contains(src, "m := rules.Hexagonal()"))
	fmt.Println(strings.Contains(src, "func TestArchitecture"))
	// Output:
	// true
	// true
}
