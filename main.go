package main

import (
	"os"

	"github.com/anirudhraja/gqllinter/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
