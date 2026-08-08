# Architecture

## The shape

One resource is one package. `internal/template/` contains everything about
templates — HTTP, business rules, SQL, types — and nothing about anything else.
Adding a resource means adding a directory.

The alternative, grouping by layer (`handlers/`, `services/`, `repositories/`),
spreads one change across every directory in the tree and makes it impossible to
tell from the layout what the service actually does.

```
cmd/api/            composition root
internal/
  template/         a resource, end to end
  health/           a resource, end to end (liveness and readiness)
  platform/         cross-cutting machinery, named for what it provides
migrations/         goose SQL
docs/               generated OpenAPI document, embedded at build time
```

`health/` sits beside `template/` rather than under `platform/` because it is a
vertical slice like any other: it owns an HTTP surface, a response contract, and
a dependency. `platform/` is for machinery a slice *uses* — it holds nothing
that answers a request on its own.

`internal/` is not decoration: Go refuses to let anything outside this module
import it, so the entire tree below it is private by construction.

## Dependency direction

```
endpoints.go  ──>  service.go  ──>  Repository (interface)
                                          ^
                                          │ implements
                                    repository.go
```

Dependencies point inward and are supplied by constructors. Nothing constructs
its own dependencies, and nothing reaches for a package-level global. The single
place anything is wired together is `run()` in `cmd/api/main.go`.

`Repository` is declared in `service.go`, next to its consumer, not next to its
implementation. This is the part of Go that most surprises people arriving from
C# or Java, so it is worth stating plainly:

- **Go interfaces are satisfied structurally.** `PostgresRepository` never names
  `Repository` and does not need to. Anything with those four methods satisfies
  it, including a test fake that has never heard of the real one.
- **The consumer owns the requirement.** `Service` says what it needs; the
  interface lists four methods rather than everything `PostgresRepository`
  happens to expose. Declared next to the implementation it would inevitably
  grow to mirror that type, which is how you end up with `IFooRepository` that
  exists only so `FooRepository` can implement it.
- **Deleting the implementation leaves the interface meaningful.** Swapping
  PostgreSQL for anything else changes one file and one line of `main.go`.

The naming follows from that: the interface is `Repository` because from inside
`template` there is only one, and the implementation carries the qualifier —
`PostgresRepository`. Go does not prefix interfaces with `I`, and the shorter
name belongs to the abstraction.

## Why files are grouped by role, not by type

There is no `IRepository.cs`-style split here. A Go package is the unit of
encapsulation, not a file — every file in `internal/template` sees every other,
identifiers are private to the package unless capitalised, and the compiler does
not care which file anything lives in. Splitting an interface away from its
implementation would buy nothing and cost a jump.

So files are cut by **role in a request** — `endpoints` → `service` →
`repository`, with `contract`, `model`, and `mapper` holding the types each
speaks. That is also the order you read them in when tracing a bug.

There is deliberately **no** interface between `Endpoints` and `Service`. The
seam worth having is at the I/O boundary, where substitution buys isolation from
a real database. An interface inside a single package, between two types that
always ship together, is indirection with nothing on the other side of it.

## Layer responsibilities

| Layer            | Owns                                                           | Never does                                             |
| ---------------- | -------------------------------------------------------------- | ------------------------------------------------------ |
| `endpoints.go`   | Parse path/query, decode+validate the body, choose the status   | Business rules, SQL, constructing errors from strings   |
| `service.go`     | Business rules, translating storage errors into `apperr`        | Touch `http`, know about JSON, know about the driver    |
| `repository.go`  | SQL, translating driver errors into package sentinels           | Know about HTTP status codes                            |
| `contract.go`    | The public API surface — the only types with `json` tags        | Carry `db` tags or leak domain internals                |
| `mapper.go`      | Explicit translation between contract and domain                | Use struct conversion (see below)                       |

Mapping is written out field by field on purpose. A struct conversion would
compile happily and silently publish any field later added to the domain type —
which is how internal state ends up in a public response. `staticcheck`'s S1016
would suggest exactly that conversion, so it is excluded for `mapper.go` alone
in `.golangci.yml`, and stays enabled everywhere else.

## Errors

There is one error type crossing layer boundaries: `apperr.Error`. It carries a
stable machine-readable `Code`, a human `Message`, the HTTP `Status` it maps to,
optional per-field `Fields`, and a wrapped cause that is **logged and never
serialised**.

The chain runs outward, translating once per boundary:

```
pgconn.PgError 23505
  └─ repository:  fmt.Errorf("insert template: %w", ErrDuplicateEmail)
       └─ service:  apperr.Conflict("a template with email … already exists")
            └─ httpx: 409 application/problem+json
```

Two consequences worth keeping:

- The service never imports the driver. `ErrNotFound` and `ErrDuplicateEmail`
  are package sentinels, so a fake repository in a test can reproduce both
  without a database, and swapping drivers touches one function.
- Every response body in the API — success or failure — has exactly one error
  shape, `application/problem+json` (RFC 7807). There is no second error type.

Only 5xx responses are logged, once, in `httpx.Responder.Error`. A 404 is a
normal outcome, not an incident.

## Request lifecycle

```
RealIP ──> Recoverer ──┬──> /health, /ready                 (unlogged)
                       └──> RequestLogger ──> Timeout ──> route
```

- **Recoverer** turns a panic into a 500 problem document with the stack in the
  log, and re-panics on `http.ErrAbortHandler`, which is a deliberate abort
  rather than a failure.
- The probes sit outside the logging group so a 1-second orchestrator poll does
  not drown the log.

There is no authentication middleware. The template ships unauthenticated, and
a service that needs it adds a middleware here rather than inheriting a no-op
seam it has to understand before it can delete it.

## API documentation

`make docs` generates `docs/swagger.json` from the annotations on the handlers.
`docs/docs.go` embeds it with `//go:embed`, so the document a running service
serves is the one that was committed — there is no runtime spec assembly, and
CI fails if the committed file has drifted.

`internal/platform/openapi` renders the Scalar UI once at startup via
`bdpiprava/scalar-go` and serves the buffered HTML. A bad UI configuration is a
boot failure, not a broken page.

## Configuration

`config.Load` reads the environment once, at startup, and reports **every**
missing required variable at once rather than failing on the first. Nothing else
in the codebase calls `os.Getenv`. `config.Database` implements
`slog.LogValuer` so the password cannot reach a log line even if the whole
config is logged.

---

# Adding a new endpoint to an existing resource

1. `contract.go` — add the request/response types. Every field gets a `json`
   tag; request fields get `validate` tags.
2. `mapper.go` — add the explicit translation to and from the domain type.
3. `service.go` — add the method to the `Repository` interface, then implement
   the business rule, translating storage sentinels into `apperr`.
4. `repository.go` — add the SQL. Wrap every error with `%w` and context
   (`fmt.Errorf("update template %s: %w", id, err)`). Translate driver errors
   into package sentinels here and nowhere else.
5. `endpoints.go` — add the handler and register it in `Routes()`. Decode with
   `httpx.Decode[T]`, pass errors to `endpoints.responder.Error` untouched, and
   never construct a fresh error in place of the one you were handed.
6. Annotate the handler for swag, including every `@Failure` status it can
   produce, then `make docs`.
7. `make migrate-create name=<name>` if the schema changes.
8. `make lint test`.

# Adding a new resource

1. Copy `internal/template/` to `internal/<resource>/`.
2. Replace `Template`/`template` throughout, including the `"Templates"` table
   name and the `db` tags.
3. Declare that resource's own `PagedResponse` in its `contract.go` over the
   shared `httpx.Pagination` — see the note in the README about why this is not
   a shared generic.
4. Mount it in `run()` in `cmd/api/main.go`.
5. `make migrate-create name=create_<resource>`, then `make docs`.
