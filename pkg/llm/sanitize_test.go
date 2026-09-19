package llm

import (
	"testing"
)

func TestStripThinkingBlocks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean string without tags",
			input:    "Hello world!",
			expected: "Hello world!",
		},
		{
			name:     "basic think block",
			input:    "<think>Evaluating user intent...</think>Hello there!",
			expected: "Hello there!",
		},
		{
			name:     "multiline thought block",
			input:    "<thought>\nLine 1\nLine 2\n</thought>\n\nResult text",
			expected: "Result text",
		},
		{
			name:     "reasoning block with leading and trailing content",
			input:    "Prefix <reasoning>deep analysis</reasoning> Suffix",
			expected: "Prefix  Suffix",
		},
		{
			name:     "unclosed thinking block at end",
			input:    "Initial answer <think>Still pondering...",
			expected: "Initial answer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripThinkingBlocks(tt.input)
			if got != tt.expected {
				t.Errorf("StripThinkingBlocks() = %q, want %q", got, tt.expected)
			}
		})
	}
}
