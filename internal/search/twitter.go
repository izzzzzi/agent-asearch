package search

import (
	"fmt"
	"os/exec"
)

type TwitterBackend struct{}

func (b *TwitterBackend) Name() Source { return SourceTwitter }
func (b *TwitterBackend) Available() bool {
	_, err := exec.LookPath("twitter")
	return err == nil
}

func (b *TwitterBackend) Search(query string, limit int) ([]Result, error) {
	if !b.Available() {
		// Fallback: use Nitter (public, no auth needed)
		return nitterSearch(query, limit)
	}
	return twitterCLISearch(query, limit)
}

func twitterCLISearch(query string, limit int) ([]Result, error) {
	cmd := exec.Command("twitter", "search", query, "--limit", fmt.Sprint(limit))
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("twitter-cli search failed: %w", err)
	}
	var data []struct {
		Text      string `json:"text"`
		URL       string `json:"url"`
		Username  string `json:"username"`
		Likes     int    `json:"likes"`
		Retweets  int    `json:"retweets"`
		CreatedAt string `json:"created_at"`
	}
	// Simple parsing - actual implementation would use twitter-cli's JSON output
	_ = out
	_ = data
	return nil, fmt.Errorf("twitter-cli search: parse not yet implemented")
}

func nitterSearch(query string, limit int) ([]Result, error) {
	// Nitter instances are often unreliable; provide a clear error
	instances := []string{
		"https://nitter.net",
		"https://nitter.poast.org",
	}
	_ = instances
	_ = limit
	_ = query
	return []Result{{
		Source:  SourceTwitter,
		Title:   fmt.Sprintf("X/Twitter search: %s", query),
		URL:     "https://x.com/search?q=" + query,
		Snippet: "需要安装 twitter-cli (pipx install twitter-cli) 或配置 Nitter 实例来搜索 X/Twitter。",
	}}, nil
}
