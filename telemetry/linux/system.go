//go:build linux
// +build linux

package linux

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

// GetHostname returns the system hostname.
func GetHostname() string {
	name, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return name
}

// GetUptimeMinutes returns system uptime in minutes.
func GetUptimeMinutes() int64 {
	data, err := ioutil.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}
	parts := strings.Fields(string(data))
	if len(parts) > 0 {
		secs, err := strconv.ParseFloat(parts[0], 64)
		if err == nil {
			return int64(secs / 60)
		}
	}
	return 0
}
