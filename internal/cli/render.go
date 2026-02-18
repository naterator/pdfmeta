package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"pdfmeta/internal/output"
)

func writeRendered(cmd *cobra.Command, asJSON bool, render func(f output.Formatter) ([]byte, error)) error {
	formatter, err := output.NewFormatter(output.ParseFormat(asJSON))
	if err != nil {
		return fmt.Errorf("create formatter: %w", err)
	}
	payload, err := render(formatter)
	if err != nil {
		return err
	}
	if _, err := cmd.OutOrStdout().Write(payload); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}
