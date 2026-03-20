package main

import (
	"os"

	"github.com/matzew/kaudy/pkg/cli"
)

func main() {
	if err := cli.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
