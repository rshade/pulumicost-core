// Package registry handles plugin discovery and lifecycle management.
//
// The registry scans for installed plugins in the standard location
// (~/.finfocus/plugins/) and manages their metadata and availability.
//
// # Plugin Directory Structure
//
// Plugins are organized as:
//
//	~/.finfocus/plugins/<name>/<version>/
//	├── finfocus-plugin-<name>     # Plugin binary
//	└── plugin.manifest.json         # Optional manifest
//
// # Discovery Process
//
//  1. Scan plugin directories for valid binaries
//  2. Validate optional manifest files
//  3. Register discovered plugins for use
//
// # Platform Detection
//
// Executable detection is platform-aware, checking Unix permissions
// or Windows .exe extensions as appropriate.
package registry
