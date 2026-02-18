package cli

import (
	"context"

	"github.com/spf13/cobra"

	"pdfmeta/internal/app"
	"pdfmeta/internal/model"
	"pdfmeta/internal/output"
	"pdfmeta/internal/validate"
)

type unsetFlags struct {
	file       string
	out        string
	inPlace    bool
	strict     bool
	asJSON     bool
	all        bool
	title      bool
	author     bool
	subject    bool
	keywords   bool
	creator    bool
	producer   bool
	createdAt  bool
	modifiedAt bool
}

func newUnsetCmd(handlers *app.Handlers) *cobra.Command {
	f := &unsetFlags{}

	cmd := &cobra.Command{
		Use:   "unset",
		Short: "Unset metadata fields",
		RunE: func(cmd *cobra.Command, args []string) error {
			fields := fieldsFromUnsetFlags(f)
			req := model.UnsetRequest{
				IO: model.IOOptions{
					InputPath:  f.file,
					OutputPath: f.out,
					InPlace:    f.inPlace,
				},
				Exec: model.ExecOptions{
					Strict: f.strict,
					JSON:   f.asJSON,
				},
				Fields: fields,
				All:    f.all,
			}
			if err := validate.UnsetRequest(req); err != nil {
				return err
			}
			result, err := handlers.Unset(context.Background(), req)
			if err != nil {
				return err
			}
			return writeRendered(cmd, f.asJSON, func(formatter output.Formatter) ([]byte, error) {
				return formatter.Show(result)
			})
		},
	}

	cmd.Flags().StringVar(&f.file, "file", "", "Input PDF file")
	cmd.Flags().StringVar(&f.out, "out", "", "Output PDF file")
	cmd.Flags().BoolVar(&f.inPlace, "in-place", false, "Modify file in place using safe atomic replace")
	cmd.Flags().BoolVar(&f.strict, "strict", false, "Reject invalid metadata instead of auto-correcting")
	cmd.Flags().BoolVar(&f.asJSON, "json", false, "Emit result JSON")

	cmd.Flags().BoolVar(&f.all, "all", false, "Unset all supported metadata fields")
	cmd.Flags().BoolVar(&f.title, "title", false, "Unset Title")
	cmd.Flags().BoolVar(&f.author, "author", false, "Unset Author")
	cmd.Flags().BoolVar(&f.subject, "subject", false, "Unset Subject")
	cmd.Flags().BoolVar(&f.keywords, "keywords", false, "Unset Keywords")
	cmd.Flags().BoolVar(&f.creator, "creator", false, "Unset Creator")
	cmd.Flags().BoolVar(&f.producer, "producer", false, "Unset Producer")
	cmd.Flags().BoolVar(&f.createdAt, "creation-date", false, "Unset Creation date")
	cmd.Flags().BoolVar(&f.modifiedAt, "mod-date", false, "Unset Modification date")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func fieldsFromUnsetFlags(f *unsetFlags) []model.Field {
	fields := make([]model.Field, 0, len(model.AllFields))
	if f.title {
		fields = append(fields, model.FieldTitle)
	}
	if f.author {
		fields = append(fields, model.FieldAuthor)
	}
	if f.subject {
		fields = append(fields, model.FieldSubject)
	}
	if f.keywords {
		fields = append(fields, model.FieldKeywords)
	}
	if f.creator {
		fields = append(fields, model.FieldCreator)
	}
	if f.producer {
		fields = append(fields, model.FieldProducer)
	}
	if f.createdAt {
		fields = append(fields, model.FieldCreationDate)
	}
	if f.modifiedAt {
		fields = append(fields, model.FieldModDate)
	}
	return fields
}
