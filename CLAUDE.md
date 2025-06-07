# CLAUDE.md
必ず日本語で回答してください。

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is `gphoto-cli`, a command-line interface tool that uses Google Photos Picker API to securely access and retrieve Google Photos information. The tool implements OAuth 2.0 authentication and provides users the ability to select photos from their entire Google Photos library using the new 2024 Picker API.

## Development Setup

This project is developed in Go 1.24.4 using Docker with Cursor Dev Containers.

### Required Setup Files
- `.env` - Environment variables for OAuth 2.0 configuration (copy from .env.example)
- `token.json` - Auto-generated OAuth token cache (created on first authentication)

### Docker Development Environment

1. Open the project in Cursor
2. Use "Dev Containers: Reopen in Container" command
3. The container will automatically set up Go 1.24 environment with Claude Code support

## Common Commands

### Building and Running
- `go build -o gphoto-cli` - Build the application (creates `gphoto-cli` binary)
- `./gphoto-cli setup` - Interactive setup for OAuth credentials
- `./gphoto-cli config show` - Show current configuration
- `./gphoto-cli config reset` - Reset configuration and authentication
- `./gphoto-cli picker` - Launch Google Photos Picker for full library access
- `./gphoto-cli view` - Quick view mode with preview and external viewer
- `./gphoto-cli --help` - Show help and available commands
- `./gphoto-cli version` - Show version information (v0.1.0)

### Development
- `go mod tidy` - Clean up dependencies
- `go fmt ./...` - Format code
- `go vet ./...` - Static analysis
- `rm -f token.json` - Clear authentication cache for re-authentication

### Authentication Reset
When switching OAuth scopes or troubleshooting authentication:
```bash
rm -f token.json && go build && ./gphoto-cli picker
```

## Architecture

### Multi-File Structure
The application is split across specialized files:
- `main.go` - Cobra CLI command definitions and entry point
- `auth.go` - OAuth 2.0 authentication flow with multiple methods (local server + manual)
- `picker.go` - Google Photos Picker API client implementation
- `photos.go` - Google Photos Library API client (limited to app-created content)

### Authentication Flow
Implements dual authentication methods in `auth.go`:
1. **Automatic**: Local server on :8080 for OAuth redirect
2. **Manual**: Out-of-band (OOB) flow with manual code entry
Uses scope `https://www.googleapis.com/auth/photospicker.mediaitems.readonly` for Picker API access

### API Implementation
**Picker API (picker.go)**:
- Session-based photo selection workflow
- Polling mechanism for user selection completion
- Full access to user's Google Photos library
- Returns detailed metadata including EXIF data, camera info, file dimensions

### CLI Commands Architecture
Uses Cobra framework with command registration in `init()`:
- `rootCmd` - Base command with help information
- `versionCmd` - Static version display
- `setupCmd` - Interactive OAuth setup
- `configCmd` - Configuration management (show/reset)
- `pickerCmd` - Full library access via Picker API
- `viewCmd` - Quick view mode with immediate preview

### OAuth 2.0 Scope Strategy
Current implementation uses `photospicker.mediaitems.readonly` scope specifically for the Picker API, which provides:
- Access to user-selected photos from entire library
- Detailed photo metadata and EXIF information
- Base URLs for image content access
- Session-based secure photo selection

### Data Structures
**Picker API Models**:
- `PickerSession` - Session management with polling configuration
- `MediaItem` - Photo metadata with nested `MediaFile` structure
- `MediaFileMetadata` - Detailed photo information including camera settings
- `PhotoMetadata` - EXIF data (ISO, aperture, focal length, exposure)

## Key Dependencies

- `github.com/spf13/cobra` - CLI framework
- `golang.org/x/oauth2` - OAuth 2.0 implementation
- `github.com/joho/godotenv` - Environment variable support (development)
- `gopkg.in/yaml.v3` - Configuration file management
- Direct HTTP client implementation for Picker API (no official Go client available)

## Configuration Management

### Interactive Setup
- `./gphoto-cli setup` provides guided OAuth 2.0 credential configuration
- Stores settings in `~/.gphoto-cli/config.yaml` with secure permissions (600)
- Supports both automatic (local server) and manual (OOB) authentication methods

### Configuration Options
- Google OAuth Client ID and Secret
- Authentication method (server/oob)
- Redirect URI configuration
- OAuth scope specification

## API Implementation

### Google Photos Picker API
This tool uses only the Google Photos Picker API for accessing user photos:
- No API enablement required in Google Cloud Console
- OAuth 2.0 authentication sufficient for access
- Session-based photo selection workflow
- Full access to user's entire Google Photos library
- Detailed metadata including EXIF, camera info, and file dimensions