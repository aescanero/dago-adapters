package openai

import (
	"context"
	"fmt"

	"github.com/aescanero/dago-libs/pkg/domain"
	"github.com/aescanero/dago-libs/pkg/ports"
	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

// Client implements the LLMClient interface for OpenAI GPT models
type Client struct {
	client *openai.Client
	logger *zap.Logger
}

// NewClient creates a new OpenAI client
func NewClient(apiKey string, logger *zap.Logger) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	client := openai.NewClient(apiKey)

	return &Client{
		client: client,
		logger: logger,
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

	// Convert messages to OpenAI format
	messages := make([]openai.ChatCompletionMessage, 0, len(llmReq.Messages))

	// Add system message if present
	if llmReq.System != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: llmReq.System,
		})
	}

	// Add conversation messages
	for _, msg := range llmReq.Messages {
		role := ""
		switch msg.Role {
		case "user":
			role = openai.ChatMessageRoleUser
		case "assistant":
			role = openai.ChatMessageRoleAssistant
		case "system":
			role = openai.ChatMessageRoleSystem
		default:
			c.logger.Warn("unknown message role, defaulting to user", zap.String("role", msg.Role))
			role = openai.ChatMessageRoleUser
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	// Build request
	chatReq := openai.ChatCompletionRequest{
		Model:    llmReq.Model,
		Messages: messages,
	}

	if llmReq.MaxTokens > 0 {
		chatReq.MaxTokens = llmReq.MaxTokens
	}

	if llmReq.Temperature > 0 {
		chatReq.Temperature = float32(llmReq.Temperature)
	}

	// Call API
	resp, err := c.client.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		c.logger.Error("API call failed", zap.Error(err))
		return nil, fmt.Errorf("API call failed: %w", err)
	}

	// Extract content
	content := ""
	if len(resp.Choices) > 0 {
		content = resp.Choices[0].Message.Content
	}

	// Convert response
	llmResp := &domain.LLMResponse{
		Content: content,
		Model:   resp.Model,
		Usage: domain.Usage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
		},
	}

	c.logger.Debug("completion generated",
		zap.Int("input_tokens", llmResp.Usage.InputTokens),
		zap.Int("output_tokens", llmResp.Usage.OutputTokens))

	return llmResp, nil
}
