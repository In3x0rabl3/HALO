package models

import (
	"fmt"
	"halo/telemetry"
	"os"
)

// ModelProvider is a generic AI provider interface.
type ModelProvider interface {
	Ask(t telemetry.Telemetry) (AIResponse, error)
}

// GetProvider returns the correct ModelProvider implementation based on the provider name.
// It checks environment variables for overrides and uses sensible defaults.
func GetProvider(name string) (ModelProvider, error) {
	// Load from env vars, or use provided/fallback values.
	if name == "" {
		name = os.Getenv("AI_PROVIDER")
		if name == "" {
			name = DefaultAIProvider
		}
	}

	apiKey := os.Getenv("AI_API_KEY")
	if apiKey == "" {
		apiKey = DefaultAPIKey
	}

	apiURL := os.Getenv("AI_API_URL")
	if apiURL == "" {
		apiURL = DefaultAPIURL
	}

	model := os.Getenv("AI_MODEL")
	if model == "" {
		model = DefaultAIModel
	}
	if model == "" {
		model = name // fallback: use provider name as model
	}

	// Set default API URLs if still empty
	if apiURL == "" {
		switch name {
		case "chatgpt":
			apiURL = "https://api.openai.com/v1/chat/completions"
		case "anthropic":
			apiURL = "https://api.anthropic.com/v1/complete"
		default:
			apiURL = "http://localhost:11434/api/generate" // Ollama default
		}
	}

	switch name {
	case "chatgpt":
		if apiKey == "" {
			return nil, fmt.Errorf("AI_API_KEY is required for chatgpt")
		}
		return &ChatGPTProvider{
			APIKey: apiKey,
		}, nil

	case "anthropic":
		if apiKey == "" {
			return nil, fmt.Errorf("AI_API_KEY is required for anthropic")
		}
		return &AnthropicProvider{
			APIKey: apiKey,
		}, nil

	case "ollama", "llama3", "mistral", "neural-chat":
		// Local Ollama models don’t require an API key
		return &OllamaProvider{
			APIURL: apiURL,
			Model:  model,
		}, nil

	default:
		return nil, fmt.Errorf("unknown model provider: %s", name)
	}
}
