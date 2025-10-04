# Copilot Instructions for smolDB

## Project Overview
smolDB is a lightweight, document-oriented database with key-based O(1) access, storing human-readable JSON files on disk. It exposes a REST API and CLI for database operations. The codebase is primarily Go, with a Next.js web frontend in `web/`.

## Architecture & Key Components
- **main.go**: Entry point for the server and CLI commands.
- **api/**: REST API handlers and tests. All HTTP endpoints are defined here.
- **db/**: On-disk storage, locking, and document management.
- **index/**: Indexing, reference resolution, and write-ahead logging (WAL).
- **log/**: Logging utilities.
- **sh/**: Interactive shell implementation.
- **web/**: Next.js frontend (optional, not required for core DB functionality).

## Developer Workflows
- **Build**: Use `make build` to build the Go binary. For cross-platform builds, use `make build-all`.
- **Run Server**: `smoldb start` (default: port 8080, folder `db`). Use `--dir`/`-d` and `--port`/`-p` to customize.
- **Shell**: `smoldb shell` for interactive exploration. Same flags as server.
- **API Usage**: See `README.md` for endpoint examples (CRUD, index regeneration, reference resolution).
- **Tests**: Run `go test ./...` for backend tests. API and index logic are covered in `api/` and `index/`.

## Project-Specific Patterns
- **Reference Resolution**: Use `REF::<key>` in JSON to embed other documents. The API and index logic auto-resolve these references.
- **Human-Readable Storage**: All documents are plain JSON files in the `db/` directory. No binary formats.
- **No Advanced Queries**: Only key-based access; no sharding, distribution, or query language.
- **Index Regeneration**: Manual via `POST /regenerate` or automatic on write operations.
- **Error Handling**: API returns clear error messages for missing keys/fields (see `README.md` for examples).

## Integration Points
- **Docker**: Run with `docker run -p 8080:8080 nishank02/smoldb:latest` for quick API access.
- **Web Frontend**: `web/` is a Next.js app for UI, but not required for DB operation.

## Conventions & Examples
- **Endpoints**: All API endpoints and expected responses are documented in `README.md`.
- **Flags**: CLI flags for directory/port are consistent across server and shell.
- **Testing**: Use test helpers in `api/test_helpers.go` and `index/test_helpers.go` for setup/teardown.

## Key Files & Directories
- `main.go`, `api/`, `db/`, `index/`, `sh/`, `README.md`
- For frontend: `web/`

---

If any section is unclear or missing important details, please provide feedback to improve these instructions.