package apperr

// Kind classifies expected service-layer failures without tying services to HTTP.
type Kind string

const (
	KindInvalidArgument  Kind = "invalid_argument"
	KindNotFound         Kind = "not_found"
	KindConflict         Kind = "conflict"
	KindForbidden        Kind = "forbidden"
	KindDependencyFailed Kind = "dependency_failed"
)

// Error is returned by services for expected business or validation failures.
type Error struct {
	Kind    Kind
	Message string
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func InvalidArgument(message string) error {
	return &Error{Kind: KindInvalidArgument, Message: message}
}

func NotFound(message string) error {
	return &Error{Kind: KindNotFound, Message: message}
}

func Conflict(message string) error {
	return &Error{Kind: KindConflict, Message: message}
}

func Forbidden(message string) error {
	return &Error{Kind: KindForbidden, Message: message}
}

func DependencyFailed(message string) error {
	return &Error{Kind: KindDependencyFailed, Message: message}
}
