package gblobs

import (
    "os"
    "path/filepath"
)

// EnsureDir creates the directory (including parents) if it does not exist.
func EnsureDir(path string) error {
    return os.MkdirAll(path, 0o755)
}

// GetDirOfBlobID returns the directory path for a blobID according to the store layout.
func GetDirOfBlobID(blobID string) string {
    p := BlobIDToPath(blobID)
    dir := filepath.Dir(p)
    return dir
}
