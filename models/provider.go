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
	// Use embedded default if env is unset
	if name == "" {
		name = DefaultAIProvider
	}

	// API key only needed for cloud providers
	apiKey := os.Getenv("AI_API_KEY")
	if apiKey == "" {
		apiKey = DefaultAPIKey
	}

	switch name {
	case "chatgpt":
		if apiKey == "" {
			return nil, fmt.Errorf("AI_API_KEY is required for chatgpt")
		}
		return &ChatGPTProvider{APIKey: apiKey}, nil

	case "anthropic":
		if apiKey == "" {
			return nil, fmt.Errorf("AI_API_KEY is required for anthropic")
		}
		return &AnthropicProvider{APIKey: apiKey}, nil

	case "ollama", "llama", "mistral":
		// Local Ollama models (llama3, mistral, codellama, etc.)
		// don’t require an API key, but we still pass one if set
		return &OllamaProvider{APIKey: apiKey}, nil

	default:
		return nil, fmt.Errorf("unknown model provider: %s", name)
	}
}
