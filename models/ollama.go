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
)

// Unified provider for all Ollama-hosted models (llama3, mistral, codellama, etc.)
type OllamaProvider struct {
	APIKey string
}

func (p *OllamaProvider) Ask(t telemetry.Telemetry) (AIResponse, error) {
	return AskOllama(t, p.APIKey)
}

func AskOllama(t telemetry.Telemetry, apiKey string) (AIResponse, error) {
	apiURL := DefaultAPIURL
	if apiURL == "" {
		apiURL = "http://localhost:11434/api/generate" // fallback Ollama default
	}

	model := DefaultAIModel
	if model == "" {
		model = "llama3"
	}

	if apiKey == "" {
		apiKey = DefaultAPIKey
	}

	// Build strict JSON-only prompt
	prompt := SystemPrompt + "\n\n" + TelemetryToJSON(t) + `
IMPORTANT:
Respond ONLY with a valid JSON object in exactly this schema:

{
  "allow": true | false,
  "conf": 0.xx,
  "reason": "short string",
  "thoughts": "short string"
}`

	payload := map[string]interface{}{
		"model":       model,
		"prompt":      prompt,
		"stream":      false,
		"temperature": 0.0,
	}

	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(data))
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 0} // 0 = no timeout (streaming)
	resp, err := client.Do(req)
	if err != nil {
		return AIResponse{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	logging.LogLine("[Ollama Raw Response] " + string(body))

	var out struct {
		Response string `json:"response"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return AIResponse{}, err
	}

	// Extract JSON object
	jsonStr := ExtractJSON(out.Response)
	if jsonStr == "" {
		return AIResponse{}, errors.New("no valid JSON found in Ollama response")
	}

	var parsed AIResponse
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return AIResponse{}, err
	}

	logging.LogLine(fmt.Sprintf("[Ollama Decision] Allow: %v Conf: %.2f Reason: %s Thoughts: %s",
		parsed.Allow, parsed.Conf, parsed.Reason, parsed.Thoughts))
	return parsed, nil
}
