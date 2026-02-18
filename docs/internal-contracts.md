# Internal Contracts

This document captures the current internal contract surface for core packages.

## Service interfaces (`internal/model/interfaces.go`)

- `Service`
  - `Show(context.Context, ShowRequest) (ShowResult, error)`
  - `Set(context.Context, SetRequest) (ShowResult, error)`
  - `Unset(context.Context, UnsetRequest) (ShowResult, error)`
  - `Batch(context.Context, BatchRequest) (BatchResult, error)`
  - `TemplateSave(context.Context, TemplateSaveRequest) (TemplateRecord, error)`
  - `TemplateApply(context.Context, TemplateApplyRequest) (ShowResult, error)`
  - `TemplateList(context.Context) ([]TemplateRecord, error)`
  - `TemplateShow(context.Context, string) (TemplateRecord, error)`
  - `TemplateDelete(context.Context, string) error`

- `MetadataStore`
  - `Read(context.Context, string) (MetadataReadResult, error)`
  - `Write(context.Context, MetadataWriteRequest) (MetadataReadResult, error)`

- `TemplateStore`
  - `Save(context.Context, TemplateRecord, bool) (TemplateRecord, error)`
  - `Get(context.Context, string) (TemplateRecord, error)`
  - `List(context.Context) ([]TemplateRecord, error)`
  - `Delete(context.Context, string) error`

## Request/response models (`internal/model/requests.go`, `internal/model/metadata.go`)

- Canonical metadata fields:
  - `title`, `author`, `subject`, `keywords`, `creator`, `producer`, `creation-date`, `mod-date`
- `MetadataPatch` uses pointer fields to represent partial updates (`nil` means unchanged).
- `ShowRequest` and `ShowResult` define single-file read shape.
- `SetRequest` and `UnsetRequest` define write and removal operations.
- `TemplateSaveRequest`, `TemplateApplyRequest`, and `TemplateRecord` define template flow.
- `BatchRequest`, `BatchItemResult`, and `BatchResult` define manifest execution and aggregate reporting.

## Validation contracts (`internal/validate/validate.go`, `internal/validate/fields.go`, `internal/validate/date.go`)

- IO rules (`SetRequest`, `UnsetRequest`, `TemplateApplyRequest`):
  - `InputPath` required
  - exactly one of `OutputPath` or `InPlace` must be set
- `SetRequest` requires at least one patch field.
- `UnsetRequest` requires either `All=true` or at least one field; cannot mix `All=true` with explicit fields.
- Field normalization:
  - unknown fields rejected
  - duplicate fields rejected
  - normalized output order follows `model.AllFields`
- Date validation:
  - strict mode accepts RFC3339 timestamps and PDF date tokens (`D:YYYY...`)
  - non-strict mode requires non-empty date strings and defers normalization/autocorrection to service layer

Validation failures return `*model.AppError` with code `ErrValidation`.

## Metadata persistence contract (`internal/metadata/store.go`)

- Reads metadata from native PDF `/Info` dictionary and catalog `/Metadata` XMP stream.
- Writes metadata via incremental update:
  - new `/Info` object
  - new `/Metadata` XML stream object
  - new catalog object referencing `/Metadata`
  - appended xref/trailer with `/Prev` link.

## Output contracts (`internal/output/contracts.go`)

- `Formatter` interface:
  - `Show(model.ShowResult) ([]byte, error)`
  - `Batch(model.BatchResult) ([]byte, error)`
  - `Template(model.TemplateRecord) ([]byte, error)`
  - `TemplateList([]model.TemplateRecord) ([]byte, error)`
  - `Err(error) ([]byte, error)`
- Supported formats:
  - `FormatText`
  - `FormatJSON`
- `ParseFormat(asJSON bool)` maps CLI `--json` toggle to formatter choice.

## Error taxonomy and exit-code mapping (`internal/model/errors.go`)

`ErrorCode` values:
- `unknown`
- `usage`
- `validation`
- `not_found`
- `conflict`
- `pdf_encrypted`
- `pdf_malformed`
- `io`
- `internal`

Exit code mapping:
- `nil` error -> `0`
- non-`AppError` -> `1`
- `usage` -> `2`
- `validation` -> `3`
- `not_found` -> `4`
- `conflict` -> `5`
- `pdf_encrypted` -> `6`
- `pdf_malformed` -> `7`
- `io` -> `8`
- `internal` -> `9`
