# pst2jmap-migration

A lightweight Go command-line utility for migrating Outlook PST mailboxes into JMAP mail servers such as Stalwart Mail Server.

The project uses:

- go-pst for PST parsing
- JMAP for modern high-performance mailbox import
- Incremental PST processing for large mailboxes

---

## Features

- Parse Outlook `.pst` files
- Read mailbox folders
- Filter Outlook system folders
- Incremental processing for large PST files
- Lightweight standalone binary
- Cross-platform builds
- JMAP-based migration architecture
- Designed for Stalwart Mail Server compatibility

---

## Current Status

Current implementation supports:

- PST parsing
- Mailbox folder discovery
- Folder filtering
- CLI tooling

Planned:

- MIME extraction
- Attachment handling
- JMAP authentication
- Mailbox creation
- Email upload/import
- Parallel migration workers
- Resume support

---

## Supported Mailbox Folders

The tool currently processes mailbox folders under:

```text
Top-of-Information-Store
```

Typical folders:

- Inbox
- Sent Items
- Deleted Items
- Junk Email
- Drafts

Outlook internal/system folders are ignored.

---

## Requirements

- Go 1.22+
- Outlook PST file
- JMAP-compatible mail server

Tested against:

- Microsoft Outlook PST exports
- Exchange Online PSTs
- Microsoft 365 mailboxes

---

## Build

### Linux AMD64

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
  -trimpath \
  -ldflags="-s -w" \
  -o pst2jmap-migration-linux-amd64
```

### Linux ARM64

```bash
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build \
  -trimpath \
  -ldflags="-s -w" \
  -o pst2jmap-migration-linux-arm64
```

### Windows AMD64

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build \
  -trimpath \
  -ldflags="-s -w" \
  -o pst2jmap-migration-windows-amd64.exe
```

---

## Usage

```bash
pst2jmap-migration \
  --pst ./backup.pst \
  --url https://mail.example.com/jmap \
  --user admin@example.com \
  --password secret
```

---

## CLI Options

| Option | Description |
|---|---|
| `--pst` | Source PST file |
| `--url` | Destination JMAP endpoint |
| `--user` | Destination mailbox username |
| `--password` | Destination mailbox password |
| `--version` | Show version |

---

## Example Output

```text
Folder: Inbox
Messages: 2369

Folder: Sent Items
Messages: 554

Folder: Deleted Items
Messages: 61

Folder: Junk Email
Messages: 20

Folder: Drafts
Messages: 5
```

---

## Project Structure

```text
internal/
├── pst/
│   ├── reader.go
│   ├── filters.go
│   └── folders.go
│
├── jmap/
│
├── migrate/
│
└── model/

main.go
```

---

## Development Roadmap

- [x] PST parsing
- [x] Folder filtering
- [x] CLI implementation
- [ ] MIME extraction
- [ ] RFC822 generation
- [ ] JMAP authentication
- [ ] Blob upload
- [ ] Email/import support
- [ ] Attachment migration
- [ ] Parallel uploads
- [ ] Resume/retry support
- [ ] Progress reporting

---

## Important Notes

This project intentionally focuses on mailbox migration only.

The following Outlook object types are currently ignored:

- Calendar
- Contacts
- Tasks
- Teams metadata
- Search folders
- Outlook application data
- Sync issue folders

---

## Dependencies

- <https://github.com/mooijtech/go-pst>
- <https://github.com/emersion/go-message>

---

## License

MIT
