# gblobs: Technical Concept

## Overview
gblobs is a Go-based object storage solution designed for both programmatic (module) and command-line (CLI) usage. It provides storage, retrieval, and management of binary large objects ("blobs") with optional encryption and transparent compression to optimize disk use. It is organized to avoid file system bottlenecks and minimize duplicate storage of identical blobs.

## Key Features

- **Dual Interface**: Usable as a Go module and via a CLI
- **Core Blob Operations**:
  - `getBlob(blobID)`: Retrieve a blob by its ID, transparently decompressing/decrypting
  - `putBlob(data, meta)`: Store a new blob, avoiding duplicates by using a hash as blob ID. Returns the new/existing blob's ID
  - `deleteBlob(blobID)`: Remove a blob by its ID
  - `existsBlob(blobID)`: Check existence of a blob
- **Store Management**:
  - `openStore(path)` / `openStore(path, key)` (encrypted)
  - `createStore(path)` / `createStore(path, key)` (encrypted)
  - `purgeStore()` (clear all blobs)
  - `stats()` (report store statistics: total count, max per level, average per level)
  - `inspectStore()` (traverse and collect metadata of all blobs)
- **Object Type**:
  - Blobs are described by `BlobType`:
    - `Name` (original file or logical name)
    - `URI` (full path or locator, protocol included)
    - `Length` (in bytes)
    - `BlobHash` (unique hash id of the content)
    - `IngestionTime` (timestamp of storage)
    - `Owner` (optional, to identify ownership)
- **Compression**: All blobs are compressed on store and decompressed on read, using gzip
- **Duplicate Avoidance**: The blob-id is a SHA-256 hash of the blob's content, ensuring byte-identical blobs share a single instance
- **Filesystem Organization**:
  - Blob IDs (hex-encoded) are split into a 3-level directory structure for scalability
  - Example: Blob ID `ba7816bf8f01cfea414140de5dae5` ⟶ stored at `ba/781/6bf/8f01cfea414140de5dae5.blob`
  - Each blob has a corresponding `.meta` file containing JSON metadata
- **Encryption (Optional)**:
  - Stores can be encrypted with a user key using AES-256
  - When provided, all blobs are encrypted before compression and decrypted after decompression
  - Both creation and opening of stores support key-based operation

## CLI Interface
- Implements all core operations:
  - `putfile <file>`: Store a file as a blob, returns blob-id
  - `putstring <string>`: Store a string as a blob, returns blob-id
  - `get <blobID>`: Retrieve a blob by ID
  - `exists <blobID>`: Check if blob exists
  - `delete <blobID>`: Delete a blob
  - `purge`: Clear all blobs from store
  - `stats`: Display store statistics (total count, max per level, average per level)
  - `inspect`: List all blobs with metadata, sorted by name then ingestion time

## Testing
- Comprehensive test suite covering all functionality
- Automated test runner producing detailed test reports
- Tests for encryption, compression, deduplication, and CLI operations

## Demos
- Comprehensive demo suite under `demo/` including:
  - Basic usage (store, retrieve, check, delete)
  - CLI scenarios (all commands)
  - Encrypted store handling
  - Blob deduplication verification
  - File and string storage examples
