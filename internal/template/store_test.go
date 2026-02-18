package template

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"pdfmeta/internal/model"
)

func testStore(t *testing.T) *FileStore {
	t.Helper()
	return NewFileStore(filepath.Join(t.TempDir(), "templates.json"))
}

func strPtr(s string) *string { return &s }

func TestSaveGetListDelete(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	in := model.TemplateRecord{
		Name: "release",
		Note: "release defaults",
		Metadata: model.MetadataPatch{
			Author: strPtr("Docs Team"),
			Title:  strPtr("Release Notes"),
		},
	}
	saved, err := store.Save(ctx, in, false)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if saved.Name != in.Name {
		t.Fatalf("saved name mismatch: %q", saved.Name)
	}

	got, err := store.Get(ctx, "release")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Note != in.Note {
		t.Fatalf("note mismatch: %q", got.Note)
	}
	if got.Metadata.Author == nil || *got.Metadata.Author != "Docs Team" {
		t.Fatalf("author mismatch: %#v", got.Metadata.Author)
	}

	list, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 || list[0].Name != "release" {
		t.Fatalf("unexpected list: %#v", list)
	}

	if err := store.Delete(ctx, "release"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := store.Get(ctx, "release"); err == nil {
		t.Fatalf("expected not found after delete")
	}
}

func TestSaveConflictAndForceOverwrite(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	first := model.TemplateRecord{Name: "default", Metadata: model.MetadataPatch{Title: strPtr("v1")}}
	if _, err := store.Save(ctx, first, false); err != nil {
		t.Fatalf("Save first: %v", err)
	}

	second := model.TemplateRecord{Name: "default", Metadata: model.MetadataPatch{Title: strPtr("v2")}}
	err := mustErr(store.Save(ctx, second, false))
	assertCode(t, err, model.ErrConflict)

	if _, err := store.Save(ctx, second, true); err != nil {
		t.Fatalf("Save force: %v", err)
	}
	got, err := store.Get(ctx, "default")
	if err != nil {
		t.Fatalf("Get after force: %v", err)
	}
	if got.Metadata.Title == nil || *got.Metadata.Title != "v2" {
		t.Fatalf("expected overwritten title, got %#v", got.Metadata.Title)
	}
}

func TestListSorted(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()
	if _, err := store.Save(ctx, model.TemplateRecord{Name: "zeta"}, false); err != nil {
		t.Fatalf("Save zeta: %v", err)
	}
	if _, err := store.Save(ctx, model.TemplateRecord{Name: "alpha"}, false); err != nil {
		t.Fatalf("Save alpha: %v", err)
	}
	list, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 || list[0].Name != "alpha" || list[1].Name != "zeta" {
		t.Fatalf("unexpected order: %#v", list)
	}
}

func TestValidationAndNotFound(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	err := mustErr(store.Save(ctx, model.TemplateRecord{Name: ""}, false))
	assertCode(t, err, model.ErrValidation)

	_, err = store.Get(ctx, "")
	assertCode(t, err, model.ErrValidation)

	_, err = store.Get(ctx, "missing")
	assertCode(t, err, model.ErrNotFound)

	err = store.Delete(ctx, "missing")
	assertCode(t, err, model.ErrNotFound)
}

func TestCanceledContext(t *testing.T) {
	store := testStore(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := store.List(ctx)
	assertCode(t, err, model.ErrInternal)
}

func TestCorruptStoreFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "templates.json")
	if err := os.WriteFile(path, []byte("{not-json"), 0o644); err != nil {
		t.Fatalf("seed corrupt file: %v", err)
	}
	store := NewFileStore(path)

	_, err := store.List(context.Background())
	assertCode(t, err, model.ErrInternal)
}

func mustErr[T any](_ T, err error) error {
	return err
}

func assertCode(t *testing.T, err error, want model.ErrorCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected %q error, got nil", want)
	}
	ae, ok := err.(*model.AppError)
	if !ok {
		t.Fatalf("error is %T, want *model.AppError", err)
	}
	if ae.Code != want {
		t.Fatalf("AppError.Code=%q want %q", ae.Code, want)
	}
}
