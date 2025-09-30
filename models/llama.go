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
)

type LlamaProvider struct {
	APIKey string
}

func (p *LlamaProvider) Ask(t telemetry.Telemetry) (AIResponse, error) {
	return AskLlama(t, p.APIKey)
}

// AskLlama mirrors AskOpenAI style: systemPrompt + telemetry JSON, then POST to local Llama/Ollama or any Llama-compatible endpoint.
// AskLlama posts telemetry + system prompt to a local Llama/Ollama instance and expects strict JSON back.
func AskLlama(t telemetry.Telemetry, apiKey string) (AIResponse, error) {
	apiURL := os.Getenv("AI_URL")
	if apiURL == "" {
		apiURL = "http://192.168.1.145:11434/api/generate" // Ollama default
	}
	model := os.Getenv("AI_MODEL")
	if model == "" {
		model = "llama3"
	}
	if apiKey == "" {
		apiKey = os.Getenv("AI_API_KEY")
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

	// Ollama expects `prompt`, not `messages`/`input`
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
	logging.LogLine("[Llama Raw Response] " + string(body))

	// Ollama returns { "response": "<text>", ... }
	var out struct {
		Response string `json:"response"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return AIResponse{}, err
	}

	// Extract JSON object from the model’s response
	jsonStr := ExtractJSON(out.Response)
	if jsonStr == "" {
		return AIResponse{}, errors.New("no valid JSON found in Llama response")
	}

	var parsed AIResponse
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return AIResponse{}, err
	}

	logging.LogLine(fmt.Sprintf("[Llama Decision] Allow: %v Conf: %.2f Reason: %s Thoughts: %s",
		parsed.Allow, parsed.Conf, parsed.Reason, parsed.Thoughts))
	return parsed, nil
}
