package analyzer

import (
	"reflect"
	"runtime/debug"
	"testing"
)

func TestMergeBuildFlags(t *testing.T) {
	settings := []debug.BuildSetting{{Key: "-tags", Value: "integration,enterprise"}, {Key: "-race", Value: "true"}, {Key: "-msan", Value: "false"}, {Key: "-ldflags", Value: "-s"}}
	for _, tc := range []struct {
		name           string
		explicit, want []string
	}{
		{"inherit relevant flags", nil, []string{"-tags=integration,enterprise", "-race=true"}},
		{"override equals form", []string{"-tags=other", "-race=false"}, []string{"-tags=other", "-race=false"}},
		{"override split form", []string{"-tags", "other"}, []string{"-race=true", "-tags", "other"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := mergeBuildFlags(settings, tc.explicit); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDisableBuildFlagInheritance(t *testing.T) {
	explicit := []string{"-tags=chosen"}
	got := loadBuildFlags(LoadOptions{BuildFlags: explicit, DisableBuildFlagInheritance: true})
	if !reflect.DeepEqual(got, explicit) {
		t.Fatalf("got %v", got)
	}
	got[0] = "modified"
	if explicit[0] != "-tags=chosen" {
		t.Fatal("options slice was aliased")
	}
}
