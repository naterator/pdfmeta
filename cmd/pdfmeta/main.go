package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"pdfmeta/internal/cli"
	"pdfmeta/internal/model"
)

func run(args []string, stdout, stderr io.Writer) int {
	deps := cli.Dependencies{
		Version: appVersion,
		Autoupdate: func(ctx context.Context, stdout io.Writer) error {
			return runAutoupdate(ctx, stdout)
		},
	}

	if err := cli.ExecuteWithDependencies(args, stdout, stderr, deps); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return model.ExitCode(err)
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
