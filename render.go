package tidyhtml

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

type tidy struct {

	// The current indentation level.
	indent int

	// The indentation level where a pre block starts. A value of -1 means
	// not currently in a pre block. A pre block is started by a <pre> element,
	// which should not be tidied at all because its whitespace is meaningful.
	preBlock int

	// The indentation level where a text block starts. A value of -1 means
	// not currently in a text block. A text block is a node that contains
	// a child node with actual text, not counting blank text nodes.
	textBlock int

	err error
}

func newTidy() tidy {
	return tidy{
		indent:    0,
		preBlock:  -1,
		textBlock: -1,
		err:       nil,
	}
}

// inNormalBlock - is the current node a regular non-pre non-text block?
func (t *tidy) inNormalBlock() bool {
	if t.inPreBlock() || t.inTextBlock() {
		return false
	}
	return true
}

// inPreBlock - is the current node within a pre block?
func (t *tidy) inPreBlock() bool {
	return t.preBlock != -1
}

// inTextBlock - is the current node within a text block?
func (t *tidy) inTextBlock() bool {
	return t.textBlock != -1
}

// isTextBlock - is the current node the start of the text block?
func (t *tidy) isTextBlock() bool {
	return t.textBlock == t.indent
}

// Render the node and all related nodes to HTML.
func (t *tidy) render(n *html.Node) (out []byte, err error) {

	buf := bytes.Buffer{}
	w := bufio.NewWriter(&buf)

	// Throw away the document node as it gets in the way.
	if n.Type == html.DocumentNode {
		n = n.FirstChild
		for s := n; s != nil; s = s.NextSibling {
			s.Parent = nil
		}
	}

	for n != nil {

		// Remove blank text nodes when not in a text/pre block.
		if t.inNormalBlock() {
			for s := n.NextSibling; isBlankText(s); s = n.NextSibling {
				n.NextSibling = s.NextSibling
			}
		}

		switch n.Type {
		case html.ElementNode:

			switch n.Data {
			case "noscript":
				// The <noscript> elements are parsed as plain text.
				// Convert them into HTML nodes so they can be tidied.
				t.err = parseTextNode(n)
			case "pre":
				if !t.inPreBlock() {
					t.preBlock = t.indent
				}
			}

			// Start a new text block?
			if t.inNormalBlock() && isTextBlock(n) {
				t.textBlock = t.indent
			}

			// Write the start of the element.
			t.writeEl(w, n)

			// Descend into children nodes.
			if n.FirstChild != nil {
				n = n.FirstChild
				t.indent++
				continue
			}

			// If there were no children, then close the element here.
			t.writeElClose(w, n)

		case html.TextNode:
			t.writeText(w, n)

		case html.CommentNode:
			t.writeComment(w, n)

		case html.DoctypeNode:
			t.writeDoctype(w, n)

		case html.DocumentNode:
			t.err = errors.New("tidyhtml: cannot render a DocumentNode node")

		case html.ErrorNode:
			t.err = errors.New("tidyhtml: cannot render an ErrorNode node")

		default:
			t.err = fmt.Errorf("tidyhtml: unknown node type: %v", n.Type)
		}

		if t.err != nil {
			err = t.err
			return
		}

		for n != nil {
			// Move across to the next sibling if there is one.
			if n.NextSibling != nil {
				n = n.NextSibling
				break
			}

			// Move upwards to the parent.
			n = n.Parent
			t.indent--
			if n != nil && n.Type == html.ElementNode {
				t.writeElClose(w, n)
			}
			if t.indent == t.textBlock {
				t.textBlock = -1
			}
			if t.indent == t.preBlock {
				t.preBlock = -1
			}
		}
	}

	if t.err != nil {
		err = t.err
		return
	}

	err = w.Flush()
	return buf.Bytes(), err
}

// Lower level functions for writing to the output:

func (t *tidy) write(w *bufio.Writer, p []byte) {
	if t.err == nil {
		_, t.err = w.Write(p)
	}
}

func (t *tidy) writeByte(w *bufio.Writer, c byte) {
	if t.err == nil {
		t.err = w.WriteByte(c)
	}
}

func (t *tidy) writeString(w *bufio.Writer, s string) {
	if t.err == nil {
		_, t.err = w.WriteString(s)
	}
}

// writeQuoted writes s to w surrounded by quotes. Normally it will use double
// quotes, but if s contains a double quote, it will use single quotes.
// It is used for writing the identifiers in a doctype declaration.
// In valid HTML, they can't contain both types of quotes.
// From https://github.com/golang/net/blob/master/html/render.go
func (t *tidy) writeQuoted(w *bufio.Writer, s string) {
	var q byte
	if strings.Contains(s, `"`) {
		q = '\''
	} else {
		q = '"'
	}
	t.writeByte(w, q)
	t.writeString(w, s)
	t.writeByte(w, q)
}

// writeIndentation adds spaces for indentation.
func (t *tidy) writeIndentation(w *bufio.Writer) {
	for i := 0; i < t.indent; i++ {
		t.writeString(w, "    ")
	}
}

// writeIndentationGuide adds a comment to help follow the level of
// indentation for <pre> tags, which have to be written without any.
func (t *tidy) writeIndentationGuide(w *bufio.Writer, guide string) {
	if t.indent >= 2 {
		t.writeString(w, "<!--")
		for i := 1; i < t.indent; i++ {
			t.writeString(w, guide)
		}
		t.writeString(w, " -->")
	}
}

// Functions for writing HTML nodes:

func (t *tidy) writeComment(w *bufio.Writer, n *html.Node) {
	if !isVeryFirstNode(n) && t.inNormalBlock() {
		t.writeIndentation(w)
	}

	t.writeString(w, "<!--")
	t.writeString(w, n.Data)
	t.writeString(w, "-->")

	if !isVeryLastNode(n) && t.inNormalBlock() {
		t.writeByte(w, '\n')
	}
}

func (t *tidy) writeDoctype(w *bufio.Writer, n *html.Node) {
	t.writeString(w, "<!doctype ")
	t.writeString(w, n.Data)
	if n.Attr != nil {
		var p, s string
		for _, a := range n.Attr {
			switch a.Key {
			case "public":
				p = a.Val
			case "system":
				s = a.Val
			}
		}
		if p != "" {
			t.writeString(w, " public ")
			t.writeQuoted(w, p)
			if s != "" {
				t.writeString(w, " ")
				t.writeQuoted(w, s)
			}
		} else if s != "" {
			t.writeString(w, " system ")
			t.writeQuoted(w, s)
		}
	}
	t.writeString(w, ">\n")
}

func (t *tidy) writeEl(w *bufio.Writer, n *html.Node) {

	if !isVeryFirstNode(n) {
		if n.Data == "pre" {
			if !isPreNode(getPrevElement(n)) {
				t.writeIndentationGuide(w, " <==")
				t.writeByte(w, '\n')
			}
		} else if !t.inPreBlock() && (!t.inTextBlock() || t.isTextBlock()) {
			t.writeIndentation(w)
		}
	}

	t.writeByte(w, '<')
	t.writeString(w, n.Data)
	for _, a := range n.Attr {
		t.writeByte(w, ' ')
		if a.Namespace != "" {
			t.writeString(w, a.Namespace)
			t.writeByte(w, ':')
		}
		t.writeString(w, a.Key)
		t.writeByte(w, '=')
		t.writeQuoted(w, html.EscapeString(a.Val))
	}
	t.writeByte(w, '>')

	if t.inNormalBlock() && hasChild(n) {
		t.writeByte(w, '\n')
	}
}

func (t *tidy) writeElClose(w *bufio.Writer, n *html.Node) {
	if t.inNormalBlock() && hasChild(n) {
		t.writeIndentation(w)
	}

	if !isVoid(n) {
		t.writeString(w, "</")
		t.writeString(w, n.Data)
		t.writeByte(w, '>')
	}

	if n.Data == "pre" {
		if !isPreNode(n.NextSibling) {
			t.writeByte(w, '\n')
			t.writeIndentationGuide(w, " ==>")
		}
	}
	if !isVeryLastNode(n) {
		if n.Data == "pre" || !t.inPreBlock() {
			if !t.inTextBlock() || t.isTextBlock() {
				t.writeByte(w, '\n')
			}
		}
	}
}

func (t *tidy) writeText(w *bufio.Writer, n *html.Node) {
	if t.inPreBlock() {
		t.writeString(w, n.Data)
		return
	}
	if !t.inTextBlock() {
		return
	}

	input := bytes.TrimSpace([]byte(n.Data))

	if len(input) == 0 {
		if hasPrev(n) || hasNext(n) {
			t.writeByte(w, ' ')
			return
		}
	}

	if hasPrev(n) && unicode.IsSpace(rune(n.Data[0])) {
		t.writeByte(w, ' ')
	}

	if hasNext(n) && unicode.IsSpace(rune(n.Data[len(n.Data)-1])) {
		defer t.writeByte(w, ' ')
	}

	for {
		i := bytes.IndexFunc(input, unicode.IsSpace)
		if i == -1 {
			// There is no more whitespace, write what is left.
			t.write(w, input)
			break
		} else if i == 0 {
			// This is whitespace, write 1 space and move
			// forward to the next non-whitespace character.
			t.writeByte(w, ' ')
			i = bytes.IndexFunc(input, isNotSpace)
			if i == -1 {
				// Only trailing whitespace is left.
				break
			}
			input = input[i:]
		} else {
			// There is some whitespace further ahead. Write the characters
			// up to that whitespace and move the position accordingly.
			t.write(w, input[:i])
			input = input[i:]
		}
	}
}

// Other helper functions:

// findContext finds the parent body or head node.
func findContext(n *html.Node) *html.Node {
	for n != nil {
		if n.Type == html.ElementNode {
			if n.Data == "body" || n.Data == "head" {
				return n
			}
		}
		n = n.Parent
	}
	return nil
}

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

// parseTextNode parses a text node's text, and replaces the
// text node, in place, with the generated nodes it contained.
func parseTextNode(n *html.Node) error {
	context := findContext(n)
	children, err := html.ParseFragment(
		strings.NewReader(n.FirstChild.Data), context,
	)
	if err != nil {
		return err
	}
	for i, c := range children {
		c.Parent = n
		if i == 0 {
			n.FirstChild = c
		} else {
			p := children[i-1]
			p.NextSibling = c
			c.PrevSibling = p
		}
	}
	return nil
}
