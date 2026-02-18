package xmp

import (
	"bytes"
	"path/filepath"
	"runtime"
	"testing"

	"pdfmeta/internal/model"
)

func fixturePath(name string) string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "testdata", "pdf", name))
}

func TestMarshalUnmarshalRoundTrip(t *testing.T) {
	in := model.Metadata{
		Title:        "Doc",
		Author:       "Author",
		Subject:      "Subject",
		Keywords:     "a,b",
		Creator:      "pdfmeta",
		Producer:     "go-test",
		CreationDate: "2026-02-17T00:00:00Z",
		ModDate:      "2026-02-17T01:00:00Z",
	}
	packet, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	out, err := Unmarshal(packet)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out != in {
		t.Fatalf("roundtrip mismatch: got %#v want %#v", out, in)
	}
}

func TestUnmarshalNoPacket(t *testing.T) {
	_, err := Unmarshal([]byte("<root/>"))
	if err == nil {
		t.Fatalf("expected error for non-xmp input")
	}
}

func TestUpsertInsertBeforeEOF(t *testing.T) {
	pdf := []byte("%PDF-1.4\ntrailer\n<< /Size 1 >>\n%%EOF")
	meta := model.Metadata{Title: "Inserted"}

	out, err := Upsert(pdf, meta)
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if !bytes.Contains(out, []byte("x:xmpmeta")) {
		t.Fatalf("expected xmp packet insertion")
	}
	if !bytes.Contains(out, []byte("Inserted")) {
		t.Fatalf("expected metadata value in output packet")
	}
}

func TestUpsertReplaceExistingPacket(t *testing.T) {
	base := []byte("%PDF-1.4\n")
	first, err := Upsert(base, model.Metadata{Title: "first"})
	if err != nil {
		t.Fatalf("Upsert first: %v", err)
	}
	second, err := Upsert(first, model.Metadata{Title: "second"})
	if err != nil {
		t.Fatalf("Upsert second: %v", err)
	}
	if bytes.Count(second, []byte("<x:xmpmeta")) != 1 {
		t.Fatalf("expected one xmp open tag after replacement")
	}
	if bytes.Contains(second, []byte("first")) {
		t.Fatalf("expected old metadata to be replaced")
	}
	if !bytes.Contains(second, []byte("second")) {
		t.Fatalf("expected new metadata value after replacement")
	}
}

func TestExtract(t *testing.T) {
	pdf := []byte("%PDF-1.4\n")
	withXMP, err := Upsert(pdf, model.Metadata{Title: "extract"})
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	packet, ok := Extract(withXMP)
	if !ok {
		t.Fatalf("expected packet extraction success")
	}
	if !bytes.Contains(packet, []byte("extract")) {
		t.Fatalf("unexpected extracted packet content")
	}
}

func TestExtractIncompletePacket(t *testing.T) {
	in := []byte("%PDF-1.4\n<x:xmpmeta xmlns:x=\"adobe:ns:meta/\">")
	if _, ok := Extract(in); ok {
		t.Fatalf("expected extract=false for incomplete packet")
	}
}

func TestUpsertWithoutEOFAppendsPacket(t *testing.T) {
	base := []byte("%PDF-1.4\n")
	out, err := Upsert(base, model.Metadata{Title: "tail"})
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if !bytes.HasSuffix(out, []byte(xpacketEnd+"\n")) {
		t.Fatalf("expected packet to be appended at end when %%EOF absent")
	}
}

func TestUnmarshalMalformedXML(t *testing.T) {
	_, err := Unmarshal([]byte(`<x:xmpmeta><rdf:RDF>`))
	if err == nil {
		t.Fatalf("expected xml decode error")
	}
}
