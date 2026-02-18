package model

import (
	"errors"
	"fmt"
	"testing"
)

func TestExitCode(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  error
		want int
	}{
		{name: "nil", err: nil, want: 0},
		{name: "plain", err: errors.New("x"), want: 1},
		{name: "usage", err: &AppError{Code: ErrUsage}, want: 2},
		{name: "validation", err: &AppError{Code: ErrValidation}, want: 3},
		{name: "not-found", err: &AppError{Code: ErrNotFound}, want: 4},
		{name: "conflict", err: &AppError{Code: ErrConflict}, want: 5},
		{name: "pdf-encrypted", err: &AppError{Code: ErrPDFEncrypted}, want: 6},
		{name: "pdf-malformed", err: &AppError{Code: ErrPDFMalformed}, want: 7},
		{name: "io", err: &AppError{Code: ErrIO}, want: 8},
		{name: "internal", err: &AppError{Code: ErrInternal}, want: 9},
		{name: "wrapped", err: fmt.Errorf("wrap: %w", &AppError{Code: ErrValidation}), want: 3},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := ExitCode(tc.err); got != tc.want {
				t.Fatalf("ExitCode()=%d want=%d err=%v", got, tc.want, tc.err)
			}
		})
	}
}
