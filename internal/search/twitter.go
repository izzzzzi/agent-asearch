package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"
)

// TwitterBackend searches X/Twitter via the official X API v2.
//
// Requires a Bearer Token from https://developer.twitter.com
// (free tier: 500k reads/month).
//
// Configure with:
//
//	asearch config set twitter "AAAAAAA..."
//	export TWITTER_BEARER_TOKEN="AAAAAAA..."
//
// Note: X has removed all anonymous/guest search APIs and the internal
// cookie-based GraphQL endpoint requires Cloudflare cookies (cf_bm)
// that expire too quickly to be practical. X API v2 is the only
// reliable option.
type TwitterBackend struct{}

func (b *TwitterBackend) Name() Source    { return SourceTwitter }
func (b *TwitterBackend) Available() bool { return true }

func (b *TwitterBackend) Search(query string, limit int) ([]Result, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}

	bearer := apiKey("twitter")
	if bearer == "" {
		return nil, fmt.Errorf("twitter: no bearer token configured")
	}
	bearer = strings.TrimPrefix(bearer, "Bearer ")

	return xAPIv2Search(query, limit, bearer)
}

// ── X API v2 ─────────────────────────────────────────────────────────────

const xAPIBase = "https://api.twitter.com/2"

func xAPIv2Search(query string, limit int, bearer string) ([]Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	u := fmt.Sprintf("%s/tweets/search/recent?query=%s&max_results=%d&tweet.fields=public_metrics,author_id,created_at&user.fields=username",
		xAPIBase, url.QueryEscape(query), minInt(limit, 100))

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+bearer)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("xapi v2: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var xres struct {
		Data     []tweetV2 `json:"data"`
		Includes *struct {
			Users []userV2 `json:"users"`
		} `json:"includes"`
		Meta struct {
			ResultCount int `json:"result_count"`
		} `json:"meta"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&xres); err != nil {
		return nil, err
	}
	if xres.Meta.ResultCount == 0 {
		return nil, fmt.Errorf("xapi v2: no results")
	}

	userMap := make(map[string]string)
	if xres.Includes != nil {
		for _, u := range xres.Includes.Users {
			userMap[u.ID] = u.Username
		}
	}

	return tweetsToResults(xres.Data, userMap), nil
}

type tweetV2 struct {
	ID        string        `json:"id"`
	Text      string        `json:"text"`
	AuthorID  string        `json:"author_id"`
	CreatedAt string        `json:"created_at"`
	Metrics   *tweetMetrics `json:"public_metrics"`
}

type tweetMetrics struct {
	Likes    int `json:"like_count"`
	Retweets int `json:"retweet_count"`
}

type userV2 struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// ── Helpers ──────────────────────────────────────────────────────────────

func tweetsToResults(tweets []tweetV2, userMap map[string]string) []Result {
	var results []Result
	for _, t := range tweets {
		snippet := truncateText(cleanTweetText(t.Text), 200)
		username := userMap[t.AuthorID]
		if username == "" {
			username = "?"
		}
		likes := 0
		retweets := 0
		if t.Metrics != nil {
			likes = t.Metrics.Likes
			retweets = t.Metrics.Retweets
		}
		results = append(results, Result{
			Source:     SourceTwitter,
			Title:      snippet,
			URL:        fmt.Sprintf("https://x.com/%s/status/%s", username, t.ID),
			Snippet:    snippet,
			Date:       formatAPIDate(t.CreatedAt),
			Score:      float64(likes),
			Engagement: fmt.Sprintf("♥%d | ↻%d | @%s", likes, retweets, username),
		})
	}
	return results
}

// truncateText safely truncates a string to maxRunes runes, adding "..." if truncated.
func truncateText(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "..."
}

func cleanTweetText(s string) string {
	var b strings.Builder
	space := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !space {
				b.WriteRune(' ')
				space = true
			}
		} else if r != '\u200e' && r != '\u200f' {
			b.WriteRune(r)
			space = false
		}
	}
	return strings.TrimSpace(b.String())
}

func formatAPIDate(iso string) string {
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return iso
	}
	return t.Format("2006-01-02")
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
