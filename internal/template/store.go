package template

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"pdfmeta/internal/filesafe"
	"pdfmeta/internal/model"
)

type FileStore struct {
	path string
}

type fileState struct {
	Templates []model.TemplateRecord `json:"templates"`
}

func NewFileStore(path string) *FileStore {
	return &FileStore{path: path}
}

var _ model.TemplateStore = (*FileStore)(nil)

func (s *FileStore) Save(ctx context.Context, rec model.TemplateRecord, force bool) (model.TemplateRecord, error) {
	if err := ctxErr(ctx); err != nil {
		return model.TemplateRecord{}, err
	}
	name := strings.TrimSpace(rec.Name)
	if name == "" {
		return model.TemplateRecord{}, &model.AppError{
			Code:    model.ErrValidation,
			Message: "template name is required",
		}
	}
	rec.Name = name

	state, err := s.load()
	if err != nil {
		return model.TemplateRecord{}, err
	}

	for i := range state.Templates {
		if state.Templates[i].Name == rec.Name {
			if !force {
				return model.TemplateRecord{}, &model.AppError{
					Code:    model.ErrConflict,
					Message: fmt.Sprintf("template %q already exists", rec.Name),
				}
			}
			state.Templates[i] = rec
			if err := s.store(state); err != nil {
				return model.TemplateRecord{}, err
			}
			return rec, nil
		}
	}

	state.Templates = append(state.Templates, rec)
	sortTemplates(state.Templates)
	if err := s.store(state); err != nil {
		return model.TemplateRecord{}, err
	}
	return rec, nil
}

func (s *FileStore) Get(ctx context.Context, name string) (model.TemplateRecord, error) {
	if err := ctxErr(ctx); err != nil {
		return model.TemplateRecord{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return model.TemplateRecord{}, &model.AppError{
			Code:    model.ErrValidation,
			Message: "template name is required",
		}
	}

	state, err := s.load()
	if err != nil {
		return model.TemplateRecord{}, err
	}
	for _, rec := range state.Templates {
		if rec.Name == name {
			return rec, nil
		}
	}
	return model.TemplateRecord{}, &model.AppError{
		Code:    model.ErrNotFound,
		Message: fmt.Sprintf("template %q not found", name),
	}
}

func (s *FileStore) List(ctx context.Context) ([]model.TemplateRecord, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, err
	}
	state, err := s.load()
	if err != nil {
		return nil, err
	}
	out := append([]model.TemplateRecord(nil), state.Templates...)
	sortTemplates(out)
	return out, nil
}

func (s *FileStore) Delete(ctx context.Context, name string) error {
	if err := ctxErr(ctx); err != nil {
		return err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return &model.AppError{
			Code:    model.ErrValidation,
			Message: "template name is required",
		}
	}

	state, err := s.load()
	if err != nil {
		return err
	}
	next := make([]model.TemplateRecord, 0, len(state.Templates))
	found := false
	for _, rec := range state.Templates {
		if rec.Name == name {
			found = true
			continue
		}
		next = append(next, rec)
	}
	if !found {
		return &model.AppError{
			Code:    model.ErrNotFound,
			Message: fmt.Sprintf("template %q not found", name),
		}
	}

	state.Templates = next
	sortTemplates(state.Templates)
	return s.store(state)
}

func (s *FileStore) load() (fileState, error) {
	path, err := s.statePath()
	if err != nil {
		return fileState{}, err
	}

	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fileState{}, nil
		}
		return fileState{}, &model.AppError{
			Code:    model.ErrIO,
			Message: fmt.Sprintf("read template store %q", path),
			Cause:   err,
		}
	}
	if len(strings.TrimSpace(string(b))) == 0 {
		return fileState{}, nil
	}

	var state fileState
	if err := json.Unmarshal(b, &state); err != nil {
		return fileState{}, &model.AppError{
			Code:    model.ErrInternal,
			Message: "decode template store",
			Cause:   err,
		}
	}
	sortTemplates(state.Templates)
	return state, nil
}

func (s *FileStore) store(state fileState) error {
	path, err := s.statePath()
	if err != nil {
		return err
	}
	sortTemplates(state.Templates)
	b, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return &model.AppError{
			Code:    model.ErrInternal,
			Message: "encode template store",
			Cause:   err,
		}
	}
	b = append(b, '\n')
	if err := filesafe.WriteAtomic(path, b, 0o644); err != nil {
		return &model.AppError{
			Code:    model.ErrIO,
			Message: fmt.Sprintf("write template store %q", path),
			Cause:   err,
		}
	}
	return nil
}

func (s *FileStore) statePath() (string, error) {
	if strings.TrimSpace(s.path) != "" {
		return s.path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", &model.AppError{
			Code:    model.ErrInternal,
			Message: "resolve home directory",
			Cause:   err,
		}
	}
	return filepath.Join(home, ".pdfmeta", "templates.json"), nil
}

func sortTemplates(records []model.TemplateRecord) {
	sort.Slice(records, func(i, j int) bool {
		return records[i].Name < records[j].Name
	})
}

func ctxErr(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return &model.AppError{
			Code:    model.ErrInternal,
			Message: "operation canceled",
			Cause:   ctx.Err(),
		}
	default:
		return nil
	}
}
