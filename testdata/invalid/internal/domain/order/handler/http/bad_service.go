package http

// VIOLATION: handler defines exported interface (should use *app.Service)
type Service interface {
	DoSomething() error
}
