package pkg

// VIOLATION: pkg must not depend on orchestration
import _ "github.com/kimtaeyun/testproject-dc-invalid/internal/orchestration"

func UseOrchestration() {}
