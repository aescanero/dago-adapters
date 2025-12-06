// Package ollama implements the LLM client adapter for Ollama local models.
//
// This adapter implements the ports.LLMClient interface defined in dago-libs,
// providing integration with Ollama for running LLMs locally.
//
// Supported models (examples):
//   - llama3.1
//   - llama3.1:70b
//   - mistral
//   - codellama
//   - phi3
//   - Any model available in your local Ollama installation
//
// Usage:
//
//	import "github.com/aescanero/dago-adapters/pkg/llm/ollama"
//
//	client, err := ollama.NewClient("http://localhost:11434", logger)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	resp, err := client.GenerateCompletion(ctx, &domain.LLMRequest{
//		Model: "llama3.1",
//		Messages: []domain.Message{
//			{Role: "user", Content: "Hello!"},
//		},
//	})
//
// Note: Ollama must be running locally or accessible at the specified endpoint.
// The default endpoint is http://localhost:11434
package ollama
