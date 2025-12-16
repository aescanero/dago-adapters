package llm

import (
	"fmt"

	"github.com/aescanero/dago-adapters/pkg/llm/anthropic"
	"github.com/aescanero/dago-adapters/pkg/llm/gemini"
	"github.com/aescanero/dago-adapters/pkg/llm/ollama"
	"github.com/aescanero/dago-adapters/pkg/llm/openai"
	"github.com/aescanero/dago-libs/pkg/ports"
	"go.uber.org/zap"
)

// Config holds LLM client configuration
type Config struct {
	Provider string
	APIKey   string
	BaseURL  string // For Ollama
	Timeout  int    // Timeout in seconds
	Logger   *zap.Logger
}

// NewClient creates a new LLM client based on provider
func NewClient(cfg *Config) (ports.LLMClient, error) {
	if cfg.Logger == nil {
		cfg.Logger = zap.NewNop()
	}

	switch cfg.Provider {
	case "anthropic", "claude":
		return anthropic.NewClient(cfg.APIKey, cfg.Logger)

	case "openai", "gpt":
		return openai.NewClient(cfg.APIKey, cfg.BaseURL, cfg.Logger)

	case "gemini", "google":
		return gemini.NewClient(cfg.APIKey, cfg.Logger)

	case "ollama", "local":
		endpoint := cfg.BaseURL
		if endpoint == "" {
			endpoint = "http://localhost:11434"
		}
		return ollama.NewClient(endpoint, cfg.Logger)

	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s (supported: anthropic, openai, gemini, ollama)", cfg.Provider)
	}
}

// GetDefaultModel returns the default model for a provider
func GetDefaultModel(provider string) string {
	switch provider {
	case "anthropic", "claude":
		return "claude-sonnet-4-20250514"
	case "openai", "gpt":
		return "gpt-4o"
	case "gemini", "google":
		return "gemini-2.0-flash-exp"
	case "ollama", "local":
		return "llama3.1"
	default:
		return ""
	}
}

// ListSupportedProviders returns a list of supported LLM providers
func ListSupportedProviders() []string {
	return []string{
		"anthropic",
		"openai",
		"gemini",
		"ollama",
	}
}
