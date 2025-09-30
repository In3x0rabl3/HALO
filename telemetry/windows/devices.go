//go:build windows
// +build windows

package windows

import (
	"strings"
	"syscall"
	"unsafe"
)

// getUSBDevices returns a list of connected USB device names.
func GetUSBDevices() []string {
	var devices []string

	setupapi := syscall.NewLazyDLL("setupapi.dll")

	procSetupDiGetClassDevs := setupapi.NewProc("SetupDiGetClassDevsW")
	procSetupDiEnumDeviceInfo := setupapi.NewProc("SetupDiEnumDeviceInfo")
	procSetupDiGetDeviceRegistryProperty := setupapi.NewProc("SetupDiGetDeviceRegistryPropertyW")
	procSetupDiDestroyDeviceInfoList := setupapi.NewProc("SetupDiDestroyDeviceInfoList")

	const DIGCF_PRESENT = 0x00000002
	const DIGCF_ALLCLASSES = 0x00000004
	const SPDRP_DEVICEDESC = 0x00000000

	type SP_DEVINFO_DATA struct {
		CbSize    uint32
		ClassGuid [16]byte
		DevInst   uint32
		Reserved  uintptr
	}

	hDevInfo, _, _ := procSetupDiGetClassDevs.Call(
		0,
		0,
		0,
		uintptr(DIGCF_PRESENT|DIGCF_ALLCLASSES),
	)
	if hDevInfo == uintptr(syscall.InvalidHandle) {
		return devices
	}
	defer procSetupDiDestroyDeviceInfoList.Call(hDevInfo)

	var devInfoData SP_DEVINFO_DATA
	devInfoData.CbSize = uint32(unsafe.Sizeof(devInfoData))

	for i := 0; ; i++ {
		ret, _, _ := procSetupDiEnumDeviceInfo.Call(hDevInfo, uintptr(i), uintptr(unsafe.Pointer(&devInfoData)))
		if ret == 0 {
			break
		}

		var buf [syscall.MAX_PATH]uint16
		var required uint32
		procSetupDiGetDeviceRegistryProperty.Call(
			hDevInfo,
			uintptr(unsafe.Pointer(&devInfoData)),
			SPDRP_DEVICEDESC,
			0,
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(len(buf)*2),
			uintptr(unsafe.Pointer(&required)),
		)

		name := syscall.UTF16ToString(buf[:])
		if strings.Contains(strings.ToLower(name), "usb") {
			devices = append(devices, name)
		}
	}

	return devices
}
