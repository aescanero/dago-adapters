package gemini

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
		name    string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "valid api key",
			apiKey:  "test-key",
			wantErr: false,
		},
		{
			name:    "empty api key",
			apiKey:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.apiKey, logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client")
			}
			if client != nil {
				_ = client.Close()
			}
		})
	}
}

func TestGenerateCompletion(t *testing.T) {
	logger := zap.NewNop()

	t.Run("invalid request type", func(t *testing.T) {
		client, _ := NewClient("test-key", logger)
		defer func() { _ = client.Close() }()

		_, err := client.GenerateCompletion(context.Background(), "invalid")
		if err == nil {
			t.Error("GenerateCompletion() expected error for invalid request type")
		}
	})

	t.Run("valid request structure", func(t *testing.T) {
		client, _ := NewClient("test-key", logger)
		defer func() { _ = client.Close() }()

		req := &domain.LLMRequest{
			Model: "gemini-2.0-flash-exp",
			Messages: []domain.Message{
				{Role: "user", Content: "Hello"},
			},
			MaxTokens:   100,
			Temperature: 0.7,
		}

		// This will fail with API error since we don't have a real key
		_, err := client.GenerateCompletion(context.Background(), req)
		if err == nil {
			t.Log("Note: API call succeeded (real API key present?)")
		}
	})
}

// Integration test - only runs with GEMINI_API_KEY environment variable
func TestGenerateCompletion_Integration(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping integration test")
	}

	logger := zap.NewNop()
	client, err := NewClient(apiKey, logger)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	req := &domain.LLMRequest{
		Model: "gemini-2.0-flash-exp",
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
