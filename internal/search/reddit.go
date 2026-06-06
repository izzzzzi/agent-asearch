package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type RedditBackend struct{}

func (b *RedditBackend) Name() Source   { return SourceReddit }
func (b *RedditBackend) Available() bool { return true }

func (b *RedditBackend) Search(query string, limit int) ([]Result, error) {
	// Use Reddit's public JSON API (no auth needed for reading)
	u := fmt.Sprintf("https://www.reddit.com/search.json?q=%s&limit=%d&sort=relevance", query, limit)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "asearch/1.0 (agent-search-cli)")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("reddit search failed: %w", err)
	}
	defer resp.Body.Close()

	var data struct {
		Data struct {
			Children []struct {
				Data struct {
					Title     string  `json:"title"`
					Permalink string  `json:"permalink"`
					Selftext  string  `json:"selftext"`
					Score     float64 `json:"score"`
					NumComments int   `json:"num_comments"`
					Subreddit string  `json:"subreddit"`
					Created   float64 `json:"created_utc"`
					URL       string  `json:"url"`
				} `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("reddit parse error: %w", err)
	}

	var results []Result
	for _, child := range data.Data.Children {
		d := child.Data
		snippet := d.Selftext
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		if snippet == "" {
			snippet = fmt.Sprintf("r/%s post", d.Subreddit)
		}
		date := time.Unix(int64(d.Created), 0).Format("2006-01-02")
		engagement := fmt.Sprintf("↑%.0f | 💬%d | r/%s", d.Score, d.NumComments, d.Subreddit)
		results = append(results, Result{
			Source:     SourceReddit,
			Title:      d.Title,
			URL:        "https://www.reddit.com" + d.Permalink,
			Snippet:    snippet,
			Date:       date,
			Score:      d.Score,
			Engagement: engagement,
		})
	}
	return results, nil
}
