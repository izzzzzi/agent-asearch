package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// JinaBackend implements Jina Reader API — converts any URL to clean markdown.
// Docs: https://jina.ai/reader
// Free tier: 1M tokens/month, no API key needed (but key lifts rate limits).
type JinaBackend struct{}

func (b *JinaBackend) Name() Source { return SourceWeb }
func (b *JinaBackend) Available() bool { return true }

func (b *JinaBackend) Search(query string, limit int) ([]Result, error) {
	// Jina Reader is for URL reading, not search.
	// For search, Jina offers a separate search endpoint.
	// We use r.jina.ai to fetch content from URLs discovered via search.
	return nil, fmt.Errorf(
		"jina reader: use for URL extraction, not search. "+
		"Usage: curl https://r.jina.ai/https://example.com -H 'Authorization: Bearer $JINA_API_KEY'")
}

// ReadURL fetches a URL as clean markdown via Jina Reader.
func ReadURL(url string) (string, error) {
	apiKey := os.Getenv("JINA_API_KEY")
	jinaURL := "https://r.jina.ai/" + url

	client := &http.Client{Timeout: 20 * time.Second}
	req, err := http.NewRequest("GET", jinaURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "text/markdown")
	req.Header.Set("X-Return-Format", "markdown")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("jina fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("jina returned status %d", resp.StatusCode)
	}

	type jinaResp struct {
		Title string `json:"title"`
		URL   string `json:"url"`
		Content string `json:"content"`
	}
	// Jina Reader returns markdown directly in body, or JSON if Accept: application/json
	var jr jinaResp
	_ = json.NewDecoder(resp.Body).Decode(&jr) // try JSON first
	if jr.Content != "" {
		return fmt.Sprintf("# %s\n\n%s", jr.Title, jr.Content), nil
	}
	return "", fmt.Errorf("jina: empty response")
}
