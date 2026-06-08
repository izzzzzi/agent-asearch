package search

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type TavilyBackend struct{}

func (b *TavilyBackend) Name() Source { return SourceTavily }
func (b *TavilyBackend) Available() bool {
	return apiKey("tavily") != ""
}

func (b *TavilyBackend) Search(query string, limit int) ([]Result, error) {
	apiKey := apiKey("tavily")
	if apiKey == "" {
		return nil, fmt.Errorf("TAVILY_API_KEY not set; get one at https://tavily.com")
	}

	if limit > 20 {
		limit = 20 // Tavily max
	}

	reqBody := map[string]any{
		"api_key":             apiKey,
		"query":               query,
		"search_depth":        "basic",
		"include_answer":      true,
		"include_raw_content": false,
		"max_results":         limit,
	}
	body, _ := json.Marshal(reqBody)

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("POST", "https://api.tavily.com/search", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tavily request failed: %w", err)
	}
	defer resp.Body.Close()

	var apiResp struct {
		Query        string  `json:"query"`
		Answer       string  `json:"answer"`
		ResponseTime float64 `json:"response_time"`
		Results      []struct {
			Title      string  `json:"title"`
			URL        string  `json:"url"`
			Content    string  `json:"content"`
			Score      float64 `json:"score"`
			RawContent string  `json:"raw_content"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("tavily parse error: %w", err)
	}

	var results []Result

	// Tavily's AI-generated answer (if available)
	if apiResp.Answer != "" {
		results = append(results, Result{
			Source:     SourceTavily,
			Title:      fmt.Sprintf("AI answer: %s", query),
			URL:        "",
			Snippet:    apiResp.Answer,
			Score:      1.0,
			Engagement: fmt.Sprintf("response_time: %.2fs", apiResp.ResponseTime),
			RawMeta: map[string]any{
				"answer":        apiResp.Answer,
				"response_time": apiResp.ResponseTime,
				"query":         apiResp.Query,
				"type":          "ai_answer",
			},
		})
	}

	// Structured results
	for _, r := range apiResp.Results {
		snippet := r.Content
		if len(snippet) > 300 {
			snippet = snippet[:300] + "..."
		}
		results = append(results, Result{
			Source:     SourceTavily,
			Title:      r.Title,
			URL:        r.URL,
			Snippet:    snippet,
			Score:      r.Score,
			Engagement: fmt.Sprintf("relevance: %.2f", r.Score),
			RawMeta: map[string]any{
				"title": r.Title,
				"url":   r.URL,
				"score": r.Score,
				"type":  "result",
			},
		})
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("tavily returned 0 results")
	}
	return results, nil
}
