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
<<<<<<< HEAD
	// Use embedded default if env is unset
	if name == "" {
		name = DefaultAIProvider
	}

	// API key only needed for ChatGPT and Anthropic
	apiKey := os.Getenv("AI_API_KEY")
	if apiKey == "" {
		apiKey = DefaultAPIKey
	}
=======
	apiKey := os.Getenv("AI_API_KEY")
>>>>>>> b190e59cb276021db88116db300eaf6555ffbf9a

	switch name {
	case "chatgpt":
		if apiKey == "" {
<<<<<<< HEAD
			return nil, fmt.Errorf("AI_API_KEY is required for chatgpt")
=======
			return nil, fmt.Errorf("AI_API_KEY not set for chatgpt")
>>>>>>> b190e59cb276021db88116db300eaf6555ffbf9a
		}
		return &ChatGPTProvider{APIKey: apiKey}, nil
	case "anthropic":
		if apiKey == "" {
<<<<<<< HEAD
			return nil, fmt.Errorf("AI_API_KEY is required for anthropic")
		}
		return &AnthropicProvider{APIKey: apiKey}, nil
	case "mistral":
		return &MistralProvider{APIKey: apiKey}, nil // optional
	case "llama":
		return &LlamaProvider{APIKey: apiKey}, nil // optional
=======
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
>>>>>>> b190e59cb276021db88116db300eaf6555ffbf9a
	default:
		return nil, fmt.Errorf("unknown model provider: %s", name)
	}
}
