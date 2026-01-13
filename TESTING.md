# Testing

## Prerequisites

- A running Anytype instance with the API enabled
- A valid `ANYTYPE_APP_KEY` (obtain via `any-vcard auth`)

## Running Tests

Set the required environment variable and run:

```bash
ANYTYPE_APP_KEY=your-app-key make test
```

Optionally, set a custom API URL (defaults to `http://localhost:31009`):

```bash
ANYTYPE_URL=http://custom:31009 ANYTYPE_APP_KEY=your-app-key make test
```

## What the Tests Do

The end-to-end tests:

1. Create a temporary test space named `anyvcard_test_XXXXX`
2. Import contacts from `examples/sample-contacts.vcf`
3. Verify all contacts were created with correct properties (name, email, phone, organization, etc.)

## Test Spaces

Test spaces are **not automatically deleted** after test runs. This allows manual inspection of imported data. You can delete them manually from Anytype when done.
