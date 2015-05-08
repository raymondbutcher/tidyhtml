// Package tidyhtml cleans up HTML input and outputs a tidy version.
package tidyhtml

import (
	"bytes"
	"io"

	"golang.org/x/net/html"
)

// Copy HTML from src to dst and tidy it up in the process.
func Copy(dst io.Writer, src io.Reader) error {

	node, err := html.Parse(src)
	if err != nil {
		return err
	}

	t := newTidy()
	b, err := t.render(node)
	if err != nil {
		return err
	}

	_, err = io.Copy(dst, bytes.NewReader(b))
	return err
}
