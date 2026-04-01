package report_test

import (
	"fmt"

	"github.com/NamhaeSusan/go-arch-guard/report"
	"github.com/NamhaeSusan/go-arch-guard/rules"
)

func ExampleMarshalJSONReport() {
	data, err := report.MarshalJSONReport([]rules.Violation{{
		Rule:     "test.rule",
		Message:  "bad import",
		Severity: rules.Error,
	}})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(data))
	// Output:
	// {
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
	//       "rule": "test.rule",
	//       "message": "bad import",
	//       "severity": "error"
	//     }
	//   ]
	// }
}
