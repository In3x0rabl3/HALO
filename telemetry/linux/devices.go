//go:build linux
// +build linux

package linux

import (
	"io/ioutil"
	"path/filepath"
	"strings"
)

// GetUSBDevices reads connected USB devices from sysfs.
func GetUSBDevices() []string {
	var devices []string
	entries, err := ioutil.ReadDir("/sys/bus/usb/devices")
	if err != nil {
		return devices
	}

	for _, entry := range entries {
		path := filepath.Join("/sys/bus/usb/devices", entry.Name(), "product")
		data, err := ioutil.ReadFile(path)
		if err == nil {
			name := strings.TrimSpace(string(data))
			if name != "" {
				devices = append(devices, name)
			}
		}
	}
	return devices
}
