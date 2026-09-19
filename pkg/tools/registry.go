package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/sashabaranov/go-openai"
)

// Registry manages tool registration, contextual exposure, and execution dispatching.
type Registry struct {
	mu    sync.RWMutex
	tools map[string]AgentTool
}

// NewRegistry initializes an empty tool registry.
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]AgentTool),
	}
}

// Register registers one or more tools into the registry.
func (r *Registry) Register(ts ...AgentTool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range ts {
		name := t.Definition().Function.Name
		r.tools[name] = t
	}
}

// Get retrieves a tool by name.
func (r *Registry) Get(name string) (AgentTool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	return t, ok
}

// AllOpenAITools returns all registered tools formatted for OpenAI chat completions.
func (r *Registry) AllOpenAITools() []openai.Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]openai.Tool, 0, len(r.tools))
	for _, t := range r.tools {
		res = append(res, t.Definition())
	}
	return res
}

// FilterByTags returns tools that match ANY of the provided tags or have no tags (universal).
// If no tags are provided, it returns all tools.
func (r *Registry) FilterByTags(tags ...string) []openai.Tool {
	if len(tags) == 0 {
		return r.AllOpenAITools()
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	tagSet := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		tagSet[tag] = struct{}{}
	}

	var matched []openai.Tool
	for _, t := range r.tools {
		toolTags := t.Tags()
		if len(toolTags) == 0 {
			// Universal tool without tag restriction
			matched = append(matched, t.Definition())
			continue
		}
		for _, tt := range toolTags {
			if _, ok := tagSet[tt]; ok {
				matched = append(matched, t.Definition())
				break
			}
		}
	}
	return matched
}

// Execute unmarshals arguments and runs the named tool, returning a normalized payload.
func (r *Registry) Execute(ctx context.Context, name string, rawArguments string) (string, error) {
	tool, ok := r.Get(name)
	if !ok {
		return MakeErrorPayload(404, fmt.Sprintf("tool %q not found", name)), fmt.Errorf("tool not found: %s", name)
	}

	args := make(map[string]any)
	if rawArguments != "" && rawArguments != "{}" {
		if err := json.Unmarshal([]byte(rawArguments), &args); err != nil {
			return MakeErrorPayload(400, fmt.Sprintf("invalid arguments json: %s", err)), err
		}
	}

	out, err := tool.Execute(ctx, args)
	if err != nil {
		return MakeErrorPayload(500, err.Error()), err
	}
	return NormalizeToolOutput(out), nil
}
