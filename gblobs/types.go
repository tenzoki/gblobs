package gblobs

import "time"

// BlobType metadata for a blob
type BlobType struct {
    Name         string    // Original name as ingested
    URI          string    // Location URI (protocol + full path)
    Length       int64     // Blob length in bytes
    BlobHash     string    // Hash of content (blob id)
    IngestionTime time.Time // Time of ingestion
    Owner        string    // Optional owner identifier
}

// Required blob store interface
type Store interface {
    GetBlob(blobID string) ([]byte, BlobType, error)
    PutBlob(data []byte, meta BlobType) (string, error)
    DeleteBlob(blobID string) error
    ExistsBlob(blobID string) (bool, error)

    OpenStore(path string, keyOpt ...string) error
    CreateStore(path string, keyOpt ...string) error
    PurgeStore() error

    Stats() (StoreStats, error)
}

// StoreStats summarizes the file organization in the blob store.
type StoreStats struct {
    TotalBlobCount       int     // number of blobs in the store
    MaxCountPerLevel     []int   // per-level file count maxima
    AverageCountPerLevel float64 // mean file count per level
}
// NowUTC returns the current time in UTC.
func NowUTC() time.Time {
    return time.Now().UTC()
}
