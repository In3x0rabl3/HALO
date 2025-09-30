package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"halo/logging"
	"halo/telemetry"
	"net/http"
	"os"
)

type MistralProvider struct {
	APIKey string
}

func (p *MistralProvider) Ask(t telemetry.Telemetry) (AIResponse, error) {
	return AskMistral(t, p.APIKey)
}

// AskMistral posts telemetry + system prompt to a Mistral-compatible endpoint (e.g., Ollama /api/generate).
func AskMistral(t telemetry.Telemetry, apiKey string) (AIResponse, error) {
	apiURL := os.Getenv("AI_URL")
	if apiURL == "" {
		apiURL = "http://192.168.1.145:11434/api/generate" // Ollama default
	}
	model := os.Getenv("AI_MODEL")
	if model == "" {
		model = "mistral"
	}
	if apiKey == "" {
		apiKey = os.Getenv("AI_API_KEY")
	}

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
		"temperature": 0.0,
		"stream":      true, // important for Ollama
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

	// Struct for streamed chunks
	type StreamResp struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
	}

	var fullResp string
	dec := json.NewDecoder(resp.Body)
	for dec.More() {
		var sr StreamResp
		if err := dec.Decode(&sr); err != nil {
			return AIResponse{}, err
		}
		fullResp += sr.Response
		if sr.Done {
			break
		}
	}

	logging.LogLine("[Mistral Full Response] " + fullResp)

	// Extract JSON object from the combined text
	jsonStr := ExtractJSON(fullResp)
	if jsonStr == "" {
		return AIResponse{}, errors.New("no valid JSON found in Mistral response")
	}

	var parsed AIResponse
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return AIResponse{}, err
	}

	logging.LogLine(fmt.Sprintf("[Mistral Decision] Allow: %v Conf: %.2f Reason: %s Thoughts: %s",
		parsed.Allow, parsed.Conf, parsed.Reason, parsed.Thoughts))
	return parsed, nil
}
