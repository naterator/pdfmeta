# CLI Reference

## Commands
- `pdfmeta show --file <pdf> [--json]`
- `pdfmeta set --file <pdf> (--out <pdf> | --in-place) [fields...] [--strict] [--json]`
- `pdfmeta unset --file <pdf> (--out <pdf> | --in-place) [--all|fields...] [--strict] [--json]`
- `pdfmeta batch --manifest <json> [--continue-on-error] [--strict] [--json]`
- `pdfmeta template save --name <name> [fields...] [--note] [--force]`
- `pdfmeta template apply --name <name> --file <pdf> (--out <pdf> | --in-place) [--strict] [--json]`
- `pdfmeta template list [--json]`
- `pdfmeta template show --name <name> [--json]`
- `pdfmeta template delete --name <name> [--force]`

## Metadata fields
- `--title`
- `--author`
- `--subject`
- `--keywords`
- `--creator`
- `--producer`
- `--creation-date`
- `--mod-date`

## Validation rules
- `set`, `unset`, `template apply`: require exactly one of `--out` or `--in-place`.
- `set`: requires at least one metadata field.
- `unset`: requires `--all` or at least one field selector.
- `--strict`: date strings must be RFC3339 or PDF date token format.
- non-strict mode: non-empty date strings are accepted and normalized where possible.

## Exit codes
- `0` success
- `3` validation
- `4` not found
- `5` conflict
- `6` encrypted PDF unsupported
- `7` malformed PDF
- `8` IO
- `9` internal
