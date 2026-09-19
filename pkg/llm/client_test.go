package llm

import (
	"context"
	"testing"

	"github.com/sashabaranov/go-openai"
)

func TestClientMockMode(t *testing.T) {
	cfg := Config{
		MockMode: true,
	}
	client := NewClient(cfg)

	ctx := context.Background()
	msg := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: "Hello!"},
	}

	reply, err := client.Chat(ctx, msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reply != "Mock narrative response." {
		t.Errorf("got %q, want %q", reply, "Mock narrative response.")
	}

	resp, err := client.ChatWithTools(ctx, msg, nil)
	if err != nil {
		t.Fatalf("unexpected error in ChatWithTools: %v", err)
	}
	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content != "Mock response with tools available." {
		t.Errorf("unexpected tool response: %+v", resp)
	}
}
