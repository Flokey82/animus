package engine

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Flokey82/animus/pkg/channel"
	"github.com/Flokey82/animus/pkg/dream"
	"github.com/Flokey82/animus/pkg/drives"
	"github.com/Flokey82/animus/pkg/llm"
	"github.com/Flokey82/animus/pkg/memory"
	"github.com/Flokey82/animus/pkg/tools"
	"github.com/Flokey82/animus/pkg/traits"
	"github.com/sashabaranov/go-openai"
)

// AgentConfig bundles agent initialization parameters.
type AgentConfig struct {
	Name         string
	BaseIdentity string
	Personality  traits.OCEAN
	LLMConfig    llm.Config
	Circadian    CircadianProfile
}

// Agent coordinates all cognitive subsystems into an embodied autonomous entity.
type Agent struct {
	mu           sync.RWMutex
	Name         string
	BaseIdentity string
	Personality  traits.OCEAN
	Circadian    CircadianProfile

	Drives    *drives.Manager
	Tools     *tools.Registry
	Memory    *memory.Store
	Episodic  *memory.EpisodicLogger
	LLMClient *llm.Client
	Dream     *dream.Cycle
	Hub       *channel.Hub

	// ContextualTagsProvider dynamically supplies tags (e.g. room fixtures, items) for tool filtering.
	ContextualTagsProvider func() []string
}

// NewAgent instantiates a fully wired autonomous agent.
func NewAgent(cfg AgentConfig) *Agent {
	if cfg.Name == "" {
		cfg.Name = "Animus"
	}
	if cfg.BaseIdentity == "" {
		cfg.BaseIdentity = "An autonomous, curious, and thoughtful companion."
	}
	if cfg.Personality == (traits.OCEAN{}) {
		cfg.Personality = traits.DefaultOCEAN()
	}
	if cfg.Circadian == (CircadianProfile{}) {
		cfg.Circadian = DefaultCircadianProfile()
	}

	client := llm.NewClient(cfg.LLMConfig)

	a := &Agent{
		Name:         cfg.Name,
		BaseIdentity: cfg.BaseIdentity,
		Personality:  cfg.Personality,
		Circadian:    cfg.Circadian,
		Drives:       drives.NewManager(),
		Tools:        tools.NewRegistry(),
		Memory:       memory.NewStore(),
		Episodic:     memory.NewEpisodicLogger(100),
		LLMClient:    client,
		Dream:        dream.NewCycle(client),
		Hub:          channel.NewHub(),
	}

	// Register default baseline drives
	a.Drives.RegisterDrive(drives.Drive{
		Name:             "Energy",
		Value:            100,
		Max:              100,
		CriticalLevel:    20,
		DecayRatePerTick: 1.0,
	})
	a.Drives.RegisterDrive(drives.Drive{
		Name:             "Curiosity",
		Value:            85,
		Max:              100,
		CriticalLevel:    25,
		DecayRatePerTick: 1.5,
	})
	a.Drives.RegisterDrive(drives.Drive{
		Name:             "Social",
		Value:            75,
		Max:              100,
		CriticalLevel:    20,
		DecayRatePerTick: 1.0,
	})

	return a
}

// BuildSystemPrompt compiles the agent's identity, expressed traits, drives, and active memories.
func (a *Agent) BuildSystemPrompt() string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("You are %s.\n", a.Name))
	sb.WriteString(fmt.Sprintf("%s\n\n", a.BaseIdentity))

	traitPrompt := traits.BuildTraitPrompt(a.Personality, nil)
	if traitPrompt != "" {
		sb.WriteString(traitPrompt + "\n")
	}

	driveSummary := a.Drives.Summary()
	if driveSummary != "" {
		sb.WriteString("### Current Motives & Needs:\n")
		sb.WriteString(driveSummary + "\n\n")
	}

	memSummary := a.Memory.ActiveMemoriesSummary()
	if memSummary != "" {
		sb.WriteString("### Known Memories & Identity Notes:\n")
		sb.WriteString(memSummary + "\n\n")
	}

	return strings.TrimSpace(sb.String())
}

// Chat processes an incoming user message, runs any necessary tool calls, and returns the response.
func (a *Agent) Chat(ctx context.Context, userInput string, sender string) (string, error) {
	a.Episodic.Log("chat", fmt.Sprintf("%s said: %q", sender, userInput))
	a.Drives.Modify("Social", 15.0) // interaction satisfies social drive

	systemPrompt := a.BuildSystemPrompt()

	// Determine contextual tools
	var tags []string
	if a.ContextualTagsProvider != nil {
		tags = a.ContextualTagsProvider()
	}
	availableTools := a.Tools.FilterByTags(tags...)

	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
		{Role: openai.ChatMessageRoleUser, Content: userInput},
	}

	// 1. First completion with tools
	resp, err := a.LLMClient.ChatWithTools(ctx, messages, availableTools)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from LLM")
	}

	choice := resp.Choices[0]

	// 2. If the LLM requested tool execution
	if len(choice.Message.ToolCalls) > 0 {
		messages = append(messages, choice.Message)
		for _, tc := range choice.Message.ToolCalls {
			toolResult, execErr := a.Tools.Execute(ctx, tc.Function.Name, tc.Function.Arguments)
			if execErr != nil {
				toolResult = tools.MakeErrorPayload(500, execErr.Error())
			}
			a.Episodic.Log("tool", fmt.Sprintf("Executed %s: %s", tc.Function.Name, tc.Function.Arguments))

			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    toolResult,
				ToolCallID: tc.ID,
			})
		}

		// Follow-up completion after tools executed
		followUp, err := a.LLMClient.Chat(ctx, messages)
		if err != nil {
			return "", err
		}
		a.Episodic.Log("chat", fmt.Sprintf("%s replied: %q", a.Name, followUp))
		a.Hub.Broadcast(channel.Message{
			Source:    "agent",
			Channel:   "agent_chat",
			Sender:    a.Name,
			Content:   followUp,
			Timestamp: time.Now(),
		})
		return followUp, nil
	}

	content := choice.Message.Content
	a.Episodic.Log("chat", fmt.Sprintf("%s replied: %q", a.Name, content))
	a.Hub.Broadcast(channel.Message{
		Source:    "agent",
		Channel:   "agent_chat",
		Sender:    a.Name,
		Content:   content,
		Timestamp: time.Now(),
	})
	return content, nil
}

// Tick applies homeostatic drive decay, evaluates circadian schedule, and logs ambient pulses.
func (a *Agent) Tick(ctx context.Context) error {
	a.Drives.Tick()

	state := a.Circadian.Evaluate(time.Now())
	a.Circadian.CurrentState = state

	return nil
}

// Sleep executes the Oneiros dream cycle, consolidates episodic memory, and rests the agent.
func (a *Agent) Sleep(ctx context.Context) (*dream.DreamResult, error) {
	logText := a.Episodic.FullLogText()
	memSummary := a.Memory.ActiveMemoriesSummary()

	result, err := a.Dream.Run(ctx, a.Name, a.BaseIdentity, logText, memSummary)
	if err != nil {
		return nil, err
	}

	// Consolidate into memory store
	if result.ConsolidatedText != "" {
		a.Memory.Add(result.ConsolidatedText, memory.TierLongTerm, 6.0, "daily_summary")
	}

	// Restore Energy drive
	a.Drives.Set("Energy", 100)

	// Clear episodic buffer for tomorrow
	a.Episodic.Clear()
	a.Episodic.Log("dream", fmt.Sprintf("Dreamt: %s", result.DreamNarrative))

	return result, nil
}
