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
	Reason   string  `json:"reason"`
	Conf     float64 `json:"confidence"`
	Thoughts string  `json:"thoughts"`
}

const systemPrompt = `
You are an AI Red Team assistant payload runner & OPSEC sentinel analyzing telemetry from a host machine.
Think in 3 steps: 
1. OBSERVE - summarize key processes, drivers, network flows, sessions, deviations, activity.
2. ANALYZE - decide if AV/EDR or monitoring is present, what traffic is allow outbound, if execution risks exposure.
3. DECIDE - output JSON with:
- allow (true/false): whether execution should proceed
- conf (0.0-1.0): confidence in that decision
- reason: concise human-readable justification
- thoughts: detailed internal reasoning showing what you checked and why you made the decision.
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
