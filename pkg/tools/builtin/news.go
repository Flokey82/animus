package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Flokey82/animus/pkg/tools"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

// HackerNewsTool fetches top discussions from Hacker News.
type HackerNewsTool struct {
	client *http.Client
}

func NewHackerNewsTool() *HackerNewsTool {
	return &HackerNewsTool{
		client: &http.Client{Timeout: 8 * time.Second},
	}
}

func (h *HackerNewsTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "get_trending_tech_news",
			Description: "Fetch the top trending technology and science headlines from Hacker News.",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"limit": {
						Type:        jsonschema.Integer,
						Description: "Number of stories to return (default 5, max 10).",
					},
				},
			},
		},
	}
}

type HNItem struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
	Score int    `json:"score"`
	By    string `json:"by"`
}

func (h *HackerNewsTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	limit := 5
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}
	if limit > 10 {
		limit = 10
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://hacker-news.firebaseio.com/v0/topstories.json", nil)
	if err != nil {
		return "", err
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var storyIDs []int
	if err := json.Unmarshal(body, &storyIDs); err != nil {
		return "", err
	}

	var stories []HNItem
	for i := 0; i < len(storyIDs) && len(stories) < limit; i++ {
		storyID := storyIDs[i]
		itemReq, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", storyID), nil)
		itemResp, err := h.client.Do(itemReq)
		if err != nil {
			continue
		}

		itemBody, _ := io.ReadAll(itemResp.Body)
		itemResp.Body.Close()

		var story HNItem
		if err := json.Unmarshal(itemBody, &story); err == nil && story.Title != "" {
			stories = append(stories, story)
		}
	}

	return tools.MakeSuccessPayload(stories), nil
}

func (h *HackerNewsTool) Tags() []string {
	return []string{"web", "tech", "news"}
}
