package app

import (
	"context"
	"strings"
	"time"

	"pdfmeta/internal/batch"
	"pdfmeta/internal/metadata"
	"pdfmeta/internal/model"
	"pdfmeta/internal/template"
	"pdfmeta/internal/validate"
)

// ServiceConfig configures the concrete service implementation.
type ServiceConfig struct {
	MetadataStore model.MetadataStore
	TemplateStore model.TemplateStore
}

// Service is the concrete runtime implementation behind CLI handlers.
type Service struct {
	metadata    model.MetadataStore
	templates   model.TemplateStore
	batchEngine *batch.Engine
}

var _ model.Service = (*Service)(nil)

// NewService creates a service with default file-backed stores when nil deps are provided.
func NewService(cfg ServiceConfig) *Service {
	svc := &Service{
		metadata:  cfg.MetadataStore,
		templates: cfg.TemplateStore,
	}
	if svc.metadata == nil {
		svc.metadata = metadata.NewStore()
	}
	if svc.templates == nil {
		svc.templates = template.NewFileStore("")
	}
	svc.batchEngine = batch.NewEngine(svc)
	return svc
}

func (s *Service) Show(ctx context.Context, req model.ShowRequest) (model.ShowResult, error) {
	rr, err := s.metadata.Read(ctx, req.InputPath)
	if err != nil {
		return model.ShowResult{}, err
	}
	meta, normalized, err := normalizeMetadata(rr.Metadata, false)
	if err != nil {
		return model.ShowResult{}, err
	}
	return model.ShowResult{
		InputPath:  req.InputPath,
		Encrypted:  rr.Encrypted,
		Metadata:   meta,
		InfoFound:  rr.InfoFound,
		XMPFound:   rr.XMPFound,
		Normalized: rr.Normalized || normalized,
	}, nil
}

func (s *Service) Set(ctx context.Context, req model.SetRequest) (model.ShowResult, error) {
	patch, err := normalizePatch(req.Changes, req.Exec.Strict)
	if err != nil {
		return model.ShowResult{}, err
	}
	rr, err := s.metadata.Write(ctx, model.MetadataWriteRequest{
		InputPath:  req.IO.InputPath,
		OutputPath: req.IO.OutputPath,
		InPlace:    req.IO.InPlace,
		Strict:     req.Exec.Strict,
		Set:        patch,
	})
	if err != nil {
		return model.ShowResult{}, err
	}
	meta, normalized, err := normalizeMetadata(rr.Metadata, req.Exec.Strict)
	if err != nil {
		return model.ShowResult{}, err
	}
	return model.ShowResult{
		InputPath:  effectiveOutputPath(req.IO),
		Encrypted:  rr.Encrypted,
		Metadata:   meta,
		InfoFound:  rr.InfoFound,
		XMPFound:   rr.XMPFound,
		Normalized: rr.Normalized || normalized,
	}, nil
}

func (s *Service) Unset(ctx context.Context, req model.UnsetRequest) (model.ShowResult, error) {
	fields, err := validate.NormalizeFields(req.Fields)
	if err != nil {
		return model.ShowResult{}, err
	}
	rr, err := s.metadata.Write(ctx, model.MetadataWriteRequest{
		InputPath:  req.IO.InputPath,
		OutputPath: req.IO.OutputPath,
		InPlace:    req.IO.InPlace,
		Strict:     req.Exec.Strict,
		Unset:      fields,
		UnsetAll:   req.All,
	})
	if err != nil {
		return model.ShowResult{}, err
	}
	meta, normalized, err := normalizeMetadata(rr.Metadata, req.Exec.Strict)
	if err != nil {
		return model.ShowResult{}, err
	}
	return model.ShowResult{
		InputPath:  effectiveOutputPath(req.IO),
		Encrypted:  rr.Encrypted,
		Metadata:   meta,
		InfoFound:  rr.InfoFound,
		XMPFound:   rr.XMPFound,
		Normalized: rr.Normalized || normalized,
	}, nil
}

func (s *Service) Batch(ctx context.Context, req model.BatchRequest) (model.BatchResult, error) {
	return s.batchEngine.Execute(ctx, req)
}

func (s *Service) TemplateSave(ctx context.Context, req model.TemplateSaveRequest) (model.TemplateRecord, error) {
	patch, err := normalizePatch(req.Metadata, false)
	if err != nil {
		return model.TemplateRecord{}, err
	}
	record := model.TemplateRecord{
		Name:     strings.TrimSpace(req.Name),
		Note:     strings.TrimSpace(req.Note),
		Metadata: patch,
	}
	return s.templates.Save(ctx, record, req.Force)
}

func (s *Service) TemplateApply(ctx context.Context, req model.TemplateApplyRequest) (model.ShowResult, error) {
	record, err := s.templates.Get(ctx, req.Name)
	if err != nil {
		return model.ShowResult{}, err
	}
	return s.Set(ctx, model.SetRequest{
		IO:      req.IO,
		Exec:    req.Exec,
		Changes: record.Metadata,
	})
}

func (s *Service) TemplateList(ctx context.Context) ([]model.TemplateRecord, error) {
	return s.templates.List(ctx)
}

func (s *Service) TemplateShow(ctx context.Context, name string) (model.TemplateRecord, error) {
	return s.templates.Get(ctx, name)
}

func (s *Service) TemplateDelete(ctx context.Context, name string) error {
	return s.templates.Delete(ctx, name)
}

func effectiveOutputPath(io model.IOOptions) string {
	if io.InPlace {
		return io.InputPath
	}
	if io.OutputPath != "" {
		return io.OutputPath
	}
	return io.InputPath
}

func normalizePatch(patch model.MetadataPatch, strict bool) (model.MetadataPatch, error) {
	var changed bool
	fix := func(v **string) {
		if *v == nil {
			return
		}
		before := **v
		after := strings.TrimSpace(before)
		if before != after {
			changed = true
		}
		*v = &after
	}
	for _, field := range []**string{
		&patch.Title,
		&patch.Author,
		&patch.Subject,
		&patch.Keywords,
		&patch.Creator,
		&patch.Producer,
	} {
		fix(field)
	}
	var err error
	patch.CreationDate, changed, err = normalizeDatePtr(patch.CreationDate, strict, changed)
	if err != nil {
		return model.MetadataPatch{}, err
	}
	patch.ModDate, changed, err = normalizeDatePtr(patch.ModDate, strict, changed)
	if err != nil {
		return model.MetadataPatch{}, err
	}
	_ = changed
	return patch, nil
}

func normalizeMetadata(meta model.Metadata, strict bool) (model.Metadata, bool, error) {
	changed := false
	normalize := func(v *string) {
		before := *v
		*v = strings.TrimSpace(*v)
		if before != *v {
			changed = true
		}
	}
	normalize(&meta.Title)
	normalize(&meta.Author)
	normalize(&meta.Subject)
	normalize(&meta.Keywords)
	normalize(&meta.Creator)
	normalize(&meta.Producer)

	next, dateChanged, err := normalizeDate(meta.CreationDate, strict)
	if err != nil {
		return model.Metadata{}, false, err
	}
	if dateChanged {
		changed = true
	}
	meta.CreationDate = next

	next, dateChanged, err = normalizeDate(meta.ModDate, strict)
	if err != nil {
		return model.Metadata{}, false, err
	}
	if dateChanged {
		changed = true
	}
	meta.ModDate = next
	return meta, changed, nil
}

func normalizeDatePtr(in *string, strict bool, changed bool) (*string, bool, error) {
	if in == nil {
		return nil, changed, nil
	}
	out, dateChanged, err := normalizeDate(*in, strict)
	if err != nil {
		return nil, changed, err
	}
	if dateChanged {
		changed = true
	}
	return &out, changed, nil
}

func normalizeDate(in string, strict bool) (string, bool, error) {
	value := strings.TrimSpace(in)
	if value == "" {
		return "", false, nil
	}
	if err := validate.DateString(value); err == nil {
		return value, value != in, nil
	}
	if strict {
		return "", false, &model.AppError{
			Code:    model.ErrValidation,
			Message: "invalid date format",
		}
	}

	layouts := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006/01/02",
		"2006/01/02 15:04:05",
	}
	for _, layout := range layouts {
		if ts, err := time.Parse(layout, value); err == nil {
			return ts.UTC().Format(time.RFC3339), true, nil
		}
	}
	return value, false, nil
}
