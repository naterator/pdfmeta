package output

import (
	"encoding/json"
	"fmt"

	"pdfmeta/internal/model"
)

// Format selects an output rendering strategy.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// Formatter renders service responses and errors for CLI output.
type Formatter interface {
	Show(model.ShowResult) ([]byte, error)
	Batch(model.BatchResult) ([]byte, error)
	Template(model.TemplateRecord) ([]byte, error)
	TemplateList([]model.TemplateRecord) ([]byte, error)
	Err(error) ([]byte, error)
}

// NewFormatter returns a formatter for the requested format.
func NewFormatter(format Format) (Formatter, error) {
	switch format {
	case FormatText:
		return textFormatter{}, nil
	case FormatJSON:
		return jsonFormatter{}, nil
	default:
		return nil, fmt.Errorf("unknown output format %q", format)
	}
}

// ParseFormat derives a rendering format from the --json toggle.
func ParseFormat(asJSON bool) Format {
	if asJSON {
		return FormatJSON
	}
	return FormatText
}

func jsonBytes(v any) ([]byte, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}
