package validate

import (
	"fmt"
	"strings"

	"pdfmeta/internal/model"
)

// ShowRequest validates single-file read input.
func ShowRequest(req model.ShowRequest) error {
	if strings.TrimSpace(req.InputPath) == "" {
		return validationError("input path is required")
	}
	return nil
}

// SetRequest validates write destination and metadata changes.
func SetRequest(req model.SetRequest) error {
	if err := ioOptions(req.IO); err != nil {
		return err
	}
	if err := metadataPatch(req.Changes, req.Exec.Strict); err != nil {
		return err
	}
	if !HasAnyPatchField(req.Changes) {
		return validationError("at least one metadata field must be set")
	}
	return nil
}

// UnsetRequest validates write destination and unset field selection.
func UnsetRequest(req model.UnsetRequest) error {
	if err := ioOptions(req.IO); err != nil {
		return err
	}
	if req.All && len(req.Fields) > 0 {
		return validationError("--all cannot be combined with explicit fields")
	}
	if !req.All && len(req.Fields) == 0 {
		return validationError("at least one field is required when --all is false")
	}
	_, err := NormalizeFields(req.Fields)
	if err != nil {
		return err
	}
	return nil
}

// TemplateSaveRequest validates persisted template payloads.
func TemplateSaveRequest(req model.TemplateSaveRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return validationError("template name is required")
	}
	if err := metadataPatch(req.Metadata, false); err != nil {
		return err
	}
	if !HasAnyPatchField(req.Metadata) {
		return validationError("template metadata must include at least one field")
	}
	return nil
}

// TemplateApplyRequest validates template name and write destination.
func TemplateApplyRequest(req model.TemplateApplyRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return validationError("template name is required")
	}
	return ioOptions(req.IO)
}

// HasAnyPatchField returns true when at least one metadata field is explicitly present.
func HasAnyPatchField(patch model.MetadataPatch) bool {
	return patch.Title != nil || patch.Author != nil || patch.Subject != nil || patch.Keywords != nil || patch.Creator != nil || patch.Producer != nil || patch.CreationDate != nil || patch.ModDate != nil
}

func ioOptions(io model.IOOptions) error {
	if strings.TrimSpace(io.InputPath) == "" {
		return validationError("input path is required")
	}
	out := strings.TrimSpace(io.OutputPath)
	if out == "" && !io.InPlace {
		return validationError("either output path or in-place mode is required")
	}
	if out != "" && io.InPlace {
		return validationError("output path and in-place mode are mutually exclusive")
	}
	return nil
}

func metadataPatch(patch model.MetadataPatch, strict bool) error {
	if patch.CreationDate != nil {
		if err := dateValue(*patch.CreationDate, strict); err != nil {
			return validationError("creation-date %v", err)
		}
	}
	if patch.ModDate != nil {
		if err := dateValue(*patch.ModDate, strict); err != nil {
			return validationError("mod-date %v", err)
		}
	}
	return nil
}

func dateValue(v string, strict bool) error {
	if strings.TrimSpace(v) == "" {
		return fmt.Errorf("must not be empty")
	}
	if strict {
		return DateString(v)
	}
	return nil
}

func validationError(format string, args ...any) error {
	return &model.AppError{
		Code:    model.ErrValidation,
		Message: fmt.Sprintf(format, args...),
	}
}
