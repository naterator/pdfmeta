# pdfmeta

`pdfmeta` is a Go CLI for reading and editing PDF metadata with:
- native PDF `/Info` dictionary updates
- native catalog `/Metadata` (XMP) stream updates
- safe output mode (`--out`) and atomic in-place writes (`--in-place`)
- batch processing via manifest
- reusable templates

## Quick start
```bash
go mod tidy
go test ./...
go run ./cmd/pdfmeta --help
```

## Basic usage
```bash
go run ./cmd/pdfmeta show --file in.pdf
go run ./cmd/pdfmeta set --file in.pdf --out out.pdf --title "Doc" --author "Team"
go run ./cmd/pdfmeta unset --file out.pdf --in-place --keywords
```

## Templates
By default templates are stored at `~/.pdfmeta/templates.json`.
Override with env var:
```bash
PDFMETA_TEMPLATE_STORE=/tmp/templates.json go run ./cmd/pdfmeta template save --name rel --title "Release"
```

## Docs
- [docs/README.md]()
- [docs/cli-reference.md]()
- [docs/usage-notes.md]()
