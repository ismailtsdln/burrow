package main

import (
	"fmt"
	"os"

	"github.com/ismailtsdln/burrow/internal/ui"
)

func main() {
	if err := ui.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
