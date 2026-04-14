//go:build darwin

package auth

import (
	"net"
	"os/exec"
	"strings"
)

// collectMachineInfo gathers hardware identifiers on macOS.
func collectMachineInfo() (*MachineInfo, error) {
	info := &MachineInfo{}

	// MAC address from en0 interface
	if iface, err := net.InterfaceByName("en0"); err == nil {
		info.MAC = iface.HardwareAddr.String()
	}

	// CPU identifier via sysctl
	if out, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output(); err == nil {
		info.CPUID = strings.TrimSpace(string(out))
	}
	// Append hardware UUID for uniqueness
	if out, err := exec.Command("sysctl", "-n", "kern.uuid").Output(); err == nil {
		info.CPUID += "|" + strings.TrimSpace(string(out))
	}

	// Disk serial from IOPlatformExpertDevice
	if out, err := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice").Output(); err == nil {
		info.DiskSerial = parseIORegValue(string(out), "IOPlatformSerialNumber")
	}

	return info, nil
}

// parseIORegValue extracts a value from ioreg output like: "key" = "value"
func parseIORegValue(output, key string) string {
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "\""+key+"\"") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				val = strings.Trim(val, "\"")
				return val
			}
		}
	}
	return ""
}
