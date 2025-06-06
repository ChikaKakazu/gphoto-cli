# CLAUDE.md
必ず日本語で回答してください。

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is `gphoto-cli`, a command-line interface tool that uses Google Photos Picker API to securely access and retrieve Google Photos information. The tool implements OAuth 2.0 authentication and provides users the ability to select photos from their entire Google Photos library using the new 2024 Picker API.

## Development Setup

This project is developed in Go 1.24.4 using Docker with Cursor Dev Containers.

### Required Setup Files
- `credentials.json` - OAuth 2.0 client credentials from Google Cloud Console
- `token.json` - Auto-generated OAuth token cache (created on first authentication)

### Docker Development Environment

1. Open the project in Cursor
2. Use "Dev Containers: Reopen in Container" command
3. The container will automatically set up Go 1.24 environment with Claude Code support

## Common Commands

### Building and Running
- `go build` - Build the application (creates `gphoto-cli` binary)
- `./gphoto-cli --help` - Show help and available commands
- `./gphoto-cli version` - Show version information (v0.1.0)
- `./gphoto-cli list` - List app-created albums only (limited by API scope)
- `./gphoto-cli photos --limit 10` - List app-created photos only
- `./gphoto-cli picker` - Launch Google Photos Picker for full library access

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

**Library API (photos.go)**:
- Limited to app-created content only (2025 API restrictions)
- Uses deprecated scope for legacy functionality
- Primarily serves albums created by this application

### CLI Commands Architecture
Uses Cobra framework with command registration in `init()`:
- `rootCmd` - Base command with help information
- `versionCmd` - Static version display
- `listCmd` - App-created albums via Library API
- `photosCmd` - App-created photos with --limit flag
- `pickerCmd` - Full library access via Picker API

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
- `github.com/gphotosuploader/google-photos-api-client-go/v3` - Legacy Library API client
- Direct HTTP client implementation for Picker API (no official Go client available)

## Important API Considerations

### 2025 Google Photos API Changes
As of March 31, 2025, Google Photos Library API scope restrictions mean:
- `photoslibrary.readonly` and similar scopes are deprecated
- Library API now only accesses app-created content
- For full library access, applications must use the new Picker API
- This tool implements both APIs to demonstrate the differences and provide maximum functionality