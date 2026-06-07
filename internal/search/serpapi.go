package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// SerpAPIBackend — 40+ search engines via SerpAPI.
// Docs: https://serpapi.com — 100 free searches/month.
type SerpAPIBackend struct{}

func (b *SerpAPIBackend) Name() Source    { return SourceWeb }
func (b *SerpAPIBackend) Available() bool { return apiKey("serpapi") != "" }

func (b *SerpAPIBackend) Search(query string, limit int) ([]Result, error) {
	apiKey := apiKey("serpapi")
	if apiKey == "" {
		return nil, fmt.Errorf("SERPAPI_API_KEY not set; get at https://serpapi.com (100 free/month)")
	}
	if limit > 100 {
		limit = 100
	}

	params := url.Values{}
	params.Set("api_key", apiKey)
	params.Set("q", query)
	params.Set("num", fmt.Sprint(limit))
	params.Set("engine", "google") // can also be: bing, yahoo, duckduckgo, youtube, etc.

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get("https://serpapi.com/search?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("serpapi: %w", err)
	}
	defer resp.Body.Close()

	var apiResp struct {
		OrganicResults []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
			Date    string `json:"date"`
		} `json:"organic_results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("serpapi parse: %w", err)
	}

	var results []Result
	for _, r := range apiResp.OrganicResults {
		snippet := r.Snippet
		if len(snippet) > 300 {
			snippet = snippet[:300] + "..."
		}
		results = append(results, Result{
			Source: SourceWeb, Title: r.Title, URL: r.Link, Snippet: snippet, Date: r.Date,
			Engagement: "via Google (SerpAPI)",
		})
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("serpapi: 0 results")
	}
	return results, nil
}
