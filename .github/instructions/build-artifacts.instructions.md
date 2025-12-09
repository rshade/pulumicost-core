---
applyTo: '**/vendor/**,**/.git/**,**/*.exe,**/*.dll,**/*.so,**/*.dylib'
excludeAgent: 'code-review'
---

# Build Artifacts and Dependencies Instructions

This covers build artifacts, dependencies, and version control files that should not be modified or analyzed.

## Excluded paths:

- `**/vendor/**` - Go vendor directory
- `**/.git/**` - Git repository metadata
- `**/*.exe` - Windows executables
- `**/*.dll` - Windows dynamic libraries
- `**/*.so` - Linux shared objects
- `**/*.dylib` - macOS dynamic libraries

## Guidelines:

- **Do not modify** any files in these locations
- **Do not analyze** or review code in these directories
- **Do not reference** files from these locations
- **Do not commit** changes to these files

## Purpose:

- Third-party dependencies (vendor/)
- Version control metadata (.git/)
- Compiled binaries and libraries
- Platform-specific executables

## When working with these files:

- Treat as generated/readonly
- Dependencies are managed by go mod
- Binaries are built by CI/CD pipelines
- Git operations handle .git directory automatically
