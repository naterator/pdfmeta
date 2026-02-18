# Examples

## Show
```bash
go run ./cmd/pdfmeta show --file testdata/pdf/minimal.pdf
go run ./cmd/pdfmeta show --file testdata/pdf/minimal.pdf --json
```

## Set
```bash
go run ./cmd/pdfmeta set --file in.pdf --out out.pdf --title "Quarterly Report" --author "Ops"
```

## Unset
```bash
go run ./cmd/pdfmeta unset --file out.pdf --in-place --keywords --subject
```

## Strict date validation
```bash
go run ./cmd/pdfmeta set --file in.pdf --out out.pdf --creation-date "2026-02-17T10:00:00Z" --strict
```

## Template flow
```bash
PDFMETA_TEMPLATE_STORE=/tmp/templates.json go run ./cmd/pdfmeta template save --name release --title "Release Doc"
PDFMETA_TEMPLATE_STORE=/tmp/templates.json go run ./cmd/pdfmeta template apply --name release --file in.pdf --out out.pdf --json
```

## Batch flow
Create manifest `manifest.json`:
```json
{
  "items": [
    {
      "op": "set",
      "input": "in.pdf",
      "output": "out.pdf",
      "set": {"title": "Batch Title"}
    },
    {
      "op": "show",
      "input": "out.pdf"
    }
  ]
}
```

Run:
```bash
go run ./cmd/pdfmeta batch --manifest manifest.json --json
```
