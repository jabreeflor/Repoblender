// Command repoblender converts a repository into llms.txt format.
package main

import (
	"os"

	"github.com/jabreeflor/repoblender/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
