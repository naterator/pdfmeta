package cli

import "errors"

func validateOutputMode(out string, inPlace bool) error {
	if out != "" && inPlace {
		return errors.New("--out and --in-place are mutually exclusive")
	}
	if out == "" && !inPlace {
		return errors.New("either --out or --in-place is required")
	}
	return nil
}
