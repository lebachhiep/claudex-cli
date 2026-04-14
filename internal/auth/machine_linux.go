//go:build linux

package auth

import (
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// collectMachineInfo gathers hardware identifiers on Linux.
func collectMachineInfo() (*MachineInfo, error) {
	info := &MachineInfo{}

	// MAC address from first non-loopback interface
	info.MAC = findFirstMAC()

	// CPU model from /proc/cpuinfo
	if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "model name") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					info.CPUID = strings.TrimSpace(parts[1])
					break
				}
			}
		}
	}

	// Disk serial from /sys or lsblk fallback
	info.DiskSerial = findDiskSerial()

	return info, nil
}

// findFirstMAC returns the MAC address of the first non-loopback interface.
func findFirstMAC() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if len(iface.HardwareAddr) > 0 {
			return iface.HardwareAddr.String()
		}
	}
	return ""
}

// findDiskSerial reads disk serial from sysfs, falls back to lsblk.
func findDiskSerial() string {
	// Try sysfs first
	matches, _ := filepath.Glob("/sys/block/*/device/serial")
	for _, m := range matches {
		if data, err := os.ReadFile(m); err == nil {
			serial := strings.TrimSpace(string(data))
			if serial != "" {
				return serial
			}
		}
	}
	// Fallback to lsblk
	if out, err := exec.Command("lsblk", "--nodeps", "-no", "serial").Output(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				return line
			}
		}
	}
	return ""
}
