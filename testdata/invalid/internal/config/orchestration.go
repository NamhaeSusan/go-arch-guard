package config

// VIOLATION: non-domain internal package imports orchestration directly
import _ "github.com/kimtaeyun/testproject-dc-invalid/internal/orchestration"

func UseOrchestration() {}
