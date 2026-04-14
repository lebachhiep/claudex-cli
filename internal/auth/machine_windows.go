//go:build windows

package auth

import (
	"os/exec"
	"strings"
)

// collectMachineInfo gathers hardware identifiers on Windows via wmic/powershell.
func collectMachineInfo() (*MachineInfo, error) {
	info := &MachineInfo{}

	// MAC address from first physical network adapter
	if out, err := runCmd("wmic", "nic", "where", "NetEnabled=true", "get", "MACAddress", "/value"); err == nil {
		info.MAC = parseWMICValue(out, "MACAddress")
	}

	// CPU ProcessorId
	if out, err := runCmd("wmic", "cpu", "get", "ProcessorId", "/value"); err == nil {
		info.CPUID = parseWMICValue(out, "ProcessorId")
	}

	// Boot disk serial number
	if out, err := runCmd("wmic", "diskdrive", "get", "SerialNumber", "/value"); err == nil {
		info.DiskSerial = parseWMICValue(out, "SerialNumber")
	}

	return info, nil
}

// runCmd executes a command and returns trimmed stdout.
func runCmd(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// parseWMICValue extracts value from "Key=Value" WMIC output.
func parseWMICValue(output, key string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		prefix := key + "="
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}
