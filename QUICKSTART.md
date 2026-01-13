# Quick Start

## Prerequisites

- Anytype desktop app running with API enabled (port 31009)

## Step 1: Build

```bash
make build
```

## Step 2: Authenticate

```bash
./build/any-vcard auth
```

1. A 4-digit code appears in your terminal
2. Open Anytype → Settings → API → enter the code
3. Copy the App Key from the output

## Step 3: Configure

```bash
export ANYTYPE_APP_KEY="your-app-key"
```

## Step 4: Find Your Space

```bash
./build/any-vcard space list
```

Copy the space ID you want to use:

```bash
export ANYTYPE_SPACE_ID="your-space-id"
```

## Step 5: Import Contacts

```bash
# Preview first
./build/any-vcard import --dry-run examples/sample-contacts.vcf

# Import
./build/any-vcard import examples/sample-contacts.vcf
```

Done! Your contacts are now in Anytype.
