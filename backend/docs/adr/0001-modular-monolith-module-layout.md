# Modular monolith: one module per bounded context, cross-module reads in a BFF

backend is one Go process split into modules that can each become a service later without being untangled. A module lives in `internal/modules/<module>/`: `module.go` builds it, `public.go` declares the only interface other code may call, `model/` holds its types (shape only, importable anywhere), and `internal/{handler,service,repository}` holds the rest, which Go's `internal/` rule stops anything outside the module from importing. An endpoint that needs two modules' data is a handler in `internal/bff/` composing their `public.go` reads in Go, never a SQL join; one module calls another directly only where a write invariant needs it, through a narrow interface the caller declares.

## Considered Options

- **Flat layered packages (`internal/{handler,service,store}`)** — rejected: less ceremony, but nothing stops one domain reaching into another's internals, and extracting a service later means untangling those seams by hand.
- **One flat package per module** — rejected: exported names would mark the boundary, but a growing module's handlers, services and queries would share a single package.

## Consequences

- Cross-module dependencies exist only through `public.go`, so they are visible in review.
- Every module carries constructor and wiring code a flat layout would skip.
- A two-module endpoint lives in neither module; [`backend/CLAUDE.md`](../../CLAUDE.md) says to look in `internal/bff/`.
- A BFF handler is named after a resource, not a screen. A screen needing several independent resources calls several endpoints in parallel, so each loads on its own.
