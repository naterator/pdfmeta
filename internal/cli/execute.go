package cli

import (
	"fmt"
	"io"
)

// Execute runs the CLI with the provided args and IO streams.
func Execute(args []string, stdout io.Writer, stderr io.Writer) error {
	root := NewRootCmd()
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)
	if err := root.Execute(); err != nil {
		return fmt.Errorf("pdfmeta: %w", err)
	}
	return nil
}
