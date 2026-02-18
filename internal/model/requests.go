package model

// IOOptions controls write destination behavior.
type IOOptions struct {
	InputPath  string `json:"inputPath"`
	OutputPath string `json:"outputPath,omitempty"`
	InPlace    bool   `json:"inPlace"`
}

// ExecOptions controls validation and output format behavior.
type ExecOptions struct {
	Strict bool `json:"strict"`
	JSON   bool `json:"json"`
}

// ShowRequest reads metadata from a single PDF.
type ShowRequest struct {
	InputPath string `json:"inputPath"`
	JSON      bool   `json:"json"`
}

// ShowResult is the display model for read operations.
type ShowResult struct {
	InputPath  string   `json:"inputPath"`
	Encrypted  bool     `json:"encrypted"`
	Metadata   Metadata `json:"metadata"`
	InfoFound  bool     `json:"infoFound"`
	XMPFound   bool     `json:"xmpFound"`
	Normalized bool     `json:"normalized"`
}

// SetRequest applies partial metadata updates.
type SetRequest struct {
	IO      IOOptions     `json:"io"`
	Exec    ExecOptions   `json:"exec"`
	Changes MetadataPatch `json:"changes"`
}

// UnsetRequest removes selected metadata fields.
type UnsetRequest struct {
	IO     IOOptions   `json:"io"`
	Exec   ExecOptions `json:"exec"`
	Fields []Field     `json:"fields"`
	All    bool        `json:"all"`
}

// TemplateSaveRequest persists a reusable metadata template.
type TemplateSaveRequest struct {
	Name     string        `json:"name"`
	Note     string        `json:"note,omitempty"`
	Force    bool          `json:"force"`
	Metadata MetadataPatch `json:"metadata"`
}

// TemplateApplyRequest applies a named template to a PDF.
type TemplateApplyRequest struct {
	Name string      `json:"name"`
	IO   IOOptions   `json:"io"`
	Exec ExecOptions `json:"exec"`
}

// TemplateRecord is the persisted template model.
type TemplateRecord struct {
	Name     string        `json:"name"`
	Note     string        `json:"note,omitempty"`
	Metadata MetadataPatch `json:"metadata"`
}
