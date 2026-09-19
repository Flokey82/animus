package dream

import (
	"context"
	"strings"
	"testing"

	"github.com/Flokey82/animus/pkg/llm"
)

func TestDreamCycleMock(t *testing.T) {
	client := llm.NewClient(llm.Config{MockMode: true})
	cycle := NewCycle(client)

	res, err := cycle.Run(context.Background(), "Lyra", "An introspective explorer.", "Explored ancient ruins and read books.", "Loves astronomy.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.DreamNarrative == "" {
		t.Errorf("expected dream narrative, got empty")
	}
	if !strings.Contains(res.ConsolidatedText, "Mock") && res.ConsolidatedText == "" {
		t.Errorf("expected consolidated text")
	}
}
