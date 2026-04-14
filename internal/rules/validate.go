// Package rules handles downloading, decrypting, extracting, and tracking rules bundles.
package rules

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// VerifyChecksum compares SHA-256 hash of data against expected "sha256:hex" string.
func VerifyChecksum(data []byte, expected string) error {
	hash := sha256.Sum256(data)
	actual := fmt.Sprintf("sha256:%x", hash)

	if !strings.EqualFold(actual, expected) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, actual)
	}
	return nil
}
