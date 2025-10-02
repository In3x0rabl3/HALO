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
	passiveEgress := PassiveEgress{
		ActiveConnections: linux.GetNetworkTraffic(),
		DefaultGateways:   linux.GetDefaultGateways(),
		ProxyEnv:          linux.GetProxyEnv(),
	}

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
		PassiveEgress:  passiveEgress, // <-- Add this!
		// Egress:      ... (leave as zero unless you're doing active checks)
	}
}

func Collect(baseline Telemetry, selfProc, parentProc string) Telemetry {
	return BuildBaseline(selfProc, parentProc)
}
