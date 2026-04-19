package root

import _ "github.com/kimtaeyun/testproject-load-broken-dep-transitive/internal/middle"

func Root() string {
	return "root"
}
