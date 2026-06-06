package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

// SearXNGBackend — self-hosted meta-search engine.
// Zero API costs, no rate limits, aggregates Google, Bing, DDG, Wikipedia + 70 engines.
// Run locally: docker run -d -p 8080:8080 searxng/searxng
// Docs: https://docs.searxng.org
type SearXNGBackend struct{}

func (b *SearXNGBackend) Name() Source { return SourceSearXNG }
func (b *SearXNGBackend) Available() bool {
	return os.Getenv("ASEARCH_SEARXNG_URL") != ""
}

func (b *SearXNGBackend) Search(query string, limit int) ([]Result, error) {
	baseURL := os.Getenv("ASEARCH_SEARXNG_URL")
	if baseURL == "" {
		return nil, fmt.Errorf(
			"searxng: set ASEARCH_SEARXNG_URL or run locally:\n"+
				"  docker run -d -p 8080:8080 searxng/searxng\n"+
				"  export ASEARCH_SEARXNG_URL=http://localhost:8080",
		)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	params := url.Values{}
	params.Set("q", query)
	params.Set("format", "json")
	params.Set("categories", "general")
	params.Set("pageno", "1")

	resp, err := client.Get(strings.TrimRight(baseURL, "/") + "/search?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("searxng request failed: %w", err)
	}
	defer resp.Body.Close()

	var apiResp struct {
		Results []struct {
			Title       string   `json:"title"`
			URL         string   `json:"url"`
			Content     string   `json:"content"`
			Engine      string   `json:"engine"`
			Score       float64  `json:"score"`
			Engines     []string `json:"engines"`
			PublishedDate string `json:"publishedDate"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("searxng parse error: %w", err)
	}

	var results []Result
	for _, r := range apiResp.Results {
		snippet := r.Content
		if len(snippet) > 300 {
			snippet = snippet[:300] + "..."
		}
		engineList := strings.Join(r.Engines, ",")
		if engineList == "" {
			engineList = r.Engine
		}
		engagement := fmt.Sprintf("via %s", engineList)
		if r.Score > 0 {
			engagement += fmt.Sprintf(" | score: %.2f", r.Score)
		}
		results = append(results, Result{
			Source:     SourceSearXNG,
			Title:      r.Title,
			URL:        r.URL,
			Snippet:    snippet,
			Date:       r.PublishedDate,
			Score:      r.Score,
			Engagement: engagement,
		})
		if len(results) >= limit {
			break
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("searxng: 0 results (check instance: %s)", baseURL)
	}
	return results, nil
}

// IsDockerRunning checks if Docker is available for local SearXNG.
func IsDockerRunning() bool {
	cmd := exec.Command("docker", "info")
	return cmd.Run() == nil
}
