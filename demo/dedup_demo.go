package main

import (
    "fmt"
    "os"
    "github.com/example/gblobs/gblobs"
)

func main() {
    storePath := "tmp_gblobs_dedup"
    _ = os.RemoveAll(storePath)
    st := &gblobs.LocalStore{}
    _ = st.CreateStore(storePath)
    data := []byte("same content")
    // First put
    id1, _ := st.PutBlob(data, gblobs.BlobType{Name: "one"})
    // Second put with identical data
    id2, _ := st.PutBlob(data, gblobs.BlobType{Name: "dup"})
    if id1 != id2 {
        fmt.Printf("Deduplication failed: got %s and %s\n", id1, id2)
    } else {
        fmt.Printf("Blob deduplicated, id: %s\n", id1)
    }
    // Count blobs (should be 1)
    stats, _ := st.Stats()
    fmt.Printf("Total blobs after dedup: %d\n", stats.TotalBlobCount)
}
