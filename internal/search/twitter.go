package search

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type TwitterBackend struct{}

func (b *TwitterBackend) Name() Source     { return SourceTwitter }
func (b *TwitterBackend) Available() bool   { return true }

func (b *TwitterBackend) Search(query string, limit int) ([]Result, error) {
	// Try twitter-cli with cookies passed via env vars
	if _, err := exec.LookPath("twitter"); err == nil {
		results, err := twitterCLI(query, limit)
		if err == nil && len(results) > 0 {
			return results, nil
		}
		// If we got here, twitter-cli is installed but failed
		// Check if cookies exist to give better error
		return nil, fmt.Errorf(
			"twitter-cli failed (API may be temporarily unavailable).\n"+
			"Run in terminal: TWITTER_AUTH_TOKEN=... TWITTER_CT0=... twitter search \"%s\"\n"+
			"Cookies are stored at ~/.asearch/twitter-cookies.txt for reference.", query)
	}

	// Check if cookie file exists (user already exported cookies)
	return nil, fmt.Errorf(
		"twitter search unavailable.\n"+
		"Install: pipx install twitter-cli\n"+
		"Then set cookies:\n"+
		"  export TWITTER_AUTH_TOKEN=\"<your_auth_token>\"\n"+
		"  export TWITTER_CT0=\"<your_ct0>\"\n"+
		"Note: Twitter's search API is currently unstable (returns 404).\n"+
		"This is a Twitter/X-side issue, not an asearch bug.",
	)
}

func twitterCLI(query string, limit int) ([]Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "twitter", "search", query, "-n", fmt.Sprint(limit), "--json")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("twitter-cli failed: %w", err)
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
		return nil, fmt.Errorf("twitter parse: %w", err)
	}

	var results []Result
	for _, t := range tweets {
		snippet := t.Text
		if len(snippet) > 200 { snippet = snippet[:200] + "..." }
		results = append(results, Result{
			Source: SourceTwitter, Title: snippet, URL: t.URL, Snippet: snippet,
			Date: t.CreatedAt, Score: float64(t.Likes),
			Engagement: fmt.Sprintf("♥%d | ↻%d | @%s", t.Likes, t.Retweets, t.Username),
		})
	}
	return results, nil
}
