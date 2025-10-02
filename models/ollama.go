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
	"strings"
)

// OllamaProvider supports all Ollama-hosted models (phi3:mini, llama3, mistral, neural-chat, phi, etc.)
type OllamaProvider struct {
	APIURL string
	Model  string
}

// Ask runs the telemetry through the selected Ollama model.
func (p *OllamaProvider) Ask(t telemetry.Telemetry) (AIResponse, error) {
	apiURL := p.APIURL
	if apiURL == "" {
		apiURL = "http://localhost:11434/api/generate" // default Ollama endpoint
	}

	model := p.Model
	if model == "" {
		model = "llama3" // fallback default
	}

	return AskOllama(t, apiURL, model)
}

// AskOllama handles the HTTP request/response flow for any Ollama model.
func AskOllama(t telemetry.Telemetry, apiURL, model string) (AIResponse, error) {
	// Normalize model input to handle shorthand like "phi3"
	if model == "" {
		model = ""
	}
	if !strings.Contains(model, ":") && model != "llama3" && model != "mistral" && model != "neural-chat" {
		// append :latest if not specified for models that default to latest
		model = model + ":latest"
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
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 0} // no timeout, supports big models
	resp, err := client.Do(req)
	if err != nil {
		return AIResponse{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	logging.LogLine(fmt.Sprintf("[Ollama Raw Response][%s] %s", model, string(body)))

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

	logging.LogLine(fmt.Sprintf("[Ollama Decision][%s] Allow: %v Conf: %.2f Reason: %s Thoughts: %s",
		model, parsed.Allow, parsed.Conf, parsed.Reason, parsed.Thoughts))
	return parsed, nil
}
