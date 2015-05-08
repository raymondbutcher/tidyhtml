# tidyhtml

This is a Go package for making HTML look tidy.

And by "tidy" I mean "exactly as I would write it by hand".

### Overview

* Parses HTML into a tree and then writes the output from scratch
    using said tree
* Assumes HTML is well formed - no effort has been made to "battle test" this
    against invalid input
* Indents elements by 4 spaces per level
* Removes unnecessary whitespace except for indentation
* Keeps elements with text as a single clump
* Outputs `<pre>` blocks with no indentation so they display correctly
* Performance has not been a priority

### Usage

Get the package:
`go get github.com/raymondbutcher/tidyhtml`

Example program, provided in the cmd directory:
```go
package main

import (
	"fmt"
	"os"

	"github.com/raymondbutcher/tidyhtml"
)

func main() {
	if err := tidyhtml.Copy(os.Stdout, os.Stdin); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stdout)
}
```

### Example

```
$ cat demo.in.html
<html>
<head>
<title>demo</title>
</head>
<body>
<div><h1>this is a demo</h1>
<p>
ok
</p>
<div><span>this will do</span></div></div>
</body></html>
```

```
$ cat demo.in.html | tidyhtml
<html>
    <head>
        <title>demo</title>
    </head>
    <body>
        <div>
            <h1>this is a demo</h1>
            <p>ok</p>
            <div>
                <span>this will do</span>
            </div>
        </div>
    </body>
</html>
```
