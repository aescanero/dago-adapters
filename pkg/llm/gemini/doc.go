// Package gemini implements the LLM client adapter for Google Gemini models.
//
// This adapter implements the ports.LLMClient interface defined in dago-libs,
// providing integration with the Google AI Gemini API.
//
// Supported models:
//   - gemini-2.0-flash-exp
//   - gemini-1.5-pro
//   - gemini-1.5-flash
//
// Usage:
//
//	import "github.com/aescanero/dago-adapters/pkg/llm/gemini"
//
//	client, err := gemini.NewClient(apiKey, logger)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
//	resp, err := client.GenerateCompletion(ctx, &domain.LLMRequest{
//		Model: "gemini-2.0-flash-exp",
//		Messages: []domain.Message{
//			{Role: "user", Content: "Hello!"},
//		},
//	})
//
// Note: Gemini uses "model" role instead of "assistant" role.
// This adapter handles the conversion automatically.
package gemini
