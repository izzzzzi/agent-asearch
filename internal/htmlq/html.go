package htmlq

import (
	"regexp"
	"strings"
)

var voidElements = map[string]bool{
	"area": true, "base": true, "br": true, "col": true, "embed": true,
	"hr": true, "img": true, "input": true, "keygen": true, "link": true,
	"meta": true, "param": true, "source": true, "track": true, "wbr": true,
}

type Element struct {
	Tag   string
	Text  string
	Attrs map[string]string
	HTML  string
}

func FindAll(html, tag, class string) []Element {
	var elems []Element
	pos := 0
	for pos < len(html) {
		tagStart := strings.Index(html[pos:], "<")
		if tagStart == -1 { break }
		tagStart += pos
		tagEnd := findGT(html, tagStart+1)
		if tagEnd == -1 { break }
		fullTag := html[tagStart+1 : tagEnd]
		pos = tagEnd + 1

		if strings.HasPrefix(fullTag, "!--") {
			if end := strings.Index(fullTag, "-->"); end >= 0 {
				pos = tagStart + end + 3
			}
			continue
		}
		if strings.HasPrefix(fullTag, "/") || strings.HasPrefix(fullTag, "!") {
			continue
		}

		tagName := extractTagName(fullTag)
		if tagName == "" { continue }
		if tag != "" && !strings.EqualFold(tagName, tag) { continue }

		attrs := parseAttrs(fullTag)
		if class != "" && !hasClass(attrs, class) { continue }

		if isSelfClose(fullTag) || voidElements[strings.ToLower(tagName)] {
			elems = append(elems, Element{
				Tag: tagName, Text: "", Attrs: attrs,
				HTML: html[tagStart : tagEnd+1],
			})
			continue
		}

		closeIdx := matchClose(html, pos, tagName)
		if closeIdx == -1 { break }
		inner := html[pos:closeIdx]
		closeLen := len(tagName) + 3 // </ + tag + >
		elems = append(elems, Element{
			Tag: tagName, Text: textContent(inner), Attrs: attrs,
			HTML: html[tagStart : closeIdx+closeLen],
		})
		pos = closeIdx + closeLen
	}
	return elems
}

func (e Element) Find(tag, class string) []Element { return FindAll(e.HTML, tag, class) }

// --- Private ---

func findGT(html string, start int) int {
	sq, dq := false, false
	for i := start; i < len(html); i++ {
		switch html[i] {
		case '\'': if !dq { sq = !sq }
		case '"': if !sq { dq = !dq }
		case '>': if !sq && !dq { return i }
		}
	}
	return -1
}

func extractTagName(tag string) string {
	end := strings.IndexAny(tag, " />\t\n")
	if end == -1 { end = len(tag) }
	return tag[:end]
}

func isSelfClose(tag string) bool {
	t := strings.TrimSpace(tag)
	return len(t) > 0 && t[len(t)-1] == '/'
}

var attrPat = regexp.MustCompile(`\s([a-zA-Z][a-zA-Z0-9_-]*)\s*=\s*"([^"]*)"`)
var attrPat2 = regexp.MustCompile(`\s([a-zA-Z][a-zA-Z0-9_-]*)\s*=\s*'([^']*)'`)

func parseAttrs(tag string) map[string]string {
	attrs := make(map[string]string)
	for _, m := range attrPat.FindAllStringSubmatch(tag, -1) {
		attrs[m[1]] = m[2]
	}
	for _, m := range attrPat2.FindAllStringSubmatch(tag, -1) {
		attrs[m[1]] = m[2]
	}
	return attrs
}

func hasClass(attrs map[string]string, class string) bool {
	c, ok := attrs["class"]
	if !ok { return false }
	for _, cl := range strings.Fields(c) {
		if strings.EqualFold(cl, class) { return true }
	}
	return false
}

func matchClose(html string, start int, tag string) int {
	depth := 0
	pos := start
	for pos < len(html) {
		nextOpen := strings.Index(html[pos:], "<")
		if nextOpen == -1 { return -1 }
		nextOpen += pos
		gt := findGT(html, nextOpen+1)
		if gt == -1 { return -1 }
		inner := html[nextOpen+1 : gt]
		nameEnd := strings.IndexAny(inner, " />\t\n")
		if nameEnd == -1 { nameEnd = len(inner) }
		name := inner[:nameEnd]
		if strings.HasPrefix(inner, "/") && strings.TrimSpace(inner[1:]) == tag {
			if depth == 0 { return nextOpen }
			depth--
		} else if !strings.HasPrefix(inner, "/") && !strings.HasPrefix(inner, "!") &&
			!isSelfClose(inner) && !voidElements[strings.ToLower(name)] &&
			name == tag {
			depth++
		}
		pos = gt + 1
	}
	return -1
}

func textContent(html string) string {
	var b strings.Builder
	inTag := false
	for pos := 0; pos < len(html); pos++ {
		if html[pos] == '<' {
			// HTML comment
			if pos+3 < len(html) && html[pos:pos+4] == "<!--" {
				end := strings.Index(html[pos+4:], "-->")
				if end >= 0 { pos = pos + 4 + end + 2; continue }
			}
			// Script/style blocks
			low := strings.ToLower(html[pos:])
			if strings.HasPrefix(low, "<script") || strings.HasPrefix(low, "<style") {
			if strings.HasPrefix(html[pos:], "<script") || strings.HasPrefix(html[pos:], "<style") {
				closeLow := "</script"
				if strings.HasPrefix(html[pos:], "<style") { closeLow = "</style" }
				ci := strings.Index(strings.ToLower(html[pos:]), closeLow)
				if ci >= 0 { pos = pos + ci + len(closeLow) - 1; continue }
			}
			}
			inTag = true
			continue
		}
		if html[pos] == '>' && inTag {
			inTag = false
			if pos+1 < len(html) && html[pos+1] != '<' && html[pos+1] != ' ' && html[pos+1] != '\n' {
				b.WriteByte(' ')
			}
			continue
		}
		if !inTag {
			b.WriteByte(html[pos])
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
