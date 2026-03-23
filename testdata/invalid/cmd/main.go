package main

// VIOLATION: cmd imports domain sub-package directly (should use alias)
import _ "github.com/kimtaeyun/testproject-dc-invalid/internal/domain/user/core/model"

func main() {}
