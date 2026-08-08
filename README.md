# go-template-service

Template service for writing APIs with Go.

The example resource is called **Template** throughout. To start a new service, 
find-and-replace `Template` / `template` with your resource name and rename the
matching files.

## Stack

| Concern     | Choice                        |
| ----------- | ----------------------------- |
| Router      | `go-chi/chi/v5` |
| Database    | PostgreSQL via `jmoiron/sqlx` |
| Driver      | `lib/pq` |
| Migrations  | `pressly/goose` (CLI)         |
| API docs    | `swaggo/swag` (Swagger 2.0)   |
| Config      | `joho/godotenv` |

## Layout

```
main.go                     composition root, router wiring, swagger annotations
routes/                     HTTP handlers
  contracts/                request/response DTOs (the public API surface)
  middleware/               panic recovery
mapping/                    contract <-> domain translation
domain/
  models/                   domain types
  services/                 business logic
persistence/
  entities/                 database row types
  readers/                  queries
  writers/                  commands
  migrations/               goose SQL migrations
exceptions/                 typed application errors
utils/                      JSON response helpers
docs/                       generated swagger output (do not edit by hand)
```

## Running locally

Requires Go 1.26+, PostgreSQL, and [goose](https://github.com/pressly/goose).

```bash
# 1. create the database
createdb template_service

# 2. apply migrations
cd persistence/migrations && goose -env=goose.env up && cd -

# 3. run
go run .
```

Swagger UI: http://localhost:8080/swagger/index.html

## Configuration

`.env` holds committed non-secret defaults. `.env.local` overrides it and is
gitignored — put your local password there:

```
DB_PASSWORD=postgres
```

| Variable      | Default          |
| ------------- | ---------------- |
| `DB_HOST` | `localhost` |
| `DB_PORT` | `5432` |
| `DB_NAME` | `template_service` |
| `DB_USER` | `postgres` |
| `DB_PASSWORD` | —                |

## Endpoints

| Method   | Path               | Description        |
| -------- | ------------------ | ------------------ |
| `GET` | `/templates` | Paged list         |
| `GET` | `/templates/{id}` | Fetch by ID        |
| `POST` | `/templates` | Create             |
| `DELETE` | `/templates/{id}` | Delete by ID       |

## Regenerating API docs

```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init
```

## Adding a migration

```bash
cd persistence/migrations
goose create <name> sql
```
