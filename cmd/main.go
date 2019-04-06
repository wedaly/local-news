package main

import (
	"fmt"
	"github.com/wedaly/local-news/internal/app"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %v DB_PATH\n", os.Args[0])
		os.Exit(1)
	}

	config := app.Config{DBPath: os.Args[1]}

	err := app.Run(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unexpected error: %v\n", err)
		os.Exit(1)
	}
}
