# Usage Notes

## MVP capabilities
- Reads metadata from native `/Info` dictionary and catalog `/Metadata` XMP stream.
- Writes metadata by appending incremental update objects for `/Info`, `/Metadata`, and catalog reference updates.
- Supports safe output file writes and atomic in-place updates.
- Supports template persistence and application.
- Supports batch manifest execution with per-item result reporting.

## Encrypted PDFs
- Read-only behavior may work depending on file structure.
- Writes are blocked for encrypted PDFs and return `ErrPDFEncrypted` (exit code `6`).

## Date behavior
- Strict mode (`--strict`): must be RFC3339 or PDF date format.
- Default mode: accepts non-empty values and normalizes common formats (`YYYY-MM-DD`, `YYYY/MM/DD`, with optional time).

## Template store location
- Default: `~/.pdfmeta/templates.json`
- Override per command: `PDFMETA_TEMPLATE_STORE=/path/templates.json`

## Known limits (non-blocking for MVP)
- Incremental writer is optimized for standard PDFs and may not handle every edge case found in complex/xref-stream-heavy files.
- No password/encrypted write support in v1.
