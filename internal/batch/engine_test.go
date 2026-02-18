package batch

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"pdfmeta/internal/model"
)

type fakeRunner struct {
	failInputs map[string]error
	calls      []string
}

func (f *fakeRunner) Show(_ context.Context, req model.ShowRequest) (model.ShowResult, error) {
	f.calls = append(f.calls, "show:"+req.InputPath)
	if err := f.failInputs[req.InputPath]; err != nil {
		return model.ShowResult{}, err
	}
	return model.ShowResult{InputPath: req.InputPath}, nil
}

func (f *fakeRunner) Set(_ context.Context, req model.SetRequest) (model.ShowResult, error) {
	f.calls = append(f.calls, "set:"+req.IO.InputPath)
	if err := f.failInputs[req.IO.InputPath]; err != nil {
		return model.ShowResult{}, err
	}
	return model.ShowResult{InputPath: req.IO.InputPath}, nil
}

func (f *fakeRunner) Unset(_ context.Context, req model.UnsetRequest) (model.ShowResult, error) {
	f.calls = append(f.calls, "unset:"+req.IO.InputPath)
	if err := f.failInputs[req.IO.InputPath]; err != nil {
		return model.ShowResult{}, err
	}
	return model.ShowResult{InputPath: req.IO.InputPath}, nil
}

func (f *fakeRunner) TemplateApply(_ context.Context, req model.TemplateApplyRequest) (model.ShowResult, error) {
	f.calls = append(f.calls, "template-apply:"+req.IO.InputPath)
	if err := f.failInputs[req.IO.InputPath]; err != nil {
		return model.ShowResult{}, err
	}
	return model.ShowResult{InputPath: req.IO.InputPath}, nil
}

func TestExecuteContinueOnError(t *testing.T) {
	manifestPath := writeManifest(t, Manifest{
		Items: []Item{
			{Op: OpShow, Input: "a.pdf"},
			{Op: OpSet, Input: "b.pdf"},
			{Op: OpUnset, Input: "c.pdf"},
		},
	})
	r := &fakeRunner{
		failInputs: map[string]error{"b.pdf": errors.New("boom")},
	}
	e := NewEngine(r)

	res, err := e.Execute(context.Background(), model.BatchRequest{
		ManifestPath:    manifestPath,
		ContinueOnError: true,
	})
	if err == nil {
		t.Fatalf("expected aggregate error")
	}
	if res.Total != 3 || res.Succeeded != 2 || res.Failed != 1 {
		t.Fatalf("unexpected aggregate: %+v", res)
	}
	if len(r.calls) != 3 {
		t.Fatalf("expected all items processed, got calls=%v", r.calls)
	}
}

func TestExecuteStopOnFirstError(t *testing.T) {
	manifestPath := writeManifest(t, Manifest{
		Items: []Item{
			{Op: OpShow, Input: "a.pdf"},
			{Op: OpSet, Input: "b.pdf"},
			{Op: OpUnset, Input: "c.pdf"},
		},
	})
	r := &fakeRunner{
		failInputs: map[string]error{"b.pdf": errors.New("boom")},
	}
	e := NewEngine(r)

	res, err := e.Execute(context.Background(), model.BatchRequest{
		ManifestPath:    manifestPath,
		ContinueOnError: false,
	})
	if err == nil {
		t.Fatalf("expected aggregate error")
	}
	if res.Total != 3 || res.Succeeded != 1 || res.Failed != 1 {
		t.Fatalf("unexpected aggregate: %+v", res)
	}
	if len(r.calls) != 2 {
		t.Fatalf("expected processing to stop after failure, calls=%v", r.calls)
	}
}

func TestLoadManifestValidation(t *testing.T) {
	if _, err := LoadManifest(""); err == nil {
		t.Fatalf("expected error for empty path")
	}

	_, err := LoadManifest(filepath.Join(t.TempDir(), "missing.json"))
	assertCode(t, err, model.ErrNotFound)

	bad := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(bad, []byte("{not-json"), 0o644); err != nil {
		t.Fatalf("write bad: %v", err)
	}
	_, err = LoadManifest(bad)
	assertCode(t, err, model.ErrValidation)

	empty := writeManifestBytes(t, []byte(`{"items":[]}`))
	_, err = LoadManifest(empty)
	assertCode(t, err, model.ErrValidation)
}

func TestExecuteUnsupportedOp(t *testing.T) {
	manifestPath := writeManifest(t, Manifest{
		Items: []Item{{Op: "unknown", Input: "in.pdf"}},
	})
	e := NewEngine(&fakeRunner{})
	res, err := e.Execute(context.Background(), model.BatchRequest{ManifestPath: manifestPath, ContinueOnError: true})
	if err == nil {
		t.Fatalf("expected aggregate error")
	}
	if res.Failed != 1 || len(res.Items) != 1 || res.Items[0].Status != "error" {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestExecuteCanceledContext(t *testing.T) {
	manifestPath := writeManifest(t, Manifest{
		Items: []Item{{Op: OpShow, Input: "a.pdf"}},
	})
	e := NewEngine(&fakeRunner{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := e.Execute(ctx, model.BatchRequest{ManifestPath: manifestPath})
	assertCode(t, err, model.ErrInternal)
}

func writeManifest(t *testing.T, m Manifest) string {
	t.Helper()
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	return writeManifestBytes(t, b)
}

func writeManifestBytes(t *testing.T, b []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	return path
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
