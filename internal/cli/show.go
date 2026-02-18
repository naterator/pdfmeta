package cli

import (
	"context"

	"github.com/spf13/cobra"

	"pdfmeta/internal/app"
	"pdfmeta/internal/model"
	"pdfmeta/internal/output"
	"pdfmeta/internal/validate"
)

type showFlags struct {
	file   string
	asJSON bool
}

func newShowCmd(handlers *app.Handlers) *cobra.Command {
	f := &showFlags{}

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show metadata from a PDF",
		RunE: func(cmd *cobra.Command, args []string) error {
			req := model.ShowRequest{
				InputPath: f.file,
				JSON:      f.asJSON,
			}
			if err := validate.ShowRequest(req); err != nil {
				return err
			}
			result, err := handlers.Show(context.Background(), req)
			if err != nil {
				return err
			}
			return writeRendered(cmd, f.asJSON, func(formatter output.Formatter) ([]byte, error) {
				return formatter.Show(result)
			})
		},
	}

	cmd.Flags().StringVar(&f.file, "file", "", "Input PDF file")
	cmd.Flags().BoolVar(&f.asJSON, "json", false, "Output JSON")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}
