package search

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// FirecrawlBackend — web scraping with JS rendering for dynamic pages.
// Docs: https://firecrawl.dev — free tier: 500 credits/month.
type FirecrawlBackend struct{}

func (b *FirecrawlBackend) Name() Source { return SourceWeb }
func (b *FirecrawlBackend) Available() bool { return os.Getenv("FIRECRAWL_API_KEY") != "" }

func (b *FirecrawlBackend) Search(query string, limit int) ([]Result, error) {
	apiKey := os.Getenv("FIRECRAWL_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("FIRECRAWL_API_KEY not set; get at https://firecrawl.dev (500 free/month)")
	}
	if limit > 20 { limit = 20 }

	baseURL := os.Getenv("FIRECRAWL_URL")
	if baseURL == "" { baseURL = "https://api.firecrawl.dev" }

	reqBody := map[string]any{
		"query": query,
		"limit": limit,
		"scrapeOptions": map[string]any{
			"formats": []string{"markdown"},
		},
	}
	body, _ := json.Marshal(reqBody)

	client := &http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest("POST", baseURL+"/v1/search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil { return nil, fmt.Errorf("firecrawl: %w", err) }
	defer resp.Body.Close()

	var apiResp struct {
		Data []struct {
			Title       string `json:"title"`
			URL         string `json:"url"`
			Description string `json:"description"`
			Markdown    string `json:"markdown"`
			PublishedDate string `json:"publishedDate"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("firecrawl parse: %w", err)
	}

	var results []Result
	for _, r := range apiResp.Data {
		snippet := r.Description
		if snippet == "" { snippet = r.Markdown }
		if len(snippet) > 300 { snippet = snippet[:300] + "..." }
		results = append(results, Result{
			Source: SourceWeb, Title: r.Title, URL: r.URL, Snippet: snippet, Date: r.PublishedDate,
			Engagement: "via Firecrawl (JS rendered)",
		})
	}
	if len(results) == 0 { return nil, fmt.Errorf("firecrawl: 0 results") }
	return results, nil
}
