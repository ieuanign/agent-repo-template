package apperr

import (
	"errors"
	"fmt"
	"testing"
)

func TestKinds(t *testing.T) {
	tests := []struct {
		kind Kind
		code string
	}{
		{NotFound, "NOT_FOUND"},
		{Unauthorized, "UNAUTHORIZED"},
		{Forbidden, "FORBIDDEN"},
		{Validation, "VALIDATION"},
		{Conflict, "CONFLICT"},
		{RateLimited, "RATE_LIMITED"},
		{Timeout, "TIMEOUT"},
		{Internal, "INTERNAL"},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			var err error = tt.kind
			wrapped := fmt.Errorf("loading the project: %w", err)
			var got Kind
			if !errors.As(wrapped, &got) {
				t.Fatalf("errors.As found no Kind in %v", wrapped)
			}
			if got.Code() != tt.code {
				t.Fatalf("Code() = %q, want %q", got.Code(), tt.code)
			}
			if got.Message() == "" {
				t.Fatal("Message() is empty")
			}
			if !errors.Is(wrapped, tt.kind) {
				t.Fatalf("errors.Is(wrapped, %s) = false", tt.code)
			}
		})
	}
}

func TestNew(t *testing.T) {
	err := fmt.Errorf("approving the estimate: %w",
		New(Conflict, "PROJECTS_PROJECT_LOCKED", "The project is locked"))

	var got *Error
	if !errors.As(err, &got) {
		t.Fatalf("errors.As found no *Error in %v", err)
	}
	if got.Kind() != Conflict || got.Code() != "PROJECTS_PROJECT_LOCKED" || got.Message() != "The project is locked" {
		t.Fatalf("got kind %s, code %q, message %q", got.Kind().Code(), got.Code(), got.Message())
	}
	if !errors.Is(err, Conflict) {
		t.Fatal("refined error does not match its kind with errors.Is")
	}
	if errors.Is(err, NotFound) {
		t.Fatal("refined error matches a different kind")
	}
}
