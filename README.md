# go-template-service

Template service for writing APIs with Go.

The example resource is called **Template** throughout. To start a new service,
copy `internal/template/` to `internal/<resource>/`, find-and-replace
`Template` / `template` with your resource name, and mount it in
`cmd/api/main.go`.

See [ARCHITECTURE.md](ARCHITECTURE.md) for the layering and the step-by-step
checklist for adding an endpoint. The rules it embodies are written down in
[.standards/go-service.md](.standards/go-service.md).

## Stack

| Concern     | Choice                                        |
| ----------- | --------------------------------------------- |
| Router      | `go-chi/chi/v5`                               |
| Database    | PostgreSQL via `jmoiron/sqlx`                 |
| Driver      | `jackc/pgx/v5` (through its `stdlib` adapter) |
| Migrations  | `pressly/goose`                               |
| API spec    | `swaggo/swag` v2 (OpenAPI 3.1)                |
| API docs UI | [Scalar](https://scalar.com) via `scalar-go`  |
| Config      | `joho/godotenv`                               |
| Validation  | `go-playground/validator/v10`                 |
| Logging     | `log/slog` (stdlib)                           |

## Layout

One resource is one package. Adding a resource means adding a directory, not
editing nine.

```
cmd/
  api/main.go               composition root: config, logger, pool, router, server, shutdown
internal/
  template/                 the example resource, end to end
    endpoints.go            HTTP handlers + OpenAPI annotations
    service.go              business logic; declares the Repository interface it needs
    repository.go           PostgreSQL queries
    model.go                domain types
    contract.go             request/response DTOs (the public API surface)
    mapper.go               contract <-> domain translation
  health/                   liveness and readiness, a slice like any other
  platform/                 cross-cutting, named for what it provides
    apperr/                 typed application errors carrying code + HTTP status
    config/                 environment loading and validation
    database/               connection pool construction and tuning
    httpx/                  response writing, request decoding, middleware
    openapi/                serves the spec and the Scalar reference UI
migrations/                 goose SQL migrations
docs/                       generated OpenAPI document (do not edit by hand)
```

## Running locally

### With Docker

```bash
docker compose up --build
```

Brings up PostgreSQL, applies migrations, and starts the service on
<http://localhost:8080>.

### Without Docker

Requires Go 1.26+ and PostgreSQL.

```bash
cp .env.sample .env.local     # .env.local is gitignored; put your password there
createdb template_service
make migrate
make run
```

`make` reads `.env` and `.env.local` itself, so migrations and the service can
never disagree about which database they are pointed at.

Run `make` with no arguments to list every target.

## API documentation

| URL                              | What                                    |
| -------------------------------- | --------------------------------------- |
| <http://localhost:8080/docs>     | Scalar reference UI                     |
| <http://localhost:8080/openapi.json> | The generated OpenAPI document      |

The document is OpenAPI 3.1, generated from the annotations on the handlers and
embedded into the binary with `//go:embed`. After changing an annotation, run
`make docs` and commit the result — CI fails if `docs/` has drifted.

The Scalar page is built by [`bdpiprava/scalar-go`](https://github.com/bdpiprava/scalar-go)
in `internal/platform/openapi/openapi.go` — theme, layout, sidebar, and default
HTTP client are all options there, applied once at startup. Scalar's JavaScript
bundle still loads from a pinned jsDelivr URL; if your environment blocks
external CDNs, self-host the bundle and point `WithCDN` at it.

## Configuration

Configuration comes from the environment. `.env` is loaded first, then
`.env.local` overrides it. Both are gitignored — neither is ever committed.
Startup fails with the full list of anything required and missing.

| Variable      | Required | Default   |
| ------------- | -------- | --------- |
| `PORT`        | no       | `8080`    |
| `LOG_LEVEL`   | no       | `info`    |
| `DB_HOST`     | yes      | —         |
| `DB_PORT`     | yes      | —         |
| `DB_NAME`     | yes      | —         |
| `DB_USER`     | yes      | —         |
| `DB_PASSWORD` | no       | —         |
| `DB_SSLMODE`  | no       | `require` |

`DB_SSLMODE` defaults to `require`. `.env.sample` sets it to `disable` because
that is what a local PostgreSQL usually needs; do not carry that value into a
deployed environment.

## Endpoints

| Method   | Path              | Description                                       |
| -------- | ----------------- | ------------------------------------------------- |
| `GET`    | `/health`         | Liveness — 200 while the process is up            |
| `GET`    | `/ready`          | Readiness — 200 when the database is reachable    |
| `GET`    | `/templates`      | Paged list; `page` and `size`, size capped at 100 |
| `GET`    | `/templates/{id}` | Fetch by ID                                       |
| `POST`   | `/templates`      | Create                                            |
| `DELETE` | `/templates/{id}` | Delete by ID                                      |

Errors use one media type across the whole API: `application/problem+json`
(RFC 7807), with a `status`, a stable machine-readable `code`, a human
`message`, and — for validation failures — a per-field `errors` array.

```json
{
  "status": 400,
  "code": "VALIDATION_FAILED",
  "message": "request body failed validation",
  "errors": [{ "field": "email", "message": "must be a valid email address" }]
}
```

The template ships **unauthenticated**. Add a middleware in `run()` when you
need one.

## Paged responses

Each resource declares its own `PagedResponse` in `contract.go` over the shared
`httpx.Pagination`:

```go
type PagedResponse struct {
	Items      []Response       `json:"items"`
	Pagination httpx.Pagination `json:"pagination"`
}
```

This is four lines per resource rather than one shared generic because `swag`
cannot resolve a cross-package generic instantiation, and its `--parseDependency`
workaround mangles every schema name in the published document.

## Development

```bash
make lint     # golangci-lint
make test     # go test -race
make ci       # everything CI runs
make docs     # regenerate the OpenAPI document
make migrate-create name=add_widgets
```

Tool versions are pinned once at the top of the `Makefile` and reused by CI, so
local and CI runs produce identical output.
