package cli

import (
	"context"

	"github.com/spf13/cobra"

	"pdfmeta/internal/app"
	"pdfmeta/internal/model"
	"pdfmeta/internal/output"
)

type batchFlags struct {
	manifest       string
	continueOnFail bool
	strict         bool
	asJSON         bool
}

func newBatchCmd(handlers *app.Handlers) *cobra.Command {
	f := &batchFlags{}

	cmd := &cobra.Command{
		Use:   "batch",
		Short: "Apply metadata operations to many PDFs",
		RunE: func(cmd *cobra.Command, args []string) error {
			req := model.BatchRequest{
				ManifestPath:    f.manifest,
				ContinueOnError: f.continueOnFail,
				Strict:          f.strict,
				JSON:            f.asJSON,
			}
			result, err := handlers.Batch(context.Background(), req)
			if err != nil {
				return err
			}
			return writeRendered(cmd, f.asJSON, func(formatter output.Formatter) ([]byte, error) {
				return formatter.Batch(result)
			})
		},
	}

	cmd.Flags().StringVar(&f.manifest, "manifest", "", "Path to batch manifest file")
	cmd.Flags().BoolVar(&f.continueOnFail, "continue-on-error", false, "Continue processing after individual file failures")
	cmd.Flags().BoolVar(&f.strict, "strict", false, "Reject invalid metadata instead of auto-correcting")
	cmd.Flags().BoolVar(&f.asJSON, "json", false, "Emit result JSON")
	_ = cmd.MarkFlagRequired("manifest")

	return cmd
}
