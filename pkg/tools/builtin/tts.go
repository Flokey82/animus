package builtin

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Flokey82/animus/pkg/tools"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

// SpeechSynthesisTool generates spoken audio using Kokoro TTS on the backend.
type SpeechSynthesisTool struct {
	Endpoint string
	Voice    string
	client   *http.Client
}

// NewSpeechSynthesisTool creates a TTS tool pointing to the Lemonade audio/speech endpoint.
func NewSpeechSynthesisTool(endpoint, voice string) *SpeechSynthesisTool {
	if endpoint == "" {
		endpoint = "http://192.168.86.208:8000/api/v1/audio/speech"
	}
	if voice == "" {
		voice = "af_bella"
	}
	return &SpeechSynthesisTool{
		Endpoint: endpoint,
		Voice:    voice,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *SpeechSynthesisTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "synthesize_speech",
			Description: "Synthesize spoken audio for a line of dialogue using the Kokoro text-to-speech engine.",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"text": {
						Type:        jsonschema.String,
						Description: "The exact sentence or dialogue to speak out loud.",
					},
					"voice": {
						Type:        jsonschema.String,
						Description: "Optional voice identifier (e.g. 'af_bella', 'af_sarah', 'am_adam').",
					},
				},
				Required: []string{"text"},
			},
		},
	}
}

func (s *SpeechSynthesisTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	text, _ := args["text"].(string)
	text = strings.TrimSpace(text)
	if text == "" {
		return tools.MakeErrorPayload(400, "text is required"), nil
	}

	voice, _ := args["voice"].(string)
	if voice == "" {
		voice = s.Voice
	}

	reqPayload := map[string]string{
		"model": "kokoro-v1",
		"input": text,
		"voice": voice,
	}
	payloadBytes, _ := json.Marshal(reqPayload)

	req, err := http.NewRequestWithContext(ctx, "POST", s.Endpoint, bytes.NewReader(payloadBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return tools.MakeErrorPayload(resp.StatusCode, fmt.Sprintf("TTS endpoint returned %d: %s", resp.StatusCode, string(b))), nil
	}

	audioBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	res := map[string]any{
		"status":      "synthesized",
		"bytes_count": len(audioBytes),
		"voice":       voice,
		"audio_b64":   base64.StdEncoding.EncodeToString(audioBytes),
	}
	return tools.MakeSuccessPayload(res), nil
}

func (s *SpeechSynthesisTool) Tags() []string {
	return []string{"voice", "audio", "speech"}
}
