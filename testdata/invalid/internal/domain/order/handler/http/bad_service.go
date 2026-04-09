package http

// OK: consumer-defined interface (non-repo naming) is allowed in handler.
type Service interface {
	DoSomething() error
}

// OK: unexported consumer interface is allowed in handler.
type auditLogger interface {
	Record(action string) error
}

// VIOLATION: repository-port interface ("*Repository") must live in core/repo/.
type OrderRepository interface {
	Find(id string) error
}
