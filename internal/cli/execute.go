package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"

	"pdfmeta/internal/output"

	"github.com/spf13/cobra"
)

// Execute runs the CLI with the provided args and IO streams.
func Execute(args []string, stdout io.Writer, stderr io.Writer) error {
	return ExecuteWithDependencies(args, stdout, stderr, Dependencies{})
}

// ExecuteWithDependencies runs the CLI with injected runtime dependencies.
func ExecuteWithDependencies(args []string, stdout io.Writer, stderr io.Writer, deps Dependencies) error {
	root := NewRootCmdWithDependencies(deps)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	root.SetContext(ctx)
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)
	executed, err := root.ExecuteC()
	if err != nil {
		if renderErr := writeCommandError(executed, stderr, err); renderErr != nil {
			return fmt.Errorf("pdfmeta: %w", renderErr)
		}
		return fmt.Errorf("pdfmeta: %w", err)
	}
	return nil
}

func commandContext(cmd *cobra.Command) context.Context {
	if cmd != nil && cmd.Context() != nil {
		return cmd.Context()
	}
	return context.Background()
}

func writeCommandError(cmd *cobra.Command, fallback io.Writer, err error) error {
	asJSON := commandRequestedJSON(cmd)
	formatter, formatErr := output.NewFormatter(output.ParseFormat(asJSON))
	if formatErr != nil {
		return fmt.Errorf("create formatter: %w", formatErr)
	}
	payload, formatErr := formatter.Err(err)
	if formatErr != nil {
		return formatErr
	}
	w := fallback
	if cmd != nil {
		w = cmd.ErrOrStderr()
	}
	if _, writeErr := w.Write(payload); writeErr != nil {
		return fmt.Errorf("write error output: %w", writeErr)
	}
	return nil
}

func commandRequestedJSON(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	flag := cmd.Flags().Lookup("json")
	return flag != nil && flag.Changed
}
