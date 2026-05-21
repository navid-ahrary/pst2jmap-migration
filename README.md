# pst2jmap-migration

CLI tool for migrating Microsoft Outlook `.pst` archives into JMAP mail servers such as Stalwart Mail Server.

The migration pipeline is optimized for large mailbox imports and uses offline PST extraction to improve performance and reliability.

---

## Architecture

```text
PST
 ↓
readpst
 ↓
EML extraction
 ↓
Go migration engine
 ↓
JMAP Upload
 ↓
Email/import
 ↓
Stalwart mailbox
```

---

## Features

Current:

- Offline PST extraction using `readpst`
- Recursive mailbox folder discovery
- EML-based processing
- Lightweight Go CLI
- Large mailbox support
- Folder hierarchy preservation

Planned:

- JMAP authentication
- Blob upload
- Email/import
- Attachment migration
- Parallel uploads
- Resume/retry
- Progress reporting
- Migration validation

---

## Requirements

### Runtime

- Linux
- Go 1.22+
- `readpst`

Install:

Ubuntu:

```bash
sudo apt install pst-utils
```

RHEL:

```bash
sudo dnf install libpst
```

Verify:

```bash
readpst -V
```

---

## Build

```bash
go build -o pst2jmap
```

Release:

```bash
CGO_ENABLED=0 \
go build \
-trimpath \
-ldflags="-s -w"
```

---

## Usage

Extract and migrate:

```bash
pst2jmap \
--pst backup.pst \
--server https://mail.example.com/jmap \
--user admin@example.com \
--password secret
```

---

## Migration Flow

### 1. Extract PST

```bash
readpst \
-r \
-e \
-u \
-o ./output \
backup.pst
```

Output:

```text
output/
├── Inbox/
├── Sent Items/
├── Deleted Items/
└── Drafts/
```

---

### 2. Import

CLI scans extracted `.eml` files.

For each message:

```text
EML
 ↓
Upload
 ↓
blobId
 ↓
Email/import
```

---

## Folder Mapping

| Outlook | JMAP |
|---|---|
| Inbox | Inbox |
| Sent Items | Sent |
| Deleted Items | Trash |
| Drafts | Drafts |
| Junk Email | Junk |

---

## Project Structure

```text
internal/
├── jmap/
├── model/
│   └── message.go
│
├── pst/
│   ├── reader.go
│   ├── messages.go
│   ├── stats.go
│   └── count.go

main.go
README.md
```

---

## Test Data

```text
testdata/
├── backup_150mb.pst
└── backup_1GB.pst
```

---

## Roadmap

### Extraction

- [ ] Replace go-pst with readpst
- [ ] Recursive extraction
- [ ] Folder normalization

### Migration

- [ ] Upload blobs
- [ ] Email/import
- [ ] Mailbox creation
- [ ] Keyword mapping

### Reliability

- [ ] Resume support
- [ ] Checkpoints
- [ ] Retry
- [ ] Validation

### Performance

- [ ] Parallel upload workers
- [ ] Streaming imports
- [ ] Memory limits

---

## Non-goals

Not migrating:

- Calendar
- Contacts
- Tasks
- Teams metadata
- Search folders
- Outlook application data

---

## License

MIT
