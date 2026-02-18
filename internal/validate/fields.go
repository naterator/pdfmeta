package validate

import (
	"pdfmeta/internal/model"
)

// NormalizeFields validates and canonicalizes field selections.
func NormalizeFields(fields []model.Field) ([]model.Field, error) {
	if len(fields) == 0 {
		return nil, nil
	}

	seen := make(map[model.Field]struct{}, len(fields))
	for _, field := range fields {
		if !isKnownField(field) {
			return nil, validationError("unknown field %q", field)
		}
		if _, ok := seen[field]; ok {
			return nil, validationError("duplicate field %q", field)
		}
		seen[field] = struct{}{}
	}

	result := make([]model.Field, 0, len(fields))
	for _, field := range model.AllFields {
		if _, ok := seen[field]; ok {
			result = append(result, field)
		}
	}
	return result, nil
}

func isKnownField(f model.Field) bool {
	for _, known := range model.AllFields {
		if f == known {
			return true
		}
	}
	return false
}

// MustField converts a string-like token to a known model.Field.
func MustField(name string) (model.Field, error) {
	field := model.Field(name)
	if !isKnownField(field) {
		return "", validationError("unknown field %q", field)
	}
	return field, nil
}
