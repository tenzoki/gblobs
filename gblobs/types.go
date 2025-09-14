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
    InspectStore() ([]BlobType, error)

    // Search functionality
    Search(query string) ([]SearchResult, error)
    SearchWithOptions(req SearchRequest) ([]SearchResult, error)
}

// StoreStats summarizes the file organization in the blob store.
type StoreStats struct {
    TotalBlobCount       int     // number of blobs in the store
    MaxCountPerLevel     []int   // per-level file count maxima
    AverageCountPerLevel float64 // mean file count per level
}

// SearchRequest defines search parameters
type SearchRequest struct {
    Query     string   // Search query string
    Limit     int      // Maximum number of results
    Offset    int      // Starting offset for pagination
    Fields    []string // Fields to return in results
    Highlight bool     // Whether to include highlighting
}

// SearchResult represents a search hit
type SearchResult struct {
    BlobID     string            // Blob identifier
    Metadata   BlobType          // Blob metadata
    Score      float64           // Search relevance score
    Highlights map[string][]string // Field -> highlighted fragments
}

// IndexDocument represents a document in the search index
type IndexDocument struct {
    BlobID        string    `json:"blobId"`
    Name          string    `json:"name"`
    URI           string    `json:"uri"`
    Owner         string    `json:"owner"`
    Content       string    `json:"content"`       // Extracted text content
    IngestionTime time.Time `json:"ingestionTime"`
    Length        int64     `json:"length"`
}
// NowUTC returns the current time in UTC.
func NowUTC() time.Time {
    return time.Now().UTC()
}
