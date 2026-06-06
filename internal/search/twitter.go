package search

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type TwitterBackend struct{}

func (b *TwitterBackend) Name() Source   { return SourceTwitter }
func (b *TwitterBackend) Available() bool { return true }

func (b *TwitterBackend) Search(query string, limit int) ([]Result, error) {
	if _, err := exec.LookPath("twitter"); err != nil {
		return nil, fmt.Errorf(
			"twitter search needs twitter-cli.\n"+
				"Install: pipx install twitter-cli\n"+
				"Then run: twitter login  (opens browser once)",
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "twitter", "search", query, "-n", fmt.Sprint(limit), "--json")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("twitter-cli failed (run in terminal: twitter login): %w", err)
	}

	var tweets []struct {
		Text      string `json:"text"`
		URL       string `json:"url"`
		Username  string `json:"username"`
		Likes     int    `json:"likes"`
		Retweets  int    `json:"retweets"`
		CreatedAt string `json:"created_at"`
	}
	if err := json.Unmarshal(out, &tweets); err != nil {
		return nil, fmt.Errorf("twitter-cli parse error: %w", err)
	}

	var results []Result
	for _, t := range tweets {
		snippet := t.Text
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		engagement := fmt.Sprintf("♥%d | ↻%d | @%s", t.Likes, t.Retweets, t.Username)
		results = append(results, Result{
			Source:     SourceTwitter,
			Title:      snippet,
			URL:        t.URL,
			Snippet:    snippet,
			Date:       t.CreatedAt,
			Score:      float64(t.Likes),
			Engagement: engagement,
		})
	}
	return results, nil
}
