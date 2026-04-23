package template

import (
	"bytes"
	"encoding/json"
	"errors"

	"pdfmeta/internal/model"
)

// MarshalRecords encodes template records in the shared import/export format.
func MarshalRecords(records []model.TemplateRecord) ([]byte, error) {
	exported := append([]model.TemplateRecord(nil), records...)
	sortTemplates(exported)

	payload, err := json.MarshalIndent(fileState{Templates: exported}, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}

// UnmarshalRecords decodes either the shared wrapper object or a raw record array.
func UnmarshalRecords(data []byte) ([]model.TemplateRecord, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil, errors.New("template input is empty")
	}

	var records []model.TemplateRecord
	if trimmed[0] == '[' {
		if err := json.Unmarshal(trimmed, &records); err != nil {
			return nil, err
		}
		sortTemplates(records)
		return records, nil
	}

	var state fileState
	if err := json.Unmarshal(trimmed, &state); err != nil {
		return nil, err
	}
	sortTemplates(state.Templates)
	return state.Templates, nil
}
