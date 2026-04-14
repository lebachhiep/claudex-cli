package crypto

import (
	"bytes"
	"testing"
)

// TestEncryptDecryptRoundTrip verifies AES-GCM encrypt/decrypt round-trip.
func TestEncryptDecryptRoundTrip(t *testing.T) {
	// Generate a 32-byte key
	key := make([]byte, keySize)
	for i := 0; i < keySize; i++ {
		key[i] = byte(i % 256)
	}

	plaintext := []byte("Hello, World!")

	// Encrypt
	ciphertext, err := EncryptGCM(key, plaintext)
	if err != nil {
		t.Fatalf("EncryptGCM failed: %v", err)
	}

	// Ciphertext should contain nonce (12) + encrypted data + tag (16)
	if len(ciphertext) < nonceSize+16 {
		t.Errorf("ciphertext too short: got %d bytes, expected at least %d", len(ciphertext), nonceSize+16)
	}

	// Decrypt
	decrypted, err := DecryptGCM(key, ciphertext)
	if err != nil {
		t.Fatalf("DecryptGCM failed: %v", err)
	}

	// Verify round-trip
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted data mismatch: expected %q, got %q", plaintext, decrypted)
	}
}

// TestEncryptDecryptEmptyPlaintext verifies encryption of empty data.
func TestEncryptDecryptEmptyPlaintext(t *testing.T) {
	key := make([]byte, keySize)
	plaintext := []byte{}

	ciphertext, err := EncryptGCM(key, plaintext)
	if err != nil {
		t.Fatalf("EncryptGCM failed: %v", err)
	}

	decrypted, err := DecryptGCM(key, ciphertext)
	if err != nil {
		t.Fatalf("DecryptGCM failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted empty data mismatch: expected %q, got %q", plaintext, decrypted)
	}
}

// TestEncryptGCMInvalidKeySize verifies error on invalid key size.
func TestEncryptGCMInvalidKeySize(t *testing.T) {
	invalidKey := []byte("short")
	plaintext := []byte("data")

	_, err := EncryptGCM(invalidKey, plaintext)
	if err == nil {
		t.Error("EncryptGCM should fail with invalid key size")
	}
}

// TestDecryptGCMInvalidKeySize verifies error on invalid key size for decryption.
func TestDecryptGCMInvalidKeySize(t *testing.T) {
	invalidKey := []byte("short")
	ciphertext := make([]byte, 32)

	_, err := DecryptGCM(invalidKey, ciphertext)
	if err == nil {
		t.Error("DecryptGCM should fail with invalid key size")
	}
}

// TestDecryptGCMTooShortCiphertext verifies error on truncated ciphertext.
func TestDecryptGCMTooShortCiphertext(t *testing.T) {
	key := make([]byte, keySize)
	tooShort := []byte("short")

	_, err := DecryptGCM(key, tooShort)
	if err == nil {
		t.Error("DecryptGCM should fail with too short ciphertext")
	}
}

// TestDecryptGCMCorruptedData verifies error on corrupted ciphertext.
func TestDecryptGCMCorruptedData(t *testing.T) {
	key := make([]byte, keySize)
	plaintext := []byte("test data")

	// Encrypt
	ciphertext, err := EncryptGCM(key, plaintext)
	if err != nil {
		t.Fatalf("EncryptGCM failed: %v", err)
	}

	// Corrupt the ciphertext (modify a byte after the nonce)
	if len(ciphertext) > nonceSize+1 {
		ciphertext[nonceSize+1] ^= 0xFF
	}

	// Decryption should fail
	_, err = DecryptGCM(key, ciphertext)
	if err == nil {
		t.Error("DecryptGCM should fail with corrupted data")
	}
}

// TestEncryptGCMDifferentNonces verifies each encryption uses a random nonce.
func TestEncryptGCMDifferentNonces(t *testing.T) {
	key := make([]byte, keySize)
	plaintext := []byte("same data")

	ct1, err := EncryptGCM(key, plaintext)
	if err != nil {
		t.Fatalf("first EncryptGCM failed: %v", err)
	}

	ct2, err := EncryptGCM(key, plaintext)
	if err != nil {
		t.Fatalf("second EncryptGCM failed: %v", err)
	}

	// Different nonces should produce different ciphertexts
	if bytes.Equal(ct1, ct2) {
		t.Error("two encryptions of same data should produce different ciphertexts (different nonces)")
	}

	// But both should decrypt to the same plaintext
	dec1, _ := DecryptGCM(key, ct1)
	dec2, _ := DecryptGCM(key, ct2)

	if !bytes.Equal(dec1, plaintext) || !bytes.Equal(dec2, plaintext) {
		t.Error("both ciphertexts should decrypt to original plaintext")
	}
}

// TestUnwrapKey verifies UnwrapKey with known test data.
func TestUnwrapKey(t *testing.T) {
	token := "test-token-12345"
	machineID := "machine-id-xyz"

	// First derive the wrapping key
	wrapKey, err := DeriveKey(
		[]byte(token),
		[]byte(machineID),
		[]byte("claudex-wrap"),
		keySize,
	)
	if err != nil {
		t.Fatalf("DeriveKey failed: %v", err)
	}

	// Encrypt a test key
	testKey := []byte("this-is-a-secret-32-byte-key-xx")
	wrappedKey, err := EncryptGCM(wrapKey, testKey)
	if err != nil {
		t.Fatalf("EncryptGCM failed: %v", err)
	}

	// Unwrap it
	unwrappedKey, err := UnwrapKey(token, machineID, wrappedKey)
	if err != nil {
		t.Fatalf("UnwrapKey failed: %v", err)
	}

	// Verify round-trip
	if !bytes.Equal(unwrappedKey, testKey) {
		t.Errorf("unwrapped key mismatch: expected %q, got %q", testKey, unwrappedKey)
	}
}

// TestUnwrapKeyWrongToken verifies UnwrapKey fails with wrong token.
func TestUnwrapKeyWrongToken(t *testing.T) {
	token := "test-token-12345"
	machineID := "machine-id-xyz"

	// Wrap with one token
	wrapKey, _ := DeriveKey(
		[]byte(token),
		[]byte(machineID),
		[]byte("claudex-wrap"),
		keySize,
	)
	testKey := []byte("this-is-a-secret-32-byte-key-xx")
	wrappedKey, _ := EncryptGCM(wrapKey, testKey)

	// Try to unwrap with different token
	_, err := UnwrapKey("wrong-token", machineID, wrappedKey)
	if err == nil {
		t.Error("UnwrapKey should fail with wrong token")
	}
}

// TestUnwrapKeyWrongMachineID verifies UnwrapKey fails with wrong machine ID.
func TestUnwrapKeyWrongMachineID(t *testing.T) {
	token := "test-token-12345"
	machineID := "machine-id-xyz"

	// Wrap with one machine ID
	wrapKey, _ := DeriveKey(
		[]byte(token),
		[]byte(machineID),
		[]byte("claudex-wrap"),
		keySize,
	)
	testKey := []byte("this-is-a-secret-32-byte-key-xx")
	wrappedKey, _ := EncryptGCM(wrapKey, testKey)

	// Try to unwrap with different machine ID
	_, err := UnwrapKey(token, "wrong-machine-id", wrappedKey)
	if err == nil {
		t.Error("UnwrapKey should fail with wrong machine ID")
	}
}
