package main

// homelabRuntimeError represents an error observed during the runtime and
// execution of the homelab business logic, as opposed to any errors
// observed during the initial command invocation, parsing the flags, etc.
type homelabRuntimeError struct {
	err error
}

func (e *homelabRuntimeError) Error() string {
	return e.err.Error()
}

func newHomelabRuntimeError(err error) *homelabRuntimeError {
	return &homelabRuntimeError{err: err}
}
