package pdf

import (
	"os"
	"path/filepath"
	"testing"

	"pdfmeta/internal/model"
)

func TestOpenFixture(t *testing.T) {
	doc, err := Open(fixturePath("minimal.pdf"))
	if err != nil {
		t.Fatalf("Open(minimal.pdf): %v", err)
	}
	if doc.Encrypted() {
		t.Fatalf("minimal.pdf should not be encrypted")
	}
	if doc.HeaderOffset() != 0 {
		t.Fatalf("HeaderOffset()=%d want 0", doc.HeaderOffset())
	}
	if doc.Version() == "" {
		t.Fatalf("Version() should not be empty")
	}
}

func TestOpenEncryptedFixture(t *testing.T) {
	doc, err := Open(fixturePath("encrypted-marker.pdf"))
	if err != nil {
		t.Fatalf("Open(encrypted-marker.pdf): %v", err)
	}
	if !doc.Encrypted() {
		t.Fatalf("encrypted-marker.pdf should be encrypted")
	}
}

func TestOpenMissingPath(t *testing.T) {
	_, err := Open(filepath.Join(t.TempDir(), "missing.pdf"))
	if err == nil {
		t.Fatalf("expected error for missing file")
	}
	assertAppErrorCode(t, err, model.ErrNotFound)
}

func TestOpenEmptyPath(t *testing.T) {
	_, err := Open("")
	if err == nil {
		t.Fatalf("expected error for empty path")
	}
	assertAppErrorCode(t, err, model.ErrValidation)
}

func TestParseBytesMalformed(t *testing.T) {
	_, err := ParseBytes("bad.pdf", []byte("not a pdf"))
	if err == nil {
		t.Fatalf("expected malformed error")
	}
	assertAppErrorCode(t, err, model.ErrPDFMalformed)
}

func TestParseBytesHeaderWindow(t *testing.T) {
	tooLate := make([]byte, headerScanLimit+8)
	copy(tooLate[headerScanLimit:], []byte("%PDF-1.7"))

	_, err := ParseBytes("late.pdf", tooLate)
	if err == nil {
		t.Fatalf("expected malformed error for late header")
	}
	assertAppErrorCode(t, err, model.ErrPDFMalformed)
}

func TestParseBytesAllowsLeadingJunkWithinWindow(t *testing.T) {
	b := readFixture(t, "leading-junk-before-header.pdf")
	doc, err := ParseBytes("leading-junk-before-header.pdf", b)
	if err != nil {
		t.Fatalf("ParseBytes leading junk: %v", err)
	}
	if doc.HeaderOffset() <= 0 {
		t.Fatalf("expected positive header offset, got %d", doc.HeaderOffset())
	}
}

func TestBytesReturnsCopy(t *testing.T) {
	doc, err := Open(fixturePath("minimal.pdf"))
	if err != nil {
		t.Fatalf("Open(minimal.pdf): %v", err)
	}
	b := doc.Bytes()
	if len(b) == 0 {
		t.Fatalf("expected non-empty bytes")
	}
	orig := b[0]
	b[0] = 'X'

	b2 := doc.Bytes()
	if b2[0] != orig {
		t.Fatalf("Bytes() should return defensive copy")
	}
}

func TestParseBytesPathRoundTrip(t *testing.T) {
	doc, err := ParseBytes("foo.pdf", readFixture(t, "minimal.pdf"))
	if err != nil {
		t.Fatalf("ParseBytes: %v", err)
	}
	if got := doc.Path(); got != "foo.pdf" {
		t.Fatalf("Path()=%q want foo.pdf", got)
	}
}

func TestParseBytesVersionParsing(t *testing.T) {
	doc, err := ParseBytes("v.pdf", []byte("%PDF-2.0\n%%EOF"))
	if err != nil {
		t.Fatalf("ParseBytes: %v", err)
	}
	if got := doc.Version(); got != "2.0" {
		t.Fatalf("Version()=%q want 2.0", got)
	}
}

func TestParseBytesNoPanicOnArbitraryInput(t *testing.T) {
	inputs := [][]byte{
		[]byte(""),
		[]byte("%PDF-"),
		[]byte("%PDF-1.4"),
		[]byte("%PDF-1.4\r\ntrailer\r\n<< /Size 1 >>"),
		make([]byte, headerScanLimit+256),
	}
	for i, in := range inputs {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("input[%d] panicked: %v", i, r)
				}
			}()
			_, _ = ParseBytes("fuzz.pdf", in)
		}()
	}
}

func TestOpenIOError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "x"), []byte("x"), 0o000); err != nil {
		t.Fatalf("setup: %v", err)
	}
	_, err := Open(filepath.Join(dir, "x", "file.pdf"))
	if err == nil {
		t.Fatalf("expected io error")
	}
	assertAppErrorCode(t, err, model.ErrIO)
}

func assertAppErrorCode(t *testing.T, err error, want model.ErrorCode) {
	t.Helper()
	ae, ok := err.(*model.AppError)
	if !ok {
		t.Fatalf("error is %T, want *model.AppError", err)
	}
	if ae.Code != want {
		t.Fatalf("AppError.Code=%q want %q", ae.Code, want)
	}
}
