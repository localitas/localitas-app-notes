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

Run these in the app's repository, `~/localitas-app-notes`.

### Version

```bash
bin/notes-server --version
```

### Build from source

```bash
# Native binary (bin/notes-server)
make build

# Linux amd64 binary the Docker image is built from (notes-server-linux-amd64)
make build-linux
```

Both run `make lint` (gofmt and go vet) first; `make test` runs the lint step and then the tests.

### Run natively

```bash
make start   # builds bin/notes-server, writes the launchd service com.localitas.app.notes for bin/notes-server and loads it on port 9202
make stop
make logs    # follows ~/.localitas/logs/notes/stdout.log
```

The app announces itself to core when it starts, and core serves it at `/apps/ext/notes/`.

### Docker

```bash
make build-docker   # builds the Linux binary and the notes:latest image
make start-docker   # runs container "notes" on port 9202, labelled localitas.app=true
make stop-docker
make logs-docker
```

Core registers containers labelled `localitas.app=true` once Docker reports them healthy.

### Publish

```bash
# Run the tests, build the image, push ghcr.io/localitas/localitas-app-notes:latest and :<VERSION>
make docker-push

# Build dist/notes-server-darwin-arm64 and dist/notes-server-darwin-amd64
# and create GitHub release vX.Y.Z on this repository with them
make release VERSION=vX.Y.Z
```

`VERSION` defaults to `git describe --tags --always --dirty`.

### Download

Pre-built macOS binaries are on the [releases page](https://github.com/localitas/localitas-app-notes/releases):
- `notes-server-darwin-arm64` (Apple Silicon)
- `notes-server-darwin-amd64` (Intel)

Every push to `main` publishes a release: `.github/workflows/release.yml` runs the tests and releases both binaries as `v0.1.<commit count>.<UTC timestamp>`. `make release` publishes one by hand.

Download with the GitHub CLI:

    gh release download --repo localitas/localitas-app-notes --pattern 'notes-server-*'
