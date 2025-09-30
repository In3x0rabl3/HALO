//go:build linux
// +build linux

package linux

import (
	"io/ioutil"
	"strings"
)

// GetDrivers lists loaded kernel modules from /proc/modules.
func GetDrivers() []string {
	var drivers []string
	data, err := ioutil.ReadFile("/proc/modules")
	if err != nil {
		return drivers
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 {
			drivers = append(drivers, fields[0])
		}
	}
	return drivers
}
