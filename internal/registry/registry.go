package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/rshade/pulumicost-core/internal/config"
	"github.com/rshade/pulumicost-core/internal/pluginhost"
)

type Registry struct {
	root     string
	launcher pluginhost.Launcher
}

func NewDefault() *Registry {
	cfg := config.New()
	return &Registry{
		root:     cfg.PluginDir,
		launcher: pluginhost.NewProcessLauncher(),
	}
}

func (r *Registry) ListPlugins() ([]PluginInfo, error) {
	var plugins []PluginInfo
	
	if _, err := os.Stat(r.root); os.IsNotExist(err) {
		return plugins, nil
	}
	
	entries, err := os.ReadDir(r.root)
	if err != nil {
		return nil, fmt.Errorf("reading plugin directory: %w", err)
	}
	
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		
		pluginPath := filepath.Join(r.root, entry.Name())
		versions, err := os.ReadDir(pluginPath)
		if err != nil {
			continue
		}
		
		for _, version := range versions {
			if !version.IsDir() {
				continue
			}
			
			versionPath := filepath.Join(pluginPath, version.Name())
			binPath := r.findBinary(versionPath)
			if binPath != "" {
				plugins = append(plugins, PluginInfo{
					Name:    entry.Name(),
					Version: version.Name(),
					Path:    binPath,
				})
			}
		}
	}
	
	return plugins, nil
}

func (r *Registry) findBinary(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		path := filepath.Join(dir, entry.Name())
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		
		if runtime.GOOS != "windows" {
			if info.Mode()&0111 != 0 {
				return path
			}
		} else {
			if filepath.Ext(path) == ".exe" {
				return path
			}
		}
	}
	
	return ""
}

func (r *Registry) Open(ctx context.Context, onlyName string) ([]*pluginhost.Client, func(), error) {
	plugins, err := r.ListPlugins()
	if err != nil {
		return nil, nil, err
	}
	
	var clients []*pluginhost.Client
	cleanup := func() {
		for _, c := range clients {
			_ = c.Close()
		}
	}
	
	for _, plugin := range plugins {
		if onlyName != "" && plugin.Name != onlyName {
			continue
		}
		
		client, err := pluginhost.NewClient(ctx, r.launcher, plugin.Path)
		if err != nil {
			continue
		}
		clients = append(clients, client)
	}
	
	return clients, cleanup, nil
}

type PluginInfo struct {
	Name    string
	Version string
	Path    string
}
