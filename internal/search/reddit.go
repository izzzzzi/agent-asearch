package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type RedditBackend struct{}

func (b *RedditBackend) Name() Source   { return SourceReddit }
func (b *RedditBackend) Available() bool { return true }

func (b *RedditBackend) Search(query string, limit int) ([]Result, error) {
	// 1. Try rdt-cli first (cookie-authenticated, no rate limits)
	if _, err := exec.LookPath("rdt"); err == nil {
		results, err := rdtSearch(query, limit)
		if err == nil && len(results) > 0 {
			return results, nil
		}
	}

	// 2. Try cookie file (~/.asearch/reddit-cookies.txt)
	if results, err := redditWithCookies(query, limit); err == nil && len(results) > 0 {
		return results, nil
	}

	// 3. Fall back to Reddit public JSON API
	return redditPublicJSON(query, limit)
}

func rdtSearch(query string, limit int) ([]Result, error) {
	cmd := exec.Command("rdt", "search", query, "--limit", fmt.Sprint(limit), "--json")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("rdt search failed: %w (run: rdt login)", err)
	}

	// rdt-cli returns {"ok":false,...} when not authenticated
	var errResp struct {
		OK    bool `json:"ok"`
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(out, &errResp) == nil && !errResp.OK {
		return nil, fmt.Errorf("rdt: %s (run: rdt login)", errResp.Error.Message)
	}

	var posts []struct {
		Title       string  `json:"title"`
		Permalink   string  `json:"permalink"`
		Selftext    string  `json:"selftext"`
		Score       float64 `json:"score"`
		NumComments int     `json:"num_comments"`
		Subreddit   string  `json:"subreddit"`
		CreatedUTC  float64 `json:"created_utc"`
	}
	if err := json.Unmarshal(out, &posts); err != nil {
		return nil, fmt.Errorf("rdt parse error: %w", err)
	}

	var results []Result
	for _, p := range posts {
		snippet := p.Selftext
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		if snippet == "" {
			snippet = fmt.Sprintf("r/%s post", p.Subreddit)
		}
		engagement := fmt.Sprintf("↑%.0f | 💬%d | r/%s", p.Score, p.NumComments, p.Subreddit)
		results = append(results, Result{
			Source:     SourceReddit,
			Title:      p.Title,
			URL:        "https://www.reddit.com" + p.Permalink,
			Snippet:    snippet,
			Score:      p.Score,
			Engagement: engagement,
		})
	}
	return results, nil
}

func redditWithCookies(query string, limit int) ([]Result, error) {
	cookieFile := os.ExpandEnv("$HOME/.asearch/reddit-cookies.txt")
	data, err := os.ReadFile(cookieFile)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	u := fmt.Sprintf("https://www.reddit.com/search.json?q=%s&limit=%d&sort=relevance&raw_json=1",
		query, limit)
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent", "asearch/1.0 (agent-search-cli)")
	req.Header.Set("Cookie", parseNetscapeCookies(string(data)))
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp struct {
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
				} `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	var results []Result
	for _, child := range apiResp.Data.Children {
		d := child.Data
		snippet := d.Selftext
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		if snippet == "" {
			snippet = fmt.Sprintf("r/%s post", d.Subreddit)
		}
		date := time.Unix(int64(d.Created), 0).Format("2006-01-02")
		engagement := fmt.Sprintf("↑%.0f | 💬%d | r/%s", d.Score, d.NumComments, d.Subreddit)
		results = append(results, Result{
			Source:     SourceReddit,
			Title:      d.Title,
			URL:        "https://www.reddit.com" + d.Permalink,
			Snippet:    snippet,
			Date:       date,
			Score:      d.Score,
			Engagement: engagement,
		})
	}
	return results, nil
}

func parseNetscapeCookies(raw string) string {
	var cookies []string
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 7 {
			cookies = append(cookies, parts[5]+"="+parts[6])
		}
	}
	return strings.Join(cookies, "; ")
}

func redditPublicJSON(query string, limit int) ([]Result, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	u := fmt.Sprintf("https://www.reddit.com/search.json?q=%s&limit=%d&sort=relevance",
		query, limit)
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent", "asearch/1.0 (agent-search-cli)")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("reddit public JSON request failed: %w", err)
	}
	defer resp.Body.Close()

	var data struct {
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
				} `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("reddit public JSON parse error: %w (try: pipx install rdt-cli && rdt login)", err)
	}

	var results []Result
	for _, child := range data.Data.Children {
		d := child.Data
		snippet := d.Selftext
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		if snippet == "" {
			snippet = fmt.Sprintf("r/%s post", d.Subreddit)
		}
		date := time.Unix(int64(d.Created), 0).Format("2006-01-02")
		engagement := fmt.Sprintf("↑%.0f | 💬%d | r/%s", d.Score, d.NumComments, d.Subreddit)
		results = append(results, Result{
			Source:     SourceReddit,
			Title:      d.Title,
			URL:        "https://www.reddit.com" + d.Permalink,
			Snippet:    snippet,
			Date:       date,
			Score:      d.Score,
			Engagement: engagement,
		})
	}
	return results, nil
}
