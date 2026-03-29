package rules

import (
	"testing"
)

func TestArchModelConsistency(t *testing.T) {
	validateModelConsistency(t, defaultModel)
}
