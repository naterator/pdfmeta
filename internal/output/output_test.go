package output

import (
	"errors"
	"strings"
	"testing"

	"pdfmeta/internal/model"
)

func TestParseFormat(t *testing.T) {
	t.Parallel()
	if got := ParseFormat(false); got != FormatText {
		t.Fatalf("ParseFormat(false) = %q, want %q", got, FormatText)
	}
	if got := ParseFormat(true); got != FormatJSON {
		t.Fatalf("ParseFormat(true) = %q, want %q", got, FormatJSON)
	}
}

func TestNewFormatter(t *testing.T) {
	t.Parallel()
	if _, err := NewFormatter(FormatText); err != nil {
		t.Fatalf("NewFormatter(text) err = %v", err)
	}
	if _, err := NewFormatter(FormatJSON); err != nil {
		t.Fatalf("NewFormatter(json) err = %v", err)
	}
	if _, err := NewFormatter("xml"); err == nil {
		t.Fatalf("NewFormatter(xml) expected error")
	}
}

func TestJSONFormatterShow(t *testing.T) {
	t.Parallel()
	f, _ := NewFormatter(FormatJSON)
	out, err := f.Show(model.ShowResult{InputPath: "in.pdf", Encrypted: false})
	if err != nil {
		t.Fatalf("Show error: %v", err)
	}
	got := string(out)
	if !strings.Contains(got, `"inputPath": "in.pdf"`) {
		t.Fatalf("Show output missing inputPath: %q", got)
	}
}

func TestTextFormatterTemplateList(t *testing.T) {
	t.Parallel()
	f, _ := NewFormatter(FormatText)
	out, err := f.TemplateList([]model.TemplateRecord{{Name: "release", Note: "v1"}})
	if err != nil {
		t.Fatalf("TemplateList error: %v", err)
	}
	if got := string(out); !strings.Contains(got, "release\tv1") {
		t.Fatalf("TemplateList output mismatch: %q", got)
	}

	emptyOut, err := f.TemplateList(nil)
	if err != nil {
		t.Fatalf("TemplateList empty error: %v", err)
	}
	if got := string(emptyOut); got != "No templates found\n" {
		t.Fatalf("TemplateList empty output mismatch: %q", got)
	}
}

func TestFormatterErr(t *testing.T) {
	t.Parallel()
	text, _ := NewFormatter(FormatText)
	json, _ := NewFormatter(FormatJSON)

	appErr := &model.AppError{Code: model.ErrValidation, Message: "bad request"}
	textOut, err := text.Err(appErr)
	if err != nil {
		t.Fatalf("text Err error: %v", err)
	}
	if got := string(textOut); !strings.Contains(got, "error[validation]: bad request") {
		t.Fatalf("text Err output mismatch: %q", got)
	}

	jsonOut, err := json.Err(appErr)
	if err != nil {
		t.Fatalf("json Err error: %v", err)
	}
	if got := string(jsonOut); !strings.Contains(got, `"code": "validation"`) {
		t.Fatalf("json Err output mismatch: %q", got)
	}

	plainOut, err := json.Err(errors.New("boom"))
	if err != nil {
		t.Fatalf("json Err plain error: %v", err)
	}
	if got := string(plainOut); !strings.Contains(got, `"error": "boom"`) {
		t.Fatalf("json Err plain output mismatch: %q", got)
	}
}

func TestTextFormatterShow(t *testing.T) {
	t.Parallel()
	f, _ := NewFormatter(FormatText)
	result := model.ShowResult{
		InputPath: "test.pdf",
		Encrypted: true,
		InfoFound: true,
		XMPFound:  false,
		Metadata: model.Metadata{
			Title:  "My Title",
			Author: "John",
		},
	}
	out, err := f.Show(result)
	if err != nil {
		t.Fatalf("Show error: %v", err)
	}
	got := string(out)
	for _, want := range []string{"Input: test.pdf", "Encrypted: true", "Title: My Title", "Author: John"} {
		if !strings.Contains(got, want) {
			t.Errorf("Show output missing %q:\n%s", want, got)
		}
	}
}

func TestTextFormatterBatch(t *testing.T) {
	t.Parallel()
	f, _ := NewFormatter(FormatText)
	result := model.BatchResult{
		Total:     2,
		Succeeded: 1,
		Failed:    1,
		Items: []model.BatchItemResult{
			{InputPath: "a.pdf", Status: "ok", OutputPath: "out.pdf"},
			{InputPath: "b.pdf", Status: "error", Error: "file not found"},
		},
	}
	out, err := f.Batch(result)
	if err != nil {
		t.Fatalf("Batch error: %v", err)
	}
	got := string(out)
	for _, want := range []string{"Total: 2", "Succeeded: 1", "Failed: 1", "a.pdf [ok]", "-> out.pdf", "b.pdf [error]: file not found"} {
		if !strings.Contains(got, want) {
			t.Errorf("Batch output missing %q:\n%s", want, got)
		}
	}
}

func TestTextFormatterTemplate(t *testing.T) {
	t.Parallel()
	f, _ := NewFormatter(FormatText)
	title := "v1"
	author := "me"
	record := model.TemplateRecord{
		Name: "release",
		Note: "Release template",
		Metadata: model.MetadataPatch{
			Title:  &title,
			Author: &author,
		},
	}
	out, err := f.Template(record)
	if err != nil {
		t.Fatalf("Template error: %v", err)
	}
	got := string(out)
	for _, want := range []string{"Name: release", "Note: Release template", "Title: v1", "Author: me"} {
		if !strings.Contains(got, want) {
			t.Errorf("Template output missing %q:\n%s", want, got)
		}
	}
}

func TestTextFormatterErr(t *testing.T) {
	t.Parallel()
	f, _ := NewFormatter(FormatText)
	out, err := f.Err(errors.New("plain error"))
	if err != nil {
		t.Fatalf("Err error: %v", err)
	}
	if got := string(out); got != "error: plain error\n" {
		t.Fatalf("Err output = %q", got)
	}
}

func TestJSONFormatterBatch(t *testing.T) {
	t.Parallel()
	f, _ := NewFormatter(FormatJSON)
	result := model.BatchResult{
		Total:     1,
		Succeeded: 1,
		Items: []model.BatchItemResult{
			{InputPath: "a.pdf", Status: "ok"},
		},
	}
	out, err := f.Batch(result)
	if err != nil {
		t.Fatalf("Batch error: %v", err)
	}
	got := string(out)
	if !strings.Contains(got, `"total": 1`) {
		t.Fatalf("Batch output missing total:\n%s", got)
	}
}

func TestJSONFormatterTemplate(t *testing.T) {
	t.Parallel()
	f, _ := NewFormatter(FormatJSON)
	record := model.TemplateRecord{Name: "test"}
	out, err := f.Template(record)
	if err != nil {
		t.Fatalf("Template error: %v", err)
	}
	if got := string(out); !strings.Contains(got, `"name": "test"`) {
		t.Fatalf("Template output missing name:\n%s", got)
	}
}

func TestJSONFormatterTemplateList(t *testing.T) {
	t.Parallel()
	f, _ := NewFormatter(FormatJSON)
	records := []model.TemplateRecord{{Name: "a"}, {Name: "b"}}
	out, err := f.TemplateList(records)
	if err != nil {
		t.Fatalf("TemplateList error: %v", err)
	}
	got := string(out)
	if !strings.Contains(got, `"name": "a"`) || !strings.Contains(got, `"name": "b"`) {
		t.Fatalf("TemplateList output:\n%s", got)
	}
}
