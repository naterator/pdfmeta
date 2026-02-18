package batch

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"pdfmeta/internal/model"
)

type Operation string

const (
	OpShow          Operation = "show"
	OpSet           Operation = "set"
	OpUnset         Operation = "unset"
	OpTemplateApply Operation = "template-apply"
)

type Manifest struct {
	Items []Item `json:"items"`
}

type Item struct {
	Op       Operation           `json:"op"`
	Input    string              `json:"input"`
	Output   string              `json:"output,omitempty"`
	InPlace  bool                `json:"inPlace,omitempty"`
	Set      model.MetadataPatch `json:"set,omitempty"`
	Unset    []model.Field       `json:"unset,omitempty"`
	UnsetAll bool                `json:"unsetAll,omitempty"`
	Template string              `json:"template,omitempty"`
}

type Runner interface {
	Show(context.Context, model.ShowRequest) (model.ShowResult, error)
	Set(context.Context, model.SetRequest) (model.ShowResult, error)
	Unset(context.Context, model.UnsetRequest) (model.ShowResult, error)
	TemplateApply(context.Context, model.TemplateApplyRequest) (model.ShowResult, error)
}

type Engine struct {
	runner Runner
}

func NewEngine(r Runner) *Engine {
	return &Engine{runner: r}
}

func (e *Engine) Execute(ctx context.Context, req model.BatchRequest) (model.BatchResult, error) {
	if e == nil || e.runner == nil {
		return model.BatchResult{}, &model.AppError{
			Code:    model.ErrInternal,
			Message: "batch runner is required",
		}
	}
	manifest, err := LoadManifest(req.ManifestPath)
	if err != nil {
		return model.BatchResult{}, err
	}

	result := model.BatchResult{
		Items: make([]model.BatchItemResult, 0, len(manifest.Items)),
		Total: len(manifest.Items),
	}

	for _, item := range manifest.Items {
		if err := ctx.Err(); err != nil {
			return result, &model.AppError{
				Code:    model.ErrInternal,
				Message: "batch canceled",
				Cause:   err,
			}
		}

		entry, itemErr := e.executeItem(ctx, item, req.Strict)
		result.Items = append(result.Items, entry)
		if itemErr != nil {
			result.Failed++
			if !req.ContinueOnError {
				return result, AggregateError(result)
			}
			continue
		}
		result.Succeeded++
	}
	return result, AggregateError(result)
}

func AggregateError(result model.BatchResult) error {
	if result.Failed == 0 {
		return nil
	}
	return &model.AppError{
		Code:    model.ErrUnknown,
		Message: fmt.Sprintf("batch completed with %d failure(s)", result.Failed),
	}
}

func LoadManifest(path string) (Manifest, error) {
	if path == "" {
		return Manifest{}, &model.AppError{
			Code:    model.ErrValidation,
			Message: "manifest path is required",
		}
	}

	b, err := os.ReadFile(path)
	if err != nil {
		code := model.ErrIO
		if os.IsNotExist(err) {
			code = model.ErrNotFound
		}
		return Manifest{}, &model.AppError{
			Code:    code,
			Message: fmt.Sprintf("read manifest %q", path),
			Cause:   err,
		}
	}

	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return Manifest{}, &model.AppError{
			Code:    model.ErrValidation,
			Message: "decode manifest json",
			Cause:   err,
		}
	}
	if len(m.Items) == 0 {
		return Manifest{}, &model.AppError{
			Code:    model.ErrValidation,
			Message: "manifest must include at least one item",
		}
	}
	return m, nil
}

func (e *Engine) executeItem(ctx context.Context, item Item, strict bool) (model.BatchItemResult, error) {
	out := model.BatchItemResult{
		InputPath:  item.Input,
		OutputPath: item.Output,
	}
	if item.Input == "" {
		out.Status = "error"
		out.Error = "input is required"
		return out, &model.AppError{Code: model.ErrValidation, Message: out.Error}
	}

	var err error
	switch item.Op {
	case OpShow:
		_, err = e.runner.Show(ctx, model.ShowRequest{InputPath: item.Input})
	case OpSet:
		_, err = e.runner.Set(ctx, model.SetRequest{
			IO: model.IOOptions{
				InputPath:  item.Input,
				OutputPath: item.Output,
				InPlace:    item.InPlace,
			},
			Exec:    model.ExecOptions{Strict: strict},
			Changes: item.Set,
		})
	case OpUnset:
		_, err = e.runner.Unset(ctx, model.UnsetRequest{
			IO: model.IOOptions{
				InputPath:  item.Input,
				OutputPath: item.Output,
				InPlace:    item.InPlace,
			},
			Exec:   model.ExecOptions{Strict: strict},
			Fields: item.Unset,
			All:    item.UnsetAll,
		})
	case OpTemplateApply:
		_, err = e.runner.TemplateApply(ctx, model.TemplateApplyRequest{
			Name: item.Template,
			IO: model.IOOptions{
				InputPath:  item.Input,
				OutputPath: item.Output,
				InPlace:    item.InPlace,
			},
			Exec: model.ExecOptions{Strict: strict},
		})
	default:
		err = &model.AppError{
			Code:    model.ErrValidation,
			Message: fmt.Sprintf("unsupported op %q", item.Op),
		}
	}

	if err != nil {
		out.Status = "error"
		out.Error = err.Error()
		return out, err
	}
	out.Status = "ok"
	return out, nil
}
