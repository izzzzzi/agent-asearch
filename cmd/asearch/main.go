package main

import (
	"os"

	"github.com/izzzzzi/agent-asearch/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
