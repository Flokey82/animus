package memory

import (
	"strings"
	"testing"
)

func TestEpisodicLogger(t *testing.T) {
	log := NewEpisodicLogger(3)

	log.Log("action", "Woke up")
	log.Log("chat", "Said good morning")
	log.Log("ambient", "Brewed tea")
	log.Log("action", "Went for a walk") // Evicts "Woke up"

	text := log.FullLogText()
	if strings.Contains(text, "Woke up") {
		t.Errorf("expected 'Woke up' to be evicted, got:\n%s", text)
	}
	if !strings.Contains(text, "Went for a walk") {
		t.Errorf("expected 'Went for a walk' to be present")
	}

	recent := log.Recent(2)
	if len(recent) != 2 || recent[1].Content != "Went for a walk" {
		t.Fatalf("unexpected recent events: %+v", recent)
	}
}

func TestStore(t *testing.T) {
	store := NewStore()

	store.Add("I love exploring space and star systems.", TierCore, 10.0, "space", "stars")
	store.Add("Visited the observatory last spring.", TierLongTerm, 7.0, "observatory")
	store.Add("The coffee tasted like vanilla today.", TierShortTerm, 2.0, "coffee")

	recalled := store.Recall("space")
	if len(recalled) != 1 || !strings.Contains(recalled[0].Content, "exploring space") {
		t.Fatalf("recall failed: %+v", recalled)
	}

	summary := store.ActiveMemoriesSummary()
	if !strings.Contains(summary, "Core Beliefs") || !strings.Contains(summary, "vanilla") {
		t.Errorf("summary missing expected contents: %s", summary)
	}
}
