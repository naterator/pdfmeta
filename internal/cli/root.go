package cli

import (
	"pdfmeta/internal/app"

	"github.com/spf13/cobra"
)

// NewRootCmd builds the top-level command tree and shared execution behavior.
func NewRootCmd() *cobra.Command {
	return NewRootCmdWithDependencies(Dependencies{})
}

// NewRootCmdWithDependencies builds the root command tree with injectable dependencies.
func NewRootCmdWithDependencies(deps Dependencies) *cobra.Command {
	deps = deps.withDefaults()
	handlers := app.NewHandlers(deps.Service)

	cmd := &cobra.Command{
		Use:           "pdfmeta",
		Short:         "Read and edit PDF metadata",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(newShowCmd(handlers))
	cmd.AddCommand(newSetCmd(handlers))
	cmd.AddCommand(newUnsetCmd(handlers))
	cmd.AddCommand(newBatchCmd(handlers))
	cmd.AddCommand(newTemplateCmd(handlers))

	return cmd
}
