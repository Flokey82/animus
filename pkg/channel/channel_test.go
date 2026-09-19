package channel

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockAgent struct {
	reply string
}

func (m *mockAgent) Chat(ctx context.Context, input, sender string) (string, error) {
	return m.reply, nil
}

func (m *mockAgent) BuildSystemPrompt() string {
	return "Mock agent prompt."
}

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

func TestWebServerEndpoints(t *testing.T) {
	agent := &mockAgent{reply: "Hello from mock agent!"}
	hub := NewHub()
	ws := NewWebServer(8081, agent, hub)

	// Test GET /
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	ws.handleIndex(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for index, got %d", w.Code)
	}

	// Test GET /api/status
	sReq := httptest.NewRequest("GET", "/api/status", nil)
	sW := httptest.NewRecorder()
	ws.handleStatus(sW, sReq)
	if sW.Code != http.StatusOK {
		t.Errorf("expected 200 for status, got %d", sW.Code)
	}

	// Test POST /api/chat
	chatBody, _ := json.Marshal(map[string]string{
		"message": "Hi!",
		"sender":  "Alice",
	})
	cReq := httptest.NewRequest("POST", "/api/chat", bytes.NewReader(chatBody))
	cW := httptest.NewRecorder()
	ws.handleChat(cW, cReq)
	if cW.Code != http.StatusOK {
		t.Errorf("expected 200 for chat, got %d", cW.Code)
	}

	var res map[string]string
	_ = json.Unmarshal(cW.Body.Bytes(), &res)
	if res["reply"] != "Hello from mock agent!" {
		t.Errorf("expected 'Hello from mock agent!', got %q", res["reply"])
	}
}
