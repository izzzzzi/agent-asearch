package htmlq

import (
	"testing"
)

func TestFindAll(t *testing.T) {
	html := `<div class="result"><a href="https://ex.com">Title</a><span class="s">Desc</span></div>`
	results := FindAll(html, "div", "result")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Text != "Title Desc" {
		t.Errorf("expected 'Title Desc', got '%s'", results[0].Text)
	}
}

func TestFindClass(t *testing.T) {
	html := `<div class="a b c">text</div>`
	results := FindAll(html, "div", "b")
	if len(results) != 1 {
		t.Fatalf("expected 1, got %d", len(results))
	}
}

func TestNested(t *testing.T) {
	html := `<div class="r"><div class="s"><span>nested</span></div></div>`
	results := FindAll(html, "div", "s")
	if len(results) != 1 {
		t.Fatalf("expected 1 nested, got %d", len(results))
	}
	if results[0].Text != "nested" {
		t.Errorf("expected 'nested', got '%s'", results[0].Text)
	}
}

func TestSelfClose(t *testing.T) {
	html := `<div class="x"><br><img src="x.jpg"><span>text</span></div>`
	results := FindAll(html, "div", "x")
	if len(results) != 1 {
		t.Fatalf("expected 1, got %d", len(results))
	}
}

func TestFindNested(t *testing.T) {
	html := `<div class="r"><a href="http://x.com">Link</a></div>`
	divs := FindAll(html, "div", "r")
	if len(divs) == 0 {
		t.Fatal("no div")
	}
	links := divs[0].Find("a", "")
	if len(links) == 0 {
		t.Fatal("no link")
	}
	if links[0].Text != "Link" {
		t.Errorf("expected 'Link', got '%s'", links[0].Text)
	}
	if links[0].Attrs["href"] != "http://x.com" {
		t.Errorf("wrong href: %s", links[0].Attrs["href"])
	}
}

func TestComment(t *testing.T) {
	html := `<div><!-- comment -->visible</div>`
	results := FindAll(html, "div", "")
	if len(results) != 1 {
		t.Fatal("expected 1")
	}
	if results[0].Text != "visible" {
		t.Errorf("expected 'visible', got '%s'", results[0].Text)
	}
}

func TestNoMatch(t *testing.T) {
	html := `<div class="x">content</div>`
	results := FindAll(html, "span", "")
	if len(results) != 0 {
		t.Errorf("expected 0, got %d", len(results))
	}
	results2 := FindAll(html, "div", "y")
	if len(results2) != 0 {
		t.Errorf("expected 0, got %d", len(results2))
	}
}
