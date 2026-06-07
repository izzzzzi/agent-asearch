package search

import (
	"testing"

	"github.com/izzzzzi/agent-asearch/internal/config"
)

func TestBackendHN(t *testing.T) {
	b := &HNBackend{}
	if !b.Available() {
		t.Skip("HN backend should always be available")
	}
	results, err := b.Search("rust programming", 3)
	if err != nil {
		t.Fatalf("HN search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("HN search returned 0 results")
	}
	if results[0].Title == "" {
		t.Error("result missing title")
	}
	if results[0].URL == "" {
		t.Error("result missing URL")
	}
}

func TestBackendRedditPublic(t *testing.T) {
	results, err := redditPublicJSON("test", 2)
	if err != nil {
		t.Skipf("Reddit public JSON unavailable: %v", err)
	}
	if len(results) > 0 && results[0].Title == "" {
		t.Error("result missing title")
	}
}

func TestBackendGitHub(t *testing.T) {
	b := &GitHubBackend{}
	if !b.Available() {
		t.Skip("gh CLI not available")
	}
	results, err := b.Search("golang cli", 2)
	if err != nil {
		t.Fatalf("GitHub search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("GitHub search returned 0 results")
	}
	if results[0].Title == "" {
		t.Error("result missing title")
	}
}

func TestBackendWikipedia(t *testing.T) {
	results := wikiSearch("rust programming language", 2)
	if results == nil {
		t.Fatal("Wikipedia search returned nil")
	}
	if len(results) == 0 {
		t.Fatal("Wikipedia search returned 0 results")
	}
	if results[0].Title == "" {
		t.Error("result missing title")
	}
	if results[0].URL == "" {
		t.Error("result missing URL")
	}
}

func TestBackendDDG(t *testing.T) {
	results := ddgSearch("golang programming", 2)
	if results == nil {
		// DDG might be blocked in CI
		t.Skip("DDG search returned nil (possibly blocked)")
	}
	if len(results) == 0 {
		t.Skip("DDG search returned 0 results (possibly blocked)")
	}
	if results[0].Title == "" {
		t.Error("result missing title")
	}
}

func TestBackendDedup(t *testing.T) {
	// Test that normalizeURL works for YouTube
	u1 := "https://www.youtube.com/watch?v=abc123&feature=shared"
	u2 := "https://www.youtube.com/watch?v=abc123"
	n1 := normalizeURL(u1)
	n2 := normalizeURL(u2)
	if n1 != n2 {
		t.Errorf("normalizeURL should dedup YouTube URLs: %s != %s", n1, n2)
	}

	// Test that normalizeURL works for Reddit
	r1 := "https://www.reddit.com/r/test/comments/123"
	r2 := "https://www.reddit.com/r/test/comments/123/?utm_source=share"
	if normalizeURL(r1) != normalizeURL(r2) {
		t.Error("normalizeURL should dedup Reddit URLs")
	}

	// Test deduplicate
	results := []Result{
		{URL: "https://example.com/a"},
		{URL: "https://www.example.com/a/"},
		{URL: "https://example.com/b"},
	}
	deduped := deduplicate(results)
	if len(deduped) != 2 {
		t.Errorf("dedup should remove 1 duplicate, got %d", len(deduped))
	}
}

func TestConfigKey(t *testing.T) {
	key := config.GetKey("tavily")
	// Just check it doesn't crash — key may be "" or set
	_ = key
	key2 := config.GetKey("nonexistent")
	if key2 != "" {
		t.Error("nonexistent key should return empty string")
	}
}
