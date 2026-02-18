package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"pdfmeta/internal/model"
	"pdfmeta/internal/template"
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
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatalf("write temp fixture: %v", err)
	}
	return path
}

func newTestService(t *testing.T) *Service {
	t.Helper()
	storePath := filepath.Join(t.TempDir(), "templates.json")
	return NewService(ServiceConfig{
		TemplateStore: template.NewFileStore(storePath),
	})
}

func TestSetAndShowRoundTrip(t *testing.T) {
	t.Parallel()

	svc := newTestService(t)
	in := copyFixture(t, "minimal.pdf")
	out := filepath.Join(t.TempDir(), "out.pdf")
	title := "Release Notes"
	author := "Doc Bot"

	_, err := svc.Set(context.Background(), model.SetRequest{
		IO: model.IOOptions{
			InputPath:  in,
			OutputPath: out,
		},
		Changes: model.MetadataPatch{
			Title:  &title,
			Author: &author,
		},
	})
	if err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, err := svc.Show(context.Background(), model.ShowRequest{InputPath: out})
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if got.Metadata.Title != title || got.Metadata.Author != author {
		t.Fatalf("unexpected metadata: %+v", got.Metadata)
	}
	if !got.InfoFound {
		t.Fatalf("expected info metadata to be present")
	}
	if !got.XMPFound {
		t.Fatalf("expected xmp metadata to be present")
	}
}

func TestTemplateSaveAndApply(t *testing.T) {
	t.Parallel()

	svc := newTestService(t)
	in := copyFixture(t, "minimal.pdf")
	out := filepath.Join(t.TempDir(), "templated.pdf")

	title := "Templated"
	if _, err := svc.TemplateSave(context.Background(), model.TemplateSaveRequest{
		Name: "release",
		Metadata: model.MetadataPatch{
			Title: &title,
		},
	}); err != nil {
		t.Fatalf("TemplateSave: %v", err)
	}

	res, err := svc.TemplateApply(context.Background(), model.TemplateApplyRequest{
		Name: "release",
		IO: model.IOOptions{
			InputPath:  in,
			OutputPath: out,
		},
	})
	if err != nil {
		t.Fatalf("TemplateApply: %v", err)
	}
	if res.Metadata.Title != title {
		t.Fatalf("expected title=%q, got %q", title, res.Metadata.Title)
	}
}

func TestBatchExecute(t *testing.T) {
	t.Parallel()

	svc := newTestService(t)
	in := copyFixture(t, "minimal.pdf")
	out := filepath.Join(t.TempDir(), "batch-out.pdf")
	title := "BatchTitle"
	manifestPath := filepath.Join(t.TempDir(), "manifest.json")

	manifest := map[string]any{
		"items": []map[string]any{
			{
				"op":     "set",
				"input":  in,
				"output": out,
				"set": map[string]any{
					"title": title,
				},
			},
			{
				"op":    "show",
				"input": out,
			},
		},
	}
	b, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := os.WriteFile(manifestPath, b, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	result, err := svc.Batch(context.Background(), model.BatchRequest{
		ManifestPath:    manifestPath,
		ContinueOnError: false,
	})
	if err != nil {
		t.Fatalf("Batch: %v", err)
	}
	if result.Failed != 0 || result.Succeeded != 2 {
		t.Fatalf("unexpected batch result: %+v", result)
	}
}
