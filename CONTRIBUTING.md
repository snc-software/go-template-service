# Contributing

## Before you open a pull request

```bash
make ci     # vet + lint + test
make docs   # if you touched a swag annotation
```

CI runs the same checks plus `govulncheck`, a secret scan, and a guard that
fails if `docs/` has drifted from the annotations. All of them can be run
locally from the `Makefile`.

## Conventions

Read [ARCHITECTURE.md](ARCHITECTURE.md) first — it explains the layering and
carries the step-by-step checklist for adding an endpoint or a resource. The
rules that are enforced rather than merely encouraged live in
[.standards/go-service.md](.standards/go-service.md) and `.golangci.yml`.

The short version:

- Wrap errors with `%w` and context at every boundary; inspect with
  `errors.Is`/`errors.As`. Never discard a cause, never replace a handed-down
  error with a freshly constructed one.
- `ctx context.Context` is the first parameter of anything doing I/O, and is
  propagated rather than created below the transport layer.
- Dependencies arrive through constructors. `run()` in `cmd/api/main.go` is the
  only place anything is wired together.
- Every exported symbol gets a doc comment starting with its own name — this is
  enforced. Beyond that, comment what a thing *is* or *does* where that is not
  obvious; do not comment why you picked one approach over another, which
  belongs in the pull request.

## Commits and branches

Branch off `main`. Keep the subject line imperative and under ~72 characters.
Squash-merge.

## Dependencies

New direct dependencies need a reason in the pull request description: what it
does, why the standard library is not enough, and how actively it is maintained.
This is a template — every dependency added here is inherited by every service
generated from it.
