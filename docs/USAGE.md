# Gblobs Usage Guide

## Overview
Gblobs provides efficient, deduplicated, optionally encrypted storage for arbitrary blobs (binary data), usable as a Go library and as a CLI tool.

---

## Go Module Usage

### Store Creation and Opening
```go
import "github.com/example/gblobs/gblobs"

store := &gblobs.LocalStore{}
// Create a store (unencrypted)
err := store.CreateStore("./path_to_store")
// Create an encrypted store
err := store.CreateStore("./path_to_store", "my-secret-key")
// Open a store (unencrypted or encrypted)
err := store.OpenStore("./path_to_store")
err := store.OpenStore("./path_to_store", "my-secret-key")
```

### Storing and Retrieving Blobs
```go
meta := gblobs.BlobType{
    Name: "my.txt",
    URI:  "file:///tmp/my.txt",
    Owner: "user01",
    IngestionTime: time.Now().UTC(),
}
blobID, err := store.PutBlob([]byte("my data"), meta)
data, meta, err := store.GetBlob(blobID)
exists, err := store.ExistsBlob(blobID)
err := store.DeleteBlob(blobID)
```

### Stats, Inspect, and Purge
```go
stats, err := store.Stats()
// stats.TotalBlobCount, stats.MaxCountPerLevel, stats.AverageCountPerLevel

blobs, err := store.InspectStore()
// Returns []gblobs.BlobType with metadata of all blobs in the store

err := store.PurgeStore()
```

---

## CLI Usage

The `gblobs` binary provides commands mirroring the library interface.

```sh
gblobs putfile <file> --store <path> [--key <encryption-key>] [--owner <str>]
gblobs putstring <string> --store <path> [--key <encryption-key>] [--name <name>] [--owner <str>]
gblobs get <blobID> --store <path> [--key <encryption-key>] [--out <file>]
gblobs exists <blobID> --store <path> [--key <encryption-key>]
gblobs delete <blobID> --store <path> [--key <encryption-key>]
gblobs purge --store <path> [--key <encryption-key>]
gblobs stats --store <path> [--key <encryption-key>]
gblobs inspect --store <path> [--key <encryption-key>]
```

### Example
```sh
gblobs putfile myimage.jpg --store ./s
ID=$(gblobs putstring "hello" --store ./s)
gblobs get "$ID" --store ./s --out output.txt
gblobs stats --store ./s
gblobs inspect --store ./s  # List all blobs with metadata, sorted by name
```

---

## Feature Highlights & Notes
- **Deduplication:** Multiple uploads of byte-identical data yield a single instance (same ID).
- **Compression:** All blobs are compressed on disk (gzip).
- **Encryption:** Specify `--key` or provide a key to enable AES-256 store encryption.
- **Stats:** The `stats` command summarizes blob counts and directory structure.
- **Inspect:** The `inspect` command traverses all paths and displays metadata for all blobs, sorted by name.
- **File layout:** Each blob is stored as `<store>/<2chars>/<3chars>/<3chars>/<rest>.blob` for 3-level scalability.
- **Metadata:** Stored in a sidecar `.meta` JSON file.
- **Blobs are immutable:** New writes of the same data result in deduplication, not overwrites.

---

## Troubleshooting
- For encrypted stores, always supply the same key or blobs cannot be read.
- Stats only count blobs, not meta files.
- If you see unexpected errors, ensure correct store/key/path for all commands.
