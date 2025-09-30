//go:build windows
// +build windows

package windows

import (
	"fmt"
	"syscall"
)

// getUptimeMinutes returns the system uptime in minutes.
func GetUptimeMinutes() int64 {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procGetTickCount64 := kernel32.NewProc("GetTickCount64")

	ret, _, _ := procGetTickCount64.Call()
	ms := int64(ret)
	return ms / 60000
}

// getHostname returns the system's hostname.
func GetHostname() string {
	buf := make([]uint16, 256)
	size := uint32(len(buf))
	err := syscall.GetComputerName(&buf[0], &size)
	if err != nil {
		return "unknown"
	}
	return syscall.UTF16ToString(buf[:size])
}

// getSystemInfo returns combined uptime + hostname for convenience.
func GetSystemInfo() string {
	return fmt.Sprintf("Hostname: %s, Uptime: %d minutes", GetHostname(), GetUptimeMinutes())
}
