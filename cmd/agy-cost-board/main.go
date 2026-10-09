package main

import (
	"github.com/julienbreux/agy-cost-board/internal/cli"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

var runCLI = cli.Execute

func main() {
	cli.Version = Version
	cli.Commit = Commit
	cli.BuildDate = BuildDate
	runCLI()
}

