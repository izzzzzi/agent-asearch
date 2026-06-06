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
// Priority: SearXNG (self-hosted, no limits) > Tavily > Exa > Brave > error.
type WebBackend struct{}

func (b *WebBackend) Name() Source   { return SourceWeb }
func (b *WebBackend) Available() bool { return true }

func (b *WebBackend) Search(query string, limit int) ([]Result, error) {
	// 0. SearXNG — self-hosted, zero cost, unlimited
	if os.Getenv("ASEARCH_SEARXNG_URL") != "" {
		if results := runBackend(&SearXNGBackend{}, query, limit); results != nil { return results, nil }
	}
	// 1. Tavily — best AI quality
	if os.Getenv("TAVILY_API_KEY") != "" {
		if results := runBackend(&TavilyBackend{}, query, limit); results != nil { return results, nil }
	}
	// 2. Perplexity — AI answers with citations
	if os.Getenv("PERPLEXITY_API_KEY") != "" {
		if results := runBackend(&PerplexityBackend{}, query, limit); results != nil { return results, nil }
	}
	// 3. Exa — neural/semantic search
	if os.Getenv("EXA_API_KEY") != "" {
		if results := runBackend(&ExaBackend{}, query, limit); results != nil { return results, nil }
	}
	// 4. Brave Search — 35B-page index
	if os.Getenv("BRAVE_API_KEY") != "" {
		if results := runBackend(&BraveBackend{}, query, limit); results != nil { return results, nil }
	}
	// 5. Serper — Google SERP
	if os.Getenv("SERPER_API_KEY") != "" {
		if results := runBackend(&SerperBackend{}, query, limit); results != nil { return results, nil }
	}
	// 6. SerpAPI — 40+ search engines
	if os.Getenv("SERPAPI_API_KEY") != "" {
		if results := runBackend(&SerpAPIBackend{}, query, limit); results != nil { return results, nil }
	}
	// 7. You.com
	if os.Getenv("YOU_API_KEY") != "" {
		if results := runBackend(&YouBackend{}, query, limit); results != nil { return results, nil }
	}
	// 8. Firecrawl — JS rendering
	if os.Getenv("FIRECRAWL_API_KEY") != "" {
		if results := runBackend(&FirecrawlBackend{}, query, limit); results != nil { return results, nil }
	}
	// 9. Parallel
	if os.Getenv("PARALLEL_API_KEY") != "" {
		if results := runBackend(&ParallelBackend{}, query, limit); results != nil { return results, nil }
	}

	// 10. Clear guidance
	return nil, fmt.Errorf(
		"web search needs a backend. Options (pick one):\n"+
			"  • SearXNG (unlimited): docker run -d searxng/searxng && export ASEARCH_SEARXNG_URL=...\n"+
			"  • Tavily (AI answers): export TAVILY_API_KEY=...    (free tier)\n"+
			"  • Perplexity (citations): export PERPLEXITY_API_KEY=...\n"+
			"  • Exa (semantic):        export EXA_API_KEY=...          (free tier)\n"+
			"  • Brave (35B index):     export BRAVE_API_KEY=...       (2000 free/mo)\n"+
			"  • Serper (Google SERP):  export SERPER_API_KEY=...      (2500 free/mo)\n"+
			"  • SerpAPI (40+ engines): export SERPAPI_API_KEY=...     (100 free/mo)\n"+
			"  • You.com:               export YOU_API_KEY=...           (free tier)\n"+
			"  • Firecrawl (JS pages):  export FIRECRAWL_API_KEY=...    (500 free/mo)\n"+
			"  • Parallel:              export PARALLEL_API_KEY=...\n"+
			"\nAll providers are optional — just set the env var for the one you want.",
	)
}

func braveSearch(query string, limit int) ([]Result, error) {
	return nil, fmt.Errorf("brave search: use BraveBackend directly")
}

func runBackend(b Backend, query string, limit int) []Result {
	results, err := b.Search(query, limit)
	if err != nil || len(results) == 0 {
		return nil
	}
	for i := range results {
		results[i].Source = SourceWeb
	}
	return results
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
