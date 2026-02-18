package cli

import (
	"bytes"
	"context"
	"testing"

	"pdfmeta/internal/model"
)

type fakeService struct {
	showReq          model.ShowRequest
	setReq           model.SetRequest
	unsetReq         model.UnsetRequest
	batchReq         model.BatchRequest
	templateSaveReq  model.TemplateSaveRequest
	templateApplyReq model.TemplateApplyRequest
	templateShowName string
	templateDelName  string
	templateListHit  bool
}

func (f *fakeService) Show(_ context.Context, req model.ShowRequest) (model.ShowResult, error) {
	f.showReq = req
	return model.ShowResult{InputPath: req.InputPath}, nil
}

func (f *fakeService) Set(_ context.Context, req model.SetRequest) (model.ShowResult, error) {
	f.setReq = req
	return model.ShowResult{InputPath: req.IO.InputPath}, nil
}

func (f *fakeService) Unset(_ context.Context, req model.UnsetRequest) (model.ShowResult, error) {
	f.unsetReq = req
	return model.ShowResult{InputPath: req.IO.InputPath}, nil
}

func (f *fakeService) Batch(_ context.Context, req model.BatchRequest) (model.BatchResult, error) {
	f.batchReq = req
	return model.BatchResult{Total: 1, Succeeded: 1, Items: []model.BatchItemResult{{InputPath: req.ManifestPath, Status: "ok"}}}, nil
}

func (f *fakeService) TemplateSave(_ context.Context, req model.TemplateSaveRequest) (model.TemplateRecord, error) {
	f.templateSaveReq = req
	return model.TemplateRecord{Name: req.Name, Metadata: req.Metadata}, nil
}

func (f *fakeService) TemplateApply(_ context.Context, req model.TemplateApplyRequest) (model.ShowResult, error) {
	f.templateApplyReq = req
	return model.ShowResult{InputPath: req.IO.InputPath}, nil
}

func (f *fakeService) TemplateList(_ context.Context) ([]model.TemplateRecord, error) {
	f.templateListHit = true
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
	cmd.SetArgs([]string{"template", "apply", "--name", "release", "--file", "in.pdf", "--out", "out.pdf", "--strict", "--json"})

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

	cmd.SetArgs([]string{"template", "delete", "--name", "release", "--force"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute template delete: %v", err)
	}
	if svc.templateDelName != "release" {
		t.Fatalf("unexpected template delete name: %q", svc.templateDelName)
	}
}
