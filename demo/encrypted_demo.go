package main

import (
    "fmt"
    "os"
    "time"
    "github.com/example/gblobs/gblobs"
)

func main() {
    // Create an encrypted store
    storePath := "tmp_gblobs_encstore"
    _ = os.RemoveAll(storePath)
    st := &gblobs.LocalStore{}
    key := "enc_key"
    err := st.CreateStore(storePath, key)
    if err != nil { panic(err) }
    data := []byte("super secret")
    meta := gblobs.BlobType{Name: "secret.txt", IngestionTime: time.Now().UTC()}
    blobID, err := st.PutBlob(data, meta)
    if err != nil { panic(err) }
    fmt.Println("Encrypted blob id:", blobID)

    // Open with correct key
    st2 := &gblobs.LocalStore{}
    err = st2.OpenStore(storePath, key)
    if err != nil { panic(err) }
    plain, meta2, err := st2.GetBlob(blobID)
    fmt.Printf("Decrypted content: %q (meta: %v)\n", plain, meta2)
}
