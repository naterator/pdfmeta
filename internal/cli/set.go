package cli

import (
	"context"

	"github.com/spf13/cobra"

	"pdfmeta/internal/app"
	"pdfmeta/internal/model"
	"pdfmeta/internal/output"
	"pdfmeta/internal/validate"
)

type setFlags struct {
	file       string
	out        string
	inPlace    bool
	strict     bool
	asJSON     bool
	title      string
	author     string
	subject    string
	keywords   string
	creator    string
	producer   string
	createdAt  string
	modifiedAt string
}

func newSetCmd(handlers *app.Handlers) *cobra.Command {
	f := &setFlags{}
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set metadata fields",
		RunE: func(cmd *cobra.Command, args []string) error {
			req := model.SetRequest{
				IO: model.IOOptions{
					InputPath:  f.file,
					OutputPath: f.out,
					InPlace:    f.inPlace,
				},
				Exec: model.ExecOptions{
					Strict: f.strict,
					JSON:   f.asJSON,
				},
				Changes: patchFromSetFlags(cmd, f),
			}
			if err := validate.SetRequest(req); err != nil {
				return err
			}
			result, err := handlers.Set(context.Background(), req)
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

	cmd.Flags().StringVar(&f.title, "title", "", "Title")
	cmd.Flags().StringVar(&f.author, "author", "", "Author")
	cmd.Flags().StringVar(&f.subject, "subject", "", "Subject")
	cmd.Flags().StringVar(&f.keywords, "keywords", "", "Keywords")
	cmd.Flags().StringVar(&f.creator, "creator", "", "Creator")
	cmd.Flags().StringVar(&f.producer, "producer", "", "Producer")
	cmd.Flags().StringVar(&f.createdAt, "creation-date", "", "Creation date")
	cmd.Flags().StringVar(&f.modifiedAt, "mod-date", "", "Modification date")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func patchFromSetFlags(cmd *cobra.Command, f *setFlags) model.MetadataPatch {
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
