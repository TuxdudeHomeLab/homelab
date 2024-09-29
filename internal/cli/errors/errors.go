package errors

// HomelabRuntimeError represents an error observed during the runtime and
// execution of the homelab business logic, as opposed to any errors
// observed during the initial command invocation, parsing the flags, etc.
type HomelabRuntimeError struct {
	err error
}

func (e *HomelabRuntimeError) Error() string {
	return e.err.Error()
}

func NewHomelabRuntimeError(err error) *HomelabRuntimeError {
	return &HomelabRuntimeError{err: err}
}
