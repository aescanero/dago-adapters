// Package anthropic implements the LLM client adapter for Anthropic Claude.
//
// This adapter implements the ports.LLMClient interface defined in dago-libs,
// providing integration with the Anthropic API for Claude models.
//
// Supported models:
//   - claude-sonnet-4-20250514
//   - claude-opus-4-20250514
//   - claude-haiku-3-5-20241022
//
// Usage:
//
//	import "github.com/aescanero/dago-adapters/pkg/llm/anthropic"
//
//	client, err := anthropic.NewClient(apiKey, logger)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	resp, err := client.GenerateCompletion(ctx, &domain.LLMRequest{
//		Model: "claude-sonnet-4-20250514",
//		Messages: []domain.Message{
//			{Role: "user", Content: "Hello!"},
//		},
//	})
package anthropic
