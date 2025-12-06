// Package llm provides LLM (Large Language Model) client adapters.
//
// This package contains implementations of the ports.LLMClient interface
// for various LLM providers including Anthropic, OpenAI, Gemini, and Ollama.
//
// All adapters implement the same interface defined in dago-libs/pkg/ports/llm.go,
// making them interchangeable.
//
// Usage:
//
//	import "github.com/aescanero/dago-adapters/pkg/llm"
//
//	// Create a client using the factory
//	client, err := llm.NewClient(&llm.Config{
//		Provider: "anthropic",
//		APIKey:   "your-api-key",
//		Logger:   logger,
//	})
//
//	// Use the client
//	resp, err := client.Complete(ctx, req)
package llm
