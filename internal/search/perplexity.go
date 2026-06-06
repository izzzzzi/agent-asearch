package search

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// PerplexityBackend — AI-powered answers via Perplexity Sonar (OpenRouter).
// Docs: https://docs.perplexity.ai — free tier available.
type PerplexityBackend struct{}

func (b *PerplexityBackend) Name() Source { return SourceWeb }
func (b *PerplexityBackend) Available() bool   { return os.Getenv("PERPLEXITY_API_KEY") != "" }

func (b *PerplexityBackend) Search(query string, limit int) ([]Result, error) {
	apiKey := os.Getenv("PERPLEXITY_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("PERPLEXITY_API_KEY not set; get at https://docs.perplexity.ai")
	}

	reqBody := map[string]any{
		"model": "sonar-pro",
		"messages": []map[string]string{
			{"role": "system", "content": "Search the web and provide accurate, cited answers."},
			{"role": "user", "content": query},
		},
		"max_tokens": 1024,
		"search_domain_filter": []string{},
		"return_images": false,
		"return_related_questions": false,
		"search_recency_filter": "month",
		"top_p": 0.9,
		"presence_penalty": 0,
		"frequency_penalty": 1,
	}
	body, _ := json.Marshal(reqBody)

	client := &http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest("POST", "https://api.perplexity.ai/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil { return nil, fmt.Errorf("perplexity: %w", err) }
	defer resp.Body.Close()

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Citations []string `json:"citations"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("perplexity parse: %w", err)
	}

	var results []Result
	if len(apiResp.Choices) > 0 && apiResp.Choices[0].Message.Content != "" {
		results = append(results, Result{
			Source: SourceWeb, Title: fmt.Sprintf("AI answer: %s", query),
			URL: "", Snippet: apiResp.Choices[0].Message.Content, Score: 1.0,
			Engagement: fmt.Sprintf("citations: %d", len(apiResp.Citations)),
		})
	}
	for _, url := range apiResp.Citations {
		results = append(results, Result{
			Source: SourceWeb, Title: url, URL: url, Snippet: "📎 citation", Engagement: "perplexity",
		})
	}
	if len(results) == 0 { return nil, fmt.Errorf("perplexity: 0 results") }
	return results, nil
}
