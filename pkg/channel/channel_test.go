package channel

import (
	"testing"
	"time"
)

func TestHubBroadcast(t *testing.T) {
	hub := NewHub()
	sub := hub.Subscribe()

	msg := Message{
		Source:    "agent",
		Channel:   "cli",
		Sender:    "Lyra",
		Content:   "Hello universe!",
		Timestamp: time.Now(),
	}

	hub.Broadcast(msg)

	select {
	case received := <-sub:
		if received.Content != "Hello universe!" {
			t.Errorf("expected 'Hello universe!', got %s", received.Content)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for broadcast message")
	}
}
