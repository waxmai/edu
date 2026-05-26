package apperr

import (
	"errors"
	"testing"
)

func TestConstructors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		kind Kind
	}{
		{name: "invalid argument", err: InvalidArgument("bad request"), kind: KindInvalidArgument},
		{name: "not found", err: NotFound("missing"), kind: KindNotFound},
		{name: "conflict", err: Conflict("duplicate"), kind: KindConflict},
		{name: "forbidden", err: Forbidden("denied"), kind: KindForbidden},
		{name: "dependency failed", err: DependencyFailed("mysql unavailable"), kind: KindDependencyFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var appErr *Error
			if !errors.As(tt.err, &appErr) {
				t.Fatalf("errors.As(%T) = false", tt.err)
			}
			if appErr.Kind != tt.kind {
				t.Fatalf("kind = %q, want %q", appErr.Kind, tt.kind)
			}
			if appErr.Error() == "" {
				t.Fatal("message is empty")
			}
		})
	}
}
