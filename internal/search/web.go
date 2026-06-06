package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

// WebBackend delegates to the best available web search API.
// Priority: Tavily > Brave > Exa > SearXNG > error with guidance.
type WebBackend struct{}

func (b *WebBackend) Name() Source   { return SourceWeb }
func (b *WebBackend) Available() bool { return true }

func (b *WebBackend) Search(query string, limit int) ([]Result, error) {
	// 1. Tavily — best quality
	if os.Getenv("TAVILY_API_KEY") != "" {
		tb := &TavilyBackend{}
		results, err := tb.Search(query, limit)
		if err == nil && len(results) > 0 {
			for i := range results {
				results[i].Source = SourceWeb
			}
			return results, nil
		}
	}

	// 2. Brave Search — 35B-page index, 2000 free/month
	if os.Getenv("BRAVE_API_KEY") != "" {
		bb := &BraveBackend{}
		results, err := bb.Search(query, limit)
		if err == nil && len(results) > 0 {
			for i := range results {
				results[i].Source = SourceWeb
			}
			return results, nil
		}
	}

	// 3. Exa — neural/semantic search
	if os.Getenv("EXA_API_KEY") != "" {
		eb := &ExaBackend{}
		results, err := eb.Search(query, limit)
		if err == nil && len(results) > 0 {
			for i := range results {
				results[i].Source = SourceWeb
			}
			return results, nil
		}
	}

	// 4. SearXNG — self-hosted
	if os.Getenv("ASEARCH_SEARXNG_URL") != "" {
		results, err := searchSearXNG(os.Getenv("ASEARCH_SEARXNG_URL"), query, limit)
		if err == nil && len(results) > 0 {
			return results, nil
		}
	}

	// 5. Clear guidance
	return nil, fmt.Errorf(
		"web search needs an API key. Options:\n"+
			"  • Tavily (best):     export TAVILY_API_KEY=\"tvly-...\"  (free tier at tavily.com)\n"+
			"  • Brave Search:      export BRAVE_API_KEY=\"BSA...\"     (2000 free/month at brave.com/search/api)\n"+
			"  • Exa (semantic):    export EXA_API_KEY=\"...\"           (free tier at exa.ai)\n"+
			"  • Self-host SearXNG: export ASEARCH_SEARXNG_URL=\"http://localhost:8080\"\n"+
			"  Or use --source tavily directly: asearch open --query \"...\" --source tavily",
	)
}

func braveSearch(query string, limit int) ([]Result, error) {
	return nil, fmt.Errorf("brave search: use BraveBackend directly")
}

func searchSearXNG(baseURL string, query string, limit int) ([]Result, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	params := url.Values{}
	params.Set("q", query)
	params.Set("format", "json")
	resp, err := client.Get(baseURL + "/search?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respData struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return nil, err
	}

	var results []Result
	for i, r := range respData.Results {
		if i >= limit {
			break
		}
		snippet := r.Content
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		results = append(results, Result{
			Source:  SourceWeb,
			Title:   r.Title,
			URL:     r.URL,
			Snippet: snippet,
		})
	}
	return results, nil
}
