package config

// VIOLATION: non-domain internal package imports a domain root directly
import _ "github.com/kimtaeyun/testproject-dc-invalid/internal/domain/user"

func UseDomainAlias() {}
