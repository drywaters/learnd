# Project Overview

Learnd is a personal learning journal application designed for capturing and reviewing learning resources. It is a full-stack Go application that leverages modern tools for specific functionalities:
- **Backend:** Go (Golang) using `chi` for routing.
- **Frontend:** Server-side rendered UI using `templ` components styled with Tailwind CSS.
- **Database:** PostgreSQL with `pgx` driver.
- **AI Integration:** Google Gemini for content summarization.
- **Enrichment:** YouTube and web page metadata extraction.

# Building and Running

The project uses a `Makefile` to manage build, run, and migration tasks.

## Prerequisites
- Go 1.24+
- PostgreSQL
- `templ` CLI (`go install github.com/a-h/templ/cmd/templ@latest`)
- `tailwindcss` CLI (standalone binary or via npm)
- `goose` CLI for migrations (`go install github.com/pressly/goose/v3/cmd/goose@latest`)

## Key Commands

*   **Run Locally:**
    ```bash
    make run
    ```
    This command generates `templ` files, builds Tailwind CSS, and runs the application.

*   **Build for Production:**
    ```bash
    make build
    ```
    Creates a binary in `bin/learnd`.

*   **Database Migrations:**
    ```bash
    make migrate          # Apply migrations
    make migrate-down     # Rollback last migration
    make migrate-status   # Check status
    ```

*   **Run Tests:**
    ```bash
    make test
    ```

*   **Generate API Key Hash:**
    ```bash
    make gen-api-key
    ```
    Helper to generate the bcrypt hash required for the `API_KEY_HASH` configuration.

## Configuration
Configuration is handled via environment variables. A `local.mk` file (not tracked by git) can be used to set these for the `make` commands (see `local.mk.example`).

*   `DATABASE_URL` (Required): PostgreSQL connection string.
*   `API_KEY_HASH` (Required): Bcrypt hash of the API key for authentication.
*   `PORT`: Server port (default: 4500).
*   `GEMINI_API_KEY`: API key for Google Gemini (optional, for summarization).
*   `YOUTUBE_API_KEY`: API key for YouTube Data API (optional, for video details).
*   `LOG_LEVEL`: Logging level (default: info).

All secrets also support a `_FILE` suffix (e.g., `DATABASE_URL_FILE`) to read the value from a file, which is useful for Docker/Kubernetes environments.

# Development Conventions

## Directory Structure
*   `cmd/learnd/`: Application entry point (`main.go`).
*   `internal/`: Private application code.
    *   `config/`: Configuration loading.
    *   `ui/`: `templ` source files (`*.templ`) and generated Go code.
    *   `handler/`: HTTP handlers.
    *   `middleware/`: HTTP middleware.
    *   `repository/`: Database access layer.
    *   `enricher/`: External data fetching (YouTube, Web).
    *   `summarizer/`: AI summarization logic.
*   `migrations/`: SQL migration files (managed by `goose`).
*   `static/`: Static assets (compiled CSS).
*   `tailwind/`: Tailwind CSS source files.

## Coding Style
*   **Go:** Follow standard Go formatting (`gofmt`).
*   **Templ:** `templ` files are located in `internal/ui`. Run `make templ` to regenerate the Go code after editing `.templ` files. Do not edit the generated `*_templ.go` files directly.
*   **CSS:** Tailwind CSS is used. Edit `tailwind/styles.css` if custom CSS is needed, but prefer utility classes in `templ` files.
*   **Migrations:** Database changes are versioned in `migrations/` using `goose`.

## Testing
*   Tests are co-located with the source code (e.g., `*_test.go`).
*   Run `make test` to execute the test suite.
