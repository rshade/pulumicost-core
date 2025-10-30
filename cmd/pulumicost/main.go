// Package main provides the pulumicost CLI tool for calculating cloud infrastructure costs.
// It supports both projected costs from Pulumi infrastructure definitions and actual historical
// costs from cloud provider APIs via a plugin-based architecture.
package main

import (
	"os"

	"github.com/rshade/pulumicost-core/internal/cli"
	"github.com/rshade/pulumicost-core/pkg/version"
)

func main() {
	root := cli.NewRootCmd(version.GetVersion())
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
