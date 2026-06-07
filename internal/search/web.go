package search

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/izzzzzi/agent-asearch/internal/htmlq"
)

// WebBackend scrapes search engines directly using stdlib HTML parser.
// Priority: SearXNG > DDG > Bing > Wikipedia > error.
type WebBackend struct{}

func (b *WebBackend) Name() Source    { return SourceWeb }
func (b *WebBackend) Available() bool { return true }

func (b *WebBackend) Search(query string, limit int) ([]Result, error) {
	// 1. SearXNG — if configured (self-hosted)
	if apiKey("searxng") != "" {
		if r := tryBackend(&SearXNGBackend{}, query, limit); r != nil {
			return r, nil
		}
	}

	// 2. API providers via config or env
	// (providers check config.GetKey internally)

	// 3. DuckDuckGo HTML — zero config, no JS
	if r := ddgSearch(query, limit); r != nil {
		return r, nil
	}

	// 4. Wikipedia API
	if r := wikiSearch(query, limit); r != nil {
		return r, nil
	}

	// 5. Bing HTML
	if r := bingSearch(query, limit); r != nil {
		return r, nil
	}

	return nil, fmt.Errorf(
		"web search available via:\n" +
			"  • Zero-config:  DuckDuckGo, Wikipedia, Bing (no key needed)\n" +
			"  • API key:       asearch config set <provider> <key>\n" +
			"                  or export TAVILY_API_KEY, EXA_API_KEY, etc.\n" +
			"  • Self-hosted:   docker run searxng/searxng + ASEARCH_SEARXNG_URL",
	)
}

func tryBackend(b Backend, query string, limit int) []Result {
	r, err := b.Search(query, limit)
	if err != nil || len(r) == 0 {
		return nil
	}
	for i := range r {
		r[i].Source = SourceWeb
	}
	return r
}

// ── DuckDuckGo HTML ────────────────────────────────────────────────────

func ddgSearch(query string, limit int) []Result {
	u := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))
	body := httpGet(u)
	if body == "" {
		return nil
	}

	var results []Result
	for _, item := range htmlq.FindAll(body, "div", "result") {
		if len(results) >= limit {
			break
		}

		// Extract title link
		links := item.Find("a", "")
		title, href := "", ""
		for _, a := range links {
			if strings.Contains(a.Attrs["class"], "result__a") {
				title = a.Text
				href = a.Attrs["href"]
				// Decode DDG redirect URL
				if strings.Contains(href, "uddg=") {
					for _, p := range strings.Split(href, "?") {
						for _, param := range strings.Split(p, "&") {
							if strings.HasPrefix(param, "uddg=") {
								if decoded, err := url.QueryUnescape(param[5:]); err == nil {
									href = decoded
								}
							}
						}
					}
				}
				break
			}
		}
		if title == "" {
			continue
		}

		// Extract snippet
		snippet := ""
		for _, a := range item.Find("a", "result__snippet") {
			snippet = a.Text
			break
		}
		if snippet == "" {
			for _, s := range item.Find("span", "result__snippet") {
				snippet = s.Text
				break
			}
		}

		results = append(results, Result{
			Source: SourceWeb, Title: title, URL: href, Snippet: snippet,
			Engagement: "via DuckDuckGo",
		})
	}

	if len(results) == 0 {
		// Try alternative: old-style result links
		for _, a := range htmlq.FindAll(body, "a", "result-link") {
			if len(results) >= limit {
				break
			}
			results = append(results, Result{
				Source: SourceWeb, Title: a.Text, URL: a.Attrs["href"],
				Engagement: "via DuckDuckGo",
			})
		}
	}
	if len(results) == 0 {
		return nil
	}
	return results
}

// ── Wikipedia API ───────────────────────────────────────────────────────

func wikiSearch(query string, limit int) []Result {
	u := fmt.Sprintf("https://en.wikipedia.org/w/api.php?action=query&list=search&srsearch=%s&format=json&srlimit=%d",
		url.QueryEscape(query), limit)
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent", "asearch/1.0 (agent-search-cli)")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var api struct {
		Query struct {
			Search []struct {
				Title   string `json:"title"`
				Snippet string `json:"snippet"`
				PageID  int    `json:"pageid"`
			} `json:"search"`
		} `json:"query"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&api); err != nil {
		return nil
	}

	var results []Result
	for _, s := range api.Query.Search {
		// Strip HTML from snippet
		snip := strings.ReplaceAll(s.Snippet, "&amp;", "&")
		snip = strings.ReplaceAll(snip, "&lt;", "<")
		snip = strings.ReplaceAll(snip, "&gt;", ">")
		snip = stripHTML(snip)
		if len(snip) > 200 {
			snip = snip[:200] + "..."
		}
		results = append(results, Result{
			Source: SourceWeb, Title: s.Title,
			URL:        fmt.Sprintf("https://en.wikipedia.org/wiki/%s", url.PathEscape(s.Title)),
			Snippet:    snip,
			Engagement: "via Wikipedia",
		})
	}
	return results
}

// ── Bing HTML ───────────────────────────────────────────────────────────

func bingSearch(query string, limit int) []Result {
	u := fmt.Sprintf("https://www.bing.com/search?q=%s", url.QueryEscape(query))
	body := httpGet(u)
	if body == "" {
		return nil
	}

	var results []Result
	for _, item := range htmlq.FindAll(body, "li", "b_algo") {
		if len(results) >= limit {
			break
		}
		// Find the title link
		links := item.Find("a", "")
		title, href := "", ""
		for _, a := range links {
			if h := a.Attrs["href"]; strings.HasPrefix(h, "http") {
				title, href = a.Text, h
				break
			}
		}
		if title == "" {
			continue
		}
		// Find description
		snippet := ""
		for _, p := range item.Find("p", "") {
			if c := p.Attrs["class"]; c == "b_lineclamp2" || c == "b_algoSlug" {
				snippet = p.Text
				break
			}
		}
		if snippet == "" {
			for _, d := range item.Find("div", "b_caption") {
				snippet = d.Text
				break
			}
		}
		results = append(results, Result{
			Source: SourceWeb, Title: title, URL: href, Snippet: snippet,
			Engagement: "via Bing",
		})
	}
	return results
}

// ── Helpers ─────────────────────────────────────────────────────────────

func httpGet(url string) string {
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return ""
	}
	var buf strings.Builder
	buf.Grow(100000)
	io.Copy(&buf, resp.Body)
	return buf.String()
}

func stripHTML(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}
