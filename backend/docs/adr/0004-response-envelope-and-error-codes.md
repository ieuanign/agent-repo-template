# One response envelope, with global and module error codes

Every response leaves backend as `{meta, status, code, message, payload}`, so web and mobile parse one shape and read every failure from the same place. A middleware at the boundary writes the envelope; handlers return typed payload objects and never set a status code. `api/openapi.yaml` describes each operation's `payload`, and its description states the envelope. The spec lists the error codes in its `ErrorCode` schema.

Errors come from two levels. `internal/platform/apperr` defines the global kinds — `NOT_FOUND`, `UNAUTHORIZED`, `FORBIDDEN`, `VALIDATION`, `CONFLICT`, `RATE_LIMITED`, `TIMEOUT`, `INTERNAL` — which any module may return as they are. A module declares its own errors in `model/errors.go`, each refining one kind under a code prefixed with the module's name (`PROJECTS_PROJECT_LOCKED`). One mapper in `internal/api` turns a kind into an HTTP status, so module code never names one, and a module split into a gRPC service maps the same kinds to gRPC codes. Anything that is not an `apperr` error becomes a 500 `INTERNAL` with a generic message, and its detail goes only to the log.

## Considered Options

- **Plain JSON bodies with RFC 9457 `application/problem+json` errors** — rejected in favour of one shape for every response, success or failure. It would have kept the spec describing the exact wire body.

## Consequences

- `code` is null on success. `message` is English, for developers; web and mobile show their own localised text chosen by `code`.
- The spec describes `payload`, not the wire body, so clients read `payload` themselves; `packages/api-client` wraps each payload type in the envelope type.
- Success messages come from one table keyed by operation, and a test fails when a mounted operation has none.
