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

### Searching Content
```go
// Simple search
results, err := store.Search("machine learning")

// Advanced search with options
req := gblobs.SearchRequest{
    Query:     "artificial intelligence",
    Limit:     10,
    Offset:    0,
    Highlight: true,
}
results, err := store.SearchWithOptions(req)

// Process results
for _, result := range results {
    fmt.Printf("Found: %s (score: %.3f)\n", result.Metadata.Name, result.Score)
    fmt.Printf("Owner: %s, Size: %d bytes\n", result.Metadata.Owner, result.Metadata.Length)

    // Display highlights if available
    for field, fragments := range result.Highlights {
        fmt.Printf("Highlight in %s: %s\n", field, strings.Join(fragments, " ... "))
    }
}
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
gblobs search <query> --store <path> [--key <encryption-key>] [--limit <n>] [--offset <n>] [--highlight]
```

### Basic Usage Examples
```sh
# Store files and strings
gblobs putfile mydocument.txt --store ./s --owner alice
gblobs putfile report.pdf --store ./s --owner bob
ID=$(gblobs putstring "Meeting notes about project planning" --store ./s --name notes.txt --owner charlie)

# Retrieve content
gblobs get "$ID" --store ./s --out output.txt

# Search for content (returns full blob IDs for retrieval)
gblobs search "meeting" --store ./s
gblobs search "alice" --store ./s --limit 5
gblobs search "project planning" --store ./s --highlight

# Use search results to retrieve documents
BLOB_ID=$(gblobs search "meeting" --store ./s | grep "Blob ID:" | cut -d' ' -f3)
gblobs get --store ./s "$BLOB_ID"

# Management commands
gblobs stats --store ./s
gblobs inspect --store ./s  # List all blobs with metadata, sorted by name then ingestion time (newest first)
```

### Search Examples
```sh
# Basic text search
gblobs search "machine learning" --store ./mystore

# Search with results limiting
gblobs search "document" --store ./mystore --limit 10

# Paginated search results
gblobs search "report" --store ./mystore --limit 5 --offset 10

# Search by owner or metadata
gblobs search "alice" --store ./mystore

# Complex queries (Bleve query syntax)
gblobs search "title:report AND owner:alice" --store ./mystore

# Disable highlighting for faster results
gblobs search "meeting notes" --store ./mystore --highlight=false

# Search in encrypted store
gblobs search "sensitive data" --store ./encrypted --key "my-secret-key"
```

---

## Feature Highlights & Notes
- **Deduplication:** Multiple uploads of byte-identical data yield a single instance (same ID).
- **Compression:** All blobs are compressed on disk (gzip).
- **Encryption:** Specify `--key` or provide a key to enable AES-256 store encryption.
- **Full-Text Search:** All stored content is automatically indexed for fast, ranked search results.
- **Smart Indexing:** Text files are fully indexed; binary files have metadata indexed for search.
- **Search Features:** Relevance scoring, result highlighting, pagination, and advanced query syntax.
- **Stats:** The `stats` command summarizes blob counts and directory structure.
- **Inspect:** The `inspect` command traverses all paths and displays metadata for all blobs, sorted by name then ingestion time (newest first).
- **File layout:** Each blob is stored as `<store>/<2chars>/<3chars>/<3chars>/<rest>.blob` for 3-level scalability.
- **Search Index:** Co-located at `<store>/index.bleve/` for full-text search capabilities.
- **Metadata:** Stored in a sidecar `.meta` JSON file.
- **Blobs are immutable:** New writes of the same data result in deduplication, not overwrites.
- **Index Consistency:** Search index is automatically maintained - additions, deletions reflected immediately.

---

## Troubleshooting
- For encrypted stores, always supply the same key or blobs cannot be read.
- Stats only count blobs, not meta files.
- If you see unexpected errors, ensure correct store/key/path for all commands.
