package llm

import (
	"testing"

	"go.uber.org/zap"
)

func TestNewClient(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name     string
		provider string
		apiKey   string
		wantErr  bool
	}{
		{
			name:     "anthropic with api key",
			provider: "anthropic",
			apiKey:   "test-key",
			wantErr:  false,
		},
		{
			name:     "claude alias",
			provider: "claude",
			apiKey:   "test-key",
			wantErr:  false,
		},
		{
			name:     "anthropic without api key",
			provider: "anthropic",
			apiKey:   "",
			wantErr:  true,
		},
		{
			name:     "openai with api key",
			provider: "openai",
			apiKey:   "test-key",
			wantErr:  false,
		},
		{
			name:     "gemini with api key",
			provider: "gemini",
			apiKey:   "test-key",
			wantErr:  false,
		},
		{
			name:     "ollama with endpoint",
			provider: "ollama",
			apiKey:   "",
			wantErr:  false,
		},
		{
			name:     "unsupported provider",
			provider: "unsupported",
			apiKey:   "test-key",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Provider: tt.provider,
				APIKey:   tt.apiKey,
				Logger:   logger,
			}

			client, err := NewClient(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client without error")
			}
		})
	}
}

func TestGetDefaultModel(t *testing.T) {
	tests := []struct {
		provider string
		want     string
	}{
		{"anthropic", "claude-sonnet-4-20250514"},
		{"claude", "claude-sonnet-4-20250514"},
		{"openai", "gpt-4o"},
		{"gpt", "gpt-4o"},
		{"gemini", "gemini-2.0-flash-exp"},
		{"google", "gemini-2.0-flash-exp"},
		{"ollama", "llama3.1"},
		{"local", "llama3.1"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			got := GetDefaultModel(tt.provider)
			if got != tt.want {
				t.Errorf("GetDefaultModel(%s) = %s, want %s", tt.provider, got, tt.want)
			}
		})
	}
}

func TestListSupportedProviders(t *testing.T) {
	providers := ListSupportedProviders()

	if len(providers) == 0 {
		t.Error("ListSupportedProviders() returned empty list")
	}

	expectedProviders := map[string]bool{
		"anthropic": true,
		"openai":    true,
		"gemini":    true,
		"ollama":    true,
	}

	for _, provider := range providers {
		if !expectedProviders[provider] {
			t.Errorf("Unexpected provider in list: %s", provider)
		}
		delete(expectedProviders, provider)
	}

	if len(expectedProviders) > 0 {
		t.Errorf("Missing providers: %v", expectedProviders)
	}
}
