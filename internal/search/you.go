package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// YouBackend — You.com web search API.
// Docs: https://you.com — free tier available.
type YouBackend struct{}

func (b *YouBackend) Name() Source    { return SourceWeb }
func (b *YouBackend) Available() bool { return apiKey("you") != "" }

func (b *YouBackend) Search(query string, limit int) ([]Result, error) {
	apiKey := apiKey("you")
	if apiKey == "" {
		return nil, fmt.Errorf("YOU_API_KEY not set; get at https://you.com/api")
	}
	if limit > 20 {
		limit = 20
	}

	params := url.Values{}
	params.Set("query", query)
	params.Set("num_web_results", fmt.Sprint(limit))

	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest("GET", "https://api.you.com/search?"+params.Encode(), nil)
	req.Header.Set("X-API-Key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("you.com: %w", err)
	}
	defer resp.Body.Close()

	var apiResp struct {
		WebResults []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Snippet string `json:"snippet"`
			Date    string `json:"date"`
		} `json:"web_results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("you.com parse: %w", err)
	}

	var results []Result
	for _, r := range apiResp.WebResults {
		snippet := r.Snippet
		if len(snippet) > 300 {
			snippet = snippet[:300] + "..."
		}
		results = append(results, Result{
			Source: SourceWeb, Title: r.Title, URL: r.URL, Snippet: snippet, Date: r.Date,
			Engagement: "via You.com",
		})
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("you.com: 0 results")
	}
	return results, nil
}
