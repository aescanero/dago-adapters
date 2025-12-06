// Package openai implements the LLM client adapter for OpenAI GPT models.
//
// This adapter implements the ports.LLMClient interface defined in dago-libs,
// providing integration with the OpenAI API for GPT models.
//
// Supported models:
//   - gpt-4o
//   - gpt-4o-mini
//   - gpt-4-turbo
//   - gpt-4
//   - gpt-3.5-turbo
//
// Usage:
//
//	import "github.com/aescanero/dago-adapters/pkg/llm/openai"
//
//	client, err := openai.NewClient(apiKey, logger)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	resp, err := client.GenerateCompletion(ctx, &domain.LLMRequest{
//		Model: "gpt-4o",
//		Messages: []domain.Message{
//			{Role: "user", Content: "Hello!"},
//		},
//	})
package openai
