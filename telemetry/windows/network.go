//go:build windows
// +build windows

package windows

import (
	"fmt"
	"net"
	"syscall"
	"unsafe"
)

const TCP_TABLE_OWNER_PID_ALL = 5

type MIB_TCPROW_OWNER_PID struct {
	State      uint32
	LocalAddr  uint32
	LocalPort  uint32
	RemoteAddr uint32
	RemotePort uint32
	OwningPid  uint32
}

// getNetworkTraffic retrieves TCP connections and owning PIDs.
func GetNetworkTraffic() []string {
	var results []string

	iphlpapi := syscall.NewLazyDLL("iphlpapi.dll")
	procGetExtendedTcpTable := iphlpapi.NewProc("GetExtendedTcpTable")

	var bufSize uint32
	procGetExtendedTcpTable.Call(
		0,
		uintptr(unsafe.Pointer(&bufSize)),
		0,
		uintptr(syscall.AF_INET),
		uintptr(TCP_TABLE_OWNER_PID_ALL),
		0,
	)

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

	count := *(*uint32)(unsafe.Pointer(&buf[0]))
	offset := unsafe.Sizeof(count)

	for i := 0; i < int(count); i++ {
		row := (*MIB_TCPROW_OWNER_PID)(unsafe.Pointer(&buf[offset]))
		localPort := ntohs(uint16(row.LocalPort >> 16))
		remotePort := ntohs(uint16(row.RemotePort >> 16))

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
				row.OwningPid, row.State,
			),
		)

		offset += unsafe.Sizeof(*row)
	}

	return results
}

// getNetworkInfo lists network interfaces and addresses.
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

func ntohs(n uint16) uint16 {
	return (n<<8)&0xff00 | n>>8
}
