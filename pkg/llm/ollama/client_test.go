package ollama

import (
	"context"
	"os"
	"testing"

	"github.com/aescanero/dago-libs/pkg/domain"
	"go.uber.org/zap"
)

func TestNewClient(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name     string
		endpoint string
		wantErr  bool
	}{
		{
			name:     "valid endpoint",
			endpoint: "http://localhost:11434",
			wantErr:  false,
		},
		{
			name:     "empty endpoint uses default",
			endpoint: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.endpoint, logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client")
			}
		})
	}
}

func TestGenerateCompletion(t *testing.T) {
	logger := zap.NewNop()

	t.Run("invalid request type", func(t *testing.T) {
		client, _ := NewClient("http://localhost:11434", logger)

		_, err := client.GenerateCompletion(context.Background(), "invalid")
		if err == nil {
			t.Error("GenerateCompletion() expected error for invalid request type")
		}
	})

	t.Run("valid request structure", func(t *testing.T) {
		client, _ := NewClient("http://localhost:11434", logger)

		req := &domain.LLMRequest{
			Model: "llama3.1",
			Messages: []domain.Message{
				{Role: "user", Content: "Hello"},
			},
			MaxTokens:   100,
			Temperature: 0.7,
		}

		// This will fail if Ollama is not running
		_, err := client.GenerateCompletion(context.Background(), req)
		if err == nil {
			t.Log("Note: API call succeeded (Ollama is running)")
		}
	})
}

// Integration test - only runs with OLLAMA_ENDPOINT environment variable
func TestGenerateCompletion_Integration(t *testing.T) {
	endpoint := os.Getenv("OLLAMA_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:11434"
		// Check if we should skip
		if os.Getenv("OLLAMA_TEST") == "" {
			t.Skip("OLLAMA_TEST not set, skipping integration test (set OLLAMA_TEST=1 to enable)")
		}
	}

	logger := zap.NewNop()
	client, err := NewClient(endpoint, logger)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	req := &domain.LLMRequest{
		Model: "llama3.1",
		Messages: []domain.Message{
			{Role: "user", Content: "Say 'Hello, World!' and nothing else."},
		},
		MaxTokens:   50,
		Temperature: 0.0,
	}

	resp, err := client.GenerateCompletion(context.Background(), req)
	if err != nil {
		t.Fatalf("GenerateCompletion() error = %v", err)
	}

	llmResp, ok := resp.(*domain.LLMResponse)
	if !ok {
		t.Fatal("Response is not *domain.LLMResponse")
	}

	if llmResp.Content == "" {
		t.Error("Response content is empty")
	}

	t.Logf("Response: %s", llmResp.Content)
	t.Logf("Usage: %d input tokens, %d output tokens",
		llmResp.Usage.InputTokens,
		llmResp.Usage.OutputTokens)
}
