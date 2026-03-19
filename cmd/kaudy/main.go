package main

import (
	"os"

	"github.com/matzew/kaudy"
)

func main() {
	if err := kaudy.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
