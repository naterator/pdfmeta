package app

import (
	"context"
	"encoding/json"
	"errors"
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

type recordingMetadataStore struct {
	readCalls  int
	writeCalls int
}

func (s *recordingMetadataStore) Read(context.Context, string) (model.MetadataReadResult, error) {
	s.readCalls++
	return model.MetadataReadResult{}, nil
}

func (s *recordingMetadataStore) Write(context.Context, model.MetadataWriteRequest) (model.MetadataReadResult, error) {
	s.writeCalls++
	return model.MetadataReadResult{}, nil
}

type recordingTemplateStore struct {
	saveCalls   int
	importCalls int
	getCalls    int
}

func (s *recordingTemplateStore) Save(_ context.Context, rec model.TemplateRecord, _ bool) (model.TemplateRecord, error) {
	s.saveCalls++
	return rec, nil
}

func (s *recordingTemplateStore) Import(_ context.Context, records []model.TemplateRecord, _ bool) ([]model.TemplateRecord, error) {
	s.importCalls++
	return records, nil
}

func (s *recordingTemplateStore) Get(context.Context, string) (model.TemplateRecord, error) {
	s.getCalls++
	return model.TemplateRecord{}, nil
}

func (s *recordingTemplateStore) List(context.Context) ([]model.TemplateRecord, error) {
	return nil, nil
}

func (s *recordingTemplateStore) Delete(context.Context, string) error {
	return nil
}

func TestServiceValidatesRequestsBeforeStores(t *testing.T) {
	t.Parallel()

	metadataStore := &recordingMetadataStore{}
	templateStore := &recordingTemplateStore{}
	svc := NewService(ServiceConfig{
		MetadataStore: metadataStore,
		TemplateStore: templateStore,
	})
	ctx := context.Background()

	_, err := svc.Show(ctx, model.ShowRequest{})
	assertAppErrorCode(t, err, model.ErrValidation)

	_, err = svc.Set(ctx, model.SetRequest{
		IO: model.IOOptions{InputPath: "in.pdf", OutputPath: "out.pdf"},
	})
	assertAppErrorCode(t, err, model.ErrValidation)

	_, err = svc.Unset(ctx, model.UnsetRequest{
		IO: model.IOOptions{InputPath: "in.pdf", OutputPath: "out.pdf"},
	})
	assertAppErrorCode(t, err, model.ErrValidation)

	_, err = svc.TemplateSave(ctx, model.TemplateSaveRequest{Name: "empty"})
	assertAppErrorCode(t, err, model.ErrValidation)

	_, err = svc.TemplateImport(ctx, model.TemplateImportRequest{})
	assertAppErrorCode(t, err, model.ErrValidation)

	_, err = svc.TemplateApply(ctx, model.TemplateApplyRequest{
		IO: model.IOOptions{InputPath: "in.pdf", OutputPath: "out.pdf"},
	})
	assertAppErrorCode(t, err, model.ErrValidation)

	if metadataStore.readCalls != 0 || metadataStore.writeCalls != 0 {
		t.Fatalf("expected validation to avoid metadata store calls, got read=%d write=%d", metadataStore.readCalls, metadataStore.writeCalls)
	}
	if templateStore.saveCalls != 0 || templateStore.importCalls != 0 || templateStore.getCalls != 0 {
		t.Fatalf("expected validation to avoid template store calls, got save=%d import=%d get=%d", templateStore.saveCalls, templateStore.importCalls, templateStore.getCalls)
	}
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

func TestTemplateApplyMergesOverrides(t *testing.T) {
	t.Parallel()

	svc := newTestService(t)
	in := copyFixture(t, "minimal.pdf")
	out := filepath.Join(t.TempDir(), "templated-overrides.pdf")

	title := "Template Title"
	author := "Docs Team"
	overrideTitle := "Hotfix Title"
	if _, err := svc.TemplateSave(context.Background(), model.TemplateSaveRequest{
		Name: "release",
		Metadata: model.MetadataPatch{
			Title:  &title,
			Author: &author,
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
		Overrides: model.MetadataPatch{
			Title: &overrideTitle,
		},
	})
	if err != nil {
		t.Fatalf("TemplateApply: %v", err)
	}
	if res.Metadata.Title != overrideTitle {
		t.Fatalf("expected override title=%q, got %q", overrideTitle, res.Metadata.Title)
	}
	if res.Metadata.Author != author {
		t.Fatalf("expected template author=%q, got %q", author, res.Metadata.Author)
	}
}

func TestUnsetClearsSelectedMetadata(t *testing.T) {
	t.Parallel()

	svc := newTestService(t)
	in := copyFixture(t, "minimal.pdf")
	withMeta := filepath.Join(t.TempDir(), "with-meta.pdf")
	cleared := filepath.Join(t.TempDir(), "cleared.pdf")

	title := "Release Notes"
	author := "Docs Team"
	if _, err := svc.Set(context.Background(), model.SetRequest{
		IO: model.IOOptions{
			InputPath:  in,
			OutputPath: withMeta,
		},
		Changes: model.MetadataPatch{
			Title:  &title,
			Author: &author,
		},
	}); err != nil {
		t.Fatalf("Set: %v", err)
	}

	res, err := svc.Unset(context.Background(), model.UnsetRequest{
		IO: model.IOOptions{
			InputPath:  withMeta,
			OutputPath: cleared,
		},
		Fields: []model.Field{model.FieldTitle},
	})
	if err != nil {
		t.Fatalf("Unset: %v", err)
	}
	if res.InputPath != cleared {
		t.Fatalf("result input path = %q, want %q", res.InputPath, cleared)
	}
	if res.Metadata.Title != "" || res.Metadata.Author != author {
		t.Fatalf("unexpected metadata after unset: %+v", res.Metadata)
	}
}

func TestTemplateListShowDelete(t *testing.T) {
	t.Parallel()

	svc := newTestService(t)
	title := "Release Notes"
	author := "Docs Team"
	if _, err := svc.TemplateSave(context.Background(), model.TemplateSaveRequest{
		Name: "release",
		Metadata: model.MetadataPatch{
			Title: &title,
		},
	}); err != nil {
		t.Fatalf("TemplateSave release: %v", err)
	}
	if _, err := svc.TemplateSave(context.Background(), model.TemplateSaveRequest{
		Name: "author",
		Metadata: model.MetadataPatch{
			Author: &author,
		},
	}); err != nil {
		t.Fatalf("TemplateSave author: %v", err)
	}

	list, err := svc.TemplateList(context.Background())
	if err != nil {
		t.Fatalf("TemplateList: %v", err)
	}
	if len(list) != 2 || list[0].Name != "author" || list[1].Name != "release" {
		t.Fatalf("unexpected template list: %+v", list)
	}

	got, err := svc.TemplateShow(context.Background(), "release")
	if err != nil {
		t.Fatalf("TemplateShow: %v", err)
	}
	if got.Metadata.Title == nil || *got.Metadata.Title != title {
		t.Fatalf("unexpected shown template: %+v", got)
	}

	if err := svc.TemplateDelete(context.Background(), "release"); err != nil {
		t.Fatalf("TemplateDelete: %v", err)
	}
	_, err = svc.TemplateShow(context.Background(), "release")
	assertAppErrorCode(t, err, model.ErrNotFound)
}

func TestSetNormalizesDateInNonStrictMode(t *testing.T) {
	t.Parallel()

	svc := newTestService(t)
	in := copyFixture(t, "minimal.pdf")
	out := filepath.Join(t.TempDir(), "dated.pdf")

	creationDate := "2026/05/12 15:04:05"
	res, err := svc.Set(context.Background(), model.SetRequest{
		IO: model.IOOptions{
			InputPath:  in,
			OutputPath: out,
		},
		Changes: model.MetadataPatch{
			CreationDate: &creationDate,
		},
	})
	if err != nil {
		t.Fatalf("Set: %v", err)
	}
	if !res.Normalized {
		t.Fatalf("expected normalized result")
	}
	if want := "2026-05-12T15:04:05Z"; res.Metadata.CreationDate != want {
		t.Fatalf("creation date = %q, want %q", res.Metadata.CreationDate, want)
	}
}

func TestSetRejectsInvalidDateInStrictMode(t *testing.T) {
	t.Parallel()

	metadataStore := &recordingMetadataStore{}
	svc := NewService(ServiceConfig{
		MetadataStore: metadataStore,
		TemplateStore: &recordingTemplateStore{},
	})

	creationDate := "2026/05/12"
	_, err := svc.Set(context.Background(), model.SetRequest{
		IO: model.IOOptions{
			InputPath:  "in.pdf",
			OutputPath: "out.pdf",
		},
		Exec: model.ExecOptions{Strict: true},
		Changes: model.MetadataPatch{
			CreationDate: &creationDate,
		},
	})
	assertAppErrorCode(t, err, model.ErrValidation)
	if metadataStore.writeCalls != 0 {
		t.Fatalf("expected strict validation to avoid metadata writes, got %d", metadataStore.writeCalls)
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

func assertAppErrorCode(t *testing.T, err error, want model.ErrorCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected %q error, got nil", want)
	}
	var appErr *model.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error is %T, want *model.AppError", err)
	}
	if appErr.Code != want {
		t.Fatalf("AppError.Code=%q want %q", appErr.Code, want)
	}
}
