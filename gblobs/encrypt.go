package gblobs

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "errors"
    "io"
)

// EncryptBlob encrypts data using AES-GCM with the provided key.
// If key is empty, no encryption is performed and data is returned unchanged.
func EncryptBlob(data []byte, key []byte) ([]byte, error) {
    if len(key) == 0 {
        // No encryption requested
        return data, nil
    }
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }
    out := gcm.Seal(nonce, nonce, data, nil)
    return out, nil
}

// DecryptBlob decrypts data encrypted by EncryptBlob with the provided key.
// If key is empty, no decryption is performed and data is returned unchanged.
func DecryptBlob(encData []byte, key []byte) ([]byte, error) {
    if len(key) == 0 {
        // No decryption requested
        return encData, nil
    }
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    nonceSize := gcm.NonceSize()
    if len(encData) < nonceSize {
        return nil, errors.New("ciphertext too short")
    }
    nonce, ciphertext := encData[:nonceSize], encData[nonceSize:]
    return gcm.Open(nil, nonce, ciphertext, nil)
}

// KeyFromPassword provides a (very basic) password-to-key function for examples; it simply pads/cuts
// the password bytes to 32 bytes (AES-256). For serious use, substitute a real KDF.
func KeyFromPassword(pw string) []byte {
    // For robust security, use scrypt/argon2, but for now simple pad/cut.
    b := []byte(pw)
    key := make([]byte, 32)
    copy(key, b)
    return key
}
