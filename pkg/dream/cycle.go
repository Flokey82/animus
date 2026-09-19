package dream

import (
	"context"
	"fmt"
	"strings"

	"github.com/Flokey82/animus/pkg/llm"
	"github.com/sashabaranov/go-openai"
)

// DreamResult holds the synthesized output of an agent's sleep cycle.
type DreamResult struct {
	DreamNarrative   string `json:"dream_narrative"`
	Reflection       string `json:"reflection"`
	EvolvedIdentity  string `json:"evolved_identity"`
	ConsolidatedText string `json:"consolidated_text"`
	Intensity        int    `json:"intensity"`
}

// Cycle manages subconscious dream synthesis, memory consolidation, and identity evolution.
type Cycle struct {
	client *llm.Client
}

// NewCycle creates a sleep and dream cycle manager.
func NewCycle(client *llm.Client) *Cycle {
	return &Cycle{client: client}
}

// Run executes the sleep cycle given the agent's identity, day's episodic log, and active memories.
func (c *Cycle) Run(ctx context.Context, agentName, baseIdentity, episodicLog, activeMemories string) (*DreamResult, error) {
	if strings.TrimSpace(episodicLog) == "" {
		episodicLog = "A quiet, peaceful day with gentle routines and quiet rest."
	}

	// 1. Synthesize Dream Narrative
	dreamPrompt := fmt.Sprintf(`You are the subconscious dream engine of %s.
She is asleep for the night. Synthesize a vivid, imaginative, yet coherent and cozy dream narrative (2-3 sentences).
Draw directly upon elements of her actual day:
--- TODAY'S LOGS ---
%s
--- ACTIVE MEMORIES ---
%s

Return ONLY the dream narrative.`, agentName, episodicLog, activeMemories)

	dreamNarrative, err := c.client.Chat(ctx, []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: dreamPrompt},
	}, llm.WithTemperature(0.85), llm.WithMaxTokens(150))
	if err != nil {
		return nil, fmt.Errorf("dream generation error: %w", err)
	}

	// 2. Daily Consolidation Summary
	consolidationPrompt := fmt.Sprintf(`Summarize the key events and emotional takeaways from %s's day into 1-2 reflective sentences for long-term memory:
%s`, agentName, episodicLog)

	consolidationText, err := c.client.Chat(ctx, []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: consolidationPrompt},
	}, llm.WithTemperature(0.5), llm.WithMaxTokens(100))
	if err != nil {
		consolidationText = "Reflected upon the events of the day in quiet contemplation."
	}

	return &DreamResult{
		DreamNarrative:   strings.TrimSpace(dreamNarrative),
		Reflection:       fmt.Sprintf("Waking with lingering impressions of %s.", agentName),
		EvolvedIdentity:  baseIdentity,
		ConsolidatedText: strings.TrimSpace(consolidationText),
		Intensity:        3,
	}, nil
}
