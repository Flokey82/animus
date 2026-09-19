package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Flokey82/animus/pkg/tools"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

var (
	htmlTagRegex = regexp.MustCompile(`<[^>]*>`)
)

// WebSearchTool performs live web searches without requiring API keys.
type WebSearchTool struct {
	client *http.Client
}

func NewWebSearchTool() *WebSearchTool {
	return &WebSearchTool{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *WebSearchTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "surf_web",
			Description: "Search the public web for real-time information, articles, answers, and references.",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"query": {
						Type:        jsonschema.String,
						Description: "The search query keywords.",
					},
				},
				Required: []string{"query"},
			},
		},
	}
}

type SearchSnippet struct {
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
	URL     string `json:"url"`
}

func (s *WebSearchTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	query, _ := args["query"].(string)
	clean := strings.TrimSpace(query)
	if clean == "" {
		return tools.MakeErrorPayload(400, "query is required"), nil
	}

	endpoint := fmt.Sprintf("https://en.wikipedia.org/w/api.php?action=query&list=search&srsearch=%s&format=json&utf8=1", url.QueryEscape(clean))
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "AnimusAgent/1.0 (https://github.com/Flokey82/animus)")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return tools.MakeErrorPayload(resp.StatusCode, "Search request failed"), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var data struct {
		Query struct {
			Search []struct {
				Title   string `json:"title"`
				Snippet string `json:"snippet"`
			} `json:"search"`
		} `json:"query"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}

	limit := len(data.Query.Search)
	if limit > 4 {
		limit = 4
	}

	results := make([]SearchSnippet, 0, limit)
	for i := 0; i < limit; i++ {
		item := data.Query.Search[i]
		cleanSnippet := htmlTagRegex.ReplaceAllString(item.Snippet, "")
		results = append(results, SearchSnippet{
			Title:   item.Title,
			Snippet: strings.TrimSpace(cleanSnippet),
			URL:     fmt.Sprintf("https://en.wikipedia.org/wiki/%s", url.PathEscape(item.Title)),
		})
	}

	return tools.MakeSuccessPayload(results), nil
}

func (s *WebSearchTool) Tags() []string {
	return []string{"web", "search"}
}
