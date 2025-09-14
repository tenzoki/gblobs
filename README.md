# gblobs

Efficient, deduplicated, optionally encrypted storage for arbitrary binary data (blobs).

## What is gblobs?

gblobs is a content-addressable storage system that automatically deduplicates, compresses, and optionally encrypts your data. It's designed for applications that need to store many files or binary objects while minimizing storage space and ensuring data integrity.

**Key Benefits:**
- **Space Efficient**: Identical files are stored only once, regardless of how many times you add them
- **Fast Retrieval**: Content-addressable storage using SHA-256 hashes for instant lookups
- **Secure**: Optional AES-256 encryption protects your data at rest
- **Scalable**: 3-level directory structure handles millions of objects efficiently
- **Flexible**: Use it as a Go library in your applications or via the command-line tool

## Features

- **Deduplication**: Identical data stored only once using SHA-256 hashing
- **Compression**: All blobs compressed with gzip to minimize disk usage
- **Optional Encryption**: AES-256 encryption with user-provided keys
- **Scalable Storage**: 3-level directory structure (`ab/cde/fgh/rest.blob`) for efficient filesystem performance
- **Dual Interface**: Available as both Go module and CLI tool
- **Rich Metadata**: Track filename, URI, owner, ingestion time, and more
- **Store Management**: Statistics, inspection, and bulk operations

## Quick Start

### 1. Build and Test
```bash
# Clone and build
git clone <repo-url>
cd gblobs
make build

# Run tests to verify everything works
make test

# Try the demo
make demo
```

### 2. CLI Usage
```bash
# Store files and get their unique IDs
gblobs putfile document.pdf --store ./mystore
gblobs putfile image.jpg --store ./mystore --owner alice

# Store strings directly
ID=$(gblobs putstring "Hello, World!" --store ./mystore --name greeting)

# Retrieve data
gblobs get $ID --store ./mystore
gblobs get $ID --store ./mystore --out retrieved.txt

# Check if blob exists
gblobs exists $ID --store ./mystore

# Get store statistics
gblobs stats --store ./mystore

# List all stored blobs with metadata
gblobs inspect --store ./mystore

# Clean up
gblobs delete $ID --store ./mystore
gblobs purge --store ./mystore  # Delete everything
```

### 3. Go Module Usage
```go
package main

import (
    "fmt"
    "time"
    "github.com/example/gblobs/gblobs"
)

func main() {
    // Create a new store
    store := &gblobs.LocalStore{}
    err := store.CreateStore("./mystore")
    if err != nil {
        panic(err)
    }

    // Store some data with metadata
    meta := gblobs.BlobType{
        Name:          "greeting.txt",
        URI:           "string://local",
        Owner:         "alice",
        IngestionTime: time.Now().UTC(),
    }

    blobID, err := store.PutBlob([]byte("Hello, World!"), meta)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Stored blob with ID: %s\n", blobID)

    // Retrieve the data
    data, retrievedMeta, err := store.GetBlob(blobID)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Retrieved: %s\n", string(data))
    fmt.Printf("Original name: %s\n", retrievedMeta.Name)

    // Get store statistics
    stats, err := store.Stats()
    if err != nil {
        panic(err)
    }

    fmt.Printf("Total blobs: %d\n", stats.TotalBlobCount)
}
```

### 4. Encrypted Storage
```bash
# Create an encrypted store
gblobs putfile secret.txt --store ./encrypted --key "my-secret-password"

# All operations need the same key
gblobs get <blobID> --store ./encrypted --key "my-secret-password"
gblobs stats --store ./encrypted --key "my-secret-password"
```

## Use Cases

- **Backup Systems**: Deduplicated storage for backup data
- **Content Management**: Store documents, images, and media files efficiently
- **Data Archival**: Long-term storage with optional encryption
- **Build Artifacts**: Cache build outputs and dependencies
- **File Deduplication**: Remove duplicate files across directories
- **Secure Document Storage**: Encrypted storage for sensitive files

## Documentation

- **[Usage Guide](docs/USAGE.md)** - Comprehensive usage examples, all CLI commands, and complete Go API reference
- **[Technical Concepts](docs/concept.md)** - Architecture, design decisions, and implementation details

## Building and Development

```bash
make build    # Build CLI tool at ./cmd/gblobs/gblobs
make test     # Run all tests
make demo     # Run CLI and Go demos
make clean    # Remove build artifacts
make help     # Show all available targets
```

## Project Structure

```
gblobs/
├── cmd/gblobs/          # CLI tool implementation
├── gblobs/              # Core Go library
├── test/                # Test suites
├── demo/                # Demo scripts and examples
├── docs/                # Documentation
└── Makefile            # Build system
```