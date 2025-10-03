//go:build windows
// +build windows

package windows

import (
	"fmt"
	"net"
	"os"
	"syscall"
	"unsafe"
)

// =============================
// Passive Network Observations
// =============================

const (
	TCP_TABLE_OWNER_PID_ALL = 5
)

type MIB_TCPROW_OWNER_PID struct {
	State      uint32
	LocalAddr  uint32
	LocalPort  uint32
	RemoteAddr uint32
	RemotePort uint32
	OwningPid  uint32
}

// GetNetworkTraffic retrieves TCP connections and owning PIDs (IPv4)
func GetNetworkTraffic() []string {
	var results []string

	iphlpapi := syscall.NewLazyDLL("iphlpapi.dll")
	procGetExtendedTcpTable := iphlpapi.NewProc("GetExtendedTcpTable")

	// First call to get buffer size
	var bufSize uint32
	procGetExtendedTcpTable.Call(
		0,
		uintptr(unsafe.Pointer(&bufSize)),
		0,
		uintptr(syscall.AF_INET),
		uintptr(TCP_TABLE_OWNER_PID_ALL),
		0,
	)

	if bufSize == 0 {
		return results
	}

	buf := make([]byte, bufSize)
	ret, _, _ := procGetExtendedTcpTable.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&bufSize)),
		0,
		uintptr(syscall.AF_INET),
		uintptr(TCP_TABLE_OWNER_PID_ALL),
		0,
	)
	if ret != 0 {
		return results
	}

	// Layout: DWORD NumEntries; followed by NumEntries * MIB_TCPROW_OWNER_PID
	count := *(*uint32)(unsafe.Pointer(&buf[0]))
	base := uintptr(unsafe.Pointer(&buf[0]))
	ptr := base + unsafe.Sizeof(count)

	for i := 0; i < int(count); i++ {
		row := (*MIB_TCPROW_OWNER_PID)(unsafe.Pointer(ptr))

		// LocalPort/RemotePort are DWORD with port in network byte order in the LOW 16 bits.
		// Using ntohs on uint16(row.LocalPort) is sufficient.
		localPort := ntohs(uint16(row.LocalPort))
		remotePort := ntohs(uint16(row.RemotePort))

		localIP := net.IPv4(
			byte(row.LocalAddr),
			byte(row.LocalAddr>>8),
			byte(row.LocalAddr>>16),
			byte(row.LocalAddr>>24),
		).String()

		remoteIP := net.IPv4(
			byte(row.RemoteAddr),
			byte(row.RemoteAddr>>8),
			byte(row.RemoteAddr>>16),
			byte(row.RemoteAddr>>24),
		).String()

		results = append(results,
			fmt.Sprintf("%s:%d -> %s:%d (PID %d, State %d)",
				localIP, localPort,
				remoteIP, remotePort,
				row.OwningPid, row.State))

		ptr += unsafe.Sizeof(*row)
	}

	return results
}

// GetNetworkInfo returns NICs + IPs
func GetNetworkInfo() []string {
	var info []string

	ifaces, err := net.Interfaces()
	if err != nil {
		return info
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			info = append(info, fmt.Sprintf("%s: %s", iface.Name, addr.String()))
		}
	}

	return info
}

// GetProxyEnv returns proxy environment variables (parity with Linux)
func GetProxyEnv() map[string]string {
	result := map[string]string{}
	for _, key := range []string{"http_proxy", "https_proxy", "HTTP_PROXY", "HTTPS_PROXY"} {
		if val := os.Getenv(key); val != "" {
			result[key] = val
		}
	}
	return result
}

func ntohs(n uint16) uint16 {
	return (n<<8)&0xff00 | n>>8
}
