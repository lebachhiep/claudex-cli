package crypto

import (
	"bytes"
	"testing"
)

// TestDeriveKeyDeterministic verifies HKDF key derivation is deterministic.
func TestDeriveKeyDeterministic(t *testing.T) {
	secret := []byte("my-secret-key")
	salt := []byte("my-salt")
	info := []byte("my-info")
	length := 32

	// First derivation
	key1, err := DeriveKey(secret, salt, info, length)
	if err != nil {
		t.Fatalf("first DeriveKey failed: %v", err)
	}

	// Second derivation with same inputs
	key2, err := DeriveKey(secret, salt, info, length)
	if err != nil {
		t.Fatalf("second DeriveKey failed: %v", err)
	}

	// Keys should be identical
	if !bytes.Equal(key1, key2) {
		t.Errorf("DeriveKey should be deterministic: got different keys for same inputs")
	}
}

// TestDeriveKeyDifferentSecrets verifies different secrets produce different keys.
func TestDeriveKeyDifferentSecrets(t *testing.T) {
	salt := []byte("my-salt")
	info := []byte("my-info")
	length := 32

	key1, err := DeriveKey([]byte("secret1"), salt, info, length)
	if err != nil {
		t.Fatalf("DeriveKey with secret1 failed: %v", err)
	}

	key2, err := DeriveKey([]byte("secret2"), salt, info, length)
	if err != nil {
		t.Fatalf("DeriveKey with secret2 failed: %v", err)
	}

	if bytes.Equal(key1, key2) {
		t.Error("DeriveKey should produce different keys for different secrets")
	}
}

// TestDeriveKeyDifferentSalts verifies different salts produce different keys.
func TestDeriveKeyDifferentSalts(t *testing.T) {
	secret := []byte("my-secret")
	info := []byte("my-info")
	length := 32

	key1, err := DeriveKey(secret, []byte("salt1"), info, length)
	if err != nil {
		t.Fatalf("DeriveKey with salt1 failed: %v", err)
	}

	key2, err := DeriveKey(secret, []byte("salt2"), info, length)
	if err != nil {
		t.Fatalf("DeriveKey with salt2 failed: %v", err)
	}

	if bytes.Equal(key1, key2) {
		t.Error("DeriveKey should produce different keys for different salts")
	}
}

// TestDeriveKeyDifferentInfo verifies different info values produce different keys.
func TestDeriveKeyDifferentInfo(t *testing.T) {
	secret := []byte("my-secret")
	salt := []byte("my-salt")
	length := 32

	key1, err := DeriveKey(secret, salt, []byte("info1"), length)
	if err != nil {
		t.Fatalf("DeriveKey with info1 failed: %v", err)
	}

	key2, err := DeriveKey(secret, salt, []byte("info2"), length)
	if err != nil {
		t.Fatalf("DeriveKey with info2 failed: %v", err)
	}

	if bytes.Equal(key1, key2) {
		t.Error("DeriveKey should produce different keys for different info values")
	}
}

// TestDeriveKeyEmptySalt verifies derivation with empty salt.
func TestDeriveKeyEmptySalt(t *testing.T) {
	secret := []byte("my-secret")
	info := []byte("my-info")
	length := 32

	key, err := DeriveKey(secret, nil, info, length)
	if err != nil {
		t.Fatalf("DeriveKey with empty salt failed: %v", err)
	}

	if len(key) != length {
		t.Errorf("key length mismatch: expected %d, got %d", length, len(key))
	}
}

// TestDeriveKeyVaryingLengths verifies derivation with different output lengths.
func TestDeriveKeyVaryingLengths(t *testing.T) {
	secret := []byte("my-secret")
	salt := []byte("my-salt")
	info := []byte("my-info")

	lengths := []int{16, 32, 64, 128}

	for _, length := range lengths {
		key, err := DeriveKey(secret, salt, info, length)
		if err != nil {
			t.Fatalf("DeriveKey with length %d failed: %v", length, err)
		}

		if len(key) != length {
			t.Errorf("key length mismatch for requested %d: expected %d, got %d", length, length, len(key))
		}
	}
}

// TestDeriveKeyZeroLength verifies derivation with zero length.
func TestDeriveKeyZeroLength(t *testing.T) {
	secret := []byte("my-secret")
	salt := []byte("my-salt")
	info := []byte("my-info")

	key, err := DeriveKey(secret, salt, info, 0)
	if err != nil {
		t.Fatalf("DeriveKey with zero length failed: %v", err)
	}

	if len(key) != 0 {
		t.Errorf("expected empty key, got %d bytes", len(key))
	}
}

// TestDeriveKeyLargeOutput verifies derivation with large output.
func TestDeriveKeyLargeOutput(t *testing.T) {
	secret := []byte("my-secret")
	salt := []byte("my-salt")
	info := []byte("my-info")
	length := 1000

	key, err := DeriveKey(secret, salt, info, length)
	if err != nil {
		t.Fatalf("DeriveKey with large length failed: %v", err)
	}

	if len(key) != length {
		t.Errorf("key length mismatch: expected %d, got %d", length, len(key))
	}
}

// TestDeriveKeyEmptySecret verifies derivation with empty secret.
func TestDeriveKeyEmptySecret(t *testing.T) {
	salt := []byte("my-salt")
	info := []byte("my-info")
	length := 32

	key, err := DeriveKey(nil, salt, info, length)
	if err != nil {
		t.Fatalf("DeriveKey with empty secret failed: %v", err)
	}

	if len(key) != length {
		t.Errorf("key length mismatch: expected %d, got %d", length, len(key))
	}
}

// TestDeriveKeyEmptyInfo verifies derivation with empty info.
func TestDeriveKeyEmptyInfo(t *testing.T) {
	secret := []byte("my-secret")
	salt := []byte("my-salt")
	length := 32

	key, err := DeriveKey(secret, salt, nil, length)
	if err != nil {
		t.Fatalf("DeriveKey with empty info failed: %v", err)
	}

	if len(key) != length {
		t.Errorf("key length mismatch: expected %d, got %d", length, len(key))
	}
}
