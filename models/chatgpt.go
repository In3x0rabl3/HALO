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
	"strconv"
	"time"
)

type ChatGPTProvider struct {
	APIKey string
}

func (p *ChatGPTProvider) Ask(t telemetry.Telemetry) (AIResponse, error) {
	return AskOpenAI(t, p.APIKey)
}

// AskOpenAI queries OpenAI's ChatGPT API
func AskOpenAI(t telemetry.Telemetry, apiKey string) (AIResponse, error) {
	apiURL := DefaultAPIURL
	if apiURL == "" {
		apiURL = "https://api.openai.com/v1/chat/completions"
	}

	model := DefaultAIModel
	if model == "" {
		model = "gpt-4o"
	}

	if apiKey == "" {
		apiKey = DefaultAPIKey
	}

	payload := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": SystemPrompt},
			{"role": "user", "content": TelemetryToJSON(t)},
		},
		"temperature": 0,
	}

	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(data))
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
	logging.LogLine("[OpenAI Raw Response] " + string(body))

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
		return AIResponse{}, errors.New("no choices returned from OpenAI")
	}

	content := result.Choices[0].Message.Content

	// Step 2: log extracted GPT content
	logging.LogLine("[OpenAI Content] " + content)

	jsonStr := ExtractJSON(content)
	if jsonStr == "" {
		return AIResponse{}, errors.New("no valid JSON found in OpenAI response")
	}

	var parsed AIResponse
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return AIResponse{}, err
	}

	// Step 3: log structured decision
	logging.LogLine(
		"[OpenAI Decision] Allow: " +
			strconv.FormatBool(parsed.Allow) +
			" Conf: " + fmt.Sprintf("%.2f", parsed.Conf) +
			" Reason: " + parsed.Reason +
			" Thoughts: " + parsed.Thoughts,
	)

	return parsed, nil
}
