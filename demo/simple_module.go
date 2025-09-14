package main

import (
    "fmt"
    "os"
    "time"
    "github.com/example/gblobs/gblobs"
)

func main() {
    // Create a new, plain (unencrypted) store
    storePath := "tmp_gblobs_store"
    _ = os.RemoveAll(storePath)
    st := &gblobs.LocalStore{}
    err := st.CreateStore(storePath)
    if err != nil { panic(err) }

    // Put a blob
    blob := []byte("hello from gblobs demo!")
    meta := gblobs.BlobType{
        Name: "demo.txt",
        URI:  "memory://demo",
        Owner: "demo-user",
        IngestionTime: time.Now().UTC(),
    }
    blobID, err := st.PutBlob(blob, meta)
    if err != nil { panic(err) }
    fmt.Println("Stored blob with id:", blobID)

    // Get blob
    got, gotMeta, err := st.GetBlob(blobID)
    if err != nil { panic(err) }
    fmt.Printf("Got blob: %q (meta: %v)\n", got, gotMeta)

    // Check existence
    exists, _ := st.ExistsBlob(blobID)
    fmt.Println("Blob exists?", exists)

    // Show stats
    stats, _ := st.Stats()
    fmt.Printf("Stats: Total=%v, MaxPerLevel=%v, AvgPerLevel=%.2f\n", stats.TotalBlobCount, stats.MaxCountPerLevel, stats.AverageCountPerLevel)

    // Delete blob
    _ = st.DeleteBlob(blobID)
    exists, _ = st.ExistsBlob(blobID)
    fmt.Println("Blob exists after delete?", exists)
}
