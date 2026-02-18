package main

import (
	"fmt"
	"io"
	"os"

	"pdfmeta/internal/cli"
	"pdfmeta/internal/model"
)

func run(args []string, stdout, stderr io.Writer) int {
	if err := cli.Execute(args, stdout, stderr); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return model.ExitCode(err)
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
