package cli

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"pdfmeta/internal/model"
)

type fakeService struct {
	showCtx          context.Context
	showReq          model.ShowRequest
	showResult       model.ShowResult
	showErr          error
	setCtx           context.Context
	setReq           model.SetRequest
	setResult        model.ShowResult
	setErr           error
	unsetCtx         context.Context
	unsetReq         model.UnsetRequest
	unsetResult      model.ShowResult
	unsetErr         error
	batchCtx         context.Context
	batchReq         model.BatchRequest
	batchResult      model.BatchResult
	batchErr         error
	templateSaveReq  model.TemplateSaveRequest
	templateSaveReqs []model.TemplateSaveRequest
	templateSaveErr  error
	templateApplyReq model.TemplateApplyRequest
	templateApplyErr error
	templateShowName string
	templateDelName  string
	templateListHit  bool
	templateListData []model.TemplateRecord
}

func (f *fakeService) Show(ctx context.Context, req model.ShowRequest) (model.ShowResult, error) {
	f.showCtx = ctx
	f.showReq = req
	if f.showErr != nil {
		return model.ShowResult{}, f.showErr
	}
	if f.showResult.InputPath != "" || f.showResult.Metadata != (model.Metadata{}) || f.showResult.Encrypted || f.showResult.InfoFound || f.showResult.XMPFound || f.showResult.Normalized {
		return f.showResult, nil
	}
	return model.ShowResult{InputPath: req.InputPath}, nil
}

func (f *fakeService) Set(ctx context.Context, req model.SetRequest) (model.ShowResult, error) {
	f.setCtx = ctx
	f.setReq = req
	if f.setErr != nil {
		return model.ShowResult{}, f.setErr
	}
	if f.setResult.InputPath != "" {
		return f.setResult, nil
	}
	return model.ShowResult{InputPath: req.IO.InputPath}, nil
}

func (f *fakeService) Unset(ctx context.Context, req model.UnsetRequest) (model.ShowResult, error) {
	f.unsetCtx = ctx
	f.unsetReq = req
	if f.unsetErr != nil {
		return model.ShowResult{}, f.unsetErr
	}
	if f.unsetResult.InputPath != "" {
		return f.unsetResult, nil
	}
	return model.ShowResult{InputPath: req.IO.InputPath}, nil
}

func (f *fakeService) Batch(ctx context.Context, req model.BatchRequest) (model.BatchResult, error) {
	f.batchCtx = ctx
	f.batchReq = req
	if f.batchResult.Total != 0 || len(f.batchResult.Items) > 0 || f.batchResult.Succeeded != 0 || f.batchResult.Failed != 0 {
		return f.batchResult, f.batchErr
	}
	return model.BatchResult{Total: 1, Succeeded: 1, Items: []model.BatchItemResult{{InputPath: req.ManifestPath, Status: "ok"}}}, f.batchErr
}

func (f *fakeService) TemplateSave(_ context.Context, req model.TemplateSaveRequest) (model.TemplateRecord, error) {
	f.templateSaveReq = req
	f.templateSaveReqs = append(f.templateSaveReqs, req)
	if f.templateSaveErr != nil {
		return model.TemplateRecord{}, f.templateSaveErr
	}
	return model.TemplateRecord{Name: req.Name, Metadata: req.Metadata}, nil
}

func (f *fakeService) TemplateApply(_ context.Context, req model.TemplateApplyRequest) (model.ShowResult, error) {
	f.templateApplyReq = req
	if f.templateApplyErr != nil {
		return model.ShowResult{}, f.templateApplyErr
	}
	return model.ShowResult{InputPath: req.IO.InputPath}, nil
}

func (f *fakeService) TemplateList(_ context.Context) ([]model.TemplateRecord, error) {
	f.templateListHit = true
	if f.templateListData != nil {
		return f.templateListData, nil
	}
	return []model.TemplateRecord{{Name: "release"}}, nil
}

func (f *fakeService) TemplateShow(_ context.Context, name string) (model.TemplateRecord, error) {
	f.templateShowName = name
	return model.TemplateRecord{Name: name}, nil
}

func (f *fakeService) TemplateDelete(_ context.Context, name string) error {
	f.templateDelName = name
	return nil
}

func TestShowCommandWiresRequest(t *testing.T) {
	t.Parallel()
	svc := &fakeService{}
	cmd := NewRootCmdWithDependencies(Dependencies{Service: svc})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"show", "--file", "doc.pdf", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute show: %v", err)
	}
	if svc.showReq.InputPath != "doc.pdf" || !svc.showReq.JSON {
		t.Fatalf("unexpected show request: %+v", svc.showReq)
	}
}

func TestShowCommandUsesCommandContext(t *testing.T) {
	t.Parallel()
	svc := &fakeService{}
	cmd := NewRootCmdWithDependencies(Dependencies{Service: svc})
	ctx := context.WithValue(context.Background(), "request-id", "abc123")
	out := &bytes.Buffer{}
	cmd.SetContext(ctx)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"show", "--file", "doc.pdf"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute show: %v", err)
	}
	if got := svc.showCtx.Value("request-id"); got != "abc123" {
		t.Fatalf("show ctx value = %v", got)
	}
}

func TestShowCommandOnlySetFiltersTextOutput(t *testing.T) {
	t.Parallel()
	svc := &fakeService{
		showResult: model.ShowResult{
			InputPath: "doc.pdf",
			Metadata: model.Metadata{
				Title: "Only Title",
			},
		},
	}
	cmd := NewRootCmdWithDependencies(Dependencies{Service: svc})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"show", "--file", "doc.pdf", "--only-set"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute show: %v", err)
	}
	got := out.String()
	if !bytes.Contains(out.Bytes(), []byte("Title: Only Title")) {
		t.Fatalf("expected title in output:\n%s", got)
	}
	if bytes.Contains(out.Bytes(), []byte("Author:")) {
		t.Fatalf("did not expect empty author line in output:\n%s", got)
	}
}

func TestShowCommandFieldFilterJSONPreservesEmptyValue(t *testing.T) {
	t.Parallel()
	svc := &fakeService{
		showResult: model.ShowResult{
			InputPath: "doc.pdf",
			Metadata:  model.Metadata{},
		},
	}
	cmd := NewRootCmdWithDependencies(Dependencies{Service: svc})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"show", "--file", "doc.pdf", "--field", "author", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute show: %v", err)
	}
	if got := out.String(); !bytes.Contains(out.Bytes(), []byte(`"author": ""`)) {
		t.Fatalf("expected empty selected field in json output:\n%s", got)
	}
}

func TestSetCommandWiresRequest(t *testing.T) {
	t.Parallel()
	svc := &fakeService{}
	cmd := NewRootCmdWithDependencies(Dependencies{Service: svc})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"set", "--file", "in.pdf", "--out", "out.pdf", "--title", "new title"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute set: %v", err)
	}
	if svc.setReq.IO.InputPath != "in.pdf" || svc.setReq.IO.OutputPath != "out.pdf" || svc.setReq.IO.InPlace {
		t.Fatalf("unexpected IO options: %+v", svc.setReq.IO)
	}
	if svc.setReq.Changes.Title == nil || *svc.setReq.Changes.Title != "new title" {
		t.Fatalf("expected title patch, got %+v", svc.setReq.Changes)
	}
}

func TestSetCommandMergesJSONInputAndFlags(t *testing.T) {
	t.Parallel()
	svc := &fakeService{}
	cmd := NewRootCmdWithDependencies(Dependencies{Service: svc})
	out := &bytes.Buffer{}
	cmd.SetIn(bytes.NewBufferString(`{"author":"json author"}`))
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"set", "--file", "in.pdf", "--out", "out.pdf", "--from-json", "-", "--title", "flag title"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute set: %v", err)
	}
	if svc.setReq.Changes.Author == nil || *svc.setReq.Changes.Author != "json author" {
		t.Fatalf("expected author from json, got %+v", svc.setReq.Changes)
	}
	if svc.setReq.Changes.Title == nil || *svc.setReq.Changes.Title != "flag title" {
		t.Fatalf("expected title from flag, got %+v", svc.setReq.Changes)
	}
}

func TestUnsetCommandWiresFields(t *testing.T) {
	t.Parallel()
	svc := &fakeService{}
	cmd := NewRootCmdWithDependencies(Dependencies{Service: svc})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"unset", "--file", "in.pdf", "--in-place", "--title", "--author"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute unset: %v", err)
	}
	if len(svc.unsetReq.Fields) != 2 || svc.unsetReq.Fields[0] != model.FieldTitle || svc.unsetReq.Fields[1] != model.FieldAuthor {
		t.Fatalf("unexpected unset fields: %+v", svc.unsetReq.Fields)
	}
}

func TestTemplateApplyWiresRequest(t *testing.T) {
	t.Parallel()
	svc := &fakeService{}
	cmd := NewRootCmdWithDependencies(Dependencies{Service: svc})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"template", "apply", "--name", "release", "--file", "in.pdf", "--out", "out.pdf", "--strict", "--json", "--title", "hotfix"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute template apply: %v", err)
	}
	if svc.templateApplyReq.Name != "release" {
		t.Fatalf("unexpected name: %+v", svc.templateApplyReq)
	}
	if !svc.templateApplyReq.Exec.Strict || !svc.templateApplyReq.Exec.JSON {
		t.Fatalf("unexpected exec options: %+v", svc.templateApplyReq.Exec)
	}
	if svc.templateApplyReq.IO.InputPath != "in.pdf" || svc.templateApplyReq.IO.OutputPath != "out.pdf" || svc.templateApplyReq.IO.InPlace {
		t.Fatalf("unexpected IO options: %+v", svc.templateApplyReq.IO)
	}
	if svc.templateApplyReq.Overrides.Title == nil || *svc.templateApplyReq.Overrides.Title != "hotfix" {
		t.Fatalf("unexpected template apply overrides: %+v", svc.templateApplyReq.Overrides)
	}
}

func TestBatchCommandWiresRequest(t *testing.T) {
	t.Parallel()
	svc := &fakeService{}
	cmd := NewRootCmdWithDependencies(Dependencies{Service: svc})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"batch", "--manifest", "jobs.json", "--continue-on-error", "--strict", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute batch: %v", err)
	}
	if svc.batchReq.ManifestPath != "jobs.json" {
		t.Fatalf("unexpected batch request: %+v", svc.batchReq)
	}
	if !svc.batchReq.ContinueOnError || !svc.batchReq.Strict || !svc.batchReq.JSON {
		t.Fatalf("unexpected batch options: %+v", svc.batchReq)
	}
}

func TestBatchCommandReadsManifestFromStdin(t *testing.T) {
	t.Parallel()
	svc := &fakeService{}
	cmd := NewRootCmdWithDependencies(Dependencies{Service: svc})
	out := &bytes.Buffer{}
	manifest := []byte(`{"items":[{"op":"show","input":"in.pdf"}]}`)
	cmd.SetIn(bytes.NewReader(manifest))
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"batch", "--manifest", "-"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute batch: %v", err)
	}
	if svc.batchReq.ManifestPath != "-" {
		t.Fatalf("unexpected manifest path: %+v", svc.batchReq)
	}
	if !bytes.Equal(svc.batchReq.ManifestBytes, manifest) {
		t.Fatalf("unexpected manifest bytes: %q", svc.batchReq.ManifestBytes)
	}
}

func TestBatchCommandRendersSummaryOnError(t *testing.T) {
	t.Parallel()
	svc := &fakeService{
		batchResult: model.BatchResult{
			Total:  1,
			Failed: 1,
			Items: []model.BatchItemResult{
				{InputPath: "in.pdf", Status: "error", Error: "boom"},
			},
		},
		batchErr: errors.New("boom"),
	}
	cmd := NewRootCmdWithDependencies(Dependencies{Service: svc})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"batch", "--manifest", "jobs.json"})

	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected batch error")
	}
	if got := out.String(); !bytes.Contains(out.Bytes(), []byte("Total: 1")) || !bytes.Contains(out.Bytes(), []byte("in.pdf [error]: boom")) {
		t.Fatalf("expected rendered batch summary, got:\n%s", got)
	}
}

func TestTemplateCommandsWireRequests(t *testing.T) {
	t.Parallel()
	svc := &fakeService{}
	cmd := NewRootCmdWithDependencies(Dependencies{Service: svc})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	cmd.SetArgs([]string{"template", "save", "--name", "release", "--title", "v1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute template save: %v", err)
	}
	if svc.templateSaveReq.Name != "release" || svc.templateSaveReq.Metadata.Title == nil || *svc.templateSaveReq.Metadata.Title != "v1" {
		t.Fatalf("unexpected template save request: %+v", svc.templateSaveReq)
	}

	cmd.SetArgs([]string{"template", "list"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute template list: %v", err)
	}
	if !svc.templateListHit {
		t.Fatalf("expected template list call")
	}

	cmd.SetArgs([]string{"template", "show", "--name", "release"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute template show: %v", err)
	}
	if svc.templateShowName != "release" {
		t.Fatalf("unexpected template show name: %q", svc.templateShowName)
	}

	cmd.SetArgs([]string{"template", "delete", "--name", "release"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute template delete: %v", err)
	}
	if svc.templateDelName != "release" {
		t.Fatalf("unexpected template delete name: %q", svc.templateDelName)
	}
}

func TestTemplateExportWritesJSON(t *testing.T) {
	t.Parallel()
	svc := &fakeService{
		templateListData: []model.TemplateRecord{{Name: "release"}},
	}
	cmd := NewRootCmdWithDependencies(Dependencies{Service: svc})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"template", "export"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute template export: %v", err)
	}
	if got := out.String(); !bytes.Contains(out.Bytes(), []byte(`"templates"`)) || !bytes.Contains(out.Bytes(), []byte(`"name": "release"`)) {
		t.Fatalf("unexpected export output:\n%s", got)
	}
}

func TestTemplateImportSavesRecords(t *testing.T) {
	t.Parallel()
	svc := &fakeService{}
	cmd := NewRootCmdWithDependencies(Dependencies{Service: svc})
	out := &bytes.Buffer{}
	cmd.SetIn(bytes.NewBufferString(`{"templates":[{"name":"a"},{"name":"b","note":"team"}]}`))
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"template", "import", "--file", "-", "--force"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute template import: %v", err)
	}
	if len(svc.templateSaveReqs) != 2 {
		t.Fatalf("expected 2 imported templates, got %d", len(svc.templateSaveReqs))
	}
	if !svc.templateSaveReqs[0].Force || !svc.templateSaveReqs[1].Force {
		t.Fatalf("expected force on imported saves: %+v", svc.templateSaveReqs)
	}
	if svc.templateSaveReqs[1].Note != "team" {
		t.Fatalf("expected imported note, got %+v", svc.templateSaveReqs[1])
	}
	if got := out.String(); !bytes.Contains(out.Bytes(), []byte("Imported 2 template(s)")) {
		t.Fatalf("unexpected import output:\n%s", got)
	}
}
