package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"halo/logging"
	"halo/telemetry"
	"io"
	"net/http"
	"os"
	"time"
)

type AnthropicProvider struct {
	APIKey string
}

func (p *AnthropicProvider) Ask(t telemetry.Telemetry) (AIResponse, error) {
	return AskAnthropic(t, p.APIKey)
}

// AskAnthropic mirrors AskOpenAI: systemPrompt + telemetry JSON, then POST to Anthropic-style endpoint.
func AskAnthropic(t telemetry.Telemetry, apiKey string) (AIResponse, error) {
	apiURL := os.Getenv("AI_URL")
	if apiURL == "" {
		apiURL = "https://api.anthropic.com/v1/complete" // override in env
	}
	model := os.Getenv("AI_MODEL")
	if model == "" {
		model = "claude-3-opus"
	}
	if apiKey == "" {
		apiKey = os.Getenv("AI_API_KEY")
	}

	// build prompt (system + telemetry)
	prompt := SystemPrompt + "\n\n" + TelemetryToJSON(t)

	// Anthropic completion-style payload (close to their "complete" endpoint)
	payload := map[string]interface{}{
		"model":          model,
		"prompt":         prompt,
		"max_tokens":     512,
		"temperature":    0.0,
		"stop_sequences": []string{},
	}

	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(data))
	if apiKey != "" {
		req.Header.Set("x-api-key", apiKey)
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return AIResponse{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	logging.LogLine("[Anthropic Raw Response] " + string(body))

	// Try likely Anthropic response structures
	var try struct {
		Completion string `json:"completion"`
		Result     string `json:"result"`
		// choices/text might appear in some variants
		Choices []struct {
			Text string `json:"text"`
		} `json:"choices"`
	}
	_ = json.Unmarshal(body, &try)

	// choose content candidate
	var content string
	if try.Completion != "" {
		content = try.Completion
	} else if try.Result != "" {
		content = try.Result
	} else if len(try.Choices) > 0 && try.Choices[0].Text != "" {
		content = try.Choices[0].Text
	}

	if content != "" {
		logging.LogLine("[Anthropic Content] " + content)
		jsonStr := ExtractJSON(content)
		if jsonStr != "" {
			var parsed AIResponse
			if err := json.Unmarshal([]byte(jsonStr), &parsed); err == nil {
				logging.LogLine(fmt.Sprintf("[Anthropic Decision] Allow: %v Conf: %.2f Reason: %s Thoughts: %s",
					parsed.Allow, parsed.Conf, parsed.Reason, parsed.Thoughts))
				return parsed, nil
			}
		}
	}

	// fallback: extract JSON from whole body
	jsonStr := ExtractJSON(string(body))
	if jsonStr == "" {
		return AIResponse{}, errors.New("no valid JSON found in Anthropic response")
	}
	var parsed AIResponse
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return AIResponse{}, err
	}
	logging.LogLine(fmt.Sprintf("[Anthropic Decision] Allow: %v Conf: %.2f Reason: %s Thoughts: %s",
		parsed.Allow, parsed.Conf, parsed.Reason, parsed.Thoughts))
	return parsed, nil
}
