package telemetry

type EgressAttempt struct {
	Dst     string `json:"dst"`
	Success bool   `json:"success"`
	RttMs   *int   `json:"rtt_ms,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ProxyInfo struct {
	Detected bool   `json:"detected"`
	Type     string `json:"type,omitempty"`
	Address  string `json:"address,omitempty"`
}

type FirewallRule struct {
	Direction string `json:"direction"`
	Port      string `json:"port"`
	Allow     bool   `json:"allow"`
}

type Egress struct {
	Attempts []EgressAttempt `json:"attempts"`
	Proxy    ProxyInfo       `json:"proxy"`
	Firewall []FirewallRule  `json:"firewall"`
	SslMitm  bool            `json:"ssl_mitm"`
}

// ===== NEW: Passive (OPSEC-friendly) egress struct =====
type PassiveEgress struct {
	// All current active TCP/UDP connections with state (including IPv6)
	ActiveConnections []string `json:"active_connections"`
	// Default gateway IPs found in /proc/net/route
	DefaultGateways []string `json:"default_gateways"`
	// Proxy environment variables detected
	ProxyEnv map[string]string `json:"proxy_env"`
}

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
	Egress         Egress
	PassiveEgress  PassiveEgress
}
