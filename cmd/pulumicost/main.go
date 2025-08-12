package main

import (
	"os"

	"github.com/rshade/pulumicost-core/internal/cli"
	"github.com/rshade/pulumicost-core/pkg/version"
)

func main() {
	root := cli.NewRootCmd(version.Version)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
