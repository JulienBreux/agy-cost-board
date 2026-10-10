package main

import (
	"os"

	"github.com/julienbreux/agy-cost-board/internal/cli"
)

var (
	Version   = "v0.4.0"
	Commit    = "none"
	BuildDate = "unknown"
)

var runCLI = cli.Execute

func main() {
	cli.Version = Version
	cli.Commit = Commit
	cli.BuildDate = BuildDate
	if err := runCLI(); err != nil {
		os.Exit(1)
	}
}

