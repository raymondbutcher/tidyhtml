package tidyhtml

import (
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

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

func isBlankText(n *html.Node) bool {
	if n != nil && n.Type == html.TextNode {
		if strings.IndexFunc(n.Data, isNotSpace) == -1 {
			return true
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
