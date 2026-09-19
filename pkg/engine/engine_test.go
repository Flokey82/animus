package engine

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Flokey82/animus/pkg/llm"
	"github.com/Flokey82/animus/pkg/tools/builtin"
)

func TestAgentChatAndTools(t *testing.T) {
	agent := NewAgent(AgentConfig{
		Name:         "Astra",
		BaseIdentity: "A philosopher and stargazer.",
		LLMConfig: llm.Config{
			MockMode: true,
		},
	})

	// Register builtin tool
	clockTool := builtin.NewClockTool(nil)
	agent.Tools.Register(clockTool)

	ctx := context.Background()

	// Initial system prompt check
	prompt := agent.BuildSystemPrompt()
	if !strings.Contains(prompt, "You are Astra.") {
		t.Errorf("expected 'You are Astra.' in prompt: %s", prompt)
	}

	// Chat test
	reply, err := agent.Chat(ctx, "What time is it?", "User")
	if err != nil {
		t.Fatalf("unexpected chat error: %v", err)
	}
	if reply == "" {
		t.Errorf("expected non-empty reply")
	}

	// Verify episodic logging
	logs := agent.Episodic.FullLogText()
	if !strings.Contains(logs, "What time is it?") {
		t.Errorf("expected user input logged in episodic memory: %s", logs)
	}

	// Tick test
	energyBefore, _ := agent.Drives.Get("Energy")
	if err := agent.Tick(ctx); err != nil {
		t.Fatalf("tick error: %v", err)
	}
	energyAfter, _ := agent.Drives.Get("Energy")
	if energyAfter.Value >= energyBefore.Value {
		t.Errorf("expected energy decay on tick, before: %f, after: %f", energyBefore.Value, energyAfter.Value)
	}

	// Sleep / dream cycle test
	dreamRes, err := agent.Sleep(ctx)
	if err != nil {
		t.Fatalf("sleep error: %v", err)
	}
	if dreamRes.DreamNarrative == "" {
		t.Errorf("expected non-empty dream narrative")
	}

	// Energy should be restored
	energyRestored, _ := agent.Drives.Get("Energy")
	if energyRestored.Value != 100 {
		t.Errorf("expected energy restored to 100, got: %f", energyRestored.Value)
	}
}

func TestCircadianSchedule(t *testing.T) {
	circ := DefaultCircadianProfile()
	// Default: Sleep 23:00 to 07:00

	dayTime := time.Date(2026, 9, 19, 14, 30, 0, 0, time.UTC)
	if state := circ.Evaluate(dayTime); state != StateAwake {
		t.Errorf("expected awake at 14:30, got %s", state)
	}

	nightTime := time.Date(2026, 9, 19, 23, 45, 0, 0, time.UTC)
	if state := circ.Evaluate(nightTime); state != StateAsleep {
		t.Errorf("expected asleep at 23:45, got %s", state)
	}

	morningPreWake := time.Date(2026, 9, 19, 6, 15, 0, 0, time.UTC)
	if state := circ.Evaluate(morningPreWake); state != StateAsleep {
		t.Errorf("expected asleep at 06:15, got %s", state)
	}
}
