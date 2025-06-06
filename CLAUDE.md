# CLAUDE.md
必ず日本語で回答してください。

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is `gphoto-cli`, a command-line interface tool that uses Google API's Picker API to retrieve and manage Google Photos information. The project uses the Cobra framework for CLI command structure.

## Development Setup

This project is developed in Go 1.24.4 using Docker with Cursor Dev Containers.

### Docker Development Environment

1. Open the project in Cursor
2. Use "Dev Containers: Reopen in Container" command
3. The container will automatically set up Go 1.24 environment with Claude Code support

## Common Commands

### Building and Running
- `go build` - Build the application (creates `gphoto-cli` binary)
- `./gphoto-cli` - Run the application
- `./gphoto-cli --help` - Show help and available commands
- `./gphoto-cli version` - Show version information
- `./gphoto-cli list` - List Google Photos albums (placeholder)

### Development
- `go mod tidy` - Clean up dependencies
- `go test ./...` - Run all tests
- `go fmt ./...` - Format code
- `go vet ./...` - Static analysis
- `staticcheck ./...` - Advanced static analysis (if available)

## Architecture

### CLI Structure
The application uses the Cobra framework for command-line interface:
- Root command (`gphoto-cli`) provides basic information and help
- Subcommands are defined as separate `cobra.Command` variables
- Commands are registered in the `init()` function using `rootCmd.AddCommand()`
- Main entry point is `rootCmd.Execute()` in `main()`

### Current Commands
- `version` - Displays version information (currently v0.1.0)
- `list` - Placeholder for listing Google Photos albums

### Dependencies
- `github.com/spf13/cobra` - CLI framework for command structure and parsing
- `github.com/spf13/pflag` - Command-line flag parsing (Cobra dependency)
- `github.com/inconshreveable/mousetrap` - Windows console helper (Cobra dependency)

## Architecture Notes for Future Development

Key considerations for Google Photos CLI tool expansion:
- Google API authentication and authorization (OAuth 2.0)
- Google Picker API integration for photo selection
- Google Photos API for retrieving photo metadata and content
- Configuration management for API credentials
- Rate limiting and API quota management
- Cross-platform compatibility