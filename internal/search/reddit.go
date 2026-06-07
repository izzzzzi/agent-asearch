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
	if results, err := searchRedditAPI(query, limit); err == nil && len(results) > 0 {
		return results, nil
	}
	return redditPublicJSON(query, limit)
}

// ── Client ──────────────────────────────────────────────────────────────

func NewRedditClient() *RedditClient {
	return &RedditClient{client: &http.Client{Timeout: 15 * time.Second}}
}

type RedditClient struct {
	client *http.Client
}

func (rc *RedditClient) cookies() string {
	data, err := os.ReadFile(os.ExpandEnv("$HOME/.asearch/reddit-cookies.txt"))
	if err != nil { return "" }
	return parseNetscapeCookies(string(data))
}

func (rc *RedditClient) get(path string, out any) error {
	u := "https://www.reddit.com" + path
	if !strings.Contains(path, "?") {
		u += "?raw_json=1"
	} else {
		u += "&raw_json=1"
	}
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent", "asearch/1.0")
	if c := rc.cookies(); c != "" {
		req.Header.Set("Cookie", c)
	}
	resp, err := rc.client.Do(req)
	if err != nil { return fmt.Errorf("reddit: %w", err) }
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("reddit: HTTP %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// ── Post types ──────────────────────────────────────────────────────────

type Post struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Permalink   string  `json:"permalink"`
	Selftext    string  `json:"selftext"`
	Score       int     `json:"score"`
	UpvoteRatio float64 `json:"upvote_ratio"`
	NumComments int     `json:"num_comments"`
	Subreddit   string  `json:"subreddit"`
	Created     float64 `json:"created_utc"`
	Author      string  `json:"author"`
	URL         string  `json:"url"`
	Domain      string  `json:"domain"`
	LinkFlair   string  `json:"link_flair_text"`
	Stickied    bool    `json:"stickied"`
	IsVideo     bool    `json:"is_video"`
	Thumbnail   string  `json:"thumbnail"`
}

type postWrap struct {
	Data struct {
		Children []struct {
			Data Post `json:"data"`
		} `json:"children"`
		Dist int `json:"dist"`
	} `json:"data"`
}

// ── Subreddit posts ─────────────────────────────────────────────────────

func (rc *RedditClient) SubredditPosts(subreddit, listing string, limit int) ([]Post, error) {
	if listing == "" { listing = "hot" }
	if limit == 0 { limit = 25 }
	if limit > 100 { limit = 100 }

	var w postWrap
	err := rc.get(fmt.Sprintf("/r/%s/%s.json?limit=%d", subreddit, listing, limit), &w)
	if err != nil { return nil, err }

	var posts []Post
	for _, c := range w.Data.Children {
		posts = append(posts, c.Data)
	}
	return posts, nil
}

// ── Post + Comments ─────────────────────────────────────────────────────

func (rc *RedditClient) PostComments(permalink string) (post Post, comments []Post, err error) {
	if !strings.HasPrefix(permalink, "/") { permalink = "/" + permalink }
	permalink = strings.TrimSuffix(permalink, "/") + ".json"

	var raw []json.RawMessage
	if err = rc.get(permalink, &raw); err != nil { return }

	if len(raw) < 2 { err = fmt.Errorf("reddit: unexpected response"); return }

	var pw postWrap
	json.Unmarshal(raw[0], &pw)
	if len(pw.Data.Children) > 0 {
		post = pw.Data.Children[0].Data
	}

	var cw struct {
		Data struct {
			Children []commentNode `json:"children"`
		} `json:"data"`
	}
	json.Unmarshal(raw[1], &cw)
	comments = flattenComments(cw.Data.Children)

	return
}

type commentNode struct {
	Kind string `json:"kind"`
	Data struct {
		ID        string          `json:"id"`
		Body      string          `json:"body"`
		Author    string          `json:"author"`
		Score     int             `json:"score"`
		Created   float64         `json:"created_utc"`
		Replies   json.RawMessage `json:"replies"`
		Permalink string          `json:"permalink"`
		Depth     int             `json:"depth"`
	} `json:"data"`
}

func flattenComments(nodes []commentNode) []Post {
	var comments []Post
	for _, n := range nodes {
		if n.Kind != "t1" { continue }
		text := n.Data.Body
		if len(text) > 200 { text = text[:200] + "..." }
		comments = append(comments, Post{
			Title: text, Author: n.Data.Author, Score: n.Data.Score,
			Created: n.Data.Created, Permalink: n.Data.Permalink,
		})
		if n.Data.Replies != nil {
			var rw struct {
				Data struct {
					Children []commentNode `json:"children"`
				} `json:"data"`
			}
			if json.Unmarshal(n.Data.Replies, &rw) == nil {
				comments = append(comments, flattenComments(rw.Data.Children)...)
			}
		}
	}
	return comments
}

// ── Subreddit info ──────────────────────────────────────────────────────

type SubredditInfo struct {
	Title           string `json:"title"`
	DisplayName     string `json:"display_name"`
	Description     string `json:"description"`
	PublicDesc      string `json:"public_description"`
	Subscribers     int    `json:"subscribers"`
	ActiveUserCount int    `json:"active_user_count"`
	Created         float64 `json:"created_utc"`
	Over18          bool   `json:"over18"`
	Lang            string `json:"lang"`
	URL             string `json:"url"`
}

func (rc *RedditClient) SubredditInfo(name string) (*SubredditInfo, error) {
	var r struct {
		Data SubredditInfo `json:"data"`
	}
	if err := rc.get("/r/"+name+"/about.json", &r); err != nil { return nil, err }
	r.Data.URL = "https://www.reddit.com" + r.Data.URL
	return &r.Data, nil
}

// ── Search ──────────────────────────────────────────────────────────────

func searchRedditAPI(query string, limit int) ([]Result, error) {
	rc := NewRedditClient()
	// Check connectivity
	if _, err := rc.SubredditPosts("all", "hot", 1); err != nil {
		return nil, err
	}
	var w postWrap
	err := rc.get(fmt.Sprintf("/search.json?q=%s&limit=%d&sort=relevance&t=year",
		url.QueryEscape(query), limit), &w)
	if err != nil { return nil, err }

	return postsToResults(w.Data.Children), nil
}

func redditPublicJSON(query string, limit int) ([]Result, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	u := fmt.Sprintf("https://www.reddit.com/search.json?q=%s&limit=%d&sort=relevance",
		url.QueryEscape(query), limit)
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent", "asearch/1.0")
	resp, err := client.Do(req)
	if err != nil { return nil, fmt.Errorf("reddit: %w", err) }
	defer resp.Body.Close()
	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("rate limited — save cookies to ~/.asearch/reddit-cookies.txt")
	}

	var w postWrap
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		return nil, fmt.Errorf("public JSON: %w (save cookies)", err)
	}
	return postsToResults(w.Data.Children), nil
}

func postsToResults(children []struct {
	Data Post `json:"data"`
}) []Result {
	var results []Result
	for _, c := range children {
		d := c.Data
		snippet := d.Selftext
		if len(snippet) > 200 { snippet = snippet[:200] + "..." }
		if snippet == "" { snippet = fmt.Sprintf("r/%s · %s", d.Subreddit, d.Author) }
		results = append(results, Result{
			Source: SourceReddit, Title: d.Title,
			URL: "https://www.reddit.com" + d.Permalink,
			Snippet: snippet, Date: time.Unix(int64(d.Created), 0).Format("2006-01-02"),
			Score: float64(d.Score),
			Engagement: fmt.Sprintf("↑%d | 💬%d | r/%s | %s", d.Score, d.NumComments, d.Subreddit, d.Author),
		})
	}
	return results
}

// ── Cookie parser ───────────────────────────────────────────────────────

func parseNetscapeCookies(raw string) string {
	var parts []string
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") { continue }
		f := strings.Split(line, "\t")
		if len(f) >= 7 { parts = append(parts, f[5]+"="+f[6]) }
	}
	return strings.Join(parts, "; ")
}
