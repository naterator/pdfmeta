package validate

import (
	"errors"
	"testing"

	"pdfmeta/internal/model"
)

func TestDateString(t *testing.T) {
	t.Parallel()

	valid := []string{
		"2026-02-17T22:00:00Z",
		"D:20260217220000Z",
		"D:20260217220000+02'00'",
	}
	for _, value := range valid {
		if err := DateString(value); err != nil {
			t.Fatalf("DateString(%q) unexpected error: %v", value, err)
		}
	}

	invalid := []string{"", "   ", "2026/02/17", "D:20"}
	for _, value := range invalid {
		if err := DateString(value); err == nil {
			t.Fatalf("DateString(%q) expected error", value)
		}
	}
}

func TestNormalizeFields(t *testing.T) {
	t.Parallel()

	got, err := NormalizeFields([]model.Field{model.FieldAuthor, model.FieldTitle})
	if err != nil {
		t.Fatalf("NormalizeFields unexpected error: %v", err)
	}
	want := []model.Field{model.FieldTitle, model.FieldAuthor}
	if len(got) != len(want) {
		t.Fatalf("NormalizeFields length got=%d want=%d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("NormalizeFields[%d] got=%q want=%q", i, got[i], want[i])
		}
	}

	if _, err := NormalizeFields([]model.Field{"unknown"}); err == nil {
		t.Fatalf("NormalizeFields expected unknown field error")
	}
	if _, err := NormalizeFields([]model.Field{model.FieldTitle, model.FieldTitle}); err == nil {
		t.Fatalf("NormalizeFields expected duplicate error")
	}
}

func TestSetRequestValidation(t *testing.T) {
	t.Parallel()

	title := "Hello"
	ok := model.SetRequest{
		IO: model.IOOptions{InputPath: "in.pdf", OutputPath: "out.pdf"},
		Changes: model.MetadataPatch{
			Title: &title,
		},
	}
	if err := SetRequest(ok); err != nil {
		t.Fatalf("SetRequest unexpected error: %v", err)
	}

	bad := model.SetRequest{IO: model.IOOptions{InputPath: "in.pdf", InPlace: true}}
	assertValidationError(t, SetRequest(bad))

	dateLoose := "2026/02/17"
	lenient := model.SetRequest{
		IO: model.IOOptions{InputPath: "in.pdf", OutputPath: "out.pdf"},
		Changes: model.MetadataPatch{
			CreationDate: &dateLoose,
		},
	}
	if err := SetRequest(lenient); err != nil {
		t.Fatalf("SetRequest(lenient) unexpected error: %v", err)
	}

	strict := lenient
	strict.Exec.Strict = true
	assertValidationError(t, SetRequest(strict))
}

func TestUnsetRequestValidation(t *testing.T) {
	t.Parallel()

	ok := model.UnsetRequest{
		IO:     model.IOOptions{InputPath: "in.pdf", InPlace: true},
		Fields: []model.Field{model.FieldTitle},
	}
	if err := UnsetRequest(ok); err != nil {
		t.Fatalf("UnsetRequest unexpected error: %v", err)
	}

	both := model.UnsetRequest{
		IO:     model.IOOptions{InputPath: "in.pdf", InPlace: true},
		All:    true,
		Fields: []model.Field{model.FieldTitle},
	}
	assertValidationError(t, UnsetRequest(both))

	none := model.UnsetRequest{
		IO: model.IOOptions{InputPath: "in.pdf", InPlace: true},
	}
	assertValidationError(t, UnsetRequest(none))
}

func TestTemplateValidation(t *testing.T) {
	t.Parallel()

	author := "alice"
	save := model.TemplateSaveRequest{Name: "release", Metadata: model.MetadataPatch{Author: &author}}
	if err := TemplateSaveRequest(save); err != nil {
		t.Fatalf("TemplateSaveRequest unexpected error: %v", err)
	}

	empty := model.TemplateSaveRequest{Name: ""}
	assertValidationError(t, TemplateSaveRequest(empty))

	apply := model.TemplateApplyRequest{Name: "release", IO: model.IOOptions{InputPath: "in.pdf", OutputPath: "out.pdf"}}
	if err := TemplateApplyRequest(apply); err != nil {
		t.Fatalf("TemplateApplyRequest unexpected error: %v", err)
	}

	badApply := model.TemplateApplyRequest{Name: "release", IO: model.IOOptions{InputPath: "in.pdf"}}
	assertValidationError(t, TemplateApplyRequest(badApply))
}

func TestShowRequestValidation(t *testing.T) {
	t.Parallel()

	if err := ShowRequest(model.ShowRequest{InputPath: "in.pdf"}); err != nil {
		t.Fatalf("ShowRequest unexpected error: %v", err)
	}

	assertValidationError(t, ShowRequest(model.ShowRequest{}))
}

func assertValidationError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected validation error")
	}
	var appErr *model.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *model.AppError, got %T", err)
	}
	if appErr.Code != model.ErrValidation {
		t.Fatalf("expected ErrValidation code, got %q", appErr.Code)
	}
}
