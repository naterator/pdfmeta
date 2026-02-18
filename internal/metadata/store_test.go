package metadata

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
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
