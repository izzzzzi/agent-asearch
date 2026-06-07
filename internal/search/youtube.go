package search

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type YouTubeBackend struct{}

func (b *YouTubeBackend) Name() Source    { return SourceYouTube }
func (b *YouTubeBackend) Available() bool { return true }

func (b *YouTubeBackend) Search(query string, limit int) ([]Result, error) {
	cookieStr, sapisid := readYTCookies()
	if sapisid != "" {
		return searchYT(query, limit, cookieStr, sapisid)
	}
	return nil, fmt.Errorf("youtube: save cookies to ~/.asearch/youtube-cookies.txt")
}

func readYTCookies() (string, string) {
	data, err := os.ReadFile(os.ExpandEnv("$HOME/.asearch/youtube-cookies.txt"))
	if err != nil {
		return "", ""
	}
	var parts []string
	var sapisid string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		f := strings.Split(line, "\t")
		if len(f) >= 7 {
			if f[5] == "SAPISID" {
				sapisid = f[6]
			}
			parts = append(parts, f[5]+"="+f[6])
		}
	}
	return strings.Join(parts, "; "), sapisid
}

func searchYT(query string, limit int, cookieStr, sapisid string) ([]Result, error) {
	ts := time.Now().Unix()
	h := sha1.New()
	h.Write([]byte(fmt.Sprintf("%d %s https://www.youtube.com", ts, sapisid)))
	auth := fmt.Sprintf("SAPISIDHASH %d_%s", ts, hex.EncodeToString(h.Sum(nil)))

	payload := fmt.Sprintf(`{"context":{"client":{"clientName":"WEB","clientVersion":"2.20250314.07.00"}},"query":%q}`, query)

	req, _ := http.NewRequest("POST",
		"https://www.youtube.com/youtubei/v1/search?key=AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8",
		strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	req.Header.Set("Authorization", auth)
	req.Header.Set("Cookie", cookieStr)
	req.Header.Set("X-YouTube-Client-Name", "1")
	req.Header.Set("X-YouTube-Client-Version", "2.20250314.07.00")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("youtube: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("youtube: HTTP %d", resp.StatusCode)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("youtube: %w", err)
	}

	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, fmt.Errorf("youtube parse: %w", err)
	}

	// Navigate: contents > twoColumnSearchResultsRenderer > primaryContents > sectionListRenderer > contents
	c1, _ := body["contents"].(map[string]any)
	if c1 == nil {
		return nil, fmt.Errorf("youtube: no contents")
	}
	c2, _ := c1["twoColumnSearchResultsRenderer"].(map[string]any)
	if c2 == nil {
		return nil, fmt.Errorf("youtube: no search renderer")
	}
	c3, _ := c2["primaryContents"].(map[string]any)
	if c3 == nil {
		return nil, fmt.Errorf("youtube: no primary contents")
	}
	c4, _ := c3["sectionListRenderer"].(map[string]any)
	if c4 == nil {
		return nil, fmt.Errorf("youtube: no section list")
	}
	sections, _ := c4["contents"].([]any)
	if sections == nil {
		return nil, fmt.Errorf("youtube: no sections")
	}

	var results []Result
	for _, secAny := range sections {
		sec, _ := secAny.(map[string]any)
		if sec == nil {
			continue
		}
		isr, _ := sec["itemSectionRenderer"].(map[string]any)
		if isr == nil {
			continue
		}
		items, _ := isr["contents"].([]any)
		if items == nil {
			continue
		}

		for _, itemAny := range items {
			if len(results) >= limit {
				break
			}
			item, _ := itemAny.(map[string]any)
			if item == nil {
				continue
			}
			vr, _ := item["videoRenderer"].(map[string]any)
			if vr == nil {
				// Count non-video items
				continue
			}

			videoID, _ := vr["videoId"].(string)
			if videoID == "" {
				continue
			}

			title := getText(vr, "title")
			if title == "" {
				continue
			}

			channel := getText(vr, "ownerText")
			if channel == "" {
				channel = getText(vr, "longBylineText")
			}
			views := getSimple(vr, "viewCountText")
			dur := getSimple(vr, "lengthText")
			pub := getSimple(vr, "publishedTimeText")

			results = append(results, Result{
				Source: SourceYouTube, Title: title,
				URL:        fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID),
				Snippet:    fmt.Sprintf("%s — %s (%s)", title, channel, pub),
				Date:       pub,
				Engagement: fmt.Sprintf("▶ %s | %s | %s", views, dur, channel),
			})
		}
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("youtube: no results")
	}
	return results, nil
}

func getText(m map[string]any, key string) string {
	obj, _ := m[key].(map[string]any)
	if obj == nil {
		return ""
	}
	runs, _ := obj["runs"].([]any)
	if len(runs) == 0 {
		return ""
	}
	r0, _ := runs[0].(map[string]any)
	if r0 == nil {
		return ""
	}
	s, _ := r0["text"].(string)
	return s
}

func getSimple(m map[string]any, key string) string {
	obj, _ := m[key].(map[string]any)
	if obj == nil {
		return ""
	}
	s, _ := obj["simpleText"].(string)
	return s
}
