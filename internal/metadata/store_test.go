package metadata

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"pdfmeta/internal/model"
)

func fixturePath(name string) string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "testdata", "pdf", name))
}

func copyFixture(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(fixturePath(name))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	dst := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(dst, b, 0o644); err != nil {
		t.Fatalf("write temp fixture: %v", err)
	}
	return dst
}

func TestReadPlainFixture(t *testing.T) {
	store := NewStore()
	res, err := store.Read(context.Background(), fixturePath("minimal.pdf"))
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if res.Encrypted {
		t.Fatalf("expected unencrypted fixture")
	}
	if res.InfoFound {
		t.Fatalf("expected no info metadata in fixture")
	}
}

func TestReadEncryptedFixture(t *testing.T) {
	store := NewStore()
	res, err := store.Read(context.Background(), fixturePath("encrypted-marker.pdf"))
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !res.Encrypted {
		t.Fatalf("expected encrypted fixture to report encrypted=true")
	}
}

func TestWriteAndReadRoundTrip(t *testing.T) {
	store := NewStore()
	in := copyFixture(t, "minimal.pdf")
	out := filepath.Join(t.TempDir(), "out.pdf")

	title := "My Title"
	author := "A. Author"
	res, err := store.Write(context.Background(), model.MetadataWriteRequest{
		InputPath:  in,
		OutputPath: out,
		Set: model.MetadataPatch{
			Title:  &title,
			Author: &author,
		},
	})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !res.InfoFound {
		t.Fatalf("expected info metadata result")
	}
	if res.Metadata.Title != title || res.Metadata.Author != author {
		t.Fatalf("unexpected metadata after write: %#v", res.Metadata)
	}

	readBack, err := store.Read(context.Background(), out)
	if err != nil {
		t.Fatalf("Read(out): %v", err)
	}
	if !readBack.InfoFound {
		t.Fatalf("expected info metadata in written output")
	}
	if readBack.Metadata.Title != title || readBack.Metadata.Author != author {
		t.Fatalf("unexpected metadata readback: %#v", readBack.Metadata)
	}
	if !readBack.XMPFound {
		t.Fatalf("expected xmp metadata in written output")
	}
}

func TestWriteEscapesPDFLiteralStringsOnce(t *testing.T) {
	store := NewStore()
	in := copyFixture(t, "minimal.pdf")
	out := filepath.Join(t.TempDir(), "escaped.pdf")

	title := "My Doc (rev 1)\nNext"
	author := `A \ B`
	if _, err := store.Write(context.Background(), model.MetadataWriteRequest{
		InputPath:  in,
		OutputPath: out,
		Set: model.MetadataPatch{
			Title:  &title,
			Author: &author,
		},
	}); err != nil {
		t.Fatalf("Write: %v", err)
	}

	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !bytes.Contains(b, []byte(`/Title (My Doc \(rev 1\)\nNext)`)) {
		t.Fatalf("expected single escaped literal title, output:\n%s", b)
	}
	if bytes.Contains(b, []byte(`/Title (My Doc \\(rev 1\\)\\nNext)`)) {
		t.Fatalf("title literal was double escaped:\n%s", b)
	}

	readBack, err := store.Read(context.Background(), out)
	if err != nil {
		t.Fatalf("Read(out): %v", err)
	}
	if readBack.Metadata.Title != title || readBack.Metadata.Author != author {
		t.Fatalf("unexpected metadata readback: %#v", readBack.Metadata)
	}
}

func TestWriteNonASCIIInfoStringsAsUTF16BEHex(t *testing.T) {
	store := NewStore()
	in := copyFixture(t, "minimal.pdf")
	out := filepath.Join(t.TempDir(), "utf16.pdf")

	title := "Caf\u00e9 \u2014 Release"
	if _, err := store.Write(context.Background(), model.MetadataWriteRequest{
		InputPath:  in,
		OutputPath: out,
		Set: model.MetadataPatch{
			Title: &title,
		},
	}); err != nil {
		t.Fatalf("Write: %v", err)
	}

	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !bytes.Contains(b, []byte(`/Title <FEFF`)) {
		t.Fatalf("expected UTF-16BE hex title in Info dict, output:\n%s", b)
	}
	if bytes.Contains(b, []byte(`/Title (`)) {
		t.Fatalf("did not expect non-ASCII title to be written as a literal string:\n%s", b)
	}

	readBack, err := store.Read(context.Background(), out)
	if err != nil {
		t.Fatalf("Read(out): %v", err)
	}
	if readBack.Metadata.Title != title {
		t.Fatalf("title readback = %q, want %q", readBack.Metadata.Title, title)
	}
}

func TestParseInfoDictHandlesNestedParens(t *testing.T) {
	meta := parseInfoDict(`<< /Title (My Doc (rev 1)) /Author (A \(B\)) >>`)
	if meta.Title != "My Doc (rev 1)" {
		t.Fatalf("title = %q", meta.Title)
	}
	if meta.Author != "A (B)" {
		t.Fatalf("author = %q", meta.Author)
	}
}

func TestParseInfoDictDecodesUTF16BEHex(t *testing.T) {
	meta := parseInfoDict(`<< /Title <FEFF00430061006600E9> >>`)
	if want := "Caf\u00e9"; meta.Title != want {
		t.Fatalf("title = %q, want %q", meta.Title, want)
	}
}

func TestWriteCreatesNativeInfoAndMetadataRefs(t *testing.T) {
	store := NewStore()
	in := copyFixture(t, "minimal.pdf")
	out := filepath.Join(t.TempDir(), "native.pdf")

	title := "Native Info"
	if _, err := store.Write(context.Background(), model.MetadataWriteRequest{
		InputPath:  in,
		OutputPath: out,
		Set: model.MetadataPatch{
			Title: &title,
		},
	}); err != nil {
		t.Fatalf("Write: %v", err)
	}

	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	if !bytes.Contains(b, []byte("/Info ")) {
		t.Fatalf("expected trailer to reference /Info object")
	}
	if !bytes.Contains(b, []byte("/Metadata ")) {
		t.Fatalf("expected catalog to reference /Metadata object")
	}
	if !bytes.Contains(b, []byte("/Subtype /XML")) {
		t.Fatalf("expected metadata stream object")
	}
}

func TestWriteUnsetInPlace(t *testing.T) {
	store := NewStore()
	in := copyFixture(t, "minimal.pdf")

	title := "Title"
	if _, err := store.Write(context.Background(), model.MetadataWriteRequest{
		InputPath: in,
		InPlace:   true,
		Set: model.MetadataPatch{
			Title: &title,
		},
	}); err != nil {
		t.Fatalf("seed write: %v", err)
	}

	if _, err := store.Write(context.Background(), model.MetadataWriteRequest{
		InputPath: in,
		InPlace:   true,
		Unset:     []model.Field{model.FieldTitle},
	}); err != nil {
		t.Fatalf("unset write: %v", err)
	}

	res, err := store.Read(context.Background(), in)
	if err != nil {
		t.Fatalf("Read(in): %v", err)
	}
	if res.Metadata.Title != "" {
		t.Fatalf("expected title to be unset, got %q", res.Metadata.Title)
	}
}

func TestWriteEncryptedFails(t *testing.T) {
	store := NewStore()
	in := copyFixture(t, "encrypted-marker.pdf")
	out := filepath.Join(t.TempDir(), "out.pdf")
	title := "x"

	_, err := store.Write(context.Background(), model.MetadataWriteRequest{
		InputPath:  in,
		OutputPath: out,
		Set: model.MetadataPatch{
			Title: &title,
		},
	})
	assertAppErrorCode(t, err, model.ErrPDFEncrypted)
}

func TestWriteNeedsOutputWhenNotInPlace(t *testing.T) {
	store := NewStore()
	in := copyFixture(t, "minimal.pdf")

	_, err := store.Write(context.Background(), model.MetadataWriteRequest{
		InputPath: in,
	})
	assertAppErrorCode(t, err, model.ErrValidation)
}

func TestReadCanceledContext(t *testing.T) {
	store := NewStore()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := store.Read(ctx, fixturePath("minimal.pdf"))
	assertAppErrorCode(t, err, model.ErrInternal)
}

func TestWriteCanceledContext(t *testing.T) {
	store := NewStore()
	in := copyFixture(t, "minimal.pdf")
	out := filepath.Join(t.TempDir(), "out.pdf")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := store.Write(ctx, model.MetadataWriteRequest{
		InputPath:  in,
		OutputPath: out,
	})
	assertAppErrorCode(t, err, model.ErrInternal)
}

func TestWriteRejectsXRefStreams(t *testing.T) {
	store := NewStore()
	in := writePDFBytes(t, "xref-stream.pdf", []byte("%PDF-1.5\n1 0 obj\n<< /Type /XRef >>\nendobj\n%%EOF\n"))
	out := filepath.Join(t.TempDir(), "out.pdf")

	_, err := store.Write(context.Background(), model.MetadataWriteRequest{
		InputPath:  in,
		OutputPath: out,
	})
	assertAppErrorCode(t, err, model.ErrPDFMalformed)
	if !strings.Contains(err.Error(), "xref streams") {
		t.Fatalf("expected xref stream message, got %v", err)
	}
}

func TestWriteRejectsObjectStreams(t *testing.T) {
	store := NewStore()
	in := writePDFBytes(t, "object-stream.pdf", []byte("%PDF-1.5\n1 0 obj\n<< /Type /ObjStm >>\nendobj\n%%EOF\n"))
	out := filepath.Join(t.TempDir(), "out.pdf")

	_, err := store.Write(context.Background(), model.MetadataWriteRequest{
		InputPath:  in,
		OutputPath: out,
	})
	assertAppErrorCode(t, err, model.ErrPDFMalformed)
	if !strings.Contains(err.Error(), "object streams") {
		t.Fatalf("expected object stream message, got %v", err)
	}
}

func writePDFBytes(t *testing.T, name string, b []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatalf("write pdf bytes: %v", err)
	}
	return path
}

func assertAppErrorCode(t *testing.T, err error, want model.ErrorCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error %q, got nil", want)
	}
	ae, ok := err.(*model.AppError)
	if !ok {
		t.Fatalf("error is %T, want *model.AppError", err)
	}
	if ae.Code != want {
		t.Fatalf("AppError.Code=%q want %q", ae.Code, want)
	}
}
