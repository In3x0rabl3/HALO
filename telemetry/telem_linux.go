//go:build linux
// +build linux

package telemetry

import (
	"halo/telemetry/linux"
)

func GetSelfAndParentNames() (string, string) {
	return linux.GetSelfAndParentNames()
}

func BuildBaseline(selfProc, parentProc string) Telemetry {
	return Telemetry{
		Processes:      linux.GetProcesses(),
		Drivers:        linux.GetDrivers(),
		NetworkTraffic: linux.GetNetworkTraffic(),
		NetworkInfo:    linux.GetNetworkInfo(),
		USBDevices:     linux.GetUSBDevices(),
		LogonSessions:  linux.GetLogonSessions(),
		UptimeMinutes:  linux.GetUptimeMinutes(),
		Hostname:       linux.GetHostname(),
		SelfProcess:    selfProc,
		ParentProcess:  parentProc,
	}
}

func Collect(baseline Telemetry, selfProc, parentProc string) Telemetry {
	return BuildBaseline(selfProc, parentProc)
}
