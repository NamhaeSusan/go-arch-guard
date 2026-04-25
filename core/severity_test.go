package core

import "testing"

func TestSeverityString(t *testing.T) {
	tests := []struct {
		name string
		s    Severity
		want string
	}{
		{"error", Error, "ERROR"},
		{"warning", Warning, "WARNING"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.String(); got != tt.want {
				t.Errorf("Severity.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
