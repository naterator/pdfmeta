package cli

import (
	"github.com/spf13/cobra"

	"pdfmeta/internal/app"
	"pdfmeta/internal/model"
	"pdfmeta/internal/output"
	"pdfmeta/internal/validate"
)

type setFlags struct {
	file     string
	out      string
	inPlace  bool
	strict   bool
	asJSON   bool
	fromJSON string
	metadata metadataStringFlags
}

func newSetCmd(handlers *app.Handlers) *cobra.Command {
	f := &setFlags{}
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set metadata fields",
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonPatch, err := readMetadataPatchInput(cmd, f.fromJSON)
			if err != nil {
				return err
			}
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
				Changes: model.MergeMetadataPatch(jsonPatch, patchFromMetadataFlags(cmd, &f.metadata)),
			}
			if err := validate.SetRequest(req); err != nil {
				return err
			}
			result, err := handlers.Set(commandContext(cmd), req)
			if err != nil {
				return err
			}
			return writeRendered(cmd, f.asJSON, func(formatter output.Formatter) ([]byte, error) {
				return formatter.Show(result)
			})
		},
	}

	cmd.Flags().StringVarP(&f.file, "file", "f", "", "Input PDF file")
	cmd.Flags().StringVarP(&f.out, "out", "o", "", "Output PDF file")
	cmd.Flags().BoolVarP(&f.inPlace, "in-place", "i", false, "Modify file in place using safe atomic replace")
	cmd.Flags().BoolVarP(&f.strict, "strict", "s", false, "Reject invalid metadata instead of auto-correcting")
	cmd.Flags().BoolVarP(&f.asJSON, "json", "j", false, "Emit result JSON")
	cmd.Flags().StringVar(&f.fromJSON, "from-json", "", "Read metadata fields from a JSON file or - for stdin")
	addMetadataPatchFlags(cmd, &f.metadata, metadataFlagUsageDefault)
	_ = cmd.MarkFlagRequired("file")

	return cmd
}
