# go-template-service

Template service for writing APIs with Go.

The example resource is called **Template** throughout. To start a new service,
copy `internal/template/` to `internal/<resource>/`, find-and-replace
`Template` / `template` with your resource name, and mount it in
`cmd/api/main.go`.

## Stack

| Concern     | Choice                        |
| ----------- | ----------------------------- |
| Router      | `go-chi/chi/v5` |
| Database    | PostgreSQL via `jmoiron/sqlx` |
| Driver      | `lib/pq` |
| Migrations  | `pressly/goose` (CLI)         |
| API docs    | `swaggo/swag` (Swagger 2.0)   |
| Config      | `joho/godotenv` |
| Validation  | `go-playground/validator/v10` |
| Logging     | `log/slog` (stdlib)           |

## Layout

One resource is one package. Adding a resource means adding a directory, not
editing nine.

```
cmd/
  api/main.go               composition root: config, logger, pool, router, server, shutdown
internal/
  template/                 the example resource, end to end
    handler.go              HTTP handlers + swagger annotations
    service.go              business logic; declares the Store interface it needs
    store.go                PostgreSQL queries
    model.go                domain types
    contract.go             request/response DTOs (the public API surface)
    mapper.go               contract <-> domain translation
  platform/                 cross-cutting, named for what it provides
    apperr/                 typed application errors carrying code + HTTP status
    httpx/                  response writing, request decoding, middleware
    config/                 environment loading and validation
    database/               connection pool construction and tuning
migrations/                 goose SQL migrations
docs/                       generated swagger output (do not edit by hand)
```

## Running locally

Requires Go 1.26+, PostgreSQL, and [goose](https://github.com/pressly/goose).

```bash
# 1. configure
cp .env.sample .env.local        # .env.local is gitignored; put your password there
cp migrations/goose.env.sample migrations/goose.env

# 2. create the database
createdb template_service

# 3. apply migrations
cd migrations && goose -env=goose.env up && cd -

# 4. run
go run ./cmd/api
```

Swagger UI: http://localhost:8080/swagger/index.html

## Configuration

Configuration comes from the environment. `.env` is loaded first, then
`.env.local` overrides it. Both are gitignored — neither is ever committed.
Startup fails with the full list of anything required and missing.

| Variable      | Required | Default            |
| ------------- | -------- | ------------------ |
| `PORT` | no | `8080` |
| `LOG_LEVEL` | no | `info` |
| `DB_HOST` | yes | — |
| `DB_PORT` | yes | — |
| `DB_NAME` | yes | — |
| `DB_USER` | yes | — |
| `DB_PASSWORD` | no | — |
| `DB_SSLMODE` | no | `require` |

`DB_SSLMODE` defaults to `require`. `.env.sample` sets it to `disable` because
that is what a local PostgreSQL usually needs; do not carry that value into a
deployed environment.

## Endpoints

| Method   | Path               | Description        |
| -------- | ------------------ | ------------------ |
| `GET` | `/templates` | Paged list; `page` and `size`, size capped at 100 |
| `GET` | `/templates/{id}` | Fetch by ID        |
| `POST` | `/templates` | Create             |
| `DELETE` | `/templates/{id}` | Delete by ID       |

Errors use one media type across the whole API: `application/problem+json`
(RFC 7807), with a `status`, a stable machine-readable `code`, a human
`message`, and — for validation failures — a per-field `errors` array.

## Regenerating API docs

```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.6
swag init -g cmd/api/main.go --parseInternal -o docs
```

## Adding a migration

```bash
cd migrations
goose create <name> sql
```
