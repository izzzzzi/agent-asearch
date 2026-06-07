package search

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// CodeBackend searches code on GitHub via gh CLI.
type CodeBackend struct{}

func (b *CodeBackend) Name() Source { return SourceCode }
func (b *CodeBackend) Available() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

func (b *CodeBackend) Search(query string, limit int) ([]Result, error) {
	if !b.Available() {
		return nil, fmt.Errorf("code search needs gh CLI: brew install gh && gh auth login")
	}

	cmd := exec.Command("gh", "search", "code", query,
		"--limit", fmt.Sprint(limit),
		"--json", "path,repository,url,textMatches")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gh code search: %w", err)
	}

	var items []struct {
		Path       string `json:"path"`
		Repository struct {
			NameWithOwner string `json:"nameWithOwner"`
			URL           string `json:"url"`
		} `json:"repository"`
		URL         string `json:"url"`
		TextMatches []struct {
			Fragment string `json:"fragment"`
			Matches  []struct {
				Text string `json:"text"`
			} `json:"matches"`
		} `json:"textMatches"`
	}
	if err := json.Unmarshal(out, &items); err != nil {
		return nil, fmt.Errorf("gh code search parse: %w", err)
	}

	var results []Result
	for _, item := range items {
		repo := item.Repository.NameWithOwner
		snippet := ""
		if len(item.TextMatches) > 0 {
			snippet = item.TextMatches[0].Fragment
			// Clean up the snippet
			snippet = strings.ReplaceAll(snippet, "\n", " ")
			snippet = strings.Join(strings.Fields(snippet), " ")
			if len(snippet) > 200 {
				snippet = snippet[:200] + "..."
			}
		}
		if snippet == "" {
			snippet = fmt.Sprintf("Code match in %s", repo)
		}

		results = append(results, Result{
			Source:     SourceCode,
			Title:      fmt.Sprintf("%s: %s", repo, item.Path),
			URL:        item.URL,
			Snippet:    snippet,
			Engagement: fmt.Sprintf("📁 %s/%s", repo, item.Path),
		})
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("code search: no results")
	}
	return results, nil
}
