package report_test

import (
	"fmt"

	"github.com/NamhaeSusan/go-arch-guard/core"
	"github.com/NamhaeSusan/go-arch-guard/report"
)

func ExampleMarshalJSONReport() {
	data, err := report.MarshalJSONReport([]core.Violation{{
		Rule:              "test.rule",
		Message:           "bad import",
		DefaultSeverity:   core.Error,
		EffectiveSeverity: core.Error,
	}})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(data))
	// Output:
	// {
	//   "schema": "go-arch-guard.report.v2",
	//   "summary": {
	//     "total": 1,
	//     "errors": 1,
	//     "warnings": 0,
	//     "files": 0,
	//     "rules": [
	//       "test.rule"
	//     ]
	//   },
	//   "violations": [
	//     {
	//       "file": "",
	//       "line": 0,
	//       "rule": "test.rule",
	//       "message": "bad import",
	//       "fix": "",
	//       "effectiveSeverity": "error",
	//       "defaultSeverity": "error"
	//     }
	//   ]
	// }
}
