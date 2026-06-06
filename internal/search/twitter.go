package search

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type TwitterBackend struct{}

func (b *TwitterBackend) Name() Source     { return SourceTwitter }
func (b *TwitterBackend) Available() bool   { return true }

func (b *TwitterBackend) Search(query string, limit int) ([]Result, error) {
	// 1. Try twitter-cli first (full-featured)
	if _, err := exec.LookPath("twitter"); err == nil {
		results, err := twitterCLI(query, limit)
		if err == nil && len(results) > 0 {
			return results, nil
		}
	}

	// 2. Cookie file exists — guide user to use twitter-cli
	if _, err := os.Stat(os.ExpandEnv("$HOME/.asearch/twitter-cookies.txt")); err == nil {
		return nil, fmt.Errorf(
			"twitter cookies found. Install twitter-cli to use them:\n"+
				"  pipx install twitter-cli && twitter login",
		)
	}

	return nil, fmt.Errorf(
		"twitter search needs auth:\n"+
			"  pipx install twitter-cli && twitter login",
	)
}

func twitterCLI(query string, limit int) ([]Result, error) {
	cmd := exec.Command("twitter", "search", query, "-n", fmt.Sprint(limit), "--json")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("twitter-cli: %w", err)
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
