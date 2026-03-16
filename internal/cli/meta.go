package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func newVersionCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show the current app version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), version)
			return err
		},
	}
}

func newAutoupdateCmd(run func(context.Context, io.Writer) error) *cobra.Command {
	return &cobra.Command{
		Use:                "autoupdate",
		Short:              "Download and install the latest release",
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(context.Background(), cmd.OutOrStdout())
		},
	}
}
