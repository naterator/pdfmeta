package cli

import (
	"github.com/spf13/cobra"

	"pdfmeta/internal/app"
	"pdfmeta/internal/model"
	"pdfmeta/internal/validate"
)

type showFlags struct {
	file    string
	asJSON  bool
	onlySet bool
	fields  []string
}

func newShowCmd(handlers *app.Handlers) *cobra.Command {
	f := &showFlags{}

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show metadata from a PDF",
		RunE: func(cmd *cobra.Command, args []string) error {
			fields, err := normalizeCLIFields(f.fields)
			if err != nil {
				return err
			}
			req := model.ShowRequest{
				InputPath: f.file,
				JSON:      f.asJSON,
			}
			if err := validate.ShowRequest(req); err != nil {
				return err
			}
			result, err := handlers.Show(commandContext(cmd), req)
			if err != nil {
				return err
			}
			return writeShowResult(cmd, f.asJSON, result, fields, f.onlySet)
		},
	}

	cmd.Flags().StringVarP(&f.file, "file", "f", "", "Input PDF file")
	cmd.Flags().BoolVarP(&f.asJSON, "json", "j", false, "Output JSON")
	cmd.Flags().BoolVar(&f.onlySet, "only-set", false, "Show only metadata fields that currently have values")
	cmd.Flags().StringSliceVar(&f.fields, "field", nil, "Limit output to specific metadata fields")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}
