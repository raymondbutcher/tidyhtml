package tidyhtml

import (
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

// getPrevElement gets the previous sibling if it is an element,
// or returns nil. It skips past blank text nodes during this check.
func getPrevElement(n *html.Node) *html.Node {
	for n != nil {
		n = n.PrevSibling
		if isBlankText(n) {
			continue
		}
		if n == nil || n.Type != html.ElementNode {
			return nil
		}
		return n
	}
	return nil
}

func isNotSpace(r rune) bool {
	return !unicode.IsSpace(r)
}

func hasChild(n *html.Node) bool {
	return n.FirstChild != nil
}

func hasText(n *html.Node) bool {
	if n.Type == html.TextNode {
		if strings.IndexFunc(n.Data, isNotSpace) != -1 {
			return true
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			if strings.IndexFunc(c.Data, isNotSpace) != -1 {
				return true
			}
		}
	}
	return false
}

func isBlankText(n *html.Node) bool {
	if n != nil && n.Type == html.TextNode {
		if strings.IndexFunc(n.Data, isNotSpace) == -1 {
			return true
		}
	}
	return false
}

func isPreNode(n *html.Node) bool {
	return n != nil && n.Type == html.ElementNode && n.Data == "pre"
}

func isTextBlock(n *html.Node) bool {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			if strings.IndexFunc(c.Data, isNotSpace) >= 0 {
				return true
			}
		}
	}
	return false
}

func isVoid(n *html.Node) bool {
	return n.Type == html.ElementNode && voidElements[n.Data]
}

// Section 12.1.2, "Elements", gives this list of void elements. Void elements
// are those that can't have any contents.
// From https://github.com/golang/net/blob/master/html/render.go
var voidElements = map[string]bool{
	"area":    true,
	"base":    true,
	"br":      true,
	"col":     true,
	"command": true,
	"embed":   true,
	"hr":      true,
	"img":     true,
	"input":   true,
	"keygen":  true,
	"link":    true,
	"meta":    true,
	"param":   true,
	"source":  true,
	"track":   true,
	"wbr":     true,
}
