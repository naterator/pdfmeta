# PDF Test Fixtures

This directory contains deterministic fixtures for `internal/pdf` unit tests.

- `minimal.pdf`: Basic PDF-shaped fixture with header and EOF markers.
- `encrypted-marker.pdf`: PDF-shaped fixture that includes a `/Encrypt` trailer token.
- `encrypt-name-prefix-only.pdf`: Includes `/Encrypted` but not exact `/Encrypt` key.
- `encrypt-in-stream-only.pdf`: Contains `/Encrypt` text in stream content, not trailer.
- `encrypt-after-startxref.pdf`: Contains `/Encrypt` after `startxref`; should not count as trailer key.
- `leading-junk-before-header.pdf`: Includes `%PDF-` after leading junk bytes.
- `truncated-no-eof.pdf`: PDF header present but intentionally missing `%%EOF`.
- `malformed-trailer.pdf`: PDF-shaped content with malformed trailer/startxref values.
- `invalid.txt`: Non-PDF fixture used for negative test cases.

These fixtures are for parser/unit test behavior and do not need to be visually renderable.
