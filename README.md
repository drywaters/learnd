# Learnd

Learnd is a personal learning journal for capturing and reviewing resources.

## Quickstart

- Copy `local.mk.example` to `local.mk` and fill in required env vars.
- Run `make run` and open `http://localhost:4500`.

## Configuration

- `DATABASE_URL` and `API_KEY_HASH` are required.
- `PORT` is optional; it defaults to `4500`.

## Health

- `GET /health` returns `200 OK` with `ok` in the body.
