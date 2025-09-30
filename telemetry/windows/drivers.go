//go:build windows
// +build windows

package windows

import (
	"strings"
	"syscall"
	"unsafe"
)

// getDrivers returns a list of loaded kernel drivers.
func GetDrivers() []string {
	var drivers []string

	psapi := syscall.NewLazyDLL("psapi.dll")
	procEnumDeviceDrivers := psapi.NewProc("EnumDeviceDrivers")
	procGetDeviceDriverBaseName := psapi.NewProc("GetDeviceDriverBaseNameW")

	var cbNeeded uint32
	var arr [1024]uintptr

	ret, _, _ := procEnumDeviceDrivers.Call(
		uintptr(unsafe.Pointer(&arr[0])),
		uintptr(len(arr))*unsafe.Sizeof(arr[0]),
		uintptr(unsafe.Pointer(&cbNeeded)),
	)
	if ret == 0 {
		return drivers
	}

	count := cbNeeded / uint32(unsafe.Sizeof(arr[0]))
	for i := 0; i < int(count); i++ {
		var name [syscall.MAX_PATH]uint16
		n, _, _ := procGetDeviceDriverBaseName.Call(
			arr[i],
			uintptr(unsafe.Pointer(&name[0])),
			uintptr(len(name)),
		)
		if n == 0 {
			continue
		}
		drivers = append(drivers, strings.ToLower(syscall.UTF16ToString(name[:])))
	}

	return drivers
}
