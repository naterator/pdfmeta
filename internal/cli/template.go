package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"pdfmeta/internal/app"
	"pdfmeta/internal/filesafe"
	"pdfmeta/internal/model"
	"pdfmeta/internal/output"
	"pdfmeta/internal/template"
	"pdfmeta/internal/validate"
)

type templateSaveFlags struct {
	name     string
	note     string
	force    bool
	metadata metadataStringFlags
}

type templateApplyFlags struct {
	name     string
	file     string
	out      string
	inPlace  bool
	strict   bool
	asJSON   bool
	metadata metadataStringFlags
}

type templateListFlags struct {
	asJSON bool
}

type templateShowFlags struct {
	name   string
	asJSON bool
}

type templateDeleteFlags struct {
	name string
}

type templateExportFlags struct {
	out string
}

type templateImportFlags struct {
	file  string
	force bool
}

func newTemplateCmd(handlers *app.Handlers) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Manage saved metadata templates",
	}
	cmd.AddCommand(newTemplateSaveCmd(handlers))
	cmd.AddCommand(newTemplateApplyCmd(handlers))
	cmd.AddCommand(newTemplateListCmd(handlers))
	cmd.AddCommand(newTemplateShowCmd(handlers))
	cmd.AddCommand(newTemplateDeleteCmd(handlers))
	cmd.AddCommand(newTemplateExportCmd(handlers))
	cmd.AddCommand(newTemplateImportCmd(handlers))
	return cmd
}

func newTemplateSaveCmd(handlers *app.Handlers) *cobra.Command {
	f := &templateSaveFlags{}

	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save a template",
		RunE: func(cmd *cobra.Command, args []string) error {
			req := model.TemplateSaveRequest{
				Name:     f.name,
				Note:     f.note,
				Force:    f.force,
				Metadata: patchFromMetadataFlags(cmd, &f.metadata),
			}
			if err := validate.TemplateSaveRequest(req); err != nil {
				return err
			}
			record, err := handlers.TemplateSave(commandContext(cmd), req)
			if err != nil {
				return err
			}
			return writeRendered(cmd, false, func(formatter output.Formatter) ([]byte, error) {
				return formatter.Template(record)
			})
		},
	}

	cmd.Flags().StringVarP(&f.name, "name", "n", "", "Template name")
	cmd.Flags().StringVar(&f.note, "note", "", "Template description")
	cmd.Flags().BoolVar(&f.force, "force", false, "Overwrite existing template")
	addMetadataPatchFlags(cmd, &f.metadata, metadataFlagUsageDefault)
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newTemplateApplyCmd(handlers *app.Handlers) *cobra.Command {
	f := &templateApplyFlags{}

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply a template",
		RunE: func(cmd *cobra.Command, args []string) error {
			req := model.TemplateApplyRequest{
				Name: f.name,
				IO: model.IOOptions{
					InputPath:  f.file,
					OutputPath: f.out,
					InPlace:    f.inPlace,
				},
				Exec: model.ExecOptions{
					Strict: f.strict,
					JSON:   f.asJSON,
				},
				Overrides: patchFromMetadataFlags(cmd, &f.metadata),
			}
			if err := validate.TemplateApplyRequest(req); err != nil {
				return err
			}
			result, err := handlers.TemplateApply(commandContext(cmd), req)
			if err != nil {
				return err
			}
			return writeRendered(cmd, f.asJSON, func(formatter output.Formatter) ([]byte, error) {
				return formatter.Show(result)
			})
		},
	}

	cmd.Flags().StringVarP(&f.name, "name", "n", "", "Template name")
	cmd.Flags().StringVarP(&f.file, "file", "f", "", "Input PDF file")
	cmd.Flags().StringVarP(&f.out, "out", "o", "", "Output PDF file")
	cmd.Flags().BoolVarP(&f.inPlace, "in-place", "i", false, "Modify file in place using safe atomic replace")
	cmd.Flags().BoolVarP(&f.strict, "strict", "s", false, "Reject invalid metadata instead of auto-correcting")
	cmd.Flags().BoolVarP(&f.asJSON, "json", "j", false, "Emit result JSON")
	addMetadataPatchFlags(cmd, &f.metadata, metadataFlagUsageOverride)
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func newTemplateListCmd(handlers *app.Handlers) *cobra.Command {
	f := &templateListFlags{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			records, err := handlers.TemplateList(commandContext(cmd))
			if err != nil {
				return err
			}
			return writeRendered(cmd, f.asJSON, func(formatter output.Formatter) ([]byte, error) {
				return formatter.TemplateList(records)
			})
		},
	}

	cmd.Flags().BoolVarP(&f.asJSON, "json", "j", false, "Emit result JSON")

	return cmd
}

func newTemplateShowCmd(handlers *app.Handlers) *cobra.Command {
	f := &templateShowFlags{}

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show template",
		RunE: func(cmd *cobra.Command, args []string) error {
			record, err := handlers.TemplateShow(commandContext(cmd), f.name)
			if err != nil {
				return err
			}
			return writeRendered(cmd, f.asJSON, func(formatter output.Formatter) ([]byte, error) {
				return formatter.Template(record)
			})
		},
	}

	cmd.Flags().StringVarP(&f.name, "name", "n", "", "Template name")
	cmd.Flags().BoolVarP(&f.asJSON, "json", "j", false, "Emit result JSON")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newTemplateDeleteCmd(handlers *app.Handlers) *cobra.Command {
	f := &templateDeleteFlags{}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete template",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.TemplateDelete(commandContext(cmd), f.name)
		},
	}

	cmd.Flags().StringVarP(&f.name, "name", "n", "", "Template name")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newTemplateExportCmd(handlers *app.Handlers) *cobra.Command {
	f := &templateExportFlags{}

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export templates as JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			records, err := handlers.TemplateList(commandContext(cmd))
			if err != nil {
				return err
			}
			payload, err := template.MarshalRecords(records)
			if err != nil {
				return &model.AppError{
					Code:    model.ErrInternal,
					Message: "encode template export",
					Cause:   err,
				}
			}
			if f.out == "" || f.out == "-" {
				_, err = cmd.OutOrStdout().Write(payload)
				return err
			}
			if err := filesafe.WriteAtomic(f.out, payload, 0o644); err != nil {
				return &model.AppError{
					Code:    model.ErrIO,
					Message: fmt.Sprintf("write template export %q", f.out),
					Cause:   err,
				}
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Exported %d template(s) to %s\n", len(records), f.out)
			return err
		},
	}

	cmd.Flags().StringVarP(&f.out, "out", "o", "", "Write export JSON to a file instead of stdout")
	return cmd
}

func newTemplateImportCmd(handlers *app.Handlers) *cobra.Command {
	f := &templateImportFlags{}

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import templates from JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := readInputBytes(cmd, f.file, "template import file")
			if err != nil {
				return err
			}
			records, err := template.UnmarshalRecords(data)
			if err != nil {
				return &model.AppError{
					Code:    model.ErrValidation,
					Message: "decode template import data",
					Cause:   err,
				}
			}
			for _, record := range records {
				if _, err := handlers.TemplateSave(commandContext(cmd), model.TemplateSaveRequest{
					Name:     record.Name,
					Note:     record.Note,
					Force:    f.force,
					Metadata: record.Metadata,
				}); err != nil {
					return err
				}
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Imported %d template(s)\n", len(records))
			return err
		},
	}

	cmd.Flags().StringVarP(&f.file, "file", "f", "", "Template import JSON file or - for stdin")
	cmd.Flags().BoolVar(&f.force, "force", false, "Overwrite existing templates with matching names")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}
