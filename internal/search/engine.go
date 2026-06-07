package search

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type Source string

const (
	SourceWeb     Source = "web"
	SourceReddit  Source = "reddit"
	SourceTwitter Source = "twitter"
	SourceHN      Source = "hn"
	SourceGitHub  Source = "github"
	SourceYouTube Source = "youtube"
	SourceTavily  Source = "tavily"
	SourceSearXNG Source = "searxng"
	SourceCode    Source = "code"
)

type Result struct {
	Seq        int     `json:"seq"`
	Source     Source  `json:"source"`
	Title      string  `json:"title"`
	URL        string  `json:"url"`
	Snippet    string  `json:"snippet"`
	Date       string  `json:"date,omitempty"`
	Score      float64 `json:"score,omitempty"`
	Engagement string  `json:"engagement,omitempty"`
	RawMeta    any     `json:"raw_meta,omitempty"`
	Related    []int   `json:"related,omitempty"`
}

type SearchRequest struct {
	Query    string
	Sources  []Source
	Limit    int
	Timeout  time.Duration
	CrossRef bool
}

type SearchResult struct {
	SID     string   `json:"sid,omitempty"`
	Query   string   `json:"query"`
	Sources []Source `json:"sources"`
	Total   int      `json:"total"`
	Results []Result `json:"results,omitempty"`
}

type Backend interface {
	Name() Source
	Search(query string, limit int) ([]Result, error)
	Available() bool
}

func Search(req SearchRequest) (*SearchResult, error) {
	if req.Timeout == 0 {
		req.Timeout = 30 * time.Second
	}
	if req.Limit == 0 {
		req.Limit = 50
	}
	if len(req.Sources) == 0 {
		req.Sources = []Source{SourceWeb}
	}

	allBackends := []Backend{
		&TavilyBackend{},
		&SearXNGBackend{},
		&SerperBackend{},
		&PerplexityBackend{},
		&SerpAPIBackend{},
		&YouBackend{},
		&FirecrawlBackend{},
		&ParallelBackend{},
		&BraveBackend{},
		&ExaBackend{},
		&WebBackend{},
		&RedditBackend{},
		&HNBackend{},
		&GitHubBackend{},
		&YouTubeBackend{},
		&CodeBackend{},
		&TwitterBackend{},
	}

	backends := make([]Backend, 0)
	for _, b := range allBackends {
		for _, s := range req.Sources {
			if b.Name() == s && b.Available() {
				backends = append(backends, b)
				break
			}
		}
	}

	var allResults []Result
	for _, b := range backends {
		results, err := b.Search(req.Query, req.Limit)
		if err != nil {
			continue
		}
		allResults = append(allResults, results...)
	}

	// Deduplicate by normalized URL
	allResults = deduplicate(allResults)

	// Re-number after dedup
	for i := range allResults {
		allResults[i].Seq = i + 1
	}

	// Cross-reference: group results by normalized entity
	if req.CrossRef {
		allResults = crossReference(allResults, req.Query)
		// Re-number again after cross-ref (may add results)
		for i := range allResults {
			allResults[i].Seq = i + 1
		}
	}

	return &SearchResult{
		Query:   req.Query,
		Sources: req.Sources,
		Total:   len(allResults),
		Results: allResults,
	}, nil
}

func SaveResults(sid string, result *SearchResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	p := resultsPath(sid)
	if err := os.MkdirAll(resultsDir(), 0700); err != nil {
		return err
	}
	return os.WriteFile(p, data, 0600)
}

func ReadResults(sid string, seq, limit int) ([]Result, int, error) {
	p := resultsPath(sid)
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, 0, fmt.Errorf("results not found for session %s", sid)
	}
	var sr SearchResult
	if err := json.Unmarshal(data, &sr); err != nil {
		return nil, 0, err
	}
	total := sr.Total
	if seq <= 0 {
		seq = 1
	}
	start := seq - 1
	if start >= len(sr.Results) {
		return nil, total, nil
	}
	end := start + limit
	if end > len(sr.Results) {
		end = len(sr.Results)
	}
	return sr.Results[start:end], total, nil
}

func FilterResults(sid string, source Source) ([]Result, int, error) {
	p := resultsPath(sid)
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, 0, fmt.Errorf("results not found for session %s", sid)
	}
	var sr SearchResult
	if err := json.Unmarshal(data, &sr); err != nil {
		return nil, 0, err
	}
	var filtered []Result
	for _, r := range sr.Results {
		if r.Source == source {
			filtered = append(filtered, r)
		}
	}
	return filtered, sr.Total, nil
}

func resultsDir() string {
	dir := os.Getenv("ASEARCH_STATE_DIR")
	if dir != "" {
		return dir + "/results"
	}
	home, _ := os.UserHomeDir()
	return home + "/.asearch/results"
}

func resultsPath(sid string) string {
	return resultsDir() + "/" + strings.ToLower(sid) + ".json"
}

// deduplicate removes results with the same normalized URL.
// First occurrence wins — earlier backends have higher priority.
func deduplicate(results []Result) []Result {
	seen := make(map[string]bool)
	var out []Result
	for _, r := range results {
		key := normalizeURL(r.URL)
		if key == "" {
			// No URL — keep it (e.g. AI answers)
			out = append(out, r)
			continue
		}
		if seen[key] {
			continue // duplicate, skip
		}
		seen[key] = true
		out = append(out, r)
	}
	return out
}

// normalizeURL strips noise from URLs for dedup comparison.
func normalizeURL(raw string) string {
	if raw == "" {
		return ""
	}
	s := strings.ToLower(raw)
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimPrefix(s, "www.")
	s = strings.TrimSuffix(s, "/")

	// For YouTube, only strip tracking params, keep video ID
	if strings.Contains(s, "youtube.com/watch") || strings.Contains(s, "youtu.be/") {
		if idx := strings.Index(s, "?"); idx >= 0 {
			query := s[idx+1:]
			var cleanParams []string
			for _, param := range strings.Split(query, "&") {
				// Keep video ID, strip tracking
				if strings.HasPrefix(param, "v=") {
					cleanParams = append(cleanParams, param)
				}
			}
			if len(cleanParams) > 0 {
				s = s[:idx] + "?" + strings.Join(cleanParams, "&")
			} else {
				s = s[:idx]
			}
		}
		return s
	}

	// Strip query params for Reddit, HN
	if strings.Contains(s, "reddit.com/") || strings.Contains(s, "ycombinator.com/") {
		if idx := strings.Index(s, "?"); idx >= 0 {
			s = s[:idx]
		}
		s = strings.TrimSuffix(s, "/")
	}

	// Strip fragment
	if idx := strings.Index(s, "#"); idx >= 0 {
		s = s[:idx]
	}
	return s
}

// splitQuery attempts to split a search query into entity parts.
// Recognises patterns: "X by Y", "X from Y", "X | Y".
func splitQuery(q string) []string {
	for _, sep := range []string{" by ", " from ", " | ", " |"} {
		if parts := strings.SplitN(q, sep, 2); len(parts) == 2 {
			return []string{strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])}
		}
	}
	return []string{q}
}

// entityName extracts a normalised entity identifier from a search result.
// Uses RawMeta fields where available, falls back to title/URL heuristics.
func entityName(r Result) string {
	raw, ok := r.RawMeta.(map[string]any)
	if ok {
		// GitHub repos: use full_name
		if fn, _ := raw["full_name"].(string); fn != "" {
			return strings.ToLower(fn)
		}
		// Reddit: use subreddit + title
		if sub, _ := raw["subreddit"].(string); sub != "" {
			return strings.ToLower(sub + "/" + r.Title)
		}
		// YouTube: use channel
		if ch, _ := raw["channel"].(string); ch != "" {
			return strings.ToLower(ch + "/" + r.Title)
		}
	}
	// GitHub URL pattern: github.com/owner/repo
	if strings.Contains(r.URL, "github.com/") {
		parts := strings.SplitN(strings.TrimPrefix(r.URL, "https://github.com/"), "/", 3)
		if len(parts) >= 2 {
			return strings.ToLower(parts[0] + "/" + parts[1])
		}
	}
	// HN URL: use story title
	if strings.Contains(r.URL, "ycombinator.com/") || strings.Contains(r.URL, "news.ycombinator.com") {
		// Normalise HN post title to match
		return strings.ToLower(strings.TrimSpace(r.Title))
	}
	// Default: normalise URL host + first path segment
	if trimmed := strings.TrimPrefix(r.URL, "https://"); trimmed != r.URL {
		parts := strings.SplitN(trimmed, "/", 3)
		if len(parts) >= 2 {
			return strings.ToLower(parts[0] + "/" + parts[1])
		}
	}
	return ""
}

// crossReference groups results by normalised entity and sets Related indices.
// Also fires sub-queries when the query has multiple entity parts.
// Requires package-level access to backends (re-registers via SearchRequest).
func crossReference(results []Result, query string) []Result {
	parts := splitQuery(query)
	if len(parts) <= 1 {
		return results // nothing to cross-reference
	}

	// Build entity → indices map
	entityIdx := make(map[string][]int)
	for i, r := range results {
		ent := entityName(r)
		if ent == "" {
			continue
		}
		entityIdx[ent] = append(entityIdx[ent], i)
	}

	// Set Related for results sharing an entity
	for _, indices := range entityIdx {
		if len(indices) < 2 {
			continue
		}
		for _, idx := range indices {
			for _, other := range indices {
				if other != idx {
					results[idx].Related = append(results[idx].Related, other+1) // 1-based Seq
				}
			}
		}
	}

	return results
}
