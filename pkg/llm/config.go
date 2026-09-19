package llm

// Config contains settings for connecting to LLM providers.
type Config struct {
	BaseURL         string `json:"base_url"`         // e.g. "http://192.168.86.208:8000/api/v1" or "http://localhost:11434/v1"
	APIKey          string `json:"api_key"`          // API key (optional for local endpoints)
	Model           string `json:"model"`            // Primary large model (e.g. Gemma-4, Qwen, Granite-4)
	ToolModel       string `json:"tool_model"`       // Secondary fast reflex model for tool decisions
	DisableThinking bool   `json:"disable_thinking"` // If true, strips reasoning blocks like <think>...</think>
	MockMode        bool   `json:"mock_mode"`        // If true, returns mock responses without making network calls
}

// DefaultConfig returns baseline configuration pointing to the local backend.
func DefaultConfig() Config {
	return Config{
		BaseURL:         "http://192.168.86.208:8000/api/v1",
		APIKey:          "dummy-key",
		Model:           "Gemma-4-26B-A4B-it-MTP-GGUF",
		ToolModel:       "granite-4.0-h-tiny-GGUF",
		DisableThinking: true,
		MockMode:        false,
	}
}
