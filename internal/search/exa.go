package search

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// ExaBackend implements Exa API — neural/semantic search for AI agents.
// Docs: https://docs.exa.ai/reference/search
type ExaBackend struct{}

func (b *ExaBackend) Name() Source { return SourceWeb }
func (b *ExaBackend) Available() bool {
	return os.Getenv("EXA_API_KEY") != ""
}

func (b *ExaBackend) Search(query string, limit int) ([]Result, error) {
	apiKey := os.Getenv("EXA_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("EXA_API_KEY not set; get one at https://exa.ai")
	}
	if limit > 25 {
		limit = 25
	}

	reqBody := map[string]any{
		"query":           query,
		"numResults":      limit,
		"useAutoprompt":   true,
		"type":            "auto",  // auto-detects neural vs keyword
		"contents": map[string]any{
			"text": map[string]any{
				"maxCharacters": 300,
			},
		},
	}
	body, _ := json.Marshal(reqBody)

	client := &http.Client{Timeout: 20 * time.Second}
	req, err := http.NewRequest("POST", "https://api.exa.ai/search", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("exa request failed: %w", err)
	}
	defer resp.Body.Close()

	var apiResp struct {
		Results []struct {
			Title      string `json:"title"`
			URL        string `json:"url"`
			Text       string `json:"text"`
			Score      float64 `json:"score"`
			PublishedDate string `json:"publishedDate"`
			Author     string `json:"author"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("exa parse error: %w", err)
	}

	var results []Result
	for _, r := range apiResp.Results {
		snippet := r.Text
		if len(snippet) > 300 {
			snippet = snippet[:300] + "..."
		}
		date := r.PublishedDate
		if len(date) > 10 {
			date = date[:10]
		}
		engagement := fmt.Sprintf("score: %.2f", r.Score)
		if r.Author != "" {
			engagement += " | " + r.Author
		}
		results = append(results, Result{
			Source:     SourceWeb,
			Title:      r.Title,
			URL:        r.URL,
			Snippet:    snippet,
			Date:       date,
			Score:      r.Score,
			Engagement: engagement,
		})
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("exa returned 0 results")
	}
	return results, nil
}
