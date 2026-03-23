package config

// VIOLATION: non-domain internal package imports a domain sub-package directly
import _ "github.com/kimtaeyun/testproject-dc-invalid/internal/domain/user/handler/http"

func Load() {}
