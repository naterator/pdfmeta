package pdf

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func fixturePath(name string) string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "testdata", "pdf", name))
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(fixturePath(name))
	if err != nil {
		t.Fatalf("read fixture %q: %v", name, err)
	}
	return b
}

func TestFixtureFilesExistAndNonEmpty(t *testing.T) {
	names := []string{
		"minimal.pdf",
		"encrypted-marker.pdf",
		"encrypt-name-prefix-only.pdf",
		"encrypt-in-stream-only.pdf",
		"encrypt-after-startxref.pdf",
		"leading-junk-before-header.pdf",
		"truncated-no-eof.pdf",
		"malformed-trailer.pdf",
		"invalid.txt",
	}
	for _, name := range names {
		info, err := os.Stat(fixturePath(name))
		if err != nil {
			t.Fatalf("stat fixture %q: %v", name, err)
		}
		if info.Size() == 0 {
			t.Fatalf("fixture %q is empty", name)
		}
	}
}

func TestPDFFixturesHavePDFHeaderAndEOF(t *testing.T) {
	names := []string{"minimal.pdf", "encrypted-marker.pdf"}
	for _, name := range names {
		b := readFixture(t, name)
		if findHeaderOffset(b) < 0 {
			t.Fatalf("%q missing PDF header", name)
		}
		if !hasEOFMarker(b) {
			t.Fatalf("%q missing EOF marker", name)
		}
	}
}

func TestEncryptedMarkerFixtureContainsEncryptToken(t *testing.T) {
	b := readFixture(t, "encrypted-marker.pdf")
	if !hasEncryptMarkerInTrailer(b) {
		t.Fatalf("encrypted-marker.pdf should include /Encrypt marker")
	}
}

func TestFixtureSignalTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		looksPDF   bool
		hasEOF     bool
		hasEncrypt bool
	}{
		{name: "minimal.pdf", looksPDF: true, hasEOF: true, hasEncrypt: false},
		{name: "encrypted-marker.pdf", looksPDF: true, hasEOF: true, hasEncrypt: true},
		{name: "encrypt-name-prefix-only.pdf", looksPDF: true, hasEOF: true, hasEncrypt: false},
		{name: "encrypt-in-stream-only.pdf", looksPDF: true, hasEOF: true, hasEncrypt: false},
		{name: "encrypt-after-startxref.pdf", looksPDF: true, hasEOF: true, hasEncrypt: false},
		{name: "leading-junk-before-header.pdf", looksPDF: true, hasEOF: true, hasEncrypt: false},
		{name: "truncated-no-eof.pdf", looksPDF: true, hasEOF: false, hasEncrypt: false},
		{name: "malformed-trailer.pdf", looksPDF: true, hasEOF: true, hasEncrypt: false},
		{name: "invalid.txt", looksPDF: false, hasEOF: false, hasEncrypt: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			b := readFixture(t, tc.name)
			if got := findHeaderOffset(b) >= 0; got != tc.looksPDF {
				t.Fatalf("findHeaderOffset(%q)>=0=%v want %v", tc.name, got, tc.looksPDF)
			}
			if got := hasEOFMarker(b); got != tc.hasEOF {
				t.Fatalf("hasEOFMarker(%q)=%v want %v", tc.name, got, tc.hasEOF)
			}
			if got := hasEncryptMarkerInTrailer(b); got != tc.hasEncrypt {
				t.Fatalf("hasEncryptMarkerInTrailer(%q)=%v want %v", tc.name, got, tc.hasEncrypt)
			}
		})
	}
}

func FuzzFixtureSignals(f *testing.F) {
	seeds := [][]byte{
		readFixtureForFuzz("minimal.pdf"),
		readFixtureForFuzz("encrypted-marker.pdf"),
		readFixtureForFuzz("encrypt-name-prefix-only.pdf"),
		readFixtureForFuzz("encrypt-in-stream-only.pdf"),
		readFixtureForFuzz("encrypt-after-startxref.pdf"),
		readFixtureForFuzz("leading-junk-before-header.pdf"),
		readFixtureForFuzz("truncated-no-eof.pdf"),
		readFixtureForFuzz("malformed-trailer.pdf"),
		readFixtureForFuzz("invalid.txt"),
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, b []byte) {
		_ = findHeaderOffset(b)
		_ = hasEOFMarker(b)
		_ = hasEncryptMarkerInTrailer(b)
	})
}

func BenchmarkFixtureSignalChecks(b *testing.B) {
	data := readFixtureForFuzz("encrypted-marker.pdf")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = findHeaderOffset(data)
		_ = hasEOFMarker(data)
		_ = hasEncryptMarkerInTrailer(data)
	}
}

func hasEOFMarker(b []byte) bool {
	return bytes.Contains(b, []byte("%%EOF"))
}

func readFixtureForFuzz(name string) []byte {
	b, err := os.ReadFile(fixturePath(name))
	if err != nil {
		panic(err)
	}
	return b
}

func TestHasEncryptMarkerBoundaries(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		want bool
	}{
		{name: "exact-name-space", in: []byte("trailer << /Encrypt /Filter >> startxref"), want: true},
		{name: "exact-name-delimiter", in: []byte("trailer<< /Encrypt>>startxref"), want: true},
		{name: "prefix-only", in: []byte("trailer << /Encrypted true >> startxref"), want: false},
		{name: "embedded", in: []byte("trailer << /EncryptName 1 >> startxref"), want: false},
		{name: "outside-trailer", in: []byte("/Encrypt trailer << /Size 2 >> startxref"), want: false},
		{name: "absent", in: []byte("trailer << /Size 2 >> startxref"), want: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := hasEncryptMarkerInTrailer(tc.in); got != tc.want {
				t.Fatalf("hasEncryptMarkerInTrailer(%q)=%v want %v", tc.in, got, tc.want)
			}
		})
	}
}
