package search

import (
	"testing"
)

func TestCrossRefQueryParsing(t *testing.T) {
	tests := []struct {
		query string
		n     int
		first string
	}{
		{"claude code by anthropic", 2, "claude code"},
		{"cursor by cursor.sh", 2, "cursor"},
		{"bolt.new from stackblitz", 2, "bolt.new"},
		{"rust programming", 1, "rust programming"},
		{"golang cli | cobra", 2, "golang cli"},
	}
	for _, tt := range tests {
		parts := splitQuery(tt.query)
		if len(parts) != tt.n {
			t.Errorf("splitQuery(%q) = %d parts, want %d", tt.query, len(parts), tt.n)
		}
		if tt.n > 0 && len(parts) > 0 && parts[0] != tt.first {
			t.Errorf("splitQuery(%q)[0] = %q, want %q", tt.query, parts[0], tt.first)
		}
	}
}

func TestCrossRefNoop(t *testing.T) {
	// Without a "by"/"from" separator, cross-ref does nothing
	results := []Result{
		{Title: "repo", URL: "https://github.com/a/b"},
		{Title: "post", URL: "https://news.ycombinator.com/item?id=1"},
	}
	got := crossReference(results, "query without separator")
	if len(got) != 2 {
		t.Errorf("crossReference changed count: %d", len(got))
	}
	if len(got[0].Related) != 0 {
		t.Errorf("expected empty Related, got %v", got[0].Related)
	}
}

func TestCrossRefWithMatchingRawMeta(t *testing.T) {
	// Two results with the same entity via RawMeta should be linked
	results := []Result{
		{
			Title: "my-tool",
			URL:   "https://github.com/owner/my-tool",
			RawMeta: map[string]any{
				"full_name": "owner/my-tool",
			},
		},
		{
			Title: "Show HN: my-tool",
			URL:   "https://news.ycombinator.com/item?id=123",
			RawMeta: map[string]any{
				"full_name": "owner/my-tool",
			},
		},
	}
	got := crossReference(results, "my-tool by owner")
	if len(got) != 2 {
		t.Errorf("crossReference changed count: %d", len(got))
	}
	// Both should have Related set since they share entity
	if len(got[0].Related) == 0 {
		t.Error("result 0 should have Related (shares entity with result 1)")
	}
	if len(got[1].Related) == 0 {
		t.Error("result 1 should have Related (shares entity with result 0)")
	}
}
