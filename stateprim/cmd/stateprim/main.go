package main

import (
	"fmt"
	"os"

	"github.com/propifly/primkit/stateprim/internal/cli"
)

func main() {
	if err := cli.NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
