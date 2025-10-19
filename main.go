package main

import (
	_ "embed"

	"github.com/sestrella/auto-commit-msg/cmd"
)

//go:embed version.txt
var version string

func main() {
	cmd.Execute(version)
}
