package builtin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClockAndCelestial(t *testing.T) {
	clock := NewClockTool(nil)
	out, err := clock.Execute(context.Background(), nil)
	if err != nil {
		t.Fatalf("clock error: %v", err)
	}
	if !strings.Contains(out, `"is_daytime"`) {
		t.Errorf("clock output missing is_daytime: %s", out)
	}

	moon := NewCelestialMoonTool()
	mOut, err := moon.Execute(context.Background(), nil)
	if err != nil {
		t.Fatalf("moon error: %v", err)
	}
	if !strings.Contains(mOut, "phase_name") {
		t.Errorf("moon output missing phase_name: %s", mOut)
	}
}

func TestSpeechSynthesisMock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "audio/wav")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("RIFFmockwavdata"))
	}))
	defer server.Close()

	tts := NewSpeechSynthesisTool(server.URL, "af_bella")
	out, err := tts.Execute(context.Background(), map[string]any{"text": "Hello test"})
	if err != nil {
		t.Fatalf("tts error: %v", err)
	}
	if !strings.Contains(out, `"bytes_count":15`) {
		t.Errorf("unexpected tts output: %s", out)
	}
}
