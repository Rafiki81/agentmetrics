package main

import (
	"os"

	"github.com/rafaelperezbeato/agentmetrics/internal/cli"
)

var version = "0.1.1"

func main() {
	os.Exit(cli.Run(os.Args[1:], version))
}
