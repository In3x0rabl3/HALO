package telemetry

type Telemetry struct {
	Processes      []string
	Drivers        []string
	NetworkTraffic []string
	NetworkInfo    []string
	USBDevices     []string
	LogonSessions  []string
	UptimeMinutes  int64
	Hostname       string
	SelfProcess    string
	ParentProcess  string
}
