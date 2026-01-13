# any-vcard

Import vCard (.vcf) files into [Anytype](https://anytype.io/).

## Quick Start

See [QUICKSTART.md](QUICKSTART.md) for a step-by-step guide.

## Installation

```bash
# From source
git clone https://github.com/rubiojr/any-vcard
cd any-vcard
make build

# Or with Go
go install github.com/rubiojr/any-vcard/cmd/any-vcard@latest
```

## Usage

### 1. Authenticate

```bash
any-vcard auth
```

Follow the prompts to get your App Key, then:

```bash
export ANYTYPE_APP_KEY="your-app-key"
```

### 2. List Spaces

```bash
any-vcard space list
```

### 3. Import Contacts

```bash
export ANYTYPE_SPACE_ID="your-space-id"

# Preview first
any-vcard import --dry-run contacts.vcf

# Import
any-vcard import contacts.vcf
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `ANYTYPE_APP_KEY` | Your Anytype App Key |
| `ANYTYPE_SPACE_ID` | Target space ID |
| `ANYTYPE_URL` | API URL (default: http://localhost:31009) |

## License

MIT
