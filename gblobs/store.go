package gblobs

import (
    "errors"
    "os"
    "sync"
    "path/filepath"
    "encoding/json"
    "fmt"
    "strings"
    "time"

    "github.com/blevesearch/bleve/v2"
    "github.com/blevesearch/bleve/v2/mapping"
)

// LocalStore provides blob storage with full-text search capabilities
type LocalStore struct {
    path      string
    key       []byte
    lock      sync.RWMutex
    index     bleve.Index  // Bleve full-text search index
    indexPath string       // Path to the search index
}


// Stats walks the folder structure counting blobs and reporting organization stats.
// - TotalBlobCount: number of blobs
// - MaxCountPerLevel: per-level max file/directory count
// - AverageCountPerLevel: average per level
func (s *LocalStore) Stats() (StoreStats, error) {
    // Traverse recursively to count .blob files and per-directory file counts
    type levelStats struct {
        counts []int
        total  int
    }
    countsPerLevel := map[int][]int{} // map[level]counts (number of entries per parent at each level)
    var totalBlobs int
    root := s.path
    maxDepth := 0
    var visit func(dir string, depth int) error
    visit = func(dir string, depth int) error {
        fis, err := os.ReadDir(dir)
        if err != nil {
            return err
        }
        cnt := 0
        for _, fi := range fis {
            full := filepath.Join(dir, fi.Name())
            if fi.IsDir() {
                cnt++ // Count directories at this level
                if depth+1 > maxDepth {
                    maxDepth = depth+1
                }
                err = visit(full, depth+1)
                if err != nil {
                    return err
                }
            } else {
                // Count all files at this level
                if filepath.Ext(fi.Name()) == ".blob" {
                    totalBlobs++
                }
                cnt++ // Count all files (both .blob and .meta)
            }
        }
        // Always record the count for this directory level
        if len(fis) > 0 {
            countsPerLevel[depth] = append(countsPerLevel[depth], cnt)
        }
        return nil
    }
    if err := visit(root, 0); err != nil {
        return StoreStats{}, err
    }
    // For each level, tally max & total for avg
    maxPerLevel := []int{}
    var totalCount, totalDirectories int
    for l := 0; l <= maxDepth; l++ {
        vals := countsPerLevel[l]
        if len(vals) == 0 {
            maxPerLevel = append(maxPerLevel, 0)
            continue
        }
        mx := vals[0]
        lvlTotal := 0
        for _, v := range vals {
            if v > mx {
                mx = v
            }
            lvlTotal += v
        }
        maxPerLevel = append(maxPerLevel, mx)
        totalCount += lvlTotal
        totalDirectories += len(vals) // Count directories at this level
    }
    avgPer := 0.0
    if totalDirectories > 0 {
        avgPer = float64(totalCount) / float64(totalDirectories)
    }
    return StoreStats{
        TotalBlobCount: totalBlobs,
        MaxCountPerLevel: maxPerLevel,
        AverageCountPerLevel: avgPer,
    }, nil
}

// createIndexMapping creates the Bleve index mapping for blob documents
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

    // Numeric field mapping
    numericFieldMapping := bleve.NewNumericFieldMapping()
    numericFieldMapping.Store = true

    // Document mapping
    docMapping := bleve.NewDocumentMapping()
    docMapping.AddFieldMappingsAt("content", textFieldMapping)
    docMapping.AddFieldMappingsAt("name", textFieldMapping)
    docMapping.AddFieldMappingsAt("uri", keywordFieldMapping)
    docMapping.AddFieldMappingsAt("owner", keywordFieldMapping)
    docMapping.AddFieldMappingsAt("blobId", keywordFieldMapping)
    docMapping.AddFieldMappingsAt("ingestionTime", dateFieldMapping)
    docMapping.AddFieldMappingsAt("length", numericFieldMapping)

    // Index mapping
    indexMapping := bleve.NewIndexMapping()
    indexMapping.DefaultMapping = docMapping
    return indexMapping
}

// closeIndex safely closes the search index
func (s *LocalStore) closeIndex() error {
    if s.index != nil {
        err := s.index.Close()
        s.index = nil
        return err
    }
    return nil
}

// isTextContent uses heuristics to determine if data contains text
func (s *LocalStore) isTextContent(data []byte, meta BlobType) bool {
    if len(data) == 0 {
        return false
    }

    // Check file extension hints
    ext := strings.ToLower(filepath.Ext(meta.Name))
    textExtensions := map[string]bool{
        ".txt": true, ".md": true, ".json": true, ".xml": true,
        ".html": true, ".css": true, ".js": true, ".py": true,
        ".go": true, ".java": true, ".c": true, ".cpp": true,
        ".h": true, ".yml": true, ".yaml": true, ".csv": true,
        ".log": true, ".sql": true, ".sh": true, ".bat": true,
        ".ini": true, ".conf": true, ".cfg": true,
    }
    if textExtensions[ext] {
        return true
    }

    // Heuristic: if most bytes are printable ASCII, treat as text
    textBytes := 0
    sampleSize := len(data)
    if sampleSize > 1024 {
        sampleSize = 1024 // Sample first 1KB
    }

    for i := 0; i < sampleSize; i++ {
        b := data[i]
        if (b >= 32 && b <= 126) || b == 9 || b == 10 || b == 13 {
            textBytes++
        }
    }

    return float64(textBytes)/float64(sampleSize) > 0.8
}

// extractTextContent extracts searchable text from blob data
func (s *LocalStore) extractTextContent(data []byte, meta BlobType) string {
    var content strings.Builder

    // Always include metadata as searchable text
    if meta.Name != "" {
        content.WriteString(meta.Name)
        content.WriteString(" ")
    }
    if meta.Owner != "" {
        content.WriteString(meta.Owner)
        content.WriteString(" ")
    }
    if meta.URI != "" {
        content.WriteString(filepath.Base(meta.URI))
        content.WriteString(" ")
    }

    // If it's text content, include the actual data
    if s.isTextContent(data, meta) {
        content.WriteString(string(data))
    }

    return content.String()
}

// createIndexDocument creates a document for the search index
func (s *LocalStore) createIndexDocument(blobID string, data []byte, meta BlobType) IndexDocument {
    content := s.extractTextContent(data, meta)
    return IndexDocument{
        BlobID:        blobID,
        Name:          meta.Name,
        URI:           meta.URI,
        Owner:         meta.Owner,
        Content:       content,
        IngestionTime: meta.IngestionTime,
        Length:        int64(len(data)),
    }
}

// Compile check
var _ Store = (*LocalStore)(nil)

// Dummy implementations for interface compliance

// CreateStore sets up directory and records key for new store
func (s *LocalStore) CreateStore(path string, keyOpt ...string) error {
    s.path = path
    if len(keyOpt) > 0 {
        s.key = KeyFromPassword(keyOpt[0])
    } else {
        s.key = nil
    }

    // Try to create base directory
    if err := EnsureDir(path); err != nil {
        return err
    }

    // Create Bleve search index
    s.indexPath = filepath.Join(path, "index.bleve")
    mapping := s.createIndexMapping()
    index, err := bleve.New(s.indexPath, mapping)
    if err != nil {
        return fmt.Errorf("failed to create search index: %w", err)
    }
    s.index = index

    return nil
}

// OpenStore loads existing store, with/without key
func (s *LocalStore) OpenStore(path string, keyOpt ...string) error {
    s.path = path
    if len(keyOpt) > 0 {
        s.key = KeyFromPassword(keyOpt[0])
    } else {
        s.key = nil
    }

    // Check base directory exists
    info, err := os.Stat(path)
    if err != nil {
        return err
    }
    if !info.IsDir() {
        return errors.New("store path exists but is not a directory")
    }

    // Open existing Bleve search index
    s.indexPath = filepath.Join(path, "index.bleve")
    if _, err := os.Stat(s.indexPath); err == nil {
        // Index exists, open it
        index, err := bleve.Open(s.indexPath)
        if err != nil {
            return fmt.Errorf("failed to open search index: %w", err)
        }
        s.index = index
    } else {
        // Index doesn't exist, create it for backward compatibility
        mapping := s.createIndexMapping()
        index, err := bleve.New(s.indexPath, mapping)
        if err != nil {
            return fmt.Errorf("failed to create search index: %w", err)
        }
        s.index = index
    }

    return nil
}

// PutBlob stores data in compressed (and optionally encrypted) form, avoids duplicates
func (s *LocalStore) PutBlob(data []byte, meta BlobType) (string, error) {
    blobID := GenerateBlobID(data)
    relPath := BlobIDToPath(blobID)
    fullPath := filepath.Join(s.path, relPath)
    dir := filepath.Dir(fullPath)
    if err := EnsureDir(dir); err != nil {
        return "", err
    }
    s.lock.Lock()
    defer s.lock.Unlock()

    // Check if blob already exists (deduplication)
    alreadyExists := false
    if _, err := os.Stat(fullPath); err == nil {
        alreadyExists = true
    }

    // If blob doesn't exist, create it
    if !alreadyExists {
        // Compress, encrypt
        content := data
        var err error
        content, err = CompressBlob(content)
        if err != nil {
            return "", err
        }
        content, err = EncryptBlob(content, s.key)
        if err != nil {
            return "", err
        }
        // Write to file
        if err := os.WriteFile(fullPath, content, 0o600); err != nil {
            return "", err
        }
        // Metadata file
        meta.BlobHash = blobID
        meta.Length = int64(len(data))
        metaFile := fullPath + ".meta"
        meta.IngestionTime = meta.IngestionTime.UTC()
        mb, err := json.MarshalIndent(meta, "", "  ")
        if err != nil {
            return "", err
        }
        if err := os.WriteFile(metaFile, mb, 0o600); err != nil {
            return "", err
        }
    } else {
        // Blob exists, but we still need to set metadata for indexing
        meta.BlobHash = blobID
        meta.Length = int64(len(data))
        if meta.IngestionTime.IsZero() {
            meta.IngestionTime = time.Now().UTC()
        }
    }

    // Always index the content (for new blobs or when metadata changes)
    if s.index != nil {
        doc := s.createIndexDocument(blobID, data, meta)
        if err := s.index.Index(blobID, doc); err != nil {
            // Log error but don't fail the blob storage operation
            fmt.Printf("Warning: failed to index blob %s: %v\n", blobID, err)
        }
    }

    return blobID, nil
}

// GetBlob loads the blob (decrypt, decompress) and returns with metadata
func (s *LocalStore) GetBlob(blobID string) ([]byte, BlobType, error) {
    relPath := BlobIDToPath(blobID)
    fullPath := filepath.Join(s.path, relPath)
    s.lock.RLock()
    defer s.lock.RUnlock()
    encContent, err := os.ReadFile(fullPath)
    if err != nil {
        return nil, BlobType{}, err
    }
    content, err := DecryptBlob(encContent, s.key)
    if err != nil {
        return nil, BlobType{}, err
    }
    plain, err := DecompressBlob(content)
    if err != nil {
        return nil, BlobType{}, err
    }
    // Metadata
    metaFile := fullPath + ".meta"
    mb, err := os.ReadFile(metaFile)
    if err != nil {
        return nil, BlobType{}, err
    }
    var meta BlobType
    if err := json.Unmarshal(mb, &meta); err != nil {
        return nil, BlobType{}, err
    }
    return plain, meta, nil
}

// ExistsBlob returns true if the main blob file exists
func (s *LocalStore) ExistsBlob(blobID string) (bool, error) {
    relPath := BlobIDToPath(blobID)
    fullPath := filepath.Join(s.path, relPath)
    _, err := os.Stat(fullPath)
    if err == nil {
        return true, nil
    }
    if os.IsNotExist(err) {
        return false, nil
    }
    return false, err
}

// DeleteBlob removes both the blob file and its metadata
func (s *LocalStore) DeleteBlob(blobID string) error {
    // Remove from search index first
    if s.index != nil {
        if err := s.index.Delete(blobID); err != nil {
            // Log error but continue with blob deletion
            fmt.Printf("Warning: failed to remove blob %s from search index: %v\n", blobID, err)
        }
    }

    // Remove blob and metadata files
    relPath := BlobIDToPath(blobID)
    fullPath := filepath.Join(s.path, relPath)
    metaFile := fullPath + ".meta"
    // Remove both files; ignore their absence
    _ = os.Remove(fullPath)
    _ = os.Remove(metaFile)
    return nil
}

// PurgeStore removes all blobs and meta files from store (very destructive)
func (s *LocalStore) PurgeStore() error {
    // Close and remove search index first
    if err := s.closeIndex(); err != nil {
        // Log error but continue with purge
        fmt.Printf("Warning: failed to close search index: %v\n", err)
    }

    // Remove all contents in store path; do not remove the root dir
    d, err := os.ReadDir(s.path)
    if err != nil {
        return err
    }
    for _, f := range d {
        err := os.RemoveAll(filepath.Join(s.path, f.Name()))
        if err != nil {
            return err
        }
    }

    // Recreate the search index after purging
    s.indexPath = filepath.Join(s.path, "index.bleve")
    mapping := s.createIndexMapping()
    index, err := bleve.New(s.indexPath, mapping)
    if err != nil {
        return fmt.Errorf("failed to recreate search index after purge: %w", err)
    }
    s.index = index

    return nil
}
// Path returns the store's root path (for testing/internal use only).
func (s *LocalStore) Path() string {
    return s.path
}

// InspectStore traverses all paths and collects metadata of all blobs in the store.
// Returns a slice of BlobType structs containing metadata for each blob.
func (s *LocalStore) InspectStore() ([]BlobType, error) {
    var blobs []BlobType
    root := s.path

    var visit func(dir string) error
    visit = func(dir string) error {
        fis, err := os.ReadDir(dir)
        if err != nil {
            return err
        }

        for _, fi := range fis {
            full := filepath.Join(dir, fi.Name())
            if fi.IsDir() {
                err = visit(full)
                if err != nil {
                    return err
                }
            } else if filepath.Ext(fi.Name()) == ".blob" {
                // Found a blob file, read its metadata
                metaFile := full + ".meta"
                mb, err := os.ReadFile(metaFile)
                if err != nil {
                    // Skip blobs without metadata files
                    continue
                }
                var meta BlobType
                if err := json.Unmarshal(mb, &meta); err != nil {
                    // Skip blobs with corrupted metadata
                    continue
                }
                blobs = append(blobs, meta)
            }
        }
        return nil
    }

    if err := visit(root); err != nil {
        return nil, err
    }

    return blobs, nil
}

// Search performs a simple text search across all indexed blobs
func (s *LocalStore) Search(query string) ([]SearchResult, error) {
    req := SearchRequest{
        Query:     query,
        Limit:     50,
        Offset:    0,
        Highlight: true,
    }
    return s.SearchWithOptions(req)
}

// SearchWithOptions performs a search with detailed configuration options
func (s *LocalStore) SearchWithOptions(req SearchRequest) ([]SearchResult, error) {
    if s.index == nil {
        return nil, errors.New("search index not available")
    }

    // Set default values
    if req.Limit <= 0 {
        req.Limit = 50
    }
    if req.Offset < 0 {
        req.Offset = 0
    }

    // Build Bleve query
    query := bleve.NewQueryStringQuery(req.Query)
    searchRequest := bleve.NewSearchRequestOptions(query, req.Limit, req.Offset, false)

    // Configure highlighting
    if req.Highlight {
        searchRequest.Highlight = bleve.NewHighlight()
        searchRequest.Highlight.AddField("content")
        searchRequest.Highlight.AddField("name")
    }

    // Configure fields to return
    if len(req.Fields) > 0 {
        searchRequest.Fields = req.Fields
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
            // Skip if blob no longer exists or can't be read
            continue
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
