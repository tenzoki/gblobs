package gblobs

import (
    "errors"
    "os"
    "sync"
    "path/filepath"
    "encoding/json"
)

// Placeholder struct for store (actual implementation to follow in later phases)
type LocalStore struct {
    path string
    key  []byte
    lock sync.RWMutex
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
    return EnsureDir(path)
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
    // If already exists, do not write (dedup)
    if _, err := os.Stat(fullPath); err == nil {
        return blobID, nil
    }
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
    relPath := BlobIDToPath(blobID)
    fullPath := filepath.Join(s.path, relPath)
    metaFile := fullPath + ".meta"
    // Remove both files; ignore their absense
    _ = os.Remove(fullPath)
    _ = os.Remove(metaFile)
    return nil
}

// PurgeStore removes all blobs and meta files from store (very destructive)
func (s *LocalStore) PurgeStore() error {
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
