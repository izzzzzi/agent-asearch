package htmlquery

import (
	"testing"
)

func TestFindAll_ByTag(t *testing.T) {
	html := `<html><body><div>one</div><div>two</div><span>three</span></body></html>`
	elems := FindAll(html, "div", "")
	if len(elems) != 2 {
		t.Fatalf("expected 2 divs, got %d", len(elems))
	}
	if elems[0].Text != "one" {
		t.Errorf("expected 'one', got %q", elems[0].Text)
	}
	if elems[1].Text != "two" {
		t.Errorf("expected 'two', got %q", elems[1].Text)
	}
}

func TestFindAll_ByTagAndClass(t *testing.T) {
	html := `<div class="result">A</div><div class="result">B</div><div class="other">C</div>`
	elems := FindAll(html, "div", "result")
	if len(elems) != 2 {
		t.Fatalf("expected 2 result divs, got %d", len(elems))
	}
	if elems[0].Text != "A" {
		t.Errorf("expected 'A', got %q", elems[0].Text)
	}
	if elems[1].Text != "B" {
		t.Errorf("expected 'B', got %q", elems[1].Text)
	}
}

func TestFindAll_ClassMultiple(t *testing.T) {
	html := `<div class="foo bar">content</div>`
	elems := FindAll(html, "div", "bar")
	if len(elems) != 1 {
		t.Fatalf("expected 1 div, got %d", len(elems))
	}
	if elems[0].Text != "content" {
		t.Errorf("expected 'content', got %q", elems[0].Text)
	}
}

func TestFindAll_ClassCaseInsensitive(t *testing.T) {
	html := `<div class="Result">content</div>`
	elems := FindAll(html, "div", "result")
	if len(elems) != 1 {
		t.Fatalf("expected 1 div, got %d", len(elems))
	}
}

func TestFindAll_NestedSameTag(t *testing.T) {
	html := `<div><div>inner</div> outer </div>`
	elems := FindAll(html, "div", "")
	if len(elems) != 2 {
		t.Fatalf("expected 2 divs, got %d", len(elems))
	}
	// Outer div should contain "inner outer"
	if elems[0].Text != "inner outer" {
		t.Errorf("expected 'inner outer', got %q", elems[0].Text)
	}
	// Inner div should be just "inner"
	if elems[1].Text != "inner" {
		t.Errorf("expected 'inner', got %q", elems[1].Text)
	}
}

func TestFindAll_Attr(t *testing.T) {
	html := `<a href="https://example.com">click</a>`
	elems := FindAll(html, "a", "")
	if len(elems) != 1 {
		t.Fatalf("expected 1 anchor, got %d", len(elems))
	}
	if elems[0].Attrs["href"] != "https://example.com" {
		t.Errorf("expected 'https://example.com', got %q", elems[0].Attrs["href"])
	}
}

func TestFindAll_SelfClosing(t *testing.T) {
	html := `<br/><br><img src="x.png"/>`
	elems := FindAll(html, "br", "")
	if len(elems) != 2 {
		t.Fatalf("expected 2 br elements, got %d", len(elems))
	}
	imgs := FindAll(html, "img", "")
	if len(imgs) != 1 {
		t.Fatalf("expected 1 img, got %d", len(imgs))
	}
	if imgs[0].Attrs["src"] != "x.png" {
		t.Errorf("expected 'x.png', got %q", imgs[0].Attrs["src"])
	}
}

func TestFindAll_VoidElement(t *testing.T) {
	html := `<img src="a.jpg"><img src="b.jpg">`
	elems := FindAll(html, "img", "")
	if len(elems) != 2 {
		t.Fatalf("expected 2 imgs, got %d", len(elems))
	}
}

func TestFindAll_NoMatch(t *testing.T) {
	html := `<div>hello</div>`
	elems := FindAll(html, "span", "")
	if len(elems) != 0 {
		t.Fatalf("expected 0 spans, got %d", len(elems))
	}
}

func TestElement_Find(t *testing.T) {
	html := `<div class="results"><a href="1">first</a><a href="2">second</a></div>`
	outer := FindAll(html, "div", "results")
	if len(outer) != 1 {
		t.Fatalf("expected 1 outer div, got %d", len(outer))
	}
	links := outer[0].Find("a", "")
	if len(links) != 2 {
		t.Fatalf("expected 2 links inside div, got %d", len(links))
	}
	if links[0].Attrs["href"] != "1" || links[1].Attrs["href"] != "2" {
		t.Errorf("unexpected href values: %q, %q", links[0].Attrs["href"], links[1].Attrs["href"])
	}
}

func TestTextContent_StripsTags(t *testing.T) {
	html := `<p>Hello <b>World</b></p>`
	elems := FindAll(html, "p", "")
	if len(elems) != 1 {
		t.Fatalf("expected 1 p, got %d", len(elems))
	}
	if elems[0].Text != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", elems[0].Text)
	}
}

func TestTextContent_SkipsScript(t *testing.T) {
	// '<div>' inside the script body must NOT be matched as a real element.
	html := `<div>visible <script>var x = "<div>";</script> text</div>`
	elems := FindAll(html, "div", "")
	if len(elems) != 1 {
		t.Fatalf("expected 1 div, got %d", len(elems))
	}
	if elems[0].Text != "visible text" {
		t.Errorf("expected 'visible text', got %q", elems[0].Text)
	}
}

func TestTextContent_DecodesEntities(t *testing.T) {
	html := `<div>AT&amp;T &lt;foo&gt;</div>`
	elems := FindAll(html, "div", "")
	if len(elems) != 1 {
		t.Fatalf("expected 1 div, got %d", len(elems))
	}
	if elems[0].Text != "AT&T <foo>" {
		t.Errorf("expected 'AT&T <foo>', got %q", elems[0].Text)
	}
}

func TestFindAll_DeeplyNested(t *testing.T) {
	html := `<ul><li>a</li><li>b<ul><li>c</li></ul></li><li>d</li></ul>`
	items := FindAll(html, "li", "")
	if len(items) != 4 {
		t.Fatalf("expected 4 li elements, got %d", len(items))
	}
	if items[0].Text != "a" {
		t.Errorf("expected 'a', got %q", items[0].Text)
	}
	if items[3].Text != "d" {
		t.Errorf("expected 'd', got %q", items[3].Text)
	}
}

func TestParseAttrs_Multiple(t *testing.T) {
	html := `<a href="/path" class="link active" id="main">text</a>`
	elems := FindAll(html, "a", "")
	if len(elems) != 1 {
		t.Fatalf("expected 1 anchor, got %d", len(elems))
	}
	m := elems[0].Attrs
	if m["href"] != "/path" {
		t.Errorf("unexpected href: %q", m["href"])
	}
	if m["class"] != "link active" {
		t.Errorf("unexpected class: %q", m["class"])
	}
	if m["id"] != "main" {
		t.Errorf("unexpected id: %q", m["id"])
	}
}

func TestFindAll_CommentIgnored(t *testing.T) {
	html := `<div><!-- comment -->text</div>`
	elems := FindAll(html, "div", "")
	if len(elems) != 1 {
		t.Fatalf("expected 1 div, got %d", len(elems))
	}
	if elems[0].Text != "text" {
		t.Errorf("expected 'text', got %q", elems[0].Text)
	}
}

func TestFindAll_UppercaseTag(t *testing.T) {
	html := `<DIV>hello</DIV>`
	elems := FindAll(html, "div", "")
	if len(elems) != 1 {
		t.Fatalf("expected 1 div, got %d", len(elems))
	}
}

func TestFindAll_HasClass_SingleQuotes(t *testing.T) {
	html := `<div class='result'>data</div>`
	elems := FindAll(html, "div", "result")
	if len(elems) != 1 {
		t.Fatalf("expected 1 div, got %d", len(elems))
	}
}

func TestFindAll_EmptyHTML(t *testing.T) {
	elems := FindAll("", "div", "")
	if len(elems) != 0 {
		t.Errorf("expected 0, got %d", len(elems))
	}
}

func TestFindAll_NoClosingTag(t *testing.T) {
	// Should not panic or loop forever.
	html := `<div>unclosed`
	elems := FindAll(html, "div", "")
	if len(elems) != 0 {
		t.Errorf("expected 0, got %d", len(elems))
	}
}

func BenchmarkFindAll(b *testing.B) {
	html := `<div class="result">first</div><div class="result">second</div><div>other</div>`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FindAll(html, "div", "result")
	}
}
