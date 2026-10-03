# backend conventions

The Go API: a modular monolith and the only PostgreSQL client. Decisions with their reasons are in
[`docs/adr/`](./docs/adr/); system-wide ones in the root [`docs/adr/`](../docs/adr/).

## Libraries (hard rules)

| Concern | Use | Never |
| --- | --- | --- |
| HTTP | gin, with oapi-codegen's `gin-server` + `strict-server` | chi, echo, net/http routing by hand |
| Database | sqlx over the `pgx/v5/stdlib` driver, hand-written SQL | an ORM, lib/pq |
| Migrations | golang-migrate | goose, atlas, migrations in Go code |
| Config | caarlos0/env | viper, config files |
| Logging | zap | slog, logrus, `fmt.Print*` / `log.*` |
| Request validation | `oapi-codegen/gin-middleware` against `api/openapi.yaml` | hand-written shape checks in handlers |
| Hot reload | air, in the `dev` Docker target only | |

Go tools (oapi-codegen, air, the golang-migrate CLI) are pinned as `tool` directives in `go.mod` and
run with `go tool …`. Tools that are not Go modules (squawk) come from devbox.

**Never copy code from the proprietary Spenmo go-sdk.** Ideas may be reimplemented from scratch.

## Layout

```
cmd/server/         # composition root: config, logger, pool, ai connection, modules, gin;
                    # `server healthcheck`
cmd/migrate/        # golang-migrate over the embedded migrations, with the guards below
cmd/seed/           # dev-only: every module's seeder, in wiring order
api/openapi.yaml    # the HTTP contract, written by hand
gen/                # generated, never hand-edited
  ai/v1/            # buf, from ../proto
  openapi/          # oapi-codegen, from api/openapi.yaml
migrations/         # <timestamp>_<module>_<desc>.{up,down}.sql, embedded by migrations.go
scripts/            # lint.sh, and migrationlint: guards 6–8 over migrations/
internal/platform/  # infrastructure, never business logic
  ai/               # the one gRPC connection to ai, and the startup report
  apperr/           # global error kinds and the error type modules refine
  buildinfo/        # version and commit, stamped with -ldflags at build time
  config/           # the root Config, parsed from the environment
  database/         # sqlx pool
  logger/           # zap, and the request-id middleware
internal/api/       # composes module handlers into the generated interface; envelope,
                    # error mapper, health, docs
internal/bff/       # endpoints composing two or more modules' public reads, one file each
internal/modules/<m>/
  module.go         # constructor; hands its handler to the composition root
  public.go         # the interface other modules and bff may call
  model/            # the module's types and errors — shape only, importable anywhere
  internal/         # private to the module, enforced by the compiler
    handler/        # this module's slice of the generated strict-server interface
    service/        # business logic
    repository/     # SQL on sqlx, only against this module's schema
    seed/           # the module's dev seed data, written through its service
```

## Modules

- Inside a module the flow is `handler → service → repository`. Repository rows map to `model` types
  at the repository boundary; generated API types map at the handler boundary.
- Cross-module calls go through the other module's `public.go`, never into its `internal/`.
- A module's SQL touches only its own PostgreSQL schema, named after the module
  ([ADR 0002](./docs/adr/0002-schema-per-module.md)). Foreign keys across schemas are allowed;
  joins are not.
- An endpoint needing two modules' data is a `bff/` handler composing their public reads. A module
  calls another directly only where a write invariant needs it, with the caller declaring the narrow
  interface it needs ([ADR 0001](./docs/adr/0001-modular-monolith-module-layout.md)).
- A `bff/` handler is named after a resource, not a screen. Public reads are batch-shaped
  (`[]id` in, map out) so composing never becomes one call per row.
- Wiring is by hand in `cmd/server/main.go`: no DI framework.
- Health, docs and auth are platform concerns, not modules.

## API contract

- `api/openapi.yaml` is the single spec; `servers` is `/api` and every route is mounted under it,
  because Traefik forwards the prefix unstripped.
- `internal/api` embeds one interface per module plus a `BFFAPI` into the struct that satisfies the
  generated `StrictServerInterface`.
- Every request is validated against the spec before a handler runs; a failure is `VALIDATION`.
  Services still enforce business rules.
- Every response is `{meta, status, code, message, payload}`, written by the envelope middleware.
  Handlers return typed payload objects and never set a status code. `204` and `304` carry no body.
- A new operation needs a success-message entry keyed `"METHOD /pattern"`; a test fails without it.
- `/api/docs` (Swagger UI, embedded and pinned — no CDN) and `/api/openapi.json` are served only when
  `APP_ENV` is `dev` or `staging`; production answers 404.
- No CORS: web and Swagger UI share the origin through Traefik, and mobile does not need it.

## Errors

- Global kinds live in `internal/platform/apperr`: `NOT_FOUND`, `UNAUTHORIZED`, `FORBIDDEN`,
  `VALIDATION`, `CONFLICT`, `RATE_LIMITED`, `TIMEOUT`, `INTERNAL`. Return them directly when nothing
  more specific applies.
- A module's own errors live in its `model/errors.go`, each refining one kind with a code prefixed by
  the module: `apperr.New(apperr.Conflict, "PROJECTS_PROJECT_LOCKED", "The project is locked")`.
- Wrap on the way up: `fmt.Errorf("approving the estimate: %w", err)`.
- Exactly one kind→HTTP mapping point, in `internal/api`. Anything that is not an `apperr` error is a
  500 `INTERNAL` with a generic message; `err.Error()` never reaches a client.
- `message` is English, for developers; clients localise by `code`
  ([ADR 0004](./docs/adr/0004-response-envelope-and-error-codes.md)).
- A module's new code joins the `ErrorCode` enum in `api/openapi.yaml` in the same change.

## Calling ai

- `platform/ai` owns the one connection to ai; the composition root passes it to the modules that
  need it.
- A module builds its own generated `ai.v1.<Module>Service` stub from that connection and wraps it in
  its own `internal/` package, behind an interface its service declares.
- Calls pass the request context, so the deadline and the request ID travel with them.
- The connection sets a 10 s deadline when the caller has none; a caller's own deadline is kept.
- Only `UNAVAILABLE` is retried: at most 3 attempts, backoff starting at 100 ms and doubling.
- Tests use a fake over `bufconn`, never a real ai.
- Mapping gRPC status codes to error kinds arrives with the first AI-backed feature.

## Data and migrations

- golang-migrate, forward-only in production
  ([ADR 0003](./docs/adr/0003-forward-only-migrations.md)). **A migration must never break the
  previous release's image**: a rename, drop, type change or new `NOT NULL` ships over two releases —
  add the new shape beside the old first, remove the old in a later release.
- One flat `migrations/` folder. Name: `<timestamp>_<module>_<desc>.{up,down}.sql`; a module's first
  migration creates its schema. Every file wraps its statements in `BEGIN; … COMMIT;`, because
  golang-migrate does not.
- Every write touching more than one row runs in one transaction.

Guards — each lives in the binary when a person could bypass `make`, in lint when it concerns files:

1. `down` and `redo` refuse unless `APP_ENV=dev`, checked by the `migrate` binary itself.
2. `force` needs `version=N` and `confirm=<database name>`, and prints the current version and dirty
   flag before acting.
3. A dirty database stops `migrate up` with the exact `make migrate-force` command to run, and stops
   the server from starting.
4. At boot the server compares the database's version with its embedded migrations: older refuses to
   start; newer (a rollback) starts with a warning. `migrate up` treats newer as nothing to do.
5. Concurrent runs are serialised by golang-migrate's PostgreSQL advisory lock.
6. Lint fails a new migration whose timestamp is not newer than the latest on `main` — golang-migrate
   silently skips one older than the current version.
7. Lint requires a `.down.sql` for every `.up.sql`, the name pattern with an existing module, and
   `BEGIN`/`COMMIT`.
8. squawk flags renames, drops, type changes and new `NOT NULL`s; a flagged statement needs an
   explicit ignore comment marking it as the second, removing release, on its own line directly
   above it: `-- squawk-ignore <rule>[, <rule>…] -- second release: <why the old shape can go>`.
9. `make rollback` swaps the image only and never runs a migration.

## Seed data

- Every module ships a seeder in `internal/seed/`, exposed as `Module.Seed(ctx)` and called only by
  `cmd/seed`. A feature is not done until its tables have seed data.
- Seeders write through the module's own service, so seeded data passes the same rules as real
  requests; data from another module comes through its `public.go`.
- Fixed natural keys (account code `DEMO`), skipping what already exists, so a re-run adds nothing.
  Fake data only: no real names, NIKs or plates.
- `cmd/seed` is built only into the `dev` image and also refuses unless `APP_ENV=dev`. A test fails
  when a directory under `internal/modules/` has no registered seeder.

## Configuration

caarlos0/env parses one root `Config` in `cmd/server`, composed of the platform's and each module's
sub-configs; boot fails listing every missing or invalid variable at once.

| Variable | Notes |
| --- | --- |
| `APP_ENV` | Required: `dev`, `staging` or `production` |
| `PORT` | Default `8080` |
| `POSTGRES_HOST`, `POSTGRES_PORT` | Set in Compose; port defaults to `5432` |
| `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` | From `secrets/postgres.env`, the same file the PostgreSQL container reads |
| `AI_HOST` | Required; set to `ai` in Compose |
| `AI_PORT` | Default `50051`, ai's own default |

Secret values live only in `secrets/*.sops.env`, split by who reads them; plain settings never go
there.

## Logging

zap from `platform/logger`: console output in dev, JSON otherwise. A gin middleware reads or assigns
`X-Request-ID`, echoes it, and puts a child logger carrying it in the request context; handlers and
services take the logger from there. It also stores the ID itself, read with
`logger.RequestIDFromContext`; calls to ai send it as `x-request-id`. Never log secrets or personal
data.

## Runtime

- `GET /api/health`: liveness, always 200 with `{version, environment}`, no dependency checked.
- `GET /api/health/ready`: readiness, 200 or 503 from PostgreSQL, naming nothing. ai is not a
  readiness dependency: an ai outage degrades AI features, not the whole API.
- `server healthcheck` calls readiness on localhost and exits 0 or 1 — the distroless image has no
  curl, and the zero-downtime deploy waits on it.
- The startup log line carries the service, version and commit.
- After it, a background startup report calls ai's `GetInfo` (up to 30 s) and logs ai's service and
  version, or a warning when ai is unreachable. Boot never waits on it or fails because of it.
- On `SIGTERM`: stop accepting, drain in-flight requests for up to 20 s, close the ai connection,
  then close the pool. Compose's `stop_grace_period` is 30 s.
- Server timeouts: 5 s to read headers, 30 s to read or write, 120 s idle.
- Nothing is held in process memory between requests: two copies run side by side during a deploy.
  A background job, when one exists, takes a PostgreSQL advisory lock so only one copy runs it.

## Build

`Dockerfile`, built from `backend/`:

- Build stage: `--platform=$BUILDPLATFORM golang:1.27-alpine` through the public ECR mirror
  (`public.ecr.aws/docker/library/`), which avoids Docker Hub's pull limits; `CGO_ENABLED=0`, so no C
  library is needed. `-ldflags -X` stamps `VERSION` and `GIT_SHA` into `internal/platform/buildinfo`
  (both default to `dev`).
- `dev` target: air over the bind-mounted source, plus the `seed` binary.
- Production target: `gcr.io/distroless/static-debian12:nonroot` with `server` and `migrate`.
- Every image reference carries its tag and its multi-arch index digest, so a rebuild never changes
  base silently. A bump changes both together, the digest read from
  `docker buildx imagetools inspect <image:tag>`.

## Tests and lint

- Colocated `foo_test.go`, table-driven. Mock through narrow, caller-declared interfaces.
- SQL is tested with go-sqlmock and handlers with `httptest`; **no test touches a real database**.
  End-to-end tests belong to web (Playwright) and mobile (Maestro).
- Lint: `gofmt` check, `go vet`, oapi-codegen regenerate-and-diff (stale `gen/openapi` fails like
  stale proto code does), squawk, and the migration-file checks above.
- Fix (`lint:fix`): `gofmt -w .`, the same scope as lint's `gofmt -l .`. The pre-commit hook runs
  it before its gate.
