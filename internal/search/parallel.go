package search

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ParallelBackend — Parallel.ai search API.
// Docs: https://parallel.ai — free tier available.
type ParallelBackend struct{}

func (b *ParallelBackend) Name() Source { return SourceWeb }
func (b *ParallelBackend) Available() bool { return apiKey("parallel") != "" }

func (b *ParallelBackend) Search(query string, limit int) ([]Result, error) {
	apiKey := apiKey("parallel")
	if apiKey == "" {
		return nil, fmt.Errorf("PARALLEL_API_KEY not set; get at https://parallel.ai")
	}
	if limit > 20 { limit = 20 }

	reqBody := map[string]any{"query": query, "num_results": limit}
	body, _ := json.Marshal(reqBody)

	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest("POST", "https://api.parallel.ai/search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil { return nil, fmt.Errorf("parallel: %w", err) }
	defer resp.Body.Close()

	var apiResp struct {
		Results []struct {
			Title   string  `json:"title"`
			URL     string  `json:"url"`
			Content string  `json:"content"`
			Score   float64 `json:"score"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("parallel parse: %w", err)
	}

	var results []Result
	for _, r := range apiResp.Results {
		snippet := r.Content
		if len(snippet) > 300 { snippet = snippet[:300] + "..." }
		results = append(results, Result{
			Source: SourceWeb, Title: r.Title, URL: r.URL, Snippet: snippet, Score: r.Score,
			Engagement: fmt.Sprintf("via Parallel | score: %.2f", r.Score),
		})
	}
	if len(results) == 0 { return nil, fmt.Errorf("parallel: 0 results") }
	return results, nil
}
