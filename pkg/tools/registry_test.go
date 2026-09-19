package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type dummyTool struct {
	name string
	tags []string
	fn   func(ctx context.Context, args map[string]any) (string, error)
}

func (d *dummyTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        d.name,
			Description: "A dummy test tool",
			Parameters:  jsonschema.Definition{Type: jsonschema.Object},
		},
	}
}

func (d *dummyTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	if d.fn != nil {
		return d.fn(ctx, args)
	}
	return "ok", nil
}

func (d *dummyTool) Tags() []string {
	return d.tags
}

func TestRegistry(t *testing.T) {
	reg := NewRegistry()

	tool1 := &dummyTool{name: "search", tags: []string{"web"}}
	tool2 := &dummyTool{name: "cook", tags: []string{"kitchen"}}
	tool3 := &dummyTool{name: "clock", tags: nil} // universal

	reg.Register(tool1, tool2, tool3)

	if len(reg.AllOpenAITools()) != 3 {
		t.Fatalf("expected 3 tools, got %d", len(reg.AllOpenAITools()))
	}

	// Filter by kitchen: should return cook and universal clock
	kitchenTools := reg.FilterByTags("kitchen")
	if len(kitchenTools) != 2 {
		t.Fatalf("expected 2 tools for kitchen, got %d", len(kitchenTools))
	}

	// Execute tool3
	out, err := reg.Execute(context.Background(), "clock", "{}")
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if out != `{"success":true,"data":"ok"}` {
		t.Errorf("unexpected output: %s", out)
	}

	// Error tool
	errTool := &dummyTool{
		name: "fail",
		fn: func(ctx context.Context, args map[string]any) (string, error) {
			return "", errors.New("boom")
		},
	}
	reg.Register(errTool)

	errOut, err := reg.Execute(context.Background(), "fail", "{}")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if errOut != `{"success":false,"code":500,"error":"boom"}` {
		t.Errorf("unexpected error payload: %s", errOut)
	}
}
