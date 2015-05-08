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
