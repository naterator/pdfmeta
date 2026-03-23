package cli

import (
	"context"
	"io"
	"os"

	"pdfmeta/internal/app"
	"pdfmeta/internal/model"
	"pdfmeta/internal/template"
)

// Dependencies are injectable CLI runtime dependencies.
type Dependencies struct {
	Service    model.Service
	Version    string
	Selfupdate func(context.Context, io.Writer) error
}

func (d Dependencies) withDefaults() Dependencies {
	if d.Service == nil {
		d.Service = app.NewService(app.ServiceConfig{
			TemplateStore: template.NewFileStore(os.Getenv("PDFMETA_TEMPLATE_STORE")),
		})
	}
	if d.Version == "" {
		d.Version = "unknown"
	}
	return d
}
