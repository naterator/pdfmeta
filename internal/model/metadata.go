package model

// Field identifies a supported metadata key.
type Field string

const (
	FieldTitle        Field = "title"
	FieldAuthor       Field = "author"
	FieldSubject      Field = "subject"
	FieldKeywords     Field = "keywords"
	FieldCreator      Field = "creator"
	FieldProducer     Field = "producer"
	FieldCreationDate Field = "creation-date"
	FieldModDate      Field = "mod-date"
)

// AllFields defines the canonical metadata field ordering used by validation and output.
var AllFields = []Field{
	FieldTitle,
	FieldAuthor,
	FieldSubject,
	FieldKeywords,
	FieldCreator,
	FieldProducer,
	FieldCreationDate,
	FieldModDate,
}

// Metadata stores normalized Info/XMP-compatible values.
type Metadata struct {
	Title        string `json:"title,omitempty"`
	Author       string `json:"author,omitempty"`
	Subject      string `json:"subject,omitempty"`
	Keywords     string `json:"keywords,omitempty"`
	Creator      string `json:"creator,omitempty"`
	Producer     string `json:"producer,omitempty"`
	CreationDate string `json:"creationDate,omitempty"`
	ModDate      string `json:"modDate,omitempty"`
}

// MetadataPatch represents partial changes where nil means untouched.
type MetadataPatch struct {
	Title        *string `json:"title,omitempty"`
	Author       *string `json:"author,omitempty"`
	Subject      *string `json:"subject,omitempty"`
	Keywords     *string `json:"keywords,omitempty"`
	Creator      *string `json:"creator,omitempty"`
	Producer     *string `json:"producer,omitempty"`
	CreationDate *string `json:"creationDate,omitempty"`
	ModDate      *string `json:"modDate,omitempty"`
}
