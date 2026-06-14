---
title: Notes
description: Note-taking with executable code blocks
---

# Notes

Create and manage notes with support for executable code blocks, variables, and data loading.

## Note Management

Create, edit, and delete notes. Each note has a title and markdown content with optional executable code blocks.

**GET /api/notes** - List all notes
**POST /api/notes** - Create a new note
**GET /api/notes/{id}** - Get a note by ID
**PUT /api/notes/{id}** - Update a note
**DELETE /api/notes/{id}** - Delete a note

## Code Execution

Notes support fenced code blocks that can be executed. The app parses code blocks from note content and runs them, returning output for each block.

**POST /api/notes/{id}/execute** - Execute all code blocks in a note
**POST /api/notes/{id}/execute/{blockID}** - Execute a single code block

Supported execution environments include JavaScript (via the built-in JS executor) and Python (via the Python client).

## Variables

Define variables within notes that persist across executions. Variables can be referenced in code blocks.

## CSV Data Loading

Load CSV files into notes for use in code blocks. The CSV loader parses tabular data and makes it available for processing.

## Search

**GET /api/search** - Full-text search across note titles and content

## HTMX Partials

The web interface uses server-rendered partials for a responsive editing experience.

**GET /partials/sidebar** - Render the notes list
**GET /partials/editor/{id}** - Render the note editor
**POST /partials/create** - Create a note and refresh the UI
**POST /partials/save/{id}** - Save a note and refresh the UI
**POST /partials/execute/{id}** - Execute code blocks and show results

## Build & Deploy

### Version

```bash
./notes-server --version
```

### Build from source

```bash
# Development (native)
cd apps/notes && go build -o bin/notes-server ./cmd/notes-server

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -trimpath -o bin/notes-server-linux-amd64 ./cmd/notes-server
```

### Docker

Build a Docker image directly from the binary:

```bash
# Default base image (debian:12-slim)
./notes-server docker-build

# Custom base image
./notes-server docker-build --base ubuntu:24.04

# Custom Dockerfile
./notes-server docker-build --dockerfile ./my.Dockerfile

# Tag and push to registry
./notes-server docker-build --tag ghcr.io/localitas/notes:latest --push
```

The `docker-build` command requires a Linux amd64 binary in the same directory. Run `make deploy-build` from the project root first.

### Download

Pre-built binaries are available on the [GitHub releases page](https://github.com/localitas/localitas/releases).

Each release includes three builds per app:
- `notes-server-darwin-arm64` (macOS Apple Silicon)
- `notes-server-linux-amd64` (Linux x86_64)
- `notes-server-linux-arm64` (Linux ARM64)

Download with the GitHub CLI:

    gh release download --repo localitas/localitas --pattern 'notes-server-*'

### Release

All app binaries are published to GitHub releases as part of `make deploy-upload-image`.
