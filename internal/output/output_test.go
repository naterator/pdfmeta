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
