# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Learnd is a personal learning journal for capturing and reviewing web resources. Users submit URLs which are asynchronously enriched with metadata (title, description, publish date) and summarized using AI.

## Development Commands

```bash
make run              # Generate templ + Tailwind, then run the app (localhost:4500)
make build            # Generate templ + Tailwind, build production binary to bin/learnd
make test             # Run Go tests (go test -v ./...)
make templ            # Generate Go code from .templ files
make templ-watch      # Watch .templ files and regenerate on change
make tail-watch       # Run Tailwind in watch mode
make tail-prod        # Build minified Tailwind CSS
make migrate          # Apply database migrations (requires DATABASE_URL)
make migrate-down     # Roll back the last migration
make migrate-status   # Show migration status
make gen-api-key      # Generate bcrypt hash for API_KEY_HASH
```

## Configuration

Copy `local.mk.example` to `local.mk` for local development. Required env vars:
- `DATABASE_URL` - PostgreSQL connection string
- `API_KEY_HASH` - bcrypt hash for authentication

Optional:
- `GEMINI_API_KEY` - enables AI summarization
- `YOUTUBE_API_KEY` - enables YouTube metadata enrichment
- `LOG_LEVEL` - debug/info/warn/error
- `PORT` - defaults to 4500

Secrets support `_FILE` variants (e.g., `API_KEY_HASH_FILE`).

## Architecture

### Request Flow
1. `cmd/learnd/main.go` - application entrypoint, initializes all components
2. `internal/server/server.go` - chi router setup, middleware, route definitions
3. `internal/handler/` - HTTP handlers for pages and API endpoints
4. `internal/repository/` - PostgreSQL data access (pgx)

### Background Processing
`internal/worker/worker.go` runs two async loops:
- **Enrichment loop** - fetches metadata for new entries using the enricher registry
- **Summarization loop** - generates AI summaries using Gemini, with URL-hash-based caching

### Enricher Registry Pattern
`internal/enricher/enricher.go` defines an `Enricher` interface with `CanHandle()`, `Enrich()`, `Name()`, and `Priority()`. The registry routes URLs to specialized enrichers (YouTube, podcast) or falls back to generic web scraping.

Current enrichers:
- `youtube.go` - YouTube Data API
- `podcast.go` - podcast RSS feeds
- `web.go` - generic HTML metadata extraction (fallback)

### UI Layer (templ + Tailwind)
- `internal/ui/*.templ` - templ source files (pages, partials, components, layout)
- `internal/ui/*_templ.go` - generated Go code (do not edit)
- `tailwind/styles.css` - Tailwind source
- `static/styles.css` - generated CSS output (do not edit)

Edit `.templ` files, then run `make templ` to regenerate.

### Database
- PostgreSQL with pgx driver
- Migrations in `migrations/` using Goose (numbered, snake-cased SQL files)

## Key Patterns

- **Processing status**: Entries track `enrichment_status` and `summary_status` (pending → processing → ok/failed/skipped)
- **URL normalization**: `internal/urlutil/` handles URL canonicalization for deduplication
- **Cookie-based auth**: `internal/middleware/auth.go` validates session cookies against `API_KEY_HASH`
