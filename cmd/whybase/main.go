package main

import (
	"os"

	"github.com/mkfeuhrer/whybase/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
