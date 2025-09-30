//go:build windows
// +build windows

package telemetry

import (
	"halo/telemetry/windows"
)

func GetSelfAndParentNames() (string, string) {
	return windows.GetSelfAndParentNames()
}

func BuildBaseline(selfProc, parentProc string) Telemetry {
	return Telemetry{
		Processes:      windows.GetProcesses(),
		Drivers:        windows.GetDrivers(),
		NetworkTraffic: windows.GetNetworkTraffic(),
		NetworkInfo:    windows.GetNetworkInfo(),
		USBDevices:     windows.GetUSBDevices(),
		LogonSessions:  windows.GetLogonSessions(),
		UptimeMinutes:  windows.GetUptimeMinutes(),
		Hostname:       windows.GetHostname(),
		SelfProcess:    selfProc,
		ParentProcess:  parentProc,
	}
}

func Collect(baseline Telemetry, selfProc, parentProc string) Telemetry {
	return BuildBaseline(selfProc, parentProc)
}
