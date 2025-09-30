package models

import (
	"fmt"
	"halo/telemetry"
	"os"
)

type ModelProvider interface {
	Ask(t telemetry.Telemetry) (AIResponse, error)
}

func GetProvider(name string) (ModelProvider, error) {
	apiKey := os.Getenv("AI_API_KEY")

	switch name {
	case "chatgpt":
		if apiKey == "" {
			return nil, fmt.Errorf("AI_API_KEY not set for chatgpt")
		}
		return &ChatGPTProvider{APIKey: apiKey}, nil
	case "anthropic":
		if apiKey == "" {
			return nil, fmt.Errorf("AI_API_KEY not set for anthropic")
		}
		return &AnthropicProvider{APIKey: apiKey}, nil
	case "mistral":
		if apiKey == "" {
			return nil, fmt.Errorf("AI_API_KEY not set for mistral")
		}
		return &MistralProvider{APIKey: apiKey}, nil
	case "llama":
		// Ollama/local llama usually doesn't need a key, so allow empty
		return &LlamaProvider{APIKey: apiKey}, nil
	default:
		return nil, fmt.Errorf("unknown model provider: %s", name)
	}
}
