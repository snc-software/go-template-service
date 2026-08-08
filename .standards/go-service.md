# Standard: Go HTTP services

**Applies to**: every Go HTTP service in this organisation, and to
`go-template-service`, which exists to make this standard the default rather
than a document people are asked to remember.

**Status**: active. Derived from Effective Go, Go Code Review Comments, the
`context`/`database/sql`/`log/slog`/`net/http` package documentation, and
RFC 7807.

Rules marked **(enforced)** fail the build via `.golangci.yml` or CI. The rest
are review criteria.

---

## 1. Package layout

1.1 Package by feature, not by layer. One resource is one package containing its
handlers, service, repository, types, and mapping. Adding a resource adds a
directory.

1.2 Everything that is not `main` lives under `internal/`. Import privacy should
be a property of the tree, not a convention.

1.3 Cross-cutting packages live under `internal/platform/` and are named for
what they provide (`httpx`, `apperr`, `config`, `database`), never for the layer
they occupy. Anything that owns an HTTP surface of its own is a slice, not
platform — health probes included.

1.4 No package named `util`, `utils`, `common`, `helpers`, `shared`, or
`models`. A package name that does not narrow what is inside it is a bag.

1.5 Package names are lowercase, single-word, and part of the identifier at the
call site. Avoid stuttering: `apperr.Error`, not `apperr.AppError`.

1.6 `cmd/<binary>/main.go` is the only composition root. It is the one file
allowed to know about every package.

## 2. Errors

2.1 **(enforced — `errorlint`)** Wrap with `%w` and inspect with
`errors.Is`/`errors.As`. Never compare errors with `==`.

2.2 Every wrap adds context naming the operation and its subject:
`fmt.Errorf("select template %s: %w", id, err)`. Lowercase, no trailing
punctuation, no "failed to" prefix.

2.3 **(enforced — `errcheck`)** Never silently drop an error. An error you
genuinely intend to ignore is assigned to `_` explicitly, so the intent is
visible.

2.4 One error type crosses layer boundaries. It carries a machine-readable code,
a human message, the HTTP status, and a wrapped cause. The cause is logged and
never serialised.

2.5 Driver and storage errors are translated into package sentinels at the
repository boundary. Layers above the repository never import the database
driver.

2.6 A handler passes on the error it was given. Constructing a new error to
replace one you were handed destroys the chain and is the single most common way
a 500 becomes a silent 404.

2.7 One error media type per API: `application/problem+json` (RFC 7807). The
`status` in the body equals the status line. Validation failures add a per-field
array; they do not get a second shape.

## 3. Context

3.1 `ctx context.Context` is the first parameter of every function that performs
I/O or can block.

3.2 Never store a context in a struct.

3.3 Below the transport layer, contexts are propagated, never created.
`context.Background()` appears in `main` and in shutdown paths only.

3.4 Anything with an external deadline gets one — readiness probes included.

## 4. HTTP

4.1 Never use `http.ListenAndServe`. Construct `http.Server` with
`ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, and `IdleTimeout` set
explicitly. **(enforced — `gosec` G114)**

4.2 Handle `SIGINT` and `SIGTERM` with `signal.NotifyContext` and drain with
`server.Shutdown`. The shutdown grace period must exceed the per-request
timeout.

4.3 Every request body is read through `http.MaxBytesReader`. Decoders set
`DisallowUnknownFields`, and reject trailing content after the first JSON value.

4.4 Response bodies are encoded into a buffer before `WriteHeader` is called. An
encoding failure must be able to change the status code.

4.5 `204` responses carry no body and no `Content-Type`.

4.6 Pagination parameters are validated and clamped, and the effective values
are echoed in the response. An unparseable parameter is a `400`, not a silent
default. Paged queries carry a total ordering, or rows move between pages.

4.7 Every service exposes a liveness probe (touches nothing) and a readiness
probe (checks dependencies with a short timeout) on separate paths. Conflating
them causes restart loops during dependency blips. Probes are excluded from
request logging.

## 5. Logging

5.1 `log/slog` with the JSON handler. No third-party logger, no `fmt.Println`,
no `log.Printf`.

5.2 The logger is constructed in `main` and injected. Level comes from
configuration.

5.3 One structured line per completed request: method, path, status, bytes,
duration.

5.4 Failures are logged once, at the outermost layer that handles them. Logging
and returning the same error produces the same incident twice.

5.5 Only 5xx responses are logged as errors. A 404 is a normal outcome.

5.6 Any type that can hold a credential implements `slog.LogValuer` to redact
it. Secrets must be unloggable by construction, not by remembering.

## 6. Configuration

6.1 Configuration is read from the environment exactly once, at startup, in one
package. Nothing else calls `os.Getenv`.

6.2 Loading fails fast and reports **every** missing required variable at once.

6.3 Security-relevant settings default to the secure value. `sslmode` defaults
to `require`; `disable` appears only in local sample files, marked as such.

6.4 Sample files carry placeholders, never working credentials. Real
environment files are gitignored, and CI scans for leaked secrets.

## 7. Persistence

7.1 The pool is opened once in `main` and injected. `sql.DB` is a pool, not a
connection — never open one per request.

7.2 Set `SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`, and
`SetConnMaxIdleTime` explicitly. The defaults are unbounded.

7.3 Every query takes a context and uses the `…Context` variant.

7.4 The database owns row timestamps (`DEFAULT now()`, `INSERT … RETURNING`).
Application clocks disagree with each other; a database clock does not disagree
with itself.

7.5 Existence is proved by the write itself — `RowsAffected`, or
`RETURNING` — never by a read before the write. A read-then-write is a race.

7.6 `github.com/jackc/pgx/v5` is the PostgreSQL driver. `lib/pq` is in
maintenance mode.

## 8. Dependency injection

8.1 Dependencies arrive through constructors. No package-level mutable state, no
service locator, no DI framework.

8.2 Interfaces are declared by the consumer, next to the code that needs them,
listing only the methods that code calls — not next to the implementation, and
never one-per-implementation. Go satisfies interfaces structurally; an
implementation does not name the interface it satisfies. Do not prefix interface
names with `I`.

8.3 Do not add an interface without a second implementation or a test double
that needs it. An interface between two types inside one package that always
ship together is indirection with nothing behind it.

## 9. API contracts

9.1 The only types with `json` tags are the contract types. Domain types and
persistence types are never serialised directly.

9.2 Translation between contract and domain is written out field by field.
Struct conversion silently publishes any field later added to the domain type.
**(`staticcheck` S1016 is excluded for `mapper.go` alone, and enabled
everywhere else.)**

9.3 Request validation is declarative on the contract type and produces
per-field errors that name the JSON field, not the Go field.

9.4 The generated OpenAPI document is OpenAPI 3.1, committed, embedded into the
binary, and CI fails if it has drifted from the annotations.

## 10. Testing

10.1 Handlers are tested through `httptest` against the real router, so
middleware, routing, and status codes are covered.

10.2 Services are tested against a fake repository returning the package
sentinels. No database required.

10.3 Repositories are tested against a real PostgreSQL — a container or a CI
service — behind an `integration` build tag. Mocking SQL tests the mock.

10.4 Every error branch that maps to a distinct HTTP status has a test asserting
that status.

10.5 `go test -race` in CI, always.

## 11. Toolchain

11.1 `.golangci.yml` is committed and CI-enforced. Baseline: `errcheck`,
`errorlint`, `govet`, `staticcheck`, `revive`, `gosec`, `bodyclose`,
`sqlclosecheck`, `rowserrcheck`, `ineffassign`, `unused`, `gocritic`,
`misspell`, `nolintlint`.

11.2 Every tool version is pinned in one place and shared by the `Makefile` and
CI, so local and CI runs produce identical output.

11.3 `go.mod` declares a minor version (`go 1.26`) with a separate `toolchain`
line.

11.4 CI runs, at minimum: build, `gofmt` check, `go mod tidy` check, vet, lint,
race tests, `govulncheck`, generated-docs staleness, and a secret scan.

11.5 Containers build `CGO_ENABLED=0` with `-trimpath`, and run non-root from a
distroless or scratch base.

## 12. Comments

12.1 **(enforced — `revive` exported, `staticcheck` ST1020)** Every exported
symbol has a doc comment, and it starts with the symbol's name. This is what
`go doc` and pkg.go.dev render; a package whose exported surface is undocumented
is not finished.

12.2 Do not comment why one approach was chosen over another. That belongs in
the pull request and in this document; in code it rots at the first refactor.

12.3 Non-obvious constraints a reader cannot derive from the code *do* get a
comment — coupling between two constants, or a deliberate re-panic.
