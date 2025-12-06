package gemini

import (
	"context"
	"fmt"

	"github.com/aescanero/dago-libs/pkg/domain"
	"github.com/aescanero/dago-libs/pkg/ports"
	"github.com/google/generative-ai-go/genai"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

// Client implements the LLMClient interface for Google Gemini models
type Client struct {
	client *genai.Client
	logger *zap.Logger
}

// NewClient creates a new Gemini client
func NewClient(apiKey string, logger *zap.Logger) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &Client{
		client: client,
		logger: logger,
	}, nil
}

// Close closes the Gemini client
func (c *Client) Close() error {
	return c.client.Close()
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

	// Create model
	model := c.client.GenerativeModel(llmReq.Model)

	// Configure generation parameters
	if llmReq.Temperature > 0 {
		temp := float32(llmReq.Temperature)
		model.Temperature = &temp
	}

	if llmReq.MaxTokens > 0 {
		maxTokens := int32(llmReq.MaxTokens)
		model.MaxOutputTokens = &maxTokens
	}

	// Start chat session
	chat := model.StartChat()

	// Add system message as first user message if present
	if llmReq.System != "" {
		chat.History = append(chat.History, &genai.Content{
			Parts: []genai.Part{
				genai.Text(llmReq.System),
			},
			Role: "user",
		})
	}

	// Add history (all messages except the last one)
	for i := 0; i < len(llmReq.Messages)-1; i++ {
		msg := llmReq.Messages[i]
		role := "user"
		if msg.Role == "assistant" {
			role = "model" // Gemini uses "model" instead of "assistant"
		}

		chat.History = append(chat.History, &genai.Content{
			Parts: []genai.Part{
				genai.Text(msg.Content),
			},
			Role: role,
		})
	}

	// Send the last message
	lastMsg := llmReq.Messages[len(llmReq.Messages)-1]
	resp, err := chat.SendMessage(ctx, genai.Text(lastMsg.Content))
	if err != nil {
		c.logger.Error("API call failed", zap.Error(err))
		return nil, fmt.Errorf("API call failed: %w", err)
	}

	// Extract content
	content := ""
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			content = string(text)
		}
	}

	// Calculate token usage from candidates
	inputTokens := 0
	outputTokens := 0
	if len(resp.Candidates) > 0 && resp.Candidates[0].TokenCount > 0 {
		outputTokens = int(resp.Candidates[0].TokenCount)
		// Estimate input tokens (Gemini v0.8.0 doesn't provide separate input count)
		// Approximate based on message sizes
		for _, msg := range llmReq.Messages {
			inputTokens += len(msg.Content) / 4 // Rough approximation
		}
	}

	// Convert response
	llmResp := &domain.LLMResponse{
		Content: content,
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
