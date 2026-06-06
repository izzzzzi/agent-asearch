package search

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type YouTubeBackend struct{}

func (b *YouTubeBackend) Name() Source { return SourceYouTube }
func (b *YouTubeBackend) Available() bool {
	_, err := exec.LookPath("yt-dlp")
	return err == nil
}

func (b *YouTubeBackend) Search(query string, limit int) ([]Result, error) {
	if !b.Available() {
		return nil, fmt.Errorf("yt-dlp not installed; run: brew install yt-dlp")
	}

	// yt-dlp flat search
	cmd := exec.Command("yt-dlp",
		"ytsearch"+fmt.Sprint(limit)+":"+query,
		"--flat-playlist",
		"--dump-json",
		"--no-warnings",
		"-q",
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("youtube search failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var results []Result
	for _, line := range lines {
		if line == "" {
			continue
		}
		var v struct {
			Title       string  `json:"title"`
			ID          string  `json:"id"`
			URL         string  `json:"webpage_url"`
			Description string  `json:"description"`
			Duration    float64 `json:"duration"`
			ViewCount   float64 `json:"view_count"`
			Channel     string  `json:"channel"`
			UploadDate  string  `json:"upload_date"`
		}
		if err := json.Unmarshal([]byte(line), &v); err != nil {
			continue
		}
		snippet := v.Description
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		duration := fmt.Sprintf("%d:%02d", int(v.Duration)/60, int(v.Duration)%60)
		engagement := fmt.Sprintf("▶%.0f views | %s | %s",
			v.ViewCount, duration, v.Channel)
		date := v.UploadDate
		if len(date) == 8 {
			date = date[0:4] + "-" + date[4:6] + "-" + date[6:8]
		}
		results = append(results, Result{
			Source:     SourceYouTube,
			Title:      v.Title,
			URL:        v.URL,
			Snippet:    snippet,
			Date:       date,
			Score:      v.ViewCount,
			Engagement: engagement,
		})
	}
	return results, nil
}
