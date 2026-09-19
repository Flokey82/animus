package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Flokey82/animus/pkg/tools"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

// WikipediaTool retrieves encyclopedic summaries of given topics.
type WikipediaTool struct {
	client *http.Client
}

func NewWikipediaTool() *WikipediaTool {
	return &WikipediaTool{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *WikipediaTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "search_wikipedia",
			Description: "Search Wikipedia for encyclopedic knowledge, summaries, definitions, and concepts.",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"query": {
						Type:        jsonschema.String,
						Description: "The topic or article title to look up (e.g. 'Quantum entanglement', 'Renaissance').",
					},
				},
				Required: []string{"query"},
			},
		},
	}
}

func (w *WikipediaTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	query, _ := args["query"].(string)
	clean := strings.TrimSpace(query)
	if clean == "" {
		return tools.MakeErrorPayload(400, "query parameter is required"), nil
	}

	apiURL := fmt.Sprintf("https://en.wikipedia.org/api/rest_v1/page/summary/%s", url.PathEscape(clean))
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "AnimusAgent/1.0 (https://github.com/Flokey82/animus)")

	resp, err := w.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return tools.MakeErrorPayload(resp.StatusCode, fmt.Sprintf("Wikipedia returned status %d for query %q", resp.StatusCode, clean)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var data struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Extract     string `json:"extract"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}

	res := map[string]any{
		"title":       data.Title,
		"description": data.Description,
		"summary":     data.Extract,
	}
	b, _ := json.Marshal(res)
	return tools.MakeSuccessPayload(string(b)), nil
}

func (w *WikipediaTool) Tags() []string {
	return []string{"web", "knowledge", "research"}
}

// HistoryTodayTool fetches historical events on today's calendar date.
type HistoryTodayTool struct {
	client *http.Client
}

func NewHistoryTodayTool() *HistoryTodayTool {
	return &HistoryTodayTool{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (h *HistoryTodayTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "get_history_today",
			Description: "Retrieve notable historical events that took place on today's calendar date.",
			Parameters:  jsonschema.Definition{Type: jsonschema.Object},
		},
	}
}

func (h *HistoryTodayTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	now := time.Now()
	apiURL := fmt.Sprintf("https://en.wikipedia.org/api/rest_v1/feed/onthisday/selected/%02d/%02d", now.Month(), now.Day())

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "AnimusAgent/1.0 (https://github.com/Flokey82/animus)")

	resp, err := h.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return tools.MakeErrorPayload(resp.StatusCode, fmt.Sprintf("Wikipedia OnThisDay returned status %d", resp.StatusCode)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var data struct {
		Selected []struct {
			Text string `json:"text"`
			Year int    `json:"year"`
		} `json:"selected"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}

	limit := len(data.Selected)
	if limit > 4 {
		limit = 4
	}
	type item struct {
		Year int    `json:"year"`
		Text string `json:"text"`
	}
	items := make([]item, 0, limit)
	for i := 0; i < limit; i++ {
		items = append(items, item{
			Year: data.Selected[i].Year,
			Text: data.Selected[i].Text,
		})
	}

	b, _ := json.Marshal(items)
	return tools.MakeSuccessPayload(string(b)), nil
}

func (h *HistoryTodayTool) Tags() []string {
	return []string{"web", "knowledge", "history"}
}
