package gblobs

import (
    "crypto/sha256"
    "encoding/hex"
)

// GenerateBlobID computes SHA-256 over input data and returns the hex-encoded blob id.
func GenerateBlobID(data []byte) string {
    h := sha256.Sum256(data)
    return hex.EncodeToString(h[:])
}

// BlobIDToPath maps a blob id to a store file path using subdirectory grouping,
// yielding a structure ba/781/6bf/xxxxxx.blob for input id ba7816bf8f01cfea414140de5.
func BlobIDToPath(blobID string) string {
    // Sanity: must be at least 8 chars for three levels
    if len(blobID) < 8 {
        return blobID + ".blob"
    }
    lvl1 := blobID[:2]
    lvl2 := blobID[2:5]
    lvl3 := blobID[5:8]
    rest := blobID[8:]
    return lvl1 + "/" + lvl2 + "/" + lvl3 + "/" + rest + ".blob"
}
