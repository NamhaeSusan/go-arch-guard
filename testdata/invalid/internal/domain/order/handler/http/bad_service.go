package http

// VIOLATION: handler defines exported interface (should use *app.Service)
type Service interface {
	DoSomething() error
}

// VIOLATION: handler defines unexported interface (hidden cross-domain dependency)
type auditLogger interface {
	Record(action string) error
}
