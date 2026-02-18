package app

import (
	"context"

	"pdfmeta/internal/model"
)

// Handlers provides app-level entrypoints used by CLI commands.
type Handlers struct {
	svc model.Service
}

// NewHandlers creates a handler facade around a metadata service.
// If svc is nil, a default runtime service is used.
func NewHandlers(svc model.Service) *Handlers {
	if svc == nil {
		svc = NewService(ServiceConfig{})
	}
	return &Handlers{svc: svc}
}

func (h *Handlers) Show(ctx context.Context, req model.ShowRequest) (model.ShowResult, error) {
	return h.svc.Show(ctx, req)
}

func (h *Handlers) Set(ctx context.Context, req model.SetRequest) (model.ShowResult, error) {
	return h.svc.Set(ctx, req)
}

func (h *Handlers) Unset(ctx context.Context, req model.UnsetRequest) (model.ShowResult, error) {
	return h.svc.Unset(ctx, req)
}

func (h *Handlers) Batch(ctx context.Context, req model.BatchRequest) (model.BatchResult, error) {
	return h.svc.Batch(ctx, req)
}

func (h *Handlers) TemplateSave(ctx context.Context, req model.TemplateSaveRequest) (model.TemplateRecord, error) {
	return h.svc.TemplateSave(ctx, req)
}

func (h *Handlers) TemplateApply(ctx context.Context, req model.TemplateApplyRequest) (model.ShowResult, error) {
	return h.svc.TemplateApply(ctx, req)
}

func (h *Handlers) TemplateList(ctx context.Context) ([]model.TemplateRecord, error) {
	return h.svc.TemplateList(ctx)
}

func (h *Handlers) TemplateShow(ctx context.Context, name string) (model.TemplateRecord, error) {
	return h.svc.TemplateShow(ctx, name)
}

func (h *Handlers) TemplateDelete(ctx context.Context, name string) error {
	return h.svc.TemplateDelete(ctx, name)
}
