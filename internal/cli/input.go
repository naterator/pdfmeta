package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"pdfmeta/internal/model"
)

func readInputBytes(cmd *cobra.Command, path, description string) ([]byte, error) {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return nil, &model.AppError{
			Code:    model.ErrValidation,
			Message: fmt.Sprintf("%s path is required", description),
		}
	}
	if trimmedPath == "-" {
		data, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return nil, &model.AppError{
				Code:    model.ErrIO,
				Message: fmt.Sprintf("read %s from stdin", description),
				Cause:   err,
			}
		}
		return data, nil
	}

	data, err := os.ReadFile(trimmedPath)
	if err != nil {
		code := model.ErrIO
		if os.IsNotExist(err) {
			code = model.ErrNotFound
		}
		return nil, &model.AppError{
			Code:    code,
			Message: fmt.Sprintf("read %s %q", description, trimmedPath),
			Cause:   err,
		}
	}
	return data, nil
}

func readMetadataPatchInput(cmd *cobra.Command, path string) (model.MetadataPatch, error) {
	if strings.TrimSpace(path) == "" {
		return model.MetadataPatch{}, nil
	}

	data, err := readInputBytes(cmd, path, "metadata json")
	if err != nil {
		return model.MetadataPatch{}, err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return model.MetadataPatch{}, &model.AppError{
			Code:    model.ErrValidation,
			Message: "metadata json input is empty",
		}
	}

	var patch model.MetadataPatch
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&patch); err != nil {
		return model.MetadataPatch{}, &model.AppError{
			Code:    model.ErrValidation,
			Message: "decode metadata json",
			Cause:   err,
		}
	}
	if err := ensureJSONEOF(dec); err != nil {
		return model.MetadataPatch{}, err
	}
	return patch, nil
}

func ensureJSONEOF(dec *json.Decoder) error {
	var extra any
	if err := dec.Decode(&extra); err == io.EOF {
		return nil
	} else if err != nil {
		return &model.AppError{
			Code:    model.ErrValidation,
			Message: "decode json input",
			Cause:   err,
		}
	}
	return &model.AppError{
		Code:    model.ErrValidation,
		Message: "decode json input: unexpected trailing content",
	}
}
