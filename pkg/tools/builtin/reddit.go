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

// RedditTool enables browsing community discussions on Reddit.
type RedditTool struct {
	client *http.Client
}

func NewRedditTool() *RedditTool {
	return &RedditTool{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (r *RedditTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "browse_reddit",
			Description: "Browse top or hot posts in any subreddit (e.g. 'science', 'space', 'technology', 'art').",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"subreddit": {
						Type:        jsonschema.String,
						Description: "Name of the subreddit (e.g. 'astronomy', 'science', 'books').",
					},
					"limit": {
						Type:        jsonschema.Integer,
						Description: "Number of posts to fetch (default 5).",
					},
				},
				Required: []string{"subreddit"},
			},
		},
	}
}

type RedditListingPost struct {
	Title       string `json:"title"`
	Author      string `json:"author"`
	Score       int    `json:"score"`
	NumComments int    `json:"num_comments"`
	URL         string `json:"url"`
}

func (r *RedditTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	sub, _ := args["subreddit"].(string)
	sub = strings.TrimPrefix(strings.TrimSpace(sub), "r/")
	if sub == "" {
		return tools.MakeErrorPayload(400, "subreddit is required"), nil
	}

	limit := 5
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}
	if limit > 10 {
		limit = 10
	}

	endpoint := fmt.Sprintf("https://www.reddit.com/r/%s/hot.json?limit=%d", url.PathEscape(sub), limit)
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return "", err
	}
	// Use desktop user agent to avoid rate-limiting on public JSON feeds
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := r.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return tools.MakeErrorPayload(resp.StatusCode, fmt.Sprintf("Reddit returned status %d for r/%s", resp.StatusCode, sub)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var wrapper struct {
		Data struct {
			Children []struct {
				Data struct {
					Title       string `json:"title"`
					Author      string `json:"author"`
					Score       int    `json:"score"`
					NumComments int    `json:"num_comments"`
					Permalink   string `json:"permalink"`
					Over18      bool   `json:"over_18"`
				} `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &wrapper); err != nil {
		return "", err
	}

	posts := make([]RedditListingPost, 0, len(wrapper.Data.Children))
	for _, child := range wrapper.Data.Children {
		d := child.Data
		if d.Over18 {
			continue // Filter NSFW
		}
		posts = append(posts, RedditListingPost{
			Title:       d.Title,
			Author:      d.Author,
			Score:       d.Score,
			NumComments: d.NumComments,
			URL:         fmt.Sprintf("https://reddit.com%s", d.Permalink),
		})
	}

	return tools.MakeSuccessPayload(posts), nil
}

func (r *RedditTool) Tags() []string {
	return []string{"web", "social", "community"}
}
