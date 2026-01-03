# Repository Guidelines

## Project Structure & Module Organization
- `cmd/learnd/main.go` is the application entrypoint.
- `internal/` holds core packages (config, server, handlers, middleware, repository, summarizer, enricher, worker).
- `templates/` contains HTML templates; `static/` holds compiled assets (e.g., `static/styles.css`).
- `tailwind/` contains source CSS for Tailwind; do not edit generated CSS directly.
- `migrations/` contains Goose SQL migrations (e.g., `001_create_entries.sql`).
- `scripts/` contains helper utilities (e.g., API key hashing).

## Build, Test, and Development Commands
- `make run`: build Tailwind CSS and run the app via `go run ./cmd/learnd`.
- `make build`: build a production binary at `bin/learnd`.
- `make tail-watch`: run Tailwind in watch mode (requires the `tailwindcss` CLI).
- `make tail-prod`: build minified Tailwind output into `static/styles.css`.
- `make migrate`, `make migrate-down`, `make migrate-status`: apply, rollback, or inspect database migrations using `goose`.
- `make test`: run Go tests (`go test -v ./...`).
- `make gen-api-key`: generate a bcrypt hash for `API_KEY_HASH`.

## Coding Style & Naming Conventions
- Go code should be formatted with `gofmt`; keep packages cohesive under `internal/`.
- Template files in `templates/` use 4-space indentation; follow existing block naming patterns.
- Migration files are numbered and snake-cased: `migrations/NNN_description.sql`.
- Treat `static/styles.css` as generated output from Tailwind.

## Testing Guidelines
- Use standard Go tests in `*_test.go` files colocated with the package under test.
- Prefer table-driven tests where multiple cases apply.
- Run `make test` before opening a PR; add tests for handlers, repositories, or enrichers when behavior changes.

## Commit & Pull Request Guidelines
- Follow Conventional Commits (`type: summary`), as seen in history (e.g., `chore: add license file`).
- Keep commit subjects short, imperative, and scoped to one change.
- PRs should include: a clear description, linked issue (if applicable), migration notes (if schema changes), and UI screenshots when templates or CSS change.

## Configuration & Secrets
- Copy `local.mk.example` to `local.mk` for local development; `local.mk` is gitignored.
- Required env vars: `DATABASE_URL` and `API_KEY_HASH`. Optional: `GEMINI_API_KEY`, `YOUTUBE_API_KEY`, `LOG_LEVEL`, and `PORT`.
- The app also supports `_FILE` variants for secrets (e.g., `API_KEY_HASH_FILE`).
