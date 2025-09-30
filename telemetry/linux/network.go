//go:build linux
// +build linux

package linux

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

// GetNetworkTraffic reads /proc/net/tcp to get active connections.
func GetNetworkTraffic() []string {
	var results []string
	f, err := os.Open("/proc/net/tcp")
	if err != nil {
		return results
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	first := true
	for scanner.Scan() {
		line := scanner.Text()
		if first {
			first = false // skip header
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			local := fields[1]
			remote := fields[2]
			results = append(results, fmt.Sprintf("%s -> %s", local, remote))
		}
	}
	return results
}

// GetNetworkInfo returns interface names + addresses (exported).
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
		for _, a := range addrs {
			info = append(info, iface.Name+": "+a.String())
		}
	}
	return info
}
