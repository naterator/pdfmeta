package validate

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var pdfDatePattern = regexp.MustCompile(`^D:\d{4}(\d{2}(\d{2}(\d{2}(\d{2}(\d{2}([Zz]|[+-]\d{2}'?\d{2}'?)?)?)?)?)?)?$`)

// DateString accepts RFC3339 or PDF date token formats.
func DateString(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return errors.New("must not be empty")
	}
	if _, err := time.Parse(time.RFC3339, trimmed); err == nil {
		return nil
	}
	if pdfDatePattern.MatchString(trimmed) {
		return nil
	}
	return errors.New("must be RFC3339 or PDF date format")
}
