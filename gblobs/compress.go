package gblobs

import (
    "bytes"
    "compress/gzip"
    "io"
)

// CompressBlob compresses data using gzip. Suitable for storing in the blob store.
func CompressBlob(data []byte) ([]byte, error) {
    var buf bytes.Buffer
    zw := gzip.NewWriter(&buf)
    _, err := zw.Write(data)
    if err != nil {
        zw.Close()
        return nil, err
    }
    if err := zw.Close(); err != nil {
        return nil, err
    }
    return buf.Bytes(), nil
}

// DecompressBlob decompresses data previously written with CompressBlob.
func DecompressBlob(compData []byte) ([]byte, error) {
    zr, err := gzip.NewReader(bytes.NewReader(compData))
    if err != nil {
        return nil, err
    }
    defer zr.Close()
    return io.ReadAll(zr)
}
