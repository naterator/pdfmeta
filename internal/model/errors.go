package model

import (
	"errors"
	"fmt"
)

// ErrorCode classifies failures for output formatting and exit mapping.
type ErrorCode string

const (
	ErrUnknown      ErrorCode = "unknown"
	ErrUsage        ErrorCode = "usage"
	ErrValidation   ErrorCode = "validation"
	ErrNotFound     ErrorCode = "not_found"
	ErrConflict     ErrorCode = "conflict"
	ErrPDFEncrypted ErrorCode = "pdf_encrypted"
	ErrPDFMalformed ErrorCode = "pdf_malformed"
	ErrIO           ErrorCode = "io"
	ErrInternal     ErrorCode = "internal"
)

// AppError carries a stable code plus wrapped cause.
type AppError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		if e.Cause != nil {
			return fmt.Sprintf("%s: %v", e.Message, e.Cause)
		}
		return e.Message
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return string(e.Code)
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// ExitCode maps model-level errors to process exit status.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}

	var ae *AppError
	if !errors.As(err, &ae) {
		return 1
	}

	switch ae.Code {
	case ErrUsage:
		return 2
	case ErrValidation:
		return 3
	case ErrNotFound:
		return 4
	case ErrConflict:
		return 5
	case ErrPDFEncrypted:
		return 6
	case ErrPDFMalformed:
		return 7
	case ErrIO:
		return 8
	case ErrInternal:
		return 9
	default:
		return 1
	}
}
