package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"pdfmeta/internal/model"
	"pdfmeta/internal/output"
	"pdfmeta/internal/validate"
)

func normalizeCLIFields(raw []string) ([]model.Field, error) {
	fields := make([]model.Field, 0, len(raw))
	for _, name := range raw {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		field, err := validate.MustField(name)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}
	return validate.NormalizeFields(fields)
}

func writeShowResult(cmd *cobra.Command, asJSON bool, result model.ShowResult, fields []model.Field, onlySet bool) error {
	if len(fields) == 0 && !onlySet {
		return writeRendered(cmd, asJSON, func(formatter output.Formatter) ([]byte, error) {
			return formatter.Show(result)
		})
	}

	var payload []byte
	var err error
	if asJSON {
		payload, err = renderFilteredShowJSON(result, fields, onlySet)
	} else {
		payload, err = renderFilteredShowText(result, fields, onlySet)
	}
	if err != nil {
		return err
	}
	_, err = cmd.OutOrStdout().Write(payload)
	if err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}

func renderFilteredShowJSON(result model.ShowResult, fields []model.Field, onlySet bool) ([]byte, error) {
	type payload struct {
		InputPath  string            `json:"inputPath"`
		Encrypted  bool              `json:"encrypted"`
		Metadata   map[string]string `json:"metadata"`
		InfoFound  bool              `json:"infoFound"`
		XMPFound   bool              `json:"xmpFound"`
		Normalized bool              `json:"normalized"`
	}

	out, err := json.MarshalIndent(payload{
		InputPath:  result.InputPath,
		Encrypted:  result.Encrypted,
		Metadata:   metadataMap(result.Metadata, fields, onlySet),
		InfoFound:  result.InfoFound,
		XMPFound:   result.XMPFound,
		Normalized: result.Normalized,
	}, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}

func renderFilteredShowText(result model.ShowResult, fields []model.Field, onlySet bool) ([]byte, error) {
	lines := []string{
		fmt.Sprintf("Input: %s", result.InputPath),
		fmt.Sprintf("Encrypted: %t", result.Encrypted),
		fmt.Sprintf("InfoPresent: %t", result.InfoFound),
		fmt.Sprintf("XMPPresent: %t", result.XMPFound),
		fmt.Sprintf("Normalized: %t", result.Normalized),
		"Metadata:",
	}

	renderedFields := fieldsToRender(result.Metadata, fields, onlySet)
	if len(renderedFields) == 0 {
		lines = append(lines, "  (none)")
		return []byte(strings.Join(lines, "\n") + "\n"), nil
	}

	for _, field := range renderedFields {
		lines = append(lines, fmt.Sprintf("  %s: %s", metadataFieldLabel(field), model.MetadataValue(result.Metadata, field)))
	}
	return []byte(strings.Join(lines, "\n") + "\n"), nil
}

func metadataMap(meta model.Metadata, fields []model.Field, onlySet bool) map[string]string {
	renderedFields := fieldsToRender(meta, fields, onlySet)
	out := make(map[string]string, len(renderedFields))
	for _, field := range renderedFields {
		out[metadataFieldJSONKey(field)] = model.MetadataValue(meta, field)
	}
	return out
}

func fieldsToRender(meta model.Metadata, fields []model.Field, onlySet bool) []model.Field {
	selected := fields
	if len(selected) == 0 {
		selected = model.AllFields
	}

	out := make([]model.Field, 0, len(selected))
	for _, field := range selected {
		if onlySet && strings.TrimSpace(model.MetadataValue(meta, field)) == "" {
			continue
		}
		out = append(out, field)
	}
	return out
}

func metadataFieldLabel(field model.Field) string {
	switch field {
	case model.FieldTitle:
		return "Title"
	case model.FieldAuthor:
		return "Author"
	case model.FieldSubject:
		return "Subject"
	case model.FieldKeywords:
		return "Keywords"
	case model.FieldCreator:
		return "Creator"
	case model.FieldProducer:
		return "Producer"
	case model.FieldCreationDate:
		return "CreationDate"
	case model.FieldModDate:
		return "ModDate"
	default:
		return string(field)
	}
}

func metadataFieldJSONKey(field model.Field) string {
	switch field {
	case model.FieldTitle:
		return "title"
	case model.FieldAuthor:
		return "author"
	case model.FieldSubject:
		return "subject"
	case model.FieldKeywords:
		return "keywords"
	case model.FieldCreator:
		return "creator"
	case model.FieldProducer:
		return "producer"
	case model.FieldCreationDate:
		return "creationDate"
	case model.FieldModDate:
		return "modDate"
	default:
		return string(field)
	}
}
