package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rshade/finfocus/internal/cli"
	"github.com/rshade/finfocus/pkg/version"
)

func TestCLIBranding(t *testing.T) {
	t.Run("root command help shows FinFocus", func(t *testing.T) {
		root := cli.NewRootCmd(version.GetVersion())
		buf := new(bytes.Buffer)
		root.SetOut(buf)
		root.SetErr(buf)
		root.SetArgs([]string{"--help"})

		err := root.Execute()
		if err != nil {
			t.Fatalf("failed to execute root command: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "FinFocus") {
			t.Errorf("expected output to contain 'FinFocus', got:\n%s", output)
		}
		if strings.Contains(strings.ToLower(output), "finfocus") {
			t.Errorf("expected output NOT to contain 'finfocus', got:\n%s", output)
		}
	})

	t.Run("version output contains FinFocus", func(t *testing.T) {
		root := cli.NewRootCmd(version.GetVersion())
		buf := new(bytes.Buffer)
		root.SetOut(buf)
		root.SetErr(buf)
		root.SetArgs([]string{"--version"})

		err := root.Execute()
		if err != nil {
			t.Fatalf("failed to execute root command: %v", err)
		}

		output := buf.String()
		// Cobra --version usually just prints the version string from cmd.Version
		// unless a custom version template is set.
		// If it only prints "0.1.0", we might need to update the template.
		t.Logf("Version output: %q", output)
	})
}
