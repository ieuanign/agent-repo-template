// Package apperr declares the global error kinds and the error type modules
// refine them with. It knows nothing of transports.
package apperr

// Kind is one of the eight global error kinds; unexported fields keep the set closed.
type Kind struct {
	code    string
	message string
}

var (
	NotFound     = Kind{"NOT_FOUND", "The requested resource was not found."}
	Unauthorized = Kind{"UNAUTHORIZED", "Authentication is required."}
	Forbidden    = Kind{"FORBIDDEN", "You do not have permission to perform this action."}
	Validation   = Kind{"VALIDATION", "The request is invalid."}
	Conflict     = Kind{"CONFLICT", "The request conflicts with the current state of the resource."}
	RateLimited  = Kind{"RATE_LIMITED", "Too many requests. Please try again later."}
	Timeout      = Kind{"TIMEOUT", "The request timed out."}
	Internal     = Kind{"INTERNAL", "An internal error occurred."}
)

func (k Kind) Error() string   { return k.code }
func (k Kind) Code() string    { return k.code }
func (k Kind) Message() string { return k.message }

// Error refines a kind with a module-prefixed code and an English message.
type Error struct {
	kind    Kind
	code    string
	message string
}

func New(kind Kind, code, message string) *Error {
	return &Error{kind: kind, code: code, message: message}
}

func (e *Error) Error() string   { return e.code + ": " + e.message }
func (e *Error) Kind() Kind      { return e.kind }
func (e *Error) Code() string    { return e.code }
func (e *Error) Message() string { return e.message }

// Unwrap lets errors.Is match a refined error against its kind.
func (e *Error) Unwrap() error { return e.kind }
