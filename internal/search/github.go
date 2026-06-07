package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

type GitHubBackend struct{}

func (b *GitHubBackend) Name() Source { return SourceGitHub }
func (b *GitHubBackend) Available() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

func (b *GitHubBackend) Search(query string, limit int) ([]Result, error) {
	if b.Available() {
		results, err := ghSearch(query, limit)
		if err == nil && len(results) > 0 {
			return results, nil
		}
	}
	return githubAPISearch(query, limit)
}

func ghSearch(query string, limit int) ([]Result, error) {
	cmd := exec.Command("gh", "search", "repos", query,
		"--sort", "stars", "--limit", fmt.Sprint(limit), "--json",
		"name,fullName,owner,url,description,stargazersCount,language,updatedAt")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gh search failed: %w", err)
	}

	var data []struct {
		Name            string                 `json:"name"`
		FullName        string                 `json:"fullName"`
		Owner           struct{ Login string } `json:"owner"`
		URL             string                 `json:"url"`
		Description     string                 `json:"description"`
		StargazersCount int                    `json:"stargazersCount"`
		Language        string                 `json:"language"`
		UpdatedAt       string                 `json:"updatedAt"`
	}
	if err := json.Unmarshal(out, &data); err != nil {
		return nil, err
	}

	var results []Result
	for _, r := range data {
		snippet := r.Description
		if snippet == "" {
			snippet = fmt.Sprintf("GitHub repository by %s", r.Owner.Login)
		}
		lang := r.Language
		if lang == "" {
			lang = "—"
		}
		date := r.UpdatedAt
		if len(date) > 10 {
			date = date[:10]
		}
		engagement := fmt.Sprintf("★%d | %s | %s", r.StargazersCount, lang, date)
		results = append(results, Result{
			Source:     SourceGitHub,
			Title:      r.FullName,
			URL:        r.URL,
			Snippet:    snippet,
			Date:       date,
			Score:      float64(r.StargazersCount),
			Engagement: engagement,
		})
	}
	return results, nil
}

func githubAPISearch(query string, limit int) ([]Result, error) {
	u := fmt.Sprintf("https://api.github.com/search/repositories?q=%s&sort=stars&per_page=%d",
		url.QueryEscape(query), limit)
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "asearch/1.0")

	if token := githubToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fallbackGitHub(query), nil
	}
	defer resp.Body.Close()

	var apiResp struct {
		Items []struct {
			FullName    string `json:"full_name"`
			HTMLURL     string `json:"html_url"`
			Description string `json:"description"`
			Stars       int    `json:"stargazers_count"`
			Language    string `json:"language"`
			UpdatedAt   string `json:"updated_at"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fallbackGitHub(query), nil
	}

	var results []Result
	for _, item := range apiResp.Items {
		snippet := item.Description
		if len(snippet) > 200 {
			snippet = snippet[:200]
		}
		lang := item.Language
		if lang == "" {
			lang = "—"
		}
		date := item.UpdatedAt
		if len(date) > 10 {
			date = date[:10]
		}
		engagement := fmt.Sprintf("★%d | %s | %s", item.Stars, lang, date)
		results = append(results, Result{
			Source:     SourceGitHub,
			Title:      item.FullName,
			URL:        item.HTMLURL,
			Snippet:    snippet,
			Date:       date,
			Score:      float64(item.Stars),
			Engagement: engagement,
		})
	}
	return results, nil
}

func githubToken() string {
	for _, env := range []string{"GITHUB_TOKEN", "GH_TOKEN"} {
		if v := strings.TrimSpace(os.Getenv(env)); v != "" {
			return v
		}
	}
	return ""
}

func fallbackGitHub(query string) []Result {
	return []Result{{
		Source:  SourceGitHub,
		Title:   fmt.Sprintf("GitHub search: %s", query),
		URL:     "https://github.com/search?q=" + url.QueryEscape(query),
		Snippet: "需要 gh CLI 或 GITHUB_TOKEN 获取完整结果。Install: brew install gh",
	}}
}
