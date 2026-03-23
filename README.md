# pdfmeta

`pdfmeta` is a Go CLI for reading and editing PDF metadata with:
- native PDF `/Info` dictionary updates
- native catalog `/Metadata` (XMP) stream updates
- safe output mode (`--out`) and atomic in-place writes (`--in-place`)
- batch processing via manifest
- reusable templates
- self-update from GitHub releases

## Install

Download a prebuilt binary from [GitHub Releases](https://github.com/naterator/pdfmeta/releases) (available for darwin, linux, freebsd, netbsd, openbsd, windows on amd64 and arm64), or build from source:

```bash
git clone git@github.com:naterator/pdfmeta.git
cd pdfmeta
go build ./cmd/pdfmeta
```

## Commands

### show

Show metadata from a PDF.

```bash
pdfmeta show --file in.pdf
pdfmeta show --file in.pdf --json
```

| Flag | Short | Description |
|------|-------|-------------|
| `--file` | `-f` | Input PDF file (required) |
| `--json` | `-j` | Output JSON |

### set

Set metadata fields. Requires `--out` or `--in-place`.

```bash
pdfmeta set --file in.pdf --out out.pdf --title "Doc" --author "Team"
pdfmeta set --file in.pdf --in-place --subject "Notes" --strict
```

| Flag | Short | Description |
|------|-------|-------------|
| `--file` | `-f` | Input PDF file (required) |
| `--out` | `-o` | Output PDF file |
| `--in-place` | `-i` | Modify file in place using safe atomic replace |
| `--strict` | `-s` | Reject invalid metadata instead of auto-correcting |
| `--json` | `-j` | Emit result JSON |
| `--title` | | Title |
| `--author` | | Author |
| `--subject` | | Subject |
| `--keywords` | | Keywords |
| `--creator` | | Creator |
| `--producer` | | Producer |
| `--creation-date` | | Creation date |
| `--mod-date` | | Modification date |

### unset

Unset metadata fields. Requires `--out` or `--in-place`.

```bash
pdfmeta unset --file out.pdf --in-place --keywords
pdfmeta unset --file out.pdf --in-place --all
```

| Flag | Short | Description |
|------|-------|-------------|
| `--file` | `-f` | Input PDF file (required) |
| `--out` | `-o` | Output PDF file |
| `--in-place` | `-i` | Modify file in place using safe atomic replace |
| `--strict` | `-s` | Reject invalid metadata instead of auto-correcting |
| `--json` | `-j` | Emit result JSON |
| `--all` | | Unset all supported metadata fields |
| `--title` | | Unset Title |
| `--author` | | Unset Author |
| `--subject` | | Unset Subject |
| `--keywords` | | Unset Keywords |
| `--creator` | | Unset Creator |
| `--producer` | | Unset Producer |
| `--creation-date` | | Unset Creation date |
| `--mod-date` | | Unset Modification date |

### batch

Apply metadata operations to many PDFs via a JSON manifest.

```bash
pdfmeta batch --manifest ops.json
pdfmeta batch --manifest ops.json --continue-on-error --json
```

| Flag | Short | Description |
|------|-------|-------------|
| `--manifest` | `-m` | Path to batch manifest file (required) |
| `--continue-on-error` | | Continue processing after individual file failures |
| `--strict` | `-s` | Reject invalid metadata instead of auto-correcting |
| `--json` | `-j` | Emit result JSON |

### template

Manage saved metadata templates. By default templates are stored at `~/.pdfmeta/templates.json`. Override with `PDFMETA_TEMPLATE_STORE`.

#### template save

```bash
pdfmeta template save --name rel --title "Release" --author "Team"
pdfmeta template save --name rel --title "v2" --force
```

| Flag | Short | Description |
|------|-------|-------------|
| `--name` | `-n` | Template name (required) |
| `--note` | | Template description |
| `--force` | | Overwrite existing template |
| `--title`, `--author`, `--subject`, `--keywords`, `--creator`, `--producer`, `--creation-date`, `--mod-date` | | Metadata fields |

#### template apply

```bash
pdfmeta template apply --name rel --file in.pdf --out out.pdf
```

| Flag | Short | Description |
|------|-------|-------------|
| `--name` | `-n` | Template name (required) |
| `--file` | `-f` | Input PDF file (required) |
| `--out` | `-o` | Output PDF file |
| `--in-place` | `-i` | Modify file in place using safe atomic replace |
| `--strict` | `-s` | Reject invalid metadata instead of auto-correcting |
| `--json` | `-j` | Emit result JSON |

#### template list

```bash
pdfmeta template list
pdfmeta template list --json
```

#### template show

```bash
pdfmeta template show --name rel
```

| Flag | Short | Description |
|------|-------|-------------|
| `--name` | `-n` | Template name (required) |
| `--json` | `-j` | Output JSON |

#### template delete

```bash
pdfmeta template delete --name rel
```

| Flag | Short | Description |
|------|-------|-------------|
| `--name` | `-n` | Template name (required) |

### version

```bash
pdfmeta version
```

### selfupdate

Download and install the latest release from GitHub. Verifies SHA-256 checksums before replacing the binary.

```bash
pdfmeta selfupdate
```

## License

[BSD 3-Clause](./LICENSE)
