package main

// Valid: cmd imports domain alias only
import (
	_ "github.com/kimtaeyun/testproject-dc/internal/domain/order"
	_ "github.com/kimtaeyun/testproject-dc/internal/domain/user"
)

func main() {}
