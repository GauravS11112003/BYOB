package providers

import (
	"fmt"
	"os"

	"github.com/GauravS11112003/BYOB/byob-backend/internal/ai"
	"github.com/GauravS11112003/BYOB/byob-backend/internal/config"
)

// Build constructs an ai.Provider from the AI configuration, reading the API key
// from the environment variable named by APIKeyEnv.
func Build(cfg config.AIConfig) (ai.Provider, error) {
	apiKey := ""
	if cfg.APIKeyEnv != "" {
		apiKey = os.Getenv(cfg.APIKeyEnv)
	}

	switch cfg.Provider {
	case "openai", "":
		return NewOpenAI(OpenAIConfig{
			APIKey:  apiKey,
			Model:   cfg.Model,
			BaseURL: cfg.BaseURL,
		}), nil
	default:
		return nil, fmt.Errorf("unknown ai provider %q", cfg.Provider)
	}
}
