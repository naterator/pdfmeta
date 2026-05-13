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
pdfmeta show --file in.pdf --only-set
pdfmeta show --file in.pdf --field title --field author
```

| Flag | Short | Description |
|------|-------|-------------|
| `--file` | `-f` | Input PDF file (required) |
| `--json` | `-j` | Output JSON |
| `--only-set` | | Show only fields that currently have values |
| `--field` | | Limit output to specific metadata fields |

### set

Set metadata fields. Requires `--out` or `--in-place`.

```bash
pdfmeta set --file in.pdf --out out.pdf --title "Doc" --author "Team"
pdfmeta set --file in.pdf --in-place --subject "Notes" --strict
pdfmeta set --file in.pdf --out out.pdf --from-json meta.json --title "Override Title"
```

| Flag | Short | Description |
|------|-------|-------------|
| `--file` | `-f` | Input PDF file (required) |
| `--out` | `-o` | Output PDF file |
| `--in-place` | `-i` | Modify file in place using safe atomic replace |
| `--strict` | `-s` | Reject invalid metadata instead of auto-correcting |
| `--json` | `-j` | Emit result JSON |
| `--from-json` | | Read metadata fields from a JSON file or `-` for stdin |
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
cat ops.json | pdfmeta batch --manifest -
```

| Flag | Short | Description |
|------|-------|-------------|
| `--manifest` | `-m` | Path to batch manifest file or `-` for stdin (required) |
| `--continue-on-error` | | Continue processing after individual file failures |
| `--strict` | `-s` | Reject invalid metadata instead of auto-correcting |
| `--json` | `-j` | Emit result JSON |

Manifest files contain an `items` array. Each item includes an `op` and an `input` PDF path.

| Key | Description |
|-----|-------------|
| `op` | One of `show`, `set`, `unset`, or `template-apply` |
| `input` | Input PDF path |
| `output` | Output PDF path for write operations |
| `inPlace` | Modify `input` in place for write operations |
| `set` | Metadata object for `set` items using keys like `title`, `author`, `creationDate`, and `modDate` |
| `unset` | Metadata fields for `unset` items, such as `title`, `keywords`, `creation-date`, or `mod-date` |
| `unsetAll` | Remove all supported metadata fields for `unset` items |
| `template` | Template name for `template-apply` items |

```json
{
  "items": [
    {
      "op": "set",
      "input": "in.pdf",
      "output": "out.pdf",
      "set": {
        "title": "Release Notes",
        "author": "Docs Team"
      }
    },
    {
      "op": "unset",
      "input": "out.pdf",
      "inPlace": true,
      "unset": ["keywords"]
    },
    {
      "op": "template-apply",
      "input": "out.pdf",
      "output": "templated.pdf",
      "template": "rel"
    }
  ]
}
```

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
pdfmeta template apply --name rel --file in.pdf --out out.pdf --title "Hotfix Notes"
```

| Flag | Short | Description |
|------|-------|-------------|
| `--name` | `-n` | Template name (required) |
| `--file` | `-f` | Input PDF file (required) |
| `--out` | `-o` | Output PDF file |
| `--in-place` | `-i` | Modify file in place using safe atomic replace |
| `--strict` | `-s` | Reject invalid metadata instead of auto-correcting |
| `--json` | `-j` | Emit result JSON |
| `--title`, `--author`, `--subject`, `--keywords`, `--creator`, `--producer`, `--creation-date`, `--mod-date` | | Override specific metadata fields after loading the template |

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

#### template export

Export templates as JSON using the same shape accepted by `template import`.

```bash
pdfmeta template export
pdfmeta template export --out templates.json
```

| Flag | Short | Description |
|------|-------|-------------|
| `--out` | `-o` | Write export JSON to a file instead of stdout |

#### template import

Import templates from JSON produced by `template export`. Supports either the wrapped `{"templates":[...]}` shape or a raw array of template records.
Each imported template must include at least one metadata field.

```bash
pdfmeta template import --file templates.json
cat templates.json | pdfmeta template import --file - --force
```

| Flag | Short | Description |
|------|-------|-------------|
| `--file` | `-f` | Template import JSON file or `-` for stdin (required) |
| `--force` | | Overwrite existing templates with matching names |

### version

```bash
pdfmeta version
```

### selfupdate

Download and install the latest release from GitHub. Verifies SHA-256 checksums before replacing the binary.
`selfupdate` is currently unavailable on Windows.

```bash
pdfmeta selfupdate
```

## License

[BSD 3-Clause](./LICENSE)
