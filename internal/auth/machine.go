// Package auth handles machine fingerprinting, token storage, and authentication flows.
package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
	"strings"
)

// MachineInfo holds hardware identifiers collected from the OS.
type MachineInfo struct {
	MAC        string
	CPUID      string
	DiskSerial string
	Hostname   string
	OS         string
	Arch       string
}

// GenerateMachineID collects hardware info and hashes it into a 32-char hex ID.
func GenerateMachineID() (string, *MachineInfo, error) {
	info, err := collectMachineInfo()
	if err != nil {
		return "", nil, fmt.Errorf("collect machine info: %w", err)
	}

	info.OS = runtime.GOOS
	info.Arch = runtime.GOARCH
	info.Hostname, _ = os.Hostname()

	raw := strings.Join([]string{
		info.MAC,
		info.CPUID,
		info.DiskSerial,
		info.Hostname,
	}, "|")

	hash := sha256.Sum256([]byte(raw))
	machineID := hex.EncodeToString(hash[:16]) // 32 hex chars
	return machineID, info, nil
}
