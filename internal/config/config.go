package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	PluginDir string
	SpecDir   string
}

func New() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		PluginDir: filepath.Join(homeDir, ".pulumicost", "plugins"),
		SpecDir:   filepath.Join(homeDir, ".pulumicost", "specs"),
	}
}

func (c *Config) PluginPath(name, version string) string {
	return filepath.Join(c.PluginDir, name, version)
}