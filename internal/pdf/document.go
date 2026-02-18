package pdf

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"pdfmeta/internal/model"
)

const (
	headerScanLimit = 1024
	pdfHeader       = "%PDF-"
)

// Document represents a parsed PDF envelope used by metadata readers/writers.
type Document struct {
	path         string
	content      []byte
	headerOffset int
	version      string
	encrypted    bool
}

// Open loads a PDF from disk and parses envelope metadata needed by callers.
func Open(path string) (*Document, error) {
	if path == "" {
		return nil, &model.AppError{
			Code:    model.ErrValidation,
			Message: "input path is required",
		}
	}

	b, err := os.ReadFile(path)
	if err != nil {
		code := model.ErrIO
		if errors.Is(err, os.ErrNotExist) {
			code = model.ErrNotFound
		}
		return nil, &model.AppError{
			Code:    code,
			Message: fmt.Sprintf("read pdf %q", path),
			Cause:   err,
		}
	}

	doc, err := ParseBytes(path, b)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// ParseBytes parses a PDF envelope from in-memory bytes.
func ParseBytes(path string, b []byte) (*Document, error) {
	if len(b) == 0 {
		return nil, &model.AppError{
			Code:    model.ErrPDFMalformed,
			Message: "pdf is empty",
		}
	}

	headerOffset := findHeaderOffset(b)
	if headerOffset < 0 {
		return nil, &model.AppError{
			Code:    model.ErrPDFMalformed,
			Message: "missing PDF header",
		}
	}

	doc := &Document{
		path:         path,
		content:      append([]byte(nil), b...),
		headerOffset: headerOffset,
		version:      parseVersionAt(b, headerOffset),
		encrypted:    hasEncryptMarkerInTrailer(b),
	}
	return doc, nil
}

func (d *Document) Path() string {
	if d == nil {
		return ""
	}
	return d.path
}

func (d *Document) HeaderOffset() int {
	if d == nil {
		return -1
	}
	return d.headerOffset
}

func (d *Document) Version() string {
	if d == nil {
		return ""
	}
	return d.version
}

func (d *Document) Encrypted() bool {
	return d != nil && d.encrypted
}

func (d *Document) Bytes() []byte {
	if d == nil {
		return nil
	}
	return append([]byte(nil), d.content...)
}

func findHeaderOffset(b []byte) int {
	limit := len(b)
	if limit > headerScanLimit {
		limit = headerScanLimit
	}
	idx := bytes.Index(b[:limit], []byte(pdfHeader))
	return idx
}

func parseVersionAt(b []byte, headerOffset int) string {
	start := headerOffset + len(pdfHeader)
	if start >= len(b) {
		return ""
	}
	end := start
	for end < len(b) {
		c := b[end]
		if c == '\n' || c == '\r' || c == ' ' || c == '\t' {
			break
		}
		end++
	}
	return string(b[start:end])
}

func hasEncryptMarkerInTrailer(b []byte) bool {
	trailerIdx := bytes.LastIndex(b, []byte("trailer"))
	if trailerIdx == -1 {
		return false
	}
	searchSpace := b[trailerIdx:]
	if sx := bytes.Index(searchSpace, []byte("startxref")); sx >= 0 {
		searchSpace = searchSpace[:sx]
	}
	return hasNameKey(searchSpace, "/Encrypt")
}

func hasNameKey(b []byte, key string) bool {
	start := 0
	for {
		idx := bytes.Index(b[start:], []byte(key))
		if idx == -1 {
			return false
		}
		idx += start
		next := idx + len(key)
		if next >= len(b) || isPDFNameTerminator(b[next]) {
			return true
		}
		start = idx + 1
	}
}

func isPDFNameTerminator(c byte) bool {
	switch c {
	case 0x00, 0x09, 0x0A, 0x0C, 0x0D, 0x20:
		return true
	case '(', ')', '<', '>', '[', ']', '{', '}', '/', '%':
		return true
	default:
		return false
	}
}
