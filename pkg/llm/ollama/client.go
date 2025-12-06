package ollama

import (
	"context"
	"fmt"

	"github.com/aescanero/dago-libs/pkg/domain"
	"github.com/aescanero/dago-libs/pkg/ports"
	"github.com/ollama/ollama/api"
	"go.uber.org/zap"
)

// Client implements the LLMClient interface for Ollama local models
type Client struct {
	client   *api.Client
	endpoint string
	logger   *zap.Logger
}

// NewClient creates a new Ollama client
// endpoint is the Ollama server URL (e.g., "http://localhost:11434")
func NewClient(endpoint string, logger *zap.Logger) (*Client, error) {
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}

	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to create Ollama client: %w", err)
	}

	return &Client{
		client:   client,
		endpoint: endpoint,
		logger:   logger,
	}, nil
}

// Complete performs a standard text completion (ports.LLMClient interface)
func (c *Client) Complete(ctx context.Context, req ports.CompletionRequest) (*ports.CompletionResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

// CompleteWithTools performs a completion with tool calling support (ports.LLMClient interface)
func (c *Client) CompleteWithTools(ctx context.Context, req ports.CompletionRequest, tools []ports.Tool) (*ports.CompletionResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

// CompleteStructured performs a completion with guaranteed JSON schema conformance (ports.LLMClient interface)
func (c *Client) CompleteStructured(ctx context.Context, req ports.CompletionRequest, schema ports.JSONSchema) (*ports.StructuredResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

// GenerateCompletion generates a completion using domain.LLMRequest (compatibility method)
func (c *Client) GenerateCompletion(ctx context.Context, req interface{}) (interface{}, error) {
	// Type assert the request
	llmReq, ok := req.(*domain.LLMRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type")
	}

	c.logger.Debug("generating completion",
		zap.String("model", llmReq.Model),
		zap.Int("message_count", len(llmReq.Messages)))

	// Build messages for Ollama
	messages := make([]api.Message, 0, len(llmReq.Messages))

	// Add system message if present
	if llmReq.System != "" {
		messages = append(messages, api.Message{
			Role:    "system",
			Content: llmReq.System,
		})
	}

	// Add conversation messages
	for _, msg := range llmReq.Messages {
		messages = append(messages, api.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Build chat request
	chatReq := &api.ChatRequest{
		Model:    llmReq.Model,
		Messages: messages,
	}

	// Set optional parameters
	if llmReq.Temperature > 0 {
		chatReq.Options = map[string]interface{}{
			"temperature": llmReq.Temperature,
		}
	}

	if llmReq.MaxTokens > 0 {
		if chatReq.Options == nil {
			chatReq.Options = make(map[string]interface{})
		}
		chatReq.Options["num_predict"] = llmReq.MaxTokens
	}

	// Make the API call
	var response api.ChatResponse
	err := c.client.Chat(ctx, chatReq, func(resp api.ChatResponse) error {
		response = resp
		return nil
	})

	if err != nil {
		c.logger.Error("API call failed", zap.Error(err))
		return nil, fmt.Errorf("API call failed: %w", err)
	}

	// Ollama provides token counts in the response
	inputTokens := 0
	outputTokens := 0

	if response.PromptEvalCount > 0 {
		inputTokens = response.PromptEvalCount
	}
	if response.EvalCount > 0 {
		outputTokens = response.EvalCount
	}

	// Convert response
	llmResp := &domain.LLMResponse{
		Content: response.Message.Content,
		Model:   llmReq.Model,
		Usage: domain.Usage{
			InputTokens:  inputTokens,
			OutputTokens: outputTokens,
		},
	}

	c.logger.Debug("completion generated",
		zap.Int("input_tokens", llmResp.Usage.InputTokens),
		zap.Int("output_tokens", llmResp.Usage.OutputTokens))

	return llmResp, nil
}
