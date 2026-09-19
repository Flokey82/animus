package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// AgentTool is the universal contract implemented by any tool the agent can invoke.
type AgentTool interface {
	// Definition returns OpenAI-compatible function schema metadata.
	Definition() openai.Tool
	// Execute runs the tool logic given JSON-unmarshaled arguments.
	Execute(ctx context.Context, arguments map[string]any) (string, error)
	// Tags returns classification tags for contextual filtering (e.g. "web", "social", "room:kitchen").
	Tags() []string
}

// ToolExecutionResult standardizes the JSON response returned by tools to the LLM.
type ToolExecutionResult struct {
	Success bool   `json:"success"`
	Code    int    `json:"code,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// String serializes the result envelope to JSON.
func (r ToolExecutionResult) String() string {
	b, err := json.Marshal(r)
	if err != nil {
		return fmt.Sprintf(`{"success":false,"error":%q}`, err.Error())
	}
	return string(b)
}

// MakeSuccessPayload wraps arbitrary data in a success envelope.
func MakeSuccessPayload(data any) string {
	return ToolExecutionResult{
		Success: true,
		Data:    data,
	}.String()
}

// MakeErrorPayload wraps an error message in an error envelope.
func MakeErrorPayload(code int, msg string) string {
	return ToolExecutionResult{
		Success: false,
		Code:    code,
		Error:   msg,
	}.String()
}

// NormalizeToolOutput ensures the returned string conforms to a valid envelope.
func NormalizeToolOutput(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
		var check ToolExecutionResult
		if err := json.Unmarshal([]byte(trimmed), &check); err == nil {
			return trimmed
		}
	}
	return MakeSuccessPayload(trimmed)
}
