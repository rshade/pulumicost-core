# Contributing to PulumiCost

Thank you for your interest in contributing to PulumiCost! This document provides guidelines and instructions for contributing code, documentation, and feedback.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Documentation](#documentation)
- [Plugin Development](#plugin-development)
- [Getting Help](#getting-help)

---

## Code of Conduct

We are committed to providing a welcoming and inclusive environment. Please read our [Code of Conduct](docs/support/code-of-conduct.md) before participating.

---

## Getting Started

### Prerequisites

- Go 1.24+ (for core development)
- Git
- Make
- For documentation: Ruby 3.2+ (for Jekyll)

### Fork and Clone

```bash
# Fork the repository on GitHub
# Clone your fork
git clone https://github.com/YOUR_USERNAME/pulumicost-core.git
cd pulumicost-core

# Add upstream remote
git remote add upstream https://github.com/rshade/pulumicost-core.git
```

---

## Development Setup

### Initial Setup

```bash
# Install dependencies
go mod download

# Verify your setup
make build
./bin/pulumicost --help
```

### Development Commands

```bash
# Build the binary
make build

# Run tests
make test

# Run linters
make lint

# Run validation
make validate

# Clean build artifacts
make clean

# View all commands
make help
```

### For Documentation Development

```bash
# Install Ruby dependencies
cd docs
bundle install
cd ..

# Lint documentation
make docs-lint

# Serve documentation locally
make docs-serve
# Visit http://localhost:4000/pulumicost-core/

# Build documentation
make docs-build

# Validate documentation
make docs-validate
```

---

## Making Changes

### Creating a Branch

```bash
# Fetch latest changes
git fetch upstream

# Create feature branch from main
git checkout -b feature/my-feature upstream/main
```

### Branch Naming Conventions

- `feature/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation only
- `refactor/description` - Code refactoring
- `test/description` - Test additions/improvements

### Code Style

#### Go Code

- Follow standard Go conventions ([Effective Go](https://golang.org/doc/effective_go))
- Use `gofmt` for formatting
- Ensure `golangci-lint` passes: `make lint`
- Write clear variable and function names
- Add comments for exported functions and complex logic

#### Markdown

- Follow Google style guide for markdown formatting
- Use `make docs-lint` to validate
- Line length: 120 characters (soft limit for prose)
- Use clear headings and structure

### Commit Messages

Write clear, descriptive commit messages:

```
feature: Add support for cost filtering by tags

This allows users to filter cost results by resource tags,
enabling cost allocation by team or environment.

- Add tag filtering to engine
- Add --filter flag to CLI
- Update tests and documentation

Closes #123
```

**Format:**
- First line: type: Short description (50 chars max)
- Blank line
- Body: Detailed explanation (wrapped at 72 chars)
- Blank line
- References: Link to related issues/PRs

**Types:**
- `feature` - New functionality
- `fix` - Bug fixes
- `docs` - Documentation changes
- `test` - Test additions/changes
- `refactor` - Code refactoring
- `perf` - Performance improvements
- `chore` - Build, dependencies, etc.

### Testing Your Changes

```bash
# Run tests for your changes
make test

# Run tests with coverage
go test -cover ./...

# Run tests in specific package
go test -v ./internal/engine/...

# Run specific test
go test -run TestCostCalculation ./...
```

All tests must pass before submitting a PR.

---

## Testing

### Writing Tests

Place tests in `*_test.go` files alongside the code being tested.

```go
// Example test
func TestMyFeature(t *testing.T) {
    // Arrange
    input := "test input"

    // Act
    result := MyFunction(input)

    // Assert
    if result != "expected" {
        t.Errorf("Expected 'expected', got '%s'", result)
    }
}
```

### Test Coverage

- Aim for >80% coverage in new code
- Run `make test` to check current coverage
- Focus on critical paths and error cases

---

## Submitting Changes

### Pull Request Process

1. **Ensure your branch is up to date:**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run all checks locally:**
   ```bash
   make test
   make lint
   make validate
   make docs-validate  # If docs changed
   ```

3. **Push your changes:**
   ```bash
   git push origin feature/my-feature
   ```

4. **Create a Pull Request:**
   - Go to the repository on GitHub
   - Click "New Pull Request"
   - Select your branch
   - Fill in the PR template
   - Submit

### PR Description Template

```markdown
## Description
Brief description of the changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Motivation and Context
Why is this change needed? What problem does it solve?

## How Has This Been Tested?
Describe how you tested your changes

## Screenshots (if applicable)
Add screenshots for UI changes

## Checklist
- [ ] My code follows the project style guidelines
- [ ] I have performed a self-review
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests passed locally with my changes
- [ ] All linters pass (`make lint`)

## Closes
Closes #(issue number)
```

### CI Checks

All pull requests must pass the following checks:

- ✅ Go Tests (`go test`)
- ✅ Code Coverage (minimum 20%)
- ✅ Lint (`golangci-lint`)
- ✅ Security (`govulncheck`)
- ✅ Documentation Validation
- ✅ Cross-Platform Build

---

## Documentation

### Where Documentation Lives

- **Code-focused docs**: `docs/` directory
- **API docs**: Protocol buffer comments in `internal/proto/`
- **Package docs**: `CLAUDE.md` files in package directories
- **Contributing**: This file and `docs/support/contributing.md`

### Documentation Standards

- All public functions should have godoc comments
- Update `docs/` when adding user-facing features
- Update `docs/reference/` for CLI/API changes
- Link to relevant documentation in code comments

### Documentation Commands

```bash
# Lint documentation
make docs-lint

# Build documentation site
make docs-build

# Serve locally for preview
make docs-serve

# Validate documentation
make docs-validate
```

### Adding a New Guide

1. Create file in appropriate `docs/` subdirectory
2. Add frontmatter:
   ```yaml
   ---
   layout: default
   title: My Guide Title
   description: Brief description for search
   ---
   ```
3. Write content following [Google style guide](https://developers.google.com/style)
4. Run `make docs-lint` to validate
5. Submit PR with documentation changes

---

## Plugin Development

To develop a PulumiCost plugin:

1. **Read the plugin development guide**: [docs/plugins/plugin-development.md](docs/plugins/plugin-development.md)
2. **Review the SDK reference**: [docs/plugins/plugin-sdk.md](docs/plugins/plugin-sdk.md)
3. **Study the Vantage plugin example**: [docs/plugins/vantage/](docs/plugins/vantage/)
4. **Use the plugin template**: `cmd/pulumicost plugin-init`

---

## Getting Help

### Documentation

- [Developer Guide](docs/guides/developer-guide.md) - Complete developer documentation
- [Architecture](docs/architecture/) - System design and architecture
- [API Reference](docs/reference/api-reference.md) - gRPC API documentation
- [Troubleshooting](docs/support/troubleshooting.md) - Common issues

### Support Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: Questions and ideas
- **Code Review**: Learn by reviewing open PRs
- **Examples**: Study examples in `examples/`

---

## Development Workflow Example

```bash
# 1. Create feature branch
git checkout -b feature/cost-filtering upstream/main

# 2. Make changes to code
# ... edit files ...

# 3. Write or update tests
# ... edit *_test.go files ...

# 4. Run tests locally
make test

# 5. Run linters
make lint

# 6. Run validation
make validate

# 7. Commit changes
git add .
git commit -m "feature: Add cost filtering by tags"

# 8. Push to fork
git push origin feature/cost-filtering

# 9. Create PR on GitHub (website)

# 10. Address review feedback
# ... make changes based on review ...

# 11. Rebase and push
git rebase upstream/main
git push -f origin feature/cost-filtering
```

---

## Code Review Process

### What to Expect

- Code reviews by maintainers and community
- Feedback on code quality, design, and testing
- Suggestions for improvement
- Timeline: Reviews typically happen within 2-3 business days

### Providing Good Reviews

When reviewing others' code:

- Be respectful and constructive
- Focus on the code, not the person
- Ask questions if something is unclear
- Suggest improvements, don't demand changes
- Acknowledge good work

---

## License

By contributing, you agree that your contributions will be licensed under the Apache-2.0 license. See [LICENSE](LICENSE) for details.

---

## Questions?

- **Development questions**: [docs/support/support-channels.md](docs/support/support-channels.md)
- **Documentation questions**: Check [docs/README.md](docs/README.md)
- **Plugin questions**: [docs/plugins/plugin-development.md](docs/plugins/plugin-development.md)

---

Thank you for contributing to PulumiCost! 🙏
