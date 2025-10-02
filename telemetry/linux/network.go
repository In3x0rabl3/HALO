//go:build linux
// +build linux

package linux

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// =============================
// Passive Network Observations
// =============================

func GetNetworkTraffic() []string {
	results := []string{}
	results = append(results, getTCPConns("/proc/net/tcp")...)
	results = append(results, getTCPConns("/proc/net/tcp6")...)
	results = append(results, getUDPConns("/proc/net/udp")...)
	results = append(results, getUDPConns("/proc/net/udp6")...)
	return results
}

func getTCPConns(path string) []string {
	var results []string
	f, err := os.Open(path)
	if err != nil {
		return results
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	first := true
	for scanner.Scan() {
		line := scanner.Text()
		if first {
			first = false
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			local := parseAddr(fields[1])
			remote := parseAddr(fields[2])
			state := parseTCPState(fields[3])
			results = append(results, fmt.Sprintf("%s -> %s (%s)", local, remote, state))
		}
	}
	return results
}

func getUDPConns(path string) []string {
	var results []string
	f, err := os.Open(path)
	if err != nil {
		return results
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	first := true
	for scanner.Scan() {
		line := scanner.Text()
		if first {
			first = false
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			local := parseAddr(fields[1])
			remote := parseAddr(fields[2])
			results = append(results, fmt.Sprintf("%s -> %s (UDP)", local, remote))
		}
	}
	return results
}

func parseAddr(addr string) string {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return addr
	}
	ipHex := parts[0]
	portHex := parts[1]
	ip := hexToIP(ipHex)
	port, _ := strconv.ParseUint(portHex, 16, 16)
	return fmt.Sprintf("%s:%d", ip, port)
}

func hexToIP(h string) string {
	if len(h) == 8 { // IPv4
		bs := make([]byte, 4)
		for i := 0; i < 4; i++ {
			b, _ := strconv.ParseUint(h[2*i:2*i+2], 16, 8)
			bs[3-i] = byte(b)
		}
		return fmt.Sprintf("%d.%d.%d.%d", bs[0], bs[1], bs[2], bs[3])
	} else if len(h) == 32 { // IPv6
		bs := make([]byte, 16)
		for i := 0; i < 16; i++ {
			b, _ := strconv.ParseUint(h[2*i:2*i+2], 16, 8)
			bs[15-i] = byte(b)
		}
		var parts []string
		for i := 0; i < 16; i += 2 {
			part := uint16(bs[i])<<8 | uint16(bs[i+1])
			parts = append(parts, fmt.Sprintf("%x", part))
		}
		return strings.Join(parts, ":")
	}
	return h
}

func parseTCPState(stateHex string) string {
	// Common states for quick reference (hex)
	states := map[string]string{
		"01": "ESTABLISHED",
		"02": "SYN_SENT",
		"03": "SYN_RECV",
		"04": "FIN_WAIT1",
		"05": "FIN_WAIT2",
		"06": "TIME_WAIT",
		"07": "CLOSE",
		"08": "CLOSE_WAIT",
		"09": "LAST_ACK",
		"0A": "LISTEN",
		"0B": "CLOSING",
	}
	state, ok := states[stateHex]
	if !ok {
		return "UNKNOWN"
	}
	return state
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

// =============================
// Passive Egress Helpers
// =============================

// GetDefaultGateways parses /proc/net/route for default gateways.
func GetDefaultGateways() []string {
	gateways := []string{}
	file, err := os.Open("/proc/net/route")
	if err != nil {
		return gateways
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	first := true
	for scanner.Scan() {
		line := scanner.Text()
		if first {
			first = false
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		if fields[1] == "00000000" { // default route
			gatewayHex := fields[2]
			ip := parseHexIP(gatewayHex)
			gateways = append(gateways, ip)
		}
	}
	return gateways
}

func parseHexIP(hexStr string) string {
	var ip string
	if len(hexStr) == 8 {
		bytes := []byte{
			hexToByte(hexStr[6:8]),
			hexToByte(hexStr[4:6]),
			hexToByte(hexStr[2:4]),
			hexToByte(hexStr[0:2]),
		}
		ip = fmt.Sprintf("%d.%d.%d.%d", bytes[0], bytes[1], bytes[2], bytes[3])
	}
	return ip
}

func hexToByte(h string) byte {
	val, _ := strconv.ParseUint(h, 16, 8)
	return byte(val)
}

// GetProxyEnv returns any proxy-related environment variables (passive).
func GetProxyEnv() map[string]string {
	result := map[string]string{}
	for _, key := range []string{"http_proxy", "https_proxy", "HTTP_PROXY", "HTTPS_PROXY"} {
		val := os.Getenv(key)
		if val != "" {
			result[key] = val
		}
	}
	return result
}
