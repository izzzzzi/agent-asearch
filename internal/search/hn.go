package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// hnHit is the raw Algolia search hit populated into RawMeta.
type hnHit struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Points      int    `json:"points"`
	NumComments int    `json:"num_comments"`
	Author      string `json:"author"`
	ObjectID    string `json:"objectID"`
	CreatedAt   string `json:"created_at"`
	StoryText   string `json:"story_text"`
}

type HNBackend struct{}

func (b *HNBackend) Name() Source    { return SourceHN }
func (b *HNBackend) Available() bool { return true }

func (b *HNBackend) Search(query string, limit int) ([]Result, error) {
	// Use Algolia HN Search API (free, no auth)
	u := fmt.Sprintf("https://hn.algolia.com/api/v1/search?query=%s&hitsPerPage=%d&tags=story",
		url.QueryEscape(query), limit)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(u)
	if err != nil {
		return nil, fmt.Errorf("hn search failed: %w", err)
	}
	defer resp.Body.Close()

	var data struct {
		Hits []hnHit `json:"hits"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("hn parse error: %w", err)
	}

	var results []Result
	for _, hit := range data.Hits {
		hnURL := hit.URL
		if hnURL == "" {
			hnURL = "https://news.ycombinator.com/item?id=" + hit.ObjectID
		}
		snippet := hit.StoryText
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		engagement := fmt.Sprintf("▲%d | 💬%d | by %s", hit.Points, hit.NumComments, hit.Author)
		results = append(results, Result{
			Source:     SourceHN,
			Title:      hit.Title,
			URL:        hnURL,
			Snippet:    snippet,
			Date:       hit.CreatedAt[:10],
			Score:      float64(hit.Points),
			Engagement: engagement,
			RawMeta:    hit,
		})
	}
	return results, nil
}
