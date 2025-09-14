package test

import (
    "os"
    "testing"
    "time"
    "strings"
    "path/filepath"
    "github.com/example/gblobs/gblobs"
)

func quickStore(t *testing.T, encrypted bool) *gblobs.LocalStore {
    dir := t.TempDir()
    store := &gblobs.LocalStore{}
    var err error
    key := ""
    if encrypted {
        key = "pw1234"
        err = store.CreateStore(dir, key)
    } else {
        err = store.CreateStore(dir)
    }
    if err != nil {
        t.Fatalf("fail create store: %v", err)
    }
    return store
}

func TestPutGetExistsDeleteBlob_Plain(t *testing.T) {
    store := quickStore(t, false)
    payload := []byte("blob123")
    meta := gblobs.BlobType{
        Name: "test123",
        URI: "test-uri",
        Owner: "me",
        IngestionTime: time.Now().UTC(),
    }
    id, err := store.PutBlob(payload, meta)
    if err != nil {
        t.Fatal(err)
    }
    exists, err := store.ExistsBlob(id)
    if err != nil || !exists {
        t.Fatal("should exist after put")
    }
    got, gotMeta, err := store.GetBlob(id)
    if err != nil {
        t.Fatal(err)
    }
    if string(got) != string(payload) {
        t.Fatalf("bad data roundtrip: got %q want %q", got, payload)
    }
    if gotMeta.Name != "test123" {
        t.Error("metadata roundtrip fail")
    }
    // Delete
    err = store.DeleteBlob(id)
    if err != nil {
        t.Fatal(err)
    }
    exists, err = store.ExistsBlob(id)
    if exists {
        t.Error("should not exist after delete")
    }
    // Delete again (absent): should not error
    err = store.DeleteBlob(id)
    if err != nil {
        t.Fatal(err)
    }
}

func TestPutGetBlob_Encrypted(t *testing.T) {
    store := quickStore(t, true)
    data := []byte("supersecretblob")
    id, err := store.PutBlob(data, gblobs.BlobType{Name: "encrypt1"})
    if err != nil {
        t.Fatal(err)
    }
    got, _, err := store.GetBlob(id)
    if err != nil {
        t.Fatal(err)
    }
    if string(got) != string(data) {
        t.Errorf("data mismatch got %q", got)
    }
}

func TestPurgeStore(t *testing.T) {
    store := quickStore(t, false)
    // Add two blobs
    _, err := store.PutBlob([]byte("a"), gblobs.BlobType{Name: "n1"})
    if err != nil { t.Fatal(err) }
    _, err = store.PutBlob([]byte("b"), gblobs.BlobType{Name: "n2"})
    if err != nil { t.Fatal(err) }
    stats, err := store.Stats()
    if err != nil { t.Fatal(err) }
    if stats.TotalBlobCount != 2 {
        t.Errorf("before purge: expect 2 blobs")
    }
    err = store.PurgeStore()
    if err != nil { t.Fatal(err) }
    stats2, err := store.Stats()
    if err != nil { t.Fatal(err) }
    if stats2.TotalBlobCount != 0 {
        t.Errorf("after purge: expect 0 blobs")
    }
}

func TestStatsVariousLevels(t *testing.T) {
    store := quickStore(t, false)
    for i := 0; i < 7; i++ {
        id := gblobs.GenerateBlobID([]byte(strings.Repeat("X", i+1)))
        rel := gblobs.BlobIDToPath(id)
        // Create nested dirs if needed
        full := filepath.Join(store.Path(), rel)
        if err := gblobs.EnsureDir(filepath.Dir(full)); err != nil { t.Fatal(err) }
        if err := os.WriteFile(full, []byte("test"), 0o600); err != nil { t.Fatal(err) }
    }
    stats, err := store.Stats()
    if err != nil { t.Fatal(err) }
    if stats.TotalBlobCount != 7 {
        t.Errorf("want 7 blobs, got %v", stats.TotalBlobCount)
    }
    // Should have some max/maxPerLevel/avg value
    if len(stats.MaxCountPerLevel) == 0 || stats.AverageCountPerLevel == 0.0 {
        t.Error("stats missing levels/avgs")
    }
}
