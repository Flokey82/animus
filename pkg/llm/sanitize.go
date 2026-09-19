package llm

import (
	"regexp"
	"strings"
)

var (
	thinkingRegex = regexp.MustCompile(`(?is)<(think|reasoning|thought)>.*?</(think|reasoning|thought)>`)
	// In case of unclosed tags at output truncation
	unclosedThinkingRegex = regexp.MustCompile(`(?is)<(think|reasoning|thought)>.*$`)
)

// StripThinkingBlocks removes <think>...</think>, <thought>...</thought>, and <reasoning>...</reasoning> blocks.
func StripThinkingBlocks(text string) string {
	text = thinkingRegex.ReplaceAllString(text, "")
	text = unclosedThinkingRegex.ReplaceAllString(text, "")
	return strings.TrimSpace(text)
}
