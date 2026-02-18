package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunSuccess(t *testing.T) {
	t.Parallel()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := run([]string{"help"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("run(help) code=%d want=0", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
	if got := stdout.String(); !strings.Contains(got, "Read and edit PDF metadata") {
		t.Fatalf("unexpected help output: %q", got)
	}
}

func TestRunFailureExitCodeAndError(t *testing.T) {
	t.Parallel()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := run([]string{"show", "--file", "missing.pdf"}, stdout, stderr)
	if code != 4 {
		t.Fatalf("run(show) code=%d want=4", code)
	}
	if got := stderr.String(); !strings.Contains(got, "read pdf") {
		t.Fatalf("expected service error in stderr, got %q", got)
	}
}
