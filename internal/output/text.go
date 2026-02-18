package output

import (
	"fmt"
	"strings"

	"pdfmeta/internal/model"
)

type textFormatter struct{}

func (textFormatter) Show(result model.ShowResult) ([]byte, error) {
	lines := []string{
		fmt.Sprintf("Input: %s", result.InputPath),
		fmt.Sprintf("Encrypted: %t", result.Encrypted),
		fmt.Sprintf("InfoPresent: %t", result.InfoFound),
		fmt.Sprintf("XMPPresent: %t", result.XMPFound),
		fmt.Sprintf("Normalized: %t", result.Normalized),
		"Metadata:",
		fmt.Sprintf("  Title: %s", result.Metadata.Title),
		fmt.Sprintf("  Author: %s", result.Metadata.Author),
		fmt.Sprintf("  Subject: %s", result.Metadata.Subject),
		fmt.Sprintf("  Keywords: %s", result.Metadata.Keywords),
		fmt.Sprintf("  Creator: %s", result.Metadata.Creator),
		fmt.Sprintf("  Producer: %s", result.Metadata.Producer),
		fmt.Sprintf("  CreationDate: %s", result.Metadata.CreationDate),
		fmt.Sprintf("  ModDate: %s", result.Metadata.ModDate),
	}
	return []byte(strings.Join(lines, "\n") + "\n"), nil
}

func (textFormatter) Batch(result model.BatchResult) ([]byte, error) {
	lines := []string{
		fmt.Sprintf("Total: %d", result.Total),
		fmt.Sprintf("Succeeded: %d", result.Succeeded),
		fmt.Sprintf("Failed: %d", result.Failed),
		"Items:",
	}
	for _, item := range result.Items {
		line := fmt.Sprintf("  - %s [%s]", item.InputPath, item.Status)
		if item.Error != "" {
			line += ": " + item.Error
		}
		if item.OutputPath != "" {
			line += fmt.Sprintf(" -> %s", item.OutputPath)
		}
		lines = append(lines, line)
	}
	return []byte(strings.Join(lines, "\n") + "\n"), nil
}

func (textFormatter) Template(record model.TemplateRecord) ([]byte, error) {
	lines := []string{
		fmt.Sprintf("Name: %s", record.Name),
		fmt.Sprintf("Note: %s", record.Note),
		"Metadata:",
	}
	if record.Metadata.Title != nil {
		lines = append(lines, fmt.Sprintf("  Title: %s", *record.Metadata.Title))
	}
	if record.Metadata.Author != nil {
		lines = append(lines, fmt.Sprintf("  Author: %s", *record.Metadata.Author))
	}
	if record.Metadata.Subject != nil {
		lines = append(lines, fmt.Sprintf("  Subject: %s", *record.Metadata.Subject))
	}
	if record.Metadata.Keywords != nil {
		lines = append(lines, fmt.Sprintf("  Keywords: %s", *record.Metadata.Keywords))
	}
	if record.Metadata.Creator != nil {
		lines = append(lines, fmt.Sprintf("  Creator: %s", *record.Metadata.Creator))
	}
	if record.Metadata.Producer != nil {
		lines = append(lines, fmt.Sprintf("  Producer: %s", *record.Metadata.Producer))
	}
	if record.Metadata.CreationDate != nil {
		lines = append(lines, fmt.Sprintf("  CreationDate: %s", *record.Metadata.CreationDate))
	}
	if record.Metadata.ModDate != nil {
		lines = append(lines, fmt.Sprintf("  ModDate: %s", *record.Metadata.ModDate))
	}
	return []byte(strings.Join(lines, "\n") + "\n"), nil
}

func (t textFormatter) TemplateList(records []model.TemplateRecord) ([]byte, error) {
	if len(records) == 0 {
		return []byte("No templates found\n"), nil
	}
	lines := make([]string, 0, len(records))
	for _, record := range records {
		lines = append(lines, fmt.Sprintf("%s\t%s", record.Name, record.Note))
	}
	return []byte(strings.Join(lines, "\n") + "\n"), nil
}

func (textFormatter) Err(err error) ([]byte, error) {
	if ae, ok := err.(*model.AppError); ok {
		return []byte(fmt.Sprintf("error[%s]: %s\n", ae.Code, ae.Error())), nil
	}
	return []byte("error: " + err.Error() + "\n"), nil
}
