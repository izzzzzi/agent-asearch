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
}

type SearchRequest struct {
	Query   string
	Sources []Source
	Limit   int
	Timeout time.Duration
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
		&BraveBackend{},
		&ExaBackend{},
		&WebBackend{},
		&RedditBackend{},
		&HNBackend{},
		&GitHubBackend{},
		&YouTubeBackend{},
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
			fmt.Fprintf(os.Stderr, `{"ok":false,"source":"%s","warning":"%s"}`+"\n", b.Name(), err.Error())
			continue
		}
		for i := range results {
			results[i].Seq = len(allResults) + 1
		}
		allResults = append(allResults, results...)
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
