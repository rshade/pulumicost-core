---
applyTo: '**/Makefile'
---

# Makefile Instructions

Makefiles in this project follow specific conventions and safety practices.

## Makefile Standards:

- Use proper target dependencies
- Include help targets and documentation
- Use safe shell commands with proper error handling
- Follow consistent naming conventions
- Include clean and validation targets

## Command Safety:

- Use proper shell escaping and quoting
- Handle command failures appropriately
- Use `set -e` for strict error checking where appropriate
- Validate command prerequisites
- Use absolute paths when necessary

## Variable Usage:

- Define variables at the top of the file
- Use consistent variable naming (UPPERCASE)
- Document complex variable usage
- Use conditional variable assignment where appropriate

## Target Organization:

- Group related targets logically
- Include `.PHONY` declarations for non-file targets
- Use standard target names (all, clean, test, etc.)
- Document target purposes and usage

## Cross-Platform Compatibility:

- Handle different operating systems appropriately
- Use portable shell commands
- Account for different executable extensions
- Test on multiple platforms when possible

## Build Process:

- Ensure reproducible builds
- Include version information embedding
- Use proper compiler flags
- Validate build artifacts

## CI/CD Integration:

- Work with GitHub Actions workflows
- Support automated builds and releases
- Include validation steps
- Generate appropriate artifacts
