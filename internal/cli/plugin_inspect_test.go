package cli_test

import (
	"bytes"
	"testing"

	"github.com/rshade/pulumicost-core/internal/cli"
	"github.com/stretchr/testify/assert"
)

func TestNewPluginInspectCmd(t *testing.T) {
	cmd := cli.NewPluginInspectCmd()
	assert.Equal(t, "inspect", cmd.Name())
	assert.Equal(t, "Inspect a plugin's capabilities and field mappings", cmd.Short)
}

func TestInspectCommand_Flags(t *testing.T) {
	cmd := cli.NewPluginInspectCmd()

	jsonFlag := cmd.Flags().Lookup("json")
	assert.NotNil(t, jsonFlag)
	assert.Equal(t, "bool", jsonFlag.Value.Type())

	verFlag := cmd.Flags().Lookup("version")
	assert.NotNil(t, verFlag)
	assert.Equal(t, "string", verFlag.Value.Type())
}

func TestInspectCommand_PluginNotFound(t *testing.T) {
	cmd := cli.NewPluginInspectCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	// Missing arguments
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.Error(t, err) // Should error on missing args

	// Non-existent plugin
	cmd.SetArgs([]string{"non-existent-plugin", "aws:ec2:Instance"})
	err = cmd.Execute()
	assert.Error(t, err)
}
