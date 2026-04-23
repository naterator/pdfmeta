package cli

import (
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
			var manifestBytes []byte
			var err error
			if f.manifest == "-" {
				manifestBytes, err = readInputBytes(cmd, f.manifest, "manifest")
				if err != nil {
					return err
				}
			}
			req := model.BatchRequest{
				ManifestPath:    f.manifest,
				ManifestBytes:   manifestBytes,
				ContinueOnError: f.continueOnFail,
				Strict:          f.strict,
				JSON:            f.asJSON,
			}
			result, err := handlers.Batch(commandContext(cmd), req)
			if result.Total > 0 {
				if renderErr := writeRendered(cmd, f.asJSON, func(formatter output.Formatter) ([]byte, error) {
					return formatter.Batch(result)
				}); renderErr != nil {
					return renderErr
				}
			}
			if err != nil {
				return err
			}
			if result.Total == 0 {
				return writeRendered(cmd, f.asJSON, func(formatter output.Formatter) ([]byte, error) {
					return formatter.Batch(result)
				})
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&f.manifest, "manifest", "m", "", "Path to batch manifest file or - for stdin")
	cmd.Flags().BoolVar(&f.continueOnFail, "continue-on-error", false, "Continue processing after individual file failures")
	cmd.Flags().BoolVarP(&f.strict, "strict", "s", false, "Reject invalid metadata instead of auto-correcting")
	cmd.Flags().BoolVarP(&f.asJSON, "json", "j", false, "Emit result JSON")
	_ = cmd.MarkFlagRequired("manifest")

	return cmd
}
