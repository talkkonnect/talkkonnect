package main

import (
	"fmt"
	"os"
)

var name = "volume"
var version = "v0.0.0"
var description = "control audio volume"
var author = "itchyny"

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
		os.Exit(1)
	}
}
