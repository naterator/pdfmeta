package model

import "strings"

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

// MetadataValue returns the value for the requested metadata field.
func MetadataValue(meta Metadata, field Field) string {
	switch field {
	case FieldTitle:
		return meta.Title
	case FieldAuthor:
		return meta.Author
	case FieldSubject:
		return meta.Subject
	case FieldKeywords:
		return meta.Keywords
	case FieldCreator:
		return meta.Creator
	case FieldProducer:
		return meta.Producer
	case FieldCreationDate:
		return meta.CreationDate
	case FieldModDate:
		return meta.ModDate
	default:
		return ""
	}
}

// SetMetadataValue writes a single field onto a metadata struct.
func SetMetadataValue(meta *Metadata, field Field, value string) {
	if meta == nil {
		return
	}
	switch field {
	case FieldTitle:
		meta.Title = value
	case FieldAuthor:
		meta.Author = value
	case FieldSubject:
		meta.Subject = value
	case FieldKeywords:
		meta.Keywords = value
	case FieldCreator:
		meta.Creator = value
	case FieldProducer:
		meta.Producer = value
	case FieldCreationDate:
		meta.CreationDate = value
	case FieldModDate:
		meta.ModDate = value
	}
}

// FilterMetadata keeps only the selected fields, optionally dropping empty values.
func FilterMetadata(meta Metadata, fields []Field, onlySet bool) Metadata {
	selected := fields
	if len(selected) == 0 {
		selected = AllFields
	}

	var filtered Metadata
	for _, field := range selected {
		value := MetadataValue(meta, field)
		if onlySet && strings.TrimSpace(value) == "" {
			continue
		}
		SetMetadataValue(&filtered, field, value)
	}
	return filtered
}

// MergeMetadataPatch overlays override onto base, keeping untouched base fields.
func MergeMetadataPatch(base, override MetadataPatch) MetadataPatch {
	merged := base
	if override.Title != nil {
		merged.Title = override.Title
	}
	if override.Author != nil {
		merged.Author = override.Author
	}
	if override.Subject != nil {
		merged.Subject = override.Subject
	}
	if override.Keywords != nil {
		merged.Keywords = override.Keywords
	}
	if override.Creator != nil {
		merged.Creator = override.Creator
	}
	if override.Producer != nil {
		merged.Producer = override.Producer
	}
	if override.CreationDate != nil {
		merged.CreationDate = override.CreationDate
	}
	if override.ModDate != nil {
		merged.ModDate = override.ModDate
	}
	return merged
}
