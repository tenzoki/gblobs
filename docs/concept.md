# gblobs: Technical Concept

## Overview
gblobs is a Go-based object storage solution designed for both programmatic (module) and command-line (CLI) usage. It provides storage, retrieval, and management of binary large objects ("blobs") with optional encryption and transparent compression to optimize disk use. It is organized to avoid file system bottlenecks and minimize duplicate storage of identical blobs.

## Key Features
### Stats and Store Analysis (New)
- Both module and CLI provide a `stats` function/command that reports:
  - Total number of blobs/files in the store
  - Maximum file count per directory level in the gblobs hierarchy
  - Average file count per level
- **Dual Interface**: Usable as a Go module and via a CLI.
- **Core Blob Operations**:
  - `getBlob(blobID)`: Retrieve a blob by its ID, transparently decompressing/decrypting.
  - `putBlob(data, meta)`: Store a new blob, avoiding duplicates by using a hash as blob ID. Returns the new/existing blob's ID.
  - `deleteBlob(blobID)`: Remove a blob by its ID.
  - `existsBlob(blobID)`: Check existence of a blob.
- **Object Type**:
  - Blobs are described by `blobtype`:
    - `name` (original file or logical name)
    - `uri` (full path or locator, protocol included)
    - `length` (in bytes)
    - `blob-hash` (unique hash id of the content)
    - `ingestion-time` (timestamp of storage)
    - `owner` (optional, to identify ownership)
- **Compression**: All blobs are compressed on store and decompressed on read, using a standard algorithm (e.g., gzip or zstd).
- **Commands on Store (Persistence Layer)**:
  - `openStore(path)` / `openStore(path, key)` (encrypted)
  - `createStore(path)` / `createStore(path, key)` (encrypted)
  - `purgeStore()` (clear all blobs)
- **Duplicate Avoidance**: The blob-id is a reliable hash (e.g., SHA-256) of the blob's content (byte[]), ensuring byte-identical blobs share a single instance.
- **Filesystem Organization**:
  - Blob IDs (hex-encoded) are split into directory segments to keep folder structures efficient.
  - Example: Blob ID `ba7816bf8f01cfea414140de5` ⟶ stored at `ba/781/6bf8f01cfea414140de5.blob`
  - Controlled depth and width for file storage scalability.
- **Encryption (Optional)**:
  - Stores can be encrypted with a user key. When provided, all blobs are encrypted before compression and decrypted after decompression.
  - Both creation and opening of stores support key-based operation.

## CLI Interface
- Implements all logical operations except direct `putBlob`.
- Additional commands:
  - `putFile(filepath)`: Store a file as a blob, returns blob-id.
  - `putString(string)`: Store a string as a blob, returns blob-id.
- Other main commands mimic the Go API (get, exists, delete, etc).
+  - Adds `stats` to display gblobs statistics (total count, max per level, average per level)
- Implements all logical operations except direct `putBlob`.
- Additional commands:
  - `putFile(filepath)`: Store a file as a blob, returns blob-id.
  - `putString(string)`: Store a string as a blob, returns blob-id.
- Other main commands mimic the Go API (get, exists, delete, etc).

## Testing
- Every file and function to have a related test case.
- Automated test runner, producing detailed test reports.

## Demos
- **Commands on Store (Persistence Layer)**:
  - `openStore(path)` / `openStore(path, key)` (encrypted)
  - `createStore(path)` / `createStore(path, key)` (encrypted)
  - `purgeStore()` (clear all blobs)
+  - `stats()` (report store and directory statistics: total, max per level, average per level)
- Comprehensive demo suite under `demo/` to explain:
  - Basic usage (store, retrieve, check, delete)
  - CLI scenarios (all commands)
  - Encrypted store handling
  - Blob deduplication
  - Handling of files and raw strings
