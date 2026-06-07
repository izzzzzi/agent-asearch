package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type RedditBackend struct{}

func (b *RedditBackend) Name() Source   { return SourceReddit }
func (b *RedditBackend) Available() bool { return true }

func (b *RedditBackend) Search(query string, limit int) ([]Result, error) {
	if results, err := redditWithCookies(query, limit); err == nil && len(results) > 0 {
		return results, nil
	}
	return redditPublicJSON(query, limit)
}

func redditWithCookies(query string, limit int) ([]Result, error) {
	cookieFile := os.ExpandEnv("$HOME/.asearch/reddit-cookies.txt")
	data, err := os.ReadFile(cookieFile)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	u := fmt.Sprintf("https://www.reddit.com/search.json?q=%s&limit=%d&sort=relevance&raw_json=1&t=year",
		url.QueryEscape(query), limit)
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent", "asearch/1.0")
	req.Header.Set("Cookie", parseNetscapeCookies(string(data)))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("reddit status %d", resp.StatusCode)
	}

	var r struct {
		Data struct {
			Children []struct {
				Data struct {
					Title       string  `json:"title"`
					Permalink   string  `json:"permalink"`
					Selftext    string  `json:"selftext"`
					Score       float64 `json:"score"`
					NumComments int     `json:"num_comments"`
					Subreddit   string  `json:"subreddit"`
					Created     float64 `json:"created_utc"`
					Author      string  `json:"author"`
				} `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	var results []Result
	for _, c := range r.Data.Children {
		d := c.Data
		snippet := d.Selftext
		if len(snippet) > 200 { snippet = snippet[:200] + "..." }
		if snippet == "" { snippet = fmt.Sprintf("r/%s · %s", d.Subreddit, d.Author) }
		date := time.Unix(int64(d.Created), 0).Format("2006-01-02")
		results = append(results, Result{
			Source: SourceReddit, Title: d.Title,
			URL: "https://www.reddit.com" + d.Permalink,
			Snippet: snippet, Date: date,
			Score: d.Score,
			Engagement: fmt.Sprintf("↑%.0f | 💬%d | r/%s | %s",
				d.Score, d.NumComments, d.Subreddit, d.Author),
		})
	}
	return results, nil
}

func redditPublicJSON(query string, limit int) ([]Result, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	u := fmt.Sprintf("https://www.reddit.com/search.json?q=%s&limit=%d&sort=relevance",
		url.QueryEscape(query), limit)
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent", "asearch/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("reddit request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("rate limited — save cookies to ~/.asearch/reddit-cookies.txt")
	}

	var r struct {
		Data struct {
			Children []struct {
				Data struct {
					Title       string  `json:"title"`
					Permalink   string  `json:"permalink"`
					Selftext    string  `json:"selftext"`
					Score       float64 `json:"score"`
					NumComments int     `json:"num_comments"`
					Subreddit   string  `json:"subreddit"`
					Created     float64 `json:"created_utc"`
					Author      string  `json:"author"`
				} `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("public JSON failed: %w (try cookies)", err)
	}

	var results []Result
	for _, c := range r.Data.Children {
		d := c.Data
		snippet := d.Selftext
		if len(snippet) > 200 { snippet = snippet[:200] + "..." }
		if snippet == "" { snippet = fmt.Sprintf("r/%s · %s", d.Subreddit, d.Author) }
		date := time.Unix(int64(d.Created), 0).Format("2006-01-02")
		results = append(results, Result{
			Source: SourceReddit, Title: d.Title,
			URL: "https://www.reddit.com" + d.Permalink,
			Snippet: snippet, Date: date,
			Score: d.Score,
			Engagement: fmt.Sprintf("↑%.0f | 💬%d | r/%s", d.Score, d.NumComments, d.Subreddit),
		})
	}
	return results, nil
}

func parseNetscapeCookies(raw string) string {
	var parts []string
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") { continue }
		fields := strings.Split(line, "\t")
		if len(fields) >= 7 {
			parts = append(parts, fields[5]+"="+fields[6])
		}
	}
	return strings.Join(parts, "; ")
}
