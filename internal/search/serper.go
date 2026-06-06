package search

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// SerperBackend — Google SERP via Serper.dev API.
// Docs: https://serper.dev — 2500 free queries/month.
type SerperBackend struct{}

func (b *SerperBackend) Name() Source { return SourceWeb }
func (b *SerperBackend) Available() bool   { return os.Getenv("SERPER_API_KEY") != "" }

func (b *SerperBackend) Search(query string, limit int) ([]Result, error) {
	apiKey := os.Getenv("SERPER_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("SERPER_API_KEY not set; get at https://serper.dev (2500 free/month)")
	}
	if limit > 100 { limit = 100 }

	reqBody := map[string]any{"q": query, "num": limit}
	body, _ := json.Marshal(reqBody)

	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest("POST", "https://google.serper.dev/search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", apiKey)

	resp, err := client.Do(req)
	if err != nil { return nil, fmt.Errorf("serper: %w", err) }
	defer resp.Body.Close()

	var apiResp struct {
		Organic []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
			Date    string `json:"date"`
		} `json:"organic"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("serper parse: %w", err)
	}

	var results []Result
	for _, r := range apiResp.Organic {
		snippet := r.Snippet
		if len(snippet) > 300 { snippet = snippet[:300] + "..." }
		results = append(results, Result{
			Source: SourceWeb, Title: r.Title, URL: r.Link, Snippet: snippet, Date: r.Date,
			Engagement: "via Google (Serper)",
		})
	}
	if len(results) == 0 { return nil, fmt.Errorf("serper: 0 results") }
	return results, nil
}
