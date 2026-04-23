package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"

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
	if err := root.Execute(); err != nil {
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
