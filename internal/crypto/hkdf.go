// Package crypto provides AES-256-GCM encryption and HKDF key derivation.
package crypto

import (
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

// DeriveKey uses HKDF-SHA256 to derive a key of the given length.
func DeriveKey(secret, salt, info []byte, length int) ([]byte, error) {
	reader := hkdf.New(sha256.New, secret, salt, info)
	key := make([]byte, length)
	if _, err := io.ReadFull(reader, key); err != nil {
		return nil, fmt.Errorf("hkdf derive: %w", err)
	}
	return key, nil
}
