# pdfmeta

`pdfmeta` is a Go CLI for reading and editing PDF metadata with:
- native PDF `/Info` dictionary updates
- native catalog `/Metadata` (XMP) stream updates
- safe output mode (`--out`) and atomic in-place writes (`--in-place`)
- batch processing via manifest
- reusable templates

## Build
```bash
git clone git@github.com:naterator/pdfmeta.git
cd pdfmeta
go build ./cmd/pdfmeta
```

## Basic usage
```bash
./pdfmeta show --file in.pdf
./pdfmeta set --file in.pdf --out out.pdf --title "Doc" --author "Team"
./pdfmeta unset --file out.pdf --in-place --keywords
```

## Templates
By default templates are stored at `~/.pdfmeta/templates.json`.
Override with env var:
```bash
PDFMETA_TEMPLATE_STORE=/tmp/templates.json ./pdfmeta template save --name rel --title "Release"
```

## Docs
- [docs/README.md]()
- [docs/cli-reference.md]()
- [docs/usage-notes.md]()
