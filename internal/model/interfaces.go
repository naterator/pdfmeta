package model

import "context"

// Service defines application-level metadata operations.
type Service interface {
	Show(context.Context, ShowRequest) (ShowResult, error)
	Set(context.Context, SetRequest) (ShowResult, error)
	Unset(context.Context, UnsetRequest) (ShowResult, error)
	Batch(context.Context, BatchRequest) (BatchResult, error)
	TemplateSave(context.Context, TemplateSaveRequest) (TemplateRecord, error)
	TemplateApply(context.Context, TemplateApplyRequest) (ShowResult, error)
	TemplateList(context.Context) ([]TemplateRecord, error)
	TemplateShow(context.Context, string) (TemplateRecord, error)
	TemplateDelete(context.Context, string) error
}

// MetadataStore handles PDF-backed metadata read/write.
type MetadataStore interface {
	Read(context.Context, string) (MetadataReadResult, error)
	Write(context.Context, MetadataWriteRequest) (MetadataReadResult, error)
}

// TemplateStore handles persistent template management.
type TemplateStore interface {
	Save(context.Context, TemplateRecord, bool) (TemplateRecord, error)
	Get(context.Context, string) (TemplateRecord, error)
	List(context.Context) ([]TemplateRecord, error)
	Delete(context.Context, string) error
}

// MetadataReadResult captures read state from Info/XMP sections.
type MetadataReadResult struct {
	Encrypted  bool
	Metadata   Metadata
	InfoFound  bool
	XMPFound   bool
	Normalized bool
}

// MetadataWriteRequest drives unified metadata writes to Info/XMP.
type MetadataWriteRequest struct {
	InputPath  string
	OutputPath string
	InPlace    bool
	Strict     bool
	Set        MetadataPatch
	Unset      []Field
	UnsetAll   bool
}

// BatchRequest coordinates operation execution across many files.
type BatchRequest struct {
	ManifestPath    string
	ContinueOnError bool
	Strict          bool
	JSON            bool
}

// BatchItemResult reports a single file outcome from batch processing.
type BatchItemResult struct {
	InputPath  string `json:"inputPath"`
	OutputPath string `json:"outputPath,omitempty"`
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
}

// BatchResult aggregates all item outcomes and final status.
type BatchResult struct {
	Items     []BatchItemResult `json:"items"`
	Total     int               `json:"total"`
	Succeeded int               `json:"succeeded"`
	Failed    int               `json:"failed"`
}
