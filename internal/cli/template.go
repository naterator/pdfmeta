package cli

import (
	"context"

	"github.com/spf13/cobra"

	"pdfmeta/internal/app"
	"pdfmeta/internal/model"
	"pdfmeta/internal/output"
	"pdfmeta/internal/validate"
)

type templateSaveFlags struct {
	name       string
	note       string
	force      bool
	title      string
	author     string
	subject    string
	keywords   string
	creator    string
	producer   string
	createdAt  string
	modifiedAt string
}

type templateApplyFlags struct {
	name    string
	file    string
	out     string
	inPlace bool
	strict  bool
	asJSON  bool
}

type templateListFlags struct {
	asJSON bool
}

type templateShowFlags struct {
	name   string
	asJSON bool
}

type templateDeleteFlags struct {
	name  string
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
				Metadata: patchFromTemplateSaveFlags(cmd, f),
			}
			if err := validate.TemplateSaveRequest(req); err != nil {
				return err
			}
			record, err := handlers.TemplateSave(context.Background(), req)
			if err != nil {
				return err
			}
			return writeRendered(cmd, false, func(formatter output.Formatter) ([]byte, error) {
				return formatter.Template(record)
			})
		},
	}

	cmd.Flags().StringVar(&f.name, "name", "", "Template name")
	cmd.Flags().StringVar(&f.note, "note", "", "Template description")
	cmd.Flags().BoolVar(&f.force, "force", false, "Overwrite existing template")
	cmd.Flags().StringVar(&f.title, "title", "", "Title")
	cmd.Flags().StringVar(&f.author, "author", "", "Author")
	cmd.Flags().StringVar(&f.subject, "subject", "", "Subject")
	cmd.Flags().StringVar(&f.keywords, "keywords", "", "Keywords")
	cmd.Flags().StringVar(&f.creator, "creator", "", "Creator")
	cmd.Flags().StringVar(&f.producer, "producer", "", "Producer")
	cmd.Flags().StringVar(&f.createdAt, "creation-date", "", "Creation date")
	cmd.Flags().StringVar(&f.modifiedAt, "mod-date", "", "Modification date")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func patchFromTemplateSaveFlags(cmd *cobra.Command, f *templateSaveFlags) model.MetadataPatch {
	var patch model.MetadataPatch
	if cmd.Flags().Changed("title") {
		patch.Title = &f.title
	}
	if cmd.Flags().Changed("author") {
		patch.Author = &f.author
	}
	if cmd.Flags().Changed("subject") {
		patch.Subject = &f.subject
	}
	if cmd.Flags().Changed("keywords") {
		patch.Keywords = &f.keywords
	}
	if cmd.Flags().Changed("creator") {
		patch.Creator = &f.creator
	}
	if cmd.Flags().Changed("producer") {
		patch.Producer = &f.producer
	}
	if cmd.Flags().Changed("creation-date") {
		patch.CreationDate = &f.createdAt
	}
	if cmd.Flags().Changed("mod-date") {
		patch.ModDate = &f.modifiedAt
	}
	return patch
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
			}
			if err := validate.TemplateApplyRequest(req); err != nil {
				return err
			}
			result, err := handlers.TemplateApply(context.Background(), req)
			if err != nil {
				return err
			}
			return writeRendered(cmd, f.asJSON, func(formatter output.Formatter) ([]byte, error) {
				return formatter.Show(result)
			})
		},
	}

	cmd.Flags().StringVar(&f.name, "name", "", "Template name")
	cmd.Flags().StringVar(&f.file, "file", "", "Input PDF file")
	cmd.Flags().StringVar(&f.out, "out", "", "Output PDF file")
	cmd.Flags().BoolVar(&f.inPlace, "in-place", false, "Modify file in place using safe atomic replace")
	cmd.Flags().BoolVar(&f.strict, "strict", false, "Reject invalid metadata instead of auto-correcting")
	cmd.Flags().BoolVar(&f.asJSON, "json", false, "Emit result JSON")
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
			records, err := handlers.TemplateList(context.Background())
			if err != nil {
				return err
			}
			return writeRendered(cmd, f.asJSON, func(formatter output.Formatter) ([]byte, error) {
				return formatter.TemplateList(records)
			})
		},
	}

	cmd.Flags().BoolVar(&f.asJSON, "json", false, "Emit result JSON")

	return cmd
}

func newTemplateShowCmd(handlers *app.Handlers) *cobra.Command {
	f := &templateShowFlags{}

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show template",
		RunE: func(cmd *cobra.Command, args []string) error {
			record, err := handlers.TemplateShow(context.Background(), f.name)
			if err != nil {
				return err
			}
			return writeRendered(cmd, f.asJSON, func(formatter output.Formatter) ([]byte, error) {
				return formatter.Template(record)
			})
		},
	}

	cmd.Flags().StringVar(&f.name, "name", "", "Template name")
	cmd.Flags().BoolVar(&f.asJSON, "json", false, "Emit result JSON")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newTemplateDeleteCmd(handlers *app.Handlers) *cobra.Command {
	f := &templateDeleteFlags{}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete template",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = f.force
			return handlers.TemplateDelete(context.Background(), f.name)
		},
	}

	cmd.Flags().StringVar(&f.name, "name", "", "Template name")
	cmd.Flags().BoolVar(&f.force, "force", false, "Delete without confirmation")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}
