package main

import (
	"os"

	"macventory/components/cli"
)

// APP VERSION - single source of truth for the application version
const VERSION = "0.1.0"

// Main
func main() {
	os.Exit(cli.Run(os.Args[1:], VERSION))
}
