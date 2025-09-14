# Bleve Full-Text Search Integration Plan

## Overview

This document outlines the integration of Bleve full-text search capabilities into gblobs, enabling content-based search across all stored binary data.

## Requirements

- All byte[] data stored in gblobs must be full-text indexed
- Index lifecycle tied to store lifecycle (create/delete store = create/delete index)
- Index located in same path as store (e.g., `store_path/index.bleve`)
- Deleting a blob removes its index entry
- Search returns ranked matches with BlobType metadata objects

## Architecture Design

### 1. Index Lifecycle Management

#### Index Location Strategy
```
store_path/
├── ab/cd/ef/...       # Blob storage (existing)
└── index.bleve/       # Bleve full-text index
```

#### Store Structure Changes
```go
type LocalStore struct {
    path     string
    key      []byte
    lock     sync.RWMutex
    index    bleve.Index  // New: Bleve index instance
    indexPath string      // New: Index directory path
}
```

### 2. Integration Points

#### Store Creation (`CreateStore`)
```go
func (s *LocalStore) CreateStore(path string, keyOpt ...string) error {
    // Existing store creation logic...

    // Create Bleve index
    s.indexPath = filepath.Join(path, "index.bleve")
    mapping := s.createIndexMapping()
    index, err := bleve.New(s.indexPath, mapping)
    if err != nil {
        return fmt.Errorf("failed to create search index: %w", err)
    }
    s.index = index
    return nil
}
```

#### Store Opening (`OpenStore`)
```go
func (s *LocalStore) OpenStore(path string, keyOpt ...string) error {
    // Existing store opening logic...

    // Open existing Bleve index
    s.indexPath = filepath.Join(path, "index.bleve")
    index, err := bleve.Open(s.indexPath)
    if err != nil {
        return fmt.Errorf("failed to open search index: %w", err)
    }
    s.index = index
    return nil
}
```

#### Store Destruction (`PurgeStore`)
```go
func (s *LocalStore) PurgeStore() error {
    // Close and remove index first
    if s.index != nil {
        s.index.Close()
        os.RemoveAll(s.indexPath)
        s.index = nil
    }

    // Existing purge logic...
    return nil
}
```

### 3. Blob Indexing Integration

#### Document Structure for Indexing
```go
type IndexDocument struct {
    BlobID        string    `json:"blobId"`
    Name          string    `json:"name"`
    URI           string    `json:"uri"`
    Owner         string    `json:"owner"`
    Content       string    `json:"content"`       // Extracted text content
    IngestionTime time.Time `json:"ingestionTime"`
    Length        int64     `json:"length"`
}
```

#### Index Mapping Configuration
```go
func (s *LocalStore) createIndexMapping() mapping.IndexMapping {
    // Text field mapping for content
    textFieldMapping := bleve.NewTextFieldMapping()
    textFieldMapping.Store = false
    textFieldMapping.IncludeInAll = true

    // Keyword field mapping for exact matches
    keywordFieldMapping := bleve.NewKeywordFieldMapping()
    keywordFieldMapping.Store = true

    // Date field mapping
    dateFieldMapping := bleve.NewDateTimeFieldMapping()
    dateFieldMapping.Store = true

    // Document mapping
    docMapping := bleve.NewDocumentMapping()
    docMapping.AddFieldMappingsAt("content", textFieldMapping)
    docMapping.AddFieldMappingsAt("name", textFieldMapping)
    docMapping.AddFieldMappingsAt("uri", keywordFieldMapping)
    docMapping.AddFieldMappingsAt("owner", keywordFieldMapping)
    docMapping.AddFieldMappingsAt("blobId", keywordFieldMapping)
    docMapping.AddFieldMappingsAt("ingestionTime", dateFieldMapping)

    // Index mapping
    indexMapping := bleve.NewIndexMapping()
    indexMapping.DefaultMapping = docMapping
    return indexMapping
}
```

#### PutBlob Integration
```go
func (s *LocalStore) PutBlob(data []byte, meta BlobType) (string, error) {
    // Existing blob storage logic...
    blobID := GenerateBlobID(data)

    // Index the content
    if s.index != nil {
        doc := s.createIndexDocument(blobID, data, meta)
        if err := s.index.Index(blobID, doc); err != nil {
            // Log error but don't fail the blob storage
            // TODO: Add logging framework
        }
    }

    return blobID, nil
}

func (s *LocalStore) createIndexDocument(blobID string, data []byte, meta BlobType) IndexDocument {
    content := s.extractTextContent(data, meta)
    return IndexDocument{
        BlobID:        blobID,
        Name:          meta.Name,
        URI:           meta.URI,
        Owner:         meta.Owner,
        Content:       content,
        IngestionTime: meta.IngestionTime,
        Length:        meta.Length,
    }
}
```

### 4. Text Extraction Strategy

#### Content Type Detection and Extraction
```go
func (s *LocalStore) extractTextContent(data []byte, meta BlobType) string {
    // Phase 1: Simple text extraction
    if s.isTextContent(data, meta) {
        return string(data)
    }

    // Phase 2: Binary content - extract filename and metadata as searchable text
    return fmt.Sprintf("%s %s %s", meta.Name, meta.URI, meta.Owner)
}

func (s *LocalStore) isTextContent(data []byte, meta BlobType) bool {
    // Simple heuristic: if most bytes are printable ASCII, treat as text
    if len(data) == 0 {
        return false
    }

    textBytes := 0
    sampleSize := min(len(data), 1024) // Sample first 1KB
    for i := 0; i < sampleSize; i++ {
        b := data[i]
        if (b >= 32 && b <= 126) || b == 9 || b == 10 || b == 13 {
            textBytes++
        }
    }

    return float64(textBytes)/float64(sampleSize) > 0.8
}
```

### 5. Search API Design

#### New Interface Methods
```go
type Store interface {
    // Existing methods...

    // New search methods
    Search(query string) ([]SearchResult, error)
    SearchWithOptions(req SearchRequest) ([]SearchResult, error)
}

type SearchRequest struct {
    Query      string
    Limit      int
    Offset     int
    Fields     []string  // Fields to return in results
    Highlight  bool      // Whether to include highlighting
}

type SearchResult struct {
    BlobID     string
    Metadata   BlobType
    Score      float64
    Highlights map[string][]string  // Field -> highlighted fragments
}
```

#### Search Implementation
```go
func (s *LocalStore) Search(query string) ([]SearchResult, error) {
    req := SearchRequest{
        Query:     query,
        Limit:     50,
        Offset:    0,
        Highlight: true,
    }
    return s.SearchWithOptions(req)
}

func (s *LocalStore) SearchWithOptions(req SearchRequest) ([]SearchResult, error) {
    if s.index == nil {
        return nil, errors.New("search index not available")
    }

    // Build Bleve query
    query := bleve.NewQueryStringQuery(req.Query)
    searchRequest := bleve.NewSearchRequestOptions(query, req.Limit, req.Offset, false)

    // Configure highlighting
    if req.Highlight {
        searchRequest.Highlight = bleve.NewHighlight()
    }

    // Execute search
    searchResult, err := s.index.Search(searchRequest)
    if err != nil {
        return nil, fmt.Errorf("search failed: %w", err)
    }

    // Convert results
    results := make([]SearchResult, 0, len(searchResult.Hits))
    for _, hit := range searchResult.Hits {
        blobID := hit.ID

        // Get metadata from blob store
        _, meta, err := s.GetBlob(blobID)
        if err != nil {
            continue // Skip if blob no longer exists
        }

        result := SearchResult{
            BlobID:     blobID,
            Metadata:   meta,
            Score:      hit.Score,
            Highlights: hit.Fragments,
        }
        results = append(results, result)
    }

    return results, nil
}
```

### 6. Delete Integration

#### DeleteBlob Enhancement
```go
func (s *LocalStore) DeleteBlob(blobID string) error {
    // Remove from index first
    if s.index != nil {
        if err := s.index.Delete(blobID); err != nil {
            // Log error but continue with blob deletion
        }
    }

    // Existing deletion logic...
    return nil
}
```

### 7. CLI Integration

#### New CLI Commands
```go
func main() {
    // Existing commands...
    case "search":
        searchCmd(os.Args[2:])
}

func searchCmd(args []string) {
    fs := flag.NewFlagSet("search", flag.ExitOnError)
    storePath := fs.String("store", "./store", "Store directory")
    key := fs.String("key", "", "Encryption key (optional)")
    limit := fs.Int("limit", 10, "Maximum results")
    highlight := fs.Bool("highlight", true, "Show highlighted matches")
    fs.Parse(args)

    if fs.NArg() < 1 {
        fmt.Println("Usage: gblobs search <query> [flags]")
        os.Exit(1)
    }

    query := fs.Arg(0)
    st := openStoreOrDie(*storePath, *key)

    results, err := st.Search(query)
    if err != nil {
        fmt.Printf("Search error: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("Found %d results:\n\n", len(results))
    for i, result := range results {
        fmt.Printf("%d. %s [%s] (score: %.3f)\n",
            i+1, result.Metadata.Name, result.BlobID[:16], result.Score)
        fmt.Printf("   %s\n", result.Metadata.URI)
        fmt.Printf("   %d bytes, %s\n",
            result.Metadata.Length,
            result.Metadata.IngestionTime.Format("2006-01-02 15:04:05"))

        if *highlight && len(result.Highlights) > 0 {
            for field, fragments := range result.Highlights {
                fmt.Printf("   %s: %s\n", field, strings.Join(fragments, " ... "))
            }
        }
        fmt.Println()
    }
}
```

## Implementation Phases

### Phase 1: Core Integration
- [ ] Add Bleve dependency to go.mod
- [ ] Extend LocalStore with index field
- [ ] Implement index lifecycle in CreateStore/OpenStore/PurgeStore
- [ ] Basic text extraction and indexing in PutBlob
- [ ] Index cleanup in DeleteBlob

### Phase 2: Search API
- [ ] Implement Search and SearchWithOptions methods
- [ ] Add SearchResult types
- [ ] CLI search command
- [ ] Basic testing

### Phase 3: Advanced Features
- [ ] Improved text extraction (PDF, Word, etc.)
- [ ] Search result highlighting
- [ ] Field-specific searches
- [ ] Date range queries
- [ ] Faceted search

### Phase 4: Performance & Production
- [ ] Index optimization
- [ ] Batch indexing improvements
- [ ] Error handling and logging
- [ ] Performance benchmarks
- [ ] Documentation updates

## Dependencies

```go
// go.mod additions
require (
    github.com/blevesearch/bleve/v2 v2.3.10
)
```

## Testing Strategy

1. **Unit Tests**
   - Index lifecycle operations
   - Text extraction logic
   - Search result conversion

2. **Integration Tests**
   - End-to-end blob storage and search
   - CLI search command testing
   - Index consistency after operations

3. **Performance Tests**
   - Large dataset indexing
   - Search performance benchmarks
   - Memory usage analysis

## Migration Strategy

- New stores automatically get search capabilities
- Existing stores require reindexing:
  ```bash
  gblobs reindex --store ./existing_store
  ```

## Error Handling

- Index failures should not prevent blob storage
- Search unavailable when index is corrupted
- Automatic index rebuilding option
- Graceful degradation when search is disabled