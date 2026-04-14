package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

const (
	nonceSize = 12 // AES-GCM standard nonce size
	keySize   = 32 // AES-256
)

// EncryptGCM encrypts plaintext with AES-256-GCM.
// Output format: nonce (12 bytes) || ciphertext || tag (16 bytes).
func EncryptGCM(key, plaintext []byte) ([]byte, error) {
	if len(key) != keySize {
		return nil, fmt.Errorf("key must be %d bytes, got %d", keySize, len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}

	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	// Seal appends ciphertext+tag after nonce
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// DecryptGCM decrypts AES-256-GCM ciphertext.
// Input format: nonce (12 bytes) || ciphertext || tag (16 bytes).
func DecryptGCM(key, ciphertext []byte) ([]byte, error) {
	if len(key) != keySize {
		return nil, fmt.Errorf("key must be %d bytes, got %d", keySize, len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}

	if len(ciphertext) < nonceSize+gcm.Overhead() {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce := ciphertext[:nonceSize]
	encrypted := ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt failed: invalid key or corrupted data")
	}
	return plaintext, nil
}

// UnwrapKey derives the wrapping key via HKDF, then decrypts the wrapped key.
// This matches the server-side key wrapping flow.
func UnwrapKey(token, machineID string, wrappedKey []byte) ([]byte, error) {
	wrapKey, err := DeriveKey(
		[]byte(token),
		[]byte(machineID),
		[]byte("claudex-wrap"),
		keySize,
	)
	if err != nil {
		return nil, fmt.Errorf("derive wrap key: %w", err)
	}

	encryptionKey, err := DecryptGCM(wrapKey, wrappedKey)
	if err != nil {
		return nil, fmt.Errorf("unwrap key: %w", err)
	}
	return encryptionKey, nil
}
