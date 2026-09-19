package llm

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// RequestOptions configures per-call parameters.
type RequestOptions struct {
	Model       string
	Temperature *float32
	MaxTokens   int
}

// Option modifies RequestOptions.
type Option func(*RequestOptions)

// WithTemperature sets sampling temperature.
func WithTemperature(t float32) Option {
	return func(o *RequestOptions) {
		o.Temperature = &t
	}
}

// WithMaxTokens sets response token limit.
func WithMaxTokens(n int) Option {
	return func(o *RequestOptions) {
		o.MaxTokens = n
	}
}

// WithModel overrides default model.
func WithModel(m string) Option {
	return func(o *RequestOptions) {
		o.Model = m
	}
}

// Client provides unified chat completion and tool calling against OpenAI-compatible backends.
type Client struct {
	cfg       Config
	apiClient *openai.Client
}

// NewClient creates an LLM client with the specified configuration.
func NewClient(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://192.168.86.208:8000/api/v1"
	}
	if cfg.APIKey == "" {
		cfg.APIKey = "dummy-key"
	}
	if cfg.Model == "" {
		cfg.Model = "Gemma-4-26B-A4B-it-MTP-GGUF"
	}
	if cfg.ToolModel == "" {
		cfg.ToolModel = cfg.Model
	}

	oaiCfg := openai.DefaultConfig(cfg.APIKey)
	oaiCfg.BaseURL = cfg.BaseURL

	return &Client{
		cfg:       cfg,
		apiClient: openai.NewClientWithConfig(oaiCfg),
	}
}

// Config returns the current client configuration.
func (c *Client) Config() Config {
	return c.cfg
}

// Chat executes a simple text chat completion.
func (c *Client) Chat(ctx context.Context, messages []openai.ChatCompletionMessage, opts ...Option) (string, error) {
	if c.cfg.MockMode {
		return "Mock narrative response.", nil
	}

	opt := RequestOptions{
		Model: c.cfg.Model,
	}
	for _, fn := range opts {
		fn(&opt)
	}

	req := openai.ChatCompletionRequest{
		Model:     opt.Model,
		Messages:  messages,
		MaxTokens: opt.MaxTokens,
	}
	if opt.Temperature != nil {
		req.Temperature = *opt.Temperature
	}

	resp, err := c.apiClient.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("llm chat error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("llm returned no choices")
	}

	content := resp.Choices[0].Message.Content
	if c.cfg.DisableThinking {
		content = StripThinkingBlocks(content)
	}

	return content, nil
}

// ChatWithTools executes a chat completion passing OpenAI-compatible tool specifications.
func (c *Client) ChatWithTools(ctx context.Context, messages []openai.ChatCompletionMessage, tools []openai.Tool, opts ...Option) (openai.ChatCompletionResponse, error) {
	if c.cfg.MockMode {
		return openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{
					Message: openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleAssistant,
						Content: "Mock response with tools available.",
					},
				},
			},
		}, nil
	}

	opt := RequestOptions{
		Model: c.cfg.Model,
	}
	for _, fn := range opts {
		fn(&opt)
	}

	req := openai.ChatCompletionRequest{
		Model:     opt.Model,
		Messages:  messages,
		Tools:     tools,
		MaxTokens: opt.MaxTokens,
	}
	if opt.Temperature != nil {
		req.Temperature = *opt.Temperature
	}

	resp, err := c.apiClient.CreateChatCompletion(ctx, req)
	if err != nil {
		return openai.ChatCompletionResponse{}, fmt.Errorf("llm chat with tools error: %w", err)
	}

	if c.cfg.DisableThinking && len(resp.Choices) > 0 {
		resp.Choices[0].Message.Content = StripThinkingBlocks(resp.Choices[0].Message.Content)
	}

	return resp, nil
}

// CallToolModel executes tool selection using the fast reflex model (e.g. Granite-4-tiny).
func (c *Client) CallToolModel(ctx context.Context, messages []openai.ChatCompletionMessage, tools []openai.Tool, opts ...Option) (openai.ChatCompletionResponse, error) {
	opts = append([]Option{WithModel(c.cfg.ToolModel)}, opts...)
	return c.ChatWithTools(ctx, messages, tools, opts...)
}
