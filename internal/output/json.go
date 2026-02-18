package output

import "pdfmeta/internal/model"

type jsonFormatter struct{}

func (jsonFormatter) Show(result model.ShowResult) ([]byte, error) {
	return jsonBytes(result)
}

func (jsonFormatter) Batch(result model.BatchResult) ([]byte, error) {
	return jsonBytes(result)
}

func (jsonFormatter) Template(record model.TemplateRecord) ([]byte, error) {
	return jsonBytes(record)
}

func (jsonFormatter) TemplateList(records []model.TemplateRecord) ([]byte, error) {
	return jsonBytes(records)
}

func (jsonFormatter) Err(err error) ([]byte, error) {
	type payload struct {
		Error string          `json:"error"`
		Code  model.ErrorCode `json:"code,omitempty"`
	}
	if ae, ok := err.(*model.AppError); ok {
		return jsonBytes(payload{Error: ae.Error(), Code: ae.Code})
	}
	return jsonBytes(payload{Error: err.Error()})
}
