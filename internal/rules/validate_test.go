package rules

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"
)

// TestVerifyChecksumMatch verifies VerifyChecksum with matching hash.
func TestVerifyChecksumMatch(t *testing.T) {
	data := []byte("test data for checksum")

	// Calculate expected hash
	hash := sha256.Sum256(data)
	expected := fmt.Sprintf("sha256:%x", hash)

	// Verify should pass
	if err := VerifyChecksum(data, expected); err != nil {
		t.Fatalf("VerifyChecksum should pass with matching hash: %v", err)
	}
}

// TestVerifyChecksumMismatch verifies VerifyChecksum fails with wrong hash.
func TestVerifyChecksumMismatch(t *testing.T) {
	data := []byte("test data")
	wrongChecksum := "sha256:0000000000000000000000000000000000000000000000000000000000000000"

	err := VerifyChecksum(data, wrongChecksum)
	if err == nil {
		t.Error("VerifyChecksum should fail with mismatched hash")
	}

	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Errorf("error message should contain 'checksum mismatch': %s", err.Error())
	}
}

// TestVerifyChecksumEmptyData verifies VerifyChecksum with empty data.
func TestVerifyChecksumEmptyData(t *testing.T) {
	data := []byte{}

	// Calculate hash of empty data
	hash := sha256.Sum256(data)
	expected := fmt.Sprintf("sha256:%x", hash)

	if err := VerifyChecksum(data, expected); err != nil {
		t.Fatalf("VerifyChecksum should work with empty data: %v", err)
	}
}

// TestVerifyChecksumCaseInsensitive verifies checksum comparison is case-insensitive.
func TestVerifyChecksumCaseInsensitive(t *testing.T) {
	data := []byte("test data")

	// Calculate hash
	hash := sha256.Sum256(data)
	expected := fmt.Sprintf("sha256:%x", hash)

	// Convert to uppercase
	expectedUpper := strings.ToUpper(expected)

	if err := VerifyChecksum(data, expectedUpper); err != nil {
		t.Fatalf("VerifyChecksum should be case-insensitive: %v", err)
	}
}

// TestVerifyChecksumMixedCase verifies checksum with mixed case.
func TestVerifyChecksumMixedCase(t *testing.T) {
	data := []byte("test data")

	// Calculate hash
	hash := sha256.Sum256(data)
	expected := fmt.Sprintf("sha256:%x", hash)

	// Create mixed case version (alternate case)
	parts := strings.Split(expected, ":")
	if len(parts) == 2 {
		hexPart := parts[1]
		mixedCase := "sha256:" + strings.ToUpper(hexPart[:16]) + strings.ToLower(hexPart[16:])

		if err := VerifyChecksum(data, mixedCase); err != nil {
			t.Fatalf("VerifyChecksum should be case-insensitive: %v", err)
		}
	}
}

// TestVerifyChecksumInvalidFormat verifies error with invalid checksum format.
func TestVerifyChecksumInvalidFormat(t *testing.T) {
	data := []byte("test data")

	// Missing sha256: prefix
	invalidChecksum := "0000000000000000000000000000000000000000000000000000000000000000"

	err := VerifyChecksum(data, invalidChecksum)
	if err == nil {
		t.Error("VerifyChecksum should fail with invalid format (missing prefix)")
	}
}

// TestVerifyChecksumWrongAlgorithm verifies error with wrong algorithm prefix.
func TestVerifyChecksumWrongAlgorithm(t *testing.T) {
	data := []byte("test data")

	hash := sha256.Sum256(data)
	wrongAlgo := fmt.Sprintf("md5:%x", hash)

	err := VerifyChecksum(data, wrongAlgo)
	if err == nil {
		t.Error("VerifyChecksum should fail with wrong algorithm prefix")
	}
}

// TestVerifyChecksumTruncatedHash verifies error with truncated hash value.
func TestVerifyChecksumTruncatedHash(t *testing.T) {
	data := []byte("test data")

	hash := sha256.Sum256(data)
	truncated := fmt.Sprintf("sha256:%x", hash)[0:20] // Truncate

	err := VerifyChecksum(data, truncated)
	if err == nil {
		t.Error("VerifyChecksum should fail with truncated hash")
	}
}

// TestVerifyChecksumLargeData verifies VerifyChecksum with large data.
func TestVerifyChecksumLargeData(t *testing.T) {
	// Create 10MB of data
	largeData := make([]byte, 10*1024*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	hash := sha256.Sum256(largeData)
	expected := fmt.Sprintf("sha256:%x", hash)

	if err := VerifyChecksum(largeData, expected); err != nil {
		t.Fatalf("VerifyChecksum should work with large data: %v", err)
	}
}

// TestVerifyChecksumKnownVector verifies against known test vector.
func TestVerifyChecksumKnownVector(t *testing.T) {
	// Known test vector: SHA256 of "abc" = ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
	data := []byte("abc")
	expectedChecksum := "sha256:ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"

	if err := VerifyChecksum(data, expectedChecksum); err != nil {
		t.Fatalf("VerifyChecksum failed on known test vector: %v", err)
	}
}

// TestVerifyChecksumErrorMessage verifies error message includes both checksums.
func TestVerifyChecksumErrorMessage(t *testing.T) {
	data := []byte("test")
	wrongChecksum := "sha256:1111111111111111111111111111111111111111111111111111111111111111"

	err := VerifyChecksum(data, wrongChecksum)
	if err == nil {
		t.Fatal("VerifyChecksum should fail with wrong hash")
	}

	errMsg := err.Error()

	// Error message should include "expected"
	if !strings.Contains(errMsg, "expected") {
		t.Errorf("error message should include 'expected': %s", errMsg)
	}

	// Error message should include "got"
	if !strings.Contains(errMsg, "got") {
		t.Errorf("error message should include 'got': %s", errMsg)
	}

	// Error message should include the wrong checksum
	if !strings.Contains(errMsg, wrongChecksum) {
		t.Errorf("error message should include expected checksum: %s", errMsg)
	}
}

// TestVerifyChecksumOneByteChange verifies detecting single byte change.
func TestVerifyChecksumOneByteChange(t *testing.T) {
	data := []byte("original data")

	// Get correct checksum
	hash := sha256.Sum256(data)
	correctChecksum := fmt.Sprintf("sha256:%x", hash)

	// Verify passes with correct checksum
	if err := VerifyChecksum(data, correctChecksum); err != nil {
		t.Fatalf("should verify with correct checksum: %v", err)
	}

	// Change one byte
	modified := make([]byte, len(data))
	copy(modified, data)
	modified[0] = modified[0] ^ 0xFF

	// Should fail with changed data
	if err := VerifyChecksum(modified, correctChecksum); err == nil {
		t.Error("should fail to verify modified data with original checksum")
	}
}

// TestVerifyChecksumMultipleCalls verifies consistency across calls.
func TestVerifyChecksumMultipleCalls(t *testing.T) {
	data := []byte("consistent data")

	hash := sha256.Sum256(data)
	expected := fmt.Sprintf("sha256:%x", hash)

	// Call multiple times - should always succeed
	for i := 0; i < 5; i++ {
		if err := VerifyChecksum(data, expected); err != nil {
			t.Fatalf("VerifyChecksum call %d failed: %v", i+1, err)
		}
	}
}

// TestVerifyChecksumBinaryData verifies with binary data (not just text).
func TestVerifyChecksumBinaryData(t *testing.T) {
	// Binary data with null bytes
	data := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD, 0xFC}

	hash := sha256.Sum256(data)
	expected := fmt.Sprintf("sha256:%x", hash)

	if err := VerifyChecksum(data, expected); err != nil {
		t.Fatalf("VerifyChecksum should work with binary data: %v", err)
	}
}

// TestVerifyChecksumSingleByteData verifies with minimal data.
func TestVerifyChecksumSingleByteData(t *testing.T) {
	data := []byte{0xFF}

	hash := sha256.Sum256(data)
	expected := fmt.Sprintf("sha256:%x", hash)

	if err := VerifyChecksum(data, expected); err != nil {
		t.Fatalf("VerifyChecksum should work with single byte: %v", err)
	}
}
