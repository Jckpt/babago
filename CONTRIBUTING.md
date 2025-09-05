# Contributing to Babago

Thank you for your interest in contributing to Babago! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Release Process](#release-process)

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/yourusername/babago.git
   cd babago
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/jckpt/babago.git
   ```

## Development Setup

### Prerequisites

- Go 1.25.1 or newer
- yt-dlp (for testing video downloads)
- Git

### Setup

1. Install dependencies:

   ```bash
   go mod download
   ```

2. Build the application:

   ```bash
   go build -o babago .
   ```

3. Run tests:
   ```bash
   go test ./...
   ```

### Development Tools

We recommend using these tools for development:

- **VS Code** with Go extension
- **GoLand** (JetBrains IDE)
- **vim/neovim** with Go plugins

## Making Changes

### Branch Naming

Use descriptive branch names:

- `feature/add-new-preset-type`
- `bugfix/fix-url-validation`
- `docs/update-readme`
- `refactor/cleanup-config-handling`

### Commit Messages

Follow conventional commit format:

```
type(scope): description

[optional body]

[optional footer]
```

Types:

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Examples:

```
feat(presets): add support for custom output formats
fix(download): handle network timeout errors
docs(readme): update installation instructions
```

### Code Style

- Follow Go standard formatting (`gofmt`)
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions small and focused
- Use `golangci-lint` for additional linting

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run specific test
go test -run TestFunctionName ./...
```

### Test Requirements

- All new features must include tests
- Bug fixes must include regression tests
- Aim for >80% code coverage
- Tests should be fast and reliable

### Manual Testing

Test the application on different platforms:

- macOS (Intel and Apple Silicon)
- Linux (various distributions)
- Windows

## Submitting Changes

### Pull Request Process

1. Create a feature branch from `main`
2. Make your changes
3. Add tests for new functionality
4. Update documentation if needed
5. Run the full test suite
6. Submit a pull request

### Pull Request Template

Use the provided PR template and fill out all relevant sections.

### Review Process

- All PRs require at least one review
- CI must pass before merging
- Address review feedback promptly
- Keep PRs focused and reasonably sized

## Release Process

### Versioning

We use [Semantic Versioning](https://semver.org/):

- `MAJOR.MINOR.PATCH`
- `1.0.0` for initial release
- `1.0.1` for bug fixes
- `1.1.0` for new features
- `2.0.0` for breaking changes

### Creating a Release

1. Update version in relevant files
2. Update CHANGELOG.md
3. Create a git tag: `git tag v1.0.0`
4. Push the tag: `git push origin v1.0.0`
5. GitHub Actions will automatically create the release

### Changelog

Update `CHANGELOG.md` with:

- New features
- Bug fixes
- Breaking changes
- Deprecations

## Project Structure

```
babago/
├── .github/              # GitHub workflows and templates
├── bin/                  # Compiled binaries
├── main.go              # Main application entry point
├── types.go             # Type definitions
├── config.go            # Configuration management
├── download.go          # Download logic
├── *_view.go            # TUI view components
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
├── Makefile             # Build automation
├── build.sh             # Build script
├── README.md            # Project documentation
├── CONTRIBUTING.md      # This file
└── LICENSE              # License file
```

## Getting Help

- Open an issue for bug reports or feature requests
- Use GitHub Discussions for questions
- Check existing issues before creating new ones

## License

By contributing to Babago, you agree that your contributions will be licensed under the MIT License.
