package test

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"
    "github.com/example/gblobs/gblobs"
)

func TestGenerateBlobID(t *testing.T) {
    c1 := []byte("test data")
    id1 := gblobs.GenerateBlobID(c1)
    id2 := gblobs.GenerateBlobID(c1)
    if id1 != id2 {
        t.Error("hash must be deterministic for same data")
    }
    c2 := []byte("other data")
    if id1 == gblobs.GenerateBlobID(c2) {
        t.Error("hash must differ for different data")
    }
    empty := []byte{}
    idEmpty := gblobs.GenerateBlobID(empty)
    if idEmpty == "" {
        t.Error("empty hash must still return a value")
    }
}

func TestBlobIDToPath(t *testing.T) {
    b := "ba7816bf8f01cfea414140de5"
    path := gblobs.BlobIDToPath(b)
    exp := "ba/781/6bf/8f01cfea414140de5.blob"
    if path != exp {
        t.Errorf("want %q got %q", exp, path)
    }
    // short id
    short := "abc"
    p2 := gblobs.BlobIDToPath(short)
    if p2 != short+".blob" {
        t.Error("short id fallback failure")
    }
}

func TestEnsureDir_Idempotent(t *testing.T) {
    tmp := t.TempDir()
    sub := filepath.Join(tmp, "foo/bar")
    if err := gblobs.EnsureDir(sub); err != nil {
        t.Fatal(err)
    }
    // Should not fail if called again
    if err := gblobs.EnsureDir(sub); err != nil {
        t.Fatal(err)
    }
    fi, err := os.Stat(sub)
    if err != nil || !fi.IsDir() {
        t.Fatal("directory creation failed")
    }
}

func TestCompressionRoundTrip(t *testing.T) {
    raw := []byte("the quick brown fox jumps over the lazy dog")
    comp, err := gblobs.CompressBlob(raw)
    if err != nil {
        t.Fatal(err)
    }
    decomp, err := gblobs.DecompressBlob(comp)
    if err != nil {
        t.Fatal(err)
    }
    if !bytes.Equal(raw, decomp) {
        t.Error("compression round trip mismatch")
    }
    // Empty case
    emp, err := gblobs.CompressBlob([]byte{})
    if err != nil {
        t.Fatal(err)
    }
    decEmp, err := gblobs.DecompressBlob(emp)
    if err != nil {
        t.Fatal(err)
    }
    if len(decEmp) != 0 {
        t.Error("decompressed empty not empty")
    }
    // Incompressible data (already compressed)
    junk := make([]byte, 1024)
    for i := range junk { junk[i] = byte(i%256) }
    c2, err := gblobs.CompressBlob(junk)
    if err != nil {
        t.Fatal(err)
    }
    d2, err := gblobs.DecompressBlob(c2)
    if err != nil {
        t.Fatal(err)
    }
    if !bytes.Equal(junk, d2) {
        t.Error("incompressible roundtrip fail")
    }
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
    pt := []byte("super secret stuff")
    key := gblobs.KeyFromPassword("hunter2")
    ct, err := gblobs.EncryptBlob(pt, key)
    if err != nil {
        t.Fatal(err)
    }
    pt2, err := gblobs.DecryptBlob(ct, key)
    if err != nil {
        t.Fatal(err)
    }
    if !bytes.Equal(pt, pt2) {
        t.Error("encryption round trip mismatch")
    }
    // Empty key = no encryption
    pt3 := []byte("barbaz")
    ct3, err := gblobs.EncryptBlob(pt3, nil)
    if err != nil {
        t.Fatal(err)
    }
    dec3, err := gblobs.DecryptBlob(ct3, nil)
    if err != nil {
        t.Fatal(err)
    }
    if !bytes.Equal(dec3, pt3) {
        t.Error("nil key roundtrip failed")
    }
    // Wrong key
    wrong := gblobs.KeyFromPassword("password")
    _, err = gblobs.DecryptBlob(ct, wrong)
    if err == nil {
        t.Error("expected failure for wrong key")
    }
    // Short ciphertext (invalid input)
    _, err = gblobs.DecryptBlob([]byte{1, 2}, key)
    if err == nil {
        t.Error("expected failure for short ciphertext")
    }
    // Empty input
    em, err := gblobs.EncryptBlob([]byte{}, key)
    if err != nil {
        t.Fatal(err)
    }
    out, err := gblobs.DecryptBlob(em, key)
    if err != nil {
        t.Fatal(err)
    }
    if len(out) != 0 {
        t.Error("expected empty output")
    }
}
