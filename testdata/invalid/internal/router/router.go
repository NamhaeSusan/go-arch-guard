package router

// VIOLATION: router imports a domain directly even though only cmd/orchestration may do so
import _ "github.com/kimtaeyun/testproject-dc-invalid/internal/domain/user"

func Setup() {}
