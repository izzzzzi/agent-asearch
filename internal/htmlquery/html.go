// Package htmlquery provides a zero-dependency HTML query library using only Go
// standard library packages. It supports finding elements by tag name and CSS class,
// extracting text content, and reading attribute values.
//
// This is NOT a full DOM parser like goquery — it scans HTML text with lightweight
// state tracking. It is designed for practical web scraping tasks (e.g. extracting
// search results) where full CSS selector support would be overkill.
//
// Limitations:
//   - Does not resolve character references inside attribute values (fine for
//     class/id/href extraction from typical HTML).
//   - Does not parse inline CSS (<style>) or JS (<script>) content for querying.
//     It does skip <script> and <style> bodies during nesting to avoid false
//     matches from "<" inside those elements.
//   - '<' or '>' inside quoted attribute values may confuse tag boundary detection.
//     This is extremely rare in real-world search engine result pages.
//   - No CSS combinators, pseudo-classes, or attribute selectors — only tag + class.
package htmlquery

import (
	"html"
	"strings"
)

// Element represents a matched HTML element.
type Element struct {
	Tag   string            // tag name (lowercase)
	HTML  string            // full outer HTML including the element's own tags
	Text  string            // extracted text content (tags stripped, entities decoded)
	Attrs map[string]string // attribute name → value (names lowered)
}

// Find returns child elements matching tag (and optional class).
// It searches within this element's inner HTML, not including the element itself.
func (e Element) Find(tag, class string) []Element {
	return FindAll(innerHTML(e.HTML, e.Tag), tag, class)
}

// FindAll returns all elements matching the given tag name.
// If class is non-empty, only elements whose class attribute contains that class
// are returned. Class matching is case-insensitive and handles multiple classes
// (e.g. "foo bar" matches class="bar").
//
// The function correctly handles:
//   - Nested elements of the same tag type (depth tracking)
//   - Self-closing tags (<br/>, <img src="..."/>)
//   - HTML void elements (<br>, <hr>, <img>, <input>, etc.)
//   - HTML comments (<!-- ... -->)
//   - <script> and <style> bodies (content is skipped during nesting)
func FindAll(h, tag, class string) []Element {
	tag = strings.ToLower(tag)
	var elems []Element
	pos := 0

	for {
		start := nextOpenTag(h, pos, tag)
		if start < 0 {
			break
		}

		// Locate the end of this opening tag.
		gt := findGT(h, start)
		if gt < 0 || gt < start {
			break
		}
		openEnd := gt + 1
		openTag := h[start:openEnd]

		// Self-closing or void — no matching close tag needed.
		if isSelfClose(openTag) || isVoid(tag) {
			if class == "" || hasClass(openTag, class) {
				elems = append(elems, Element{
					Tag: tag, HTML: openTag, Text: "",
					Attrs: parseAttrs(openTag),
				})
			}
			pos = openEnd
			continue
		}

		// Find matching close tag, handling nested tags of the same name.
		closeEnd := matchClose(h, openEnd, tag)
		if closeEnd < 0 {
			pos = openEnd
			continue
		}

		// Inner content excludes the outer opening and closing tags.
		// closeEnd points past the closing '>', so the close tag occupies
		// closeEnd-(len(tag)+3) .. closeEnd  (where +3 is for '</' and '>').
		inner := h[openEnd : closeEnd-len(tag)-3]

		if class == "" || hasClass(openTag, class) {
			elems = append(elems, Element{
				Tag: tag, HTML: h[start:closeEnd], Text: textContent(inner),
				Attrs: parseAttrs(openTag),
			})
		}

		// Advance past the opening tag so we can find subsequent siblings
		// (e.g. the second <li> after the first one finishes).
		pos = openEnd
	}
	return elems
}

// ---------------------------------------------------------------------------
// Tag scanning
// ---------------------------------------------------------------------------

// nextOpenTag finds the next '<tagName' occurrence that is:
//   - not a closing tag (</...)
//   - not a comment (<!--) or doctype (<!)
//   - not inside a <script> or <style> body
//   - not in the middle of a longer tag name
func nextOpenTag(h string, start int, tag string) int {
	for start < len(h) {
		lt := strings.IndexByte(h[start:], '<')
		if lt < 0 {
			return -1
		}
		i := start + lt
		if i+1 >= len(h) {
			return -1
		}

		c := h[i+1]

		// Handle comments and doctype declarations properly so their
		// contents (which may contain '<tagName' patterns) are skipped.
		if c == '!' {
			if hasPrefixAt(h, i, "<!--") {
				end := strings.Index(h[i:], "-->")
				if end < 0 {
					return -1
				}
				start = i + end + 3
				continue
			}
			// Doctype <!DOCTYPE ...> or other <!...> markup.
			gt := strings.IndexByte(h[i:], '>')
			if gt < 0 {
				return -1
			}
			start = i + gt + 1
			continue
		}

		switch {
		case c == '/':
			// Closing tag — skip.
			start = i + 2
			continue
		case !isLetter(c):
			// '<' followed by a non-letter is not a valid tag (e.g. '<  text').
			start = i + 2
			continue
		}

		// Read the full tag name.
		j := i + 1
		for j < len(h) && isIdent(h[j]) {
			j++
		}
		name := strings.ToLower(h[i+1 : j])
		if name == tag {
			return i
		}

		// Skip <script> and <style> bodies entirely so that '<' inside
		// JavaScript/CSS (e.g. comparisons, regex, strings) does not
		// produce false tag matches during subsequent scanning.
		if name == "script" || name == "style" {
			ct := "</" + name + ">"
			ci := strings.Index(strings.ToLower(h[i:]), ct)
			if ci < 0 {
				return -1
			}
			start = i + ci + len(ct)
			continue
		}

		start = j
	}
	return -1
}

// matchClose finds the position after the matching </tag> for an opening tag
// at position openStart. It correctly handles nested tags of the same name.
func matchClose(h string, pos int, tag string) int {
	depth := 1
	for pos < len(h) && depth > 0 {
		lt := strings.IndexByte(h[pos:], '<')
		if lt < 0 {
			return -1
		}
		pos += lt
		if pos+1 >= len(h) {
			return -1
		}

		switch {
		case h[pos+1] == '/':
			// Closing tag — read name and check.
			j := pos + 2
			for j < len(h) && isIdent(h[j]) {
				j++
			}
			if strings.ToLower(h[pos+2:j]) == tag {
				depth--
				if depth == 0 {
					gt := findGT(h, pos)
					if gt < 0 {
						return -1
					}
					return gt + 1
				}
			}
			pos = j

		case hasPrefixAt(h, pos, "<!--"):
			// HTML comment — skip to -->.
			end := strings.Index(h[pos:], "-->")
			if end < 0 {
				return -1
			}
			pos += end + 3

		case h[pos+1] == '!':
			// Other markup (doctype, CDATA) — skip to >.
			gt := findGT(h, pos)
			if gt < 0 {
				return -1
			}
			pos = gt + 1

		default:
			// Opening tag — read name.
			j := pos + 1
			for j < len(h) && isIdent(h[j]) {
				j++
			}
			name := strings.ToLower(h[pos+1 : j])

			// <script> and <style> have content that may contain '<', so
			// skip their entire body to avoid confusing the depth counter.
			if name == "script" || name == "style" {
				ct := "</" + name + ">"
				ci := strings.Index(strings.ToLower(h[pos:]), ct)
				if ci < 0 {
					return -1
				}
				pos += ci + len(ct)
				continue
			}

			gt := findGT(h, pos)
			if gt < 0 {
				return -1
			}

			// If this is another opening tag of our target type, increment depth
			// (unless it's self-closing or a void element).
			if name == tag {
				tagHTML := h[pos : gt+1]
				if !isSelfClose(tagHTML) && !isVoid(name) {
					depth++
				}
			}
			pos = gt + 1
		}
	}
	if depth == 0 {
		return pos
	}
	return -1
}

// findGT finds the first unquoted '>' starting from start.
// This prevents '>' inside quoted attribute values from being treated
// as the end of a tag.
func findGT(h string, start int) int {
	sq, dq := false, false
	for i := start; i < len(h); i++ {
		switch h[i] {
		case '\'':
			if !dq {
				sq = !sq
			}
		case '"':
			if !sq {
				dq = !dq
			}
		case '>':
			if !sq && !dq {
				return i
			}
		}
	}
	return -1
}

// hasPrefixAt reports whether h[i:] starts with prefix.
func hasPrefixAt(h string, i int, prefix string) bool {
	return i+len(prefix) <= len(h) && h[i:i+len(prefix)] == prefix
}

// ---------------------------------------------------------------------------
// Character classification
// ---------------------------------------------------------------------------

func isLetter(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z'
}

func isIdent(b byte) bool {
	return isLetter(b) || b >= '0' && b <= '9' || b == '-' || b == '_'
}

// ---------------------------------------------------------------------------
// Self-closing / void element detection
// ---------------------------------------------------------------------------

// isSelfClose reports whether an opening tag string ends with /> (possibly
// with whitespace before the slash).
func isSelfClose(s string) bool {
	s = strings.TrimRight(s, " \t\r\n")
	return len(s) >= 2 && s[len(s)-2] == '/'
}

// isVoid returns true for HTML void elements that never have a closing tag.
func isVoid(tag string) bool {
	switch tag {
	case "area", "base", "br", "col", "embed", "hr", "img", "input",
		"keygen", "link", "meta", "param", "source", "track", "wbr":
		return true
	}
	return false
}

// ---------------------------------------------------------------------------
// Class matching
// ---------------------------------------------------------------------------

// hasClass checks whether the opening tag HTML contains the given class name
// in its class attribute. Multiple classes are supported (e.g. class="foo bar"
// matches class="bar"). Matching is case-insensitive.
func hasClass(openTag, class string) bool {
	// Normalise to lower for case-insensitive matching.
	lower := strings.ToLower(openTag)

	// Try class="..."
	idx := strings.Index(lower, "class=\"")
	if idx < 0 {
		// Try class='...'
		idx = strings.Index(lower, "class='")
		if idx < 0 {
			return false
		}
		start := idx + 7 // len("class='")
		end := strings.IndexByte(openTag[start:], '\'')
		if end < 0 {
			return false
		}
		return classInList(openTag[start:start+end], class)
	}
	start := idx + 7 // len("class=\"")
	end := strings.IndexByte(openTag[start:], '"')
	if end < 0 {
		return false
	}
	return classInList(openTag[start:start+end], class)
}

// classInList reports whether class appears in a space-separated class list.
func classInList(list, class string) bool {
	for _, c := range strings.Fields(list) {
		if strings.EqualFold(c, class) {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Attribute parsing
// ---------------------------------------------------------------------------

// parseAttrs extracts all key-value pairs from an opening tag.
// Boolean attributes (no =value) are stored with an empty string value.
// Attribute names are lowercased; values are kept as-is.
// Does not decode HTML entities in attribute values.
func parseAttrs(openTag string) map[string]string {
	m := make(map[string]string)
	gt := strings.IndexByte(openTag, '>')
	if gt < 0 {
		return m
	}

	i := 0
	if openTag[0] == '<' {
		i = 1
	}
	// Skip past tag name.
	for i < gt && openTag[i] != ' ' && openTag[i] != '\t' && openTag[i] != '\n' {
		i++
	}

	for i < gt {
		// Skip whitespace before attribute.
		for i < gt && (openTag[i] == ' ' || openTag[i] == '\t' || openTag[i] == '\n') {
			i++
		}
		if i >= gt {
			break
		}

		// Read attribute name.
		ks := i
		for i < gt && openTag[i] != '=' && openTag[i] != ' ' &&
			openTag[i] != '\t' && openTag[i] != '\n' && openTag[i] != '>' {
			i++
		}
		key := openTag[ks:i]

		// No value — boolean attribute.
		if i >= gt || openTag[i] == '>' || openTag[i] == ' ' ||
			openTag[i] == '\t' || openTag[i] == '\n' {
			m[strings.ToLower(key)] = ""
			continue
		}

		// Skip '='.
		if openTag[i] == '=' {
			i++
		}
		// Skip whitespace before value.
		for i < gt && (openTag[i] == ' ' || openTag[i] == '\t' || openTag[i] == '\n') {
			i++
		}
		if i >= gt {
			m[strings.ToLower(key)] = ""
			break
		}

		// Quoted or unquoted value.
		if openTag[i] == '"' || openTag[i] == '\'' {
			quote := openTag[i]
			i++
			vs := i
			for i < gt && openTag[i] != quote {
				i++
			}
			m[strings.ToLower(key)] = openTag[vs:i]
			if i < gt {
				i++ // skip closing quote
			}
		} else {
			vs := i
			for i < gt && openTag[i] != ' ' && openTag[i] != '\t' &&
				openTag[i] != '\n' && openTag[i] != '>' {
				i++
			}
			m[strings.ToLower(key)] = openTag[vs:i]
		}
	}
	return m
}

// ---------------------------------------------------------------------------
// Text extraction
// ---------------------------------------------------------------------------

// textContent strips all HTML tags, decodes HTML entities, collapses
// whitespace, and trims the result. <script> and <style> bodies are
// excluded from the output.
func textContent(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '<' {
			// Read tag name.
			j := i + 1
			for j < len(s) && isIdent(s[j]) {
				j++
			}
			name := strings.ToLower(s[i+1 : j])

			gt := findGT(s, i)
			if gt < 0 {
				break
			}

			// Skip script/style bodies entirely.
			if name == "script" || name == "style" {
				ct := "</" + name + ">"
				ci := strings.Index(strings.ToLower(s[i:]), ct)
				if ci < 0 {
					break
				}
				i += ci + len(ct)
				continue
			}
			i = gt + 1
			continue
		}
		n := i
		for n < len(s) && s[n] != '<' {
			n++
		}
		b.WriteString(s[i:n])
		i = n
	}

	result := html.UnescapeString(b.String())
	parts := strings.Fields(result)
	return strings.Join(parts, " ")
}

// ---------------------------------------------------------------------------
// Inner HTML
// ---------------------------------------------------------------------------

// innerHTML returns the content between the outer opening and closing tags.
// It uses LastIndex to find the matching close tag, which works correctly
// when nested tags of the same type are properly balanced.
func innerHTML(h, tag string) string {
	gt := findGT(h, 0)
	if gt < 0 {
		return ""
	}
	ct := "</" + tag + ">"
	ci := strings.LastIndex(strings.ToLower(h), ct)
	if ci < 0 {
		return h[gt+1:]
	}
	return h[gt+1 : ci]
}
