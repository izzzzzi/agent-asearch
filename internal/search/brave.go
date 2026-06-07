package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// BraveBackend implements Brave Search API (free tier: 2000 queries/month).
// Docs: https://api.search.brave.com/app/documentation/web-search
type BraveBackend struct{}

func (b *BraveBackend) Name() Source { return SourceWeb }
func (b *BraveBackend) Available() bool {
	return apiKey("brave") != ""
}

func (b *BraveBackend) Search(query string, limit int) ([]Result, error) {
	apiKey := apiKey("brave")
	if apiKey == "" {
		return nil, fmt.Errorf("BRAVE_API_KEY not set; get one at https://brave.com/search/api/")
	}
	if limit > 20 {
		limit = 20
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET",
		fmt.Sprintf("https://api.search.brave.com/res/v1/web/search?q=%s&count=%d",
			urlEncode(query), limit), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("X-Subscription-Token", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("brave request failed: %w", err)
	}
	defer resp.Body.Close()

	var apiResp struct {
		Web struct {
			Results []struct {
				Title       string `json:"title"`
				URL         string `json:"url"`
				Description string `json:"description"`
				Age         string `json:"age"`
			} `json:"results"`
		} `json:"web"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("brave parse error: %w", err)
	}

	var results []Result
	for _, r := range apiResp.Web.Results {
		snippet := r.Description
		if len(snippet) > 300 {
			snippet = snippet[:300] + "..."
		}
		results = append(results, Result{
			Source:  SourceWeb,
			Title:   r.Title,
			URL:     r.URL,
			Snippet: snippet,
			Date:    r.Age,
		})
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("brave returned 0 results")
	}
	return results, nil
}

func urlEncode(s string) string {
	// Basic URL encoding — Go's url.QueryEscape is in net/url
	return urlEncodeSimple(s)
}

func urlEncodeSimple(s string) string {
	// Use a simple replace for the common case
	result := ""
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' || c == '~' {
			result += string(c)
		} else if c == ' ' {
			result += "+"
		} else {
			result += fmt.Sprintf("%%%02X", c)
		}
	}
	return result
}
