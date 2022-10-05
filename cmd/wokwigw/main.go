package main

import (
	"fmt"
	"os"
)

func main() {
	if err := execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
