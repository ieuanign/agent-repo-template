# Independent Changesets versioning

backend, ai, web and mobile each carry their own version, bumped independently with Changesets, so a release of one service never forces a release of the others. `packages/design-tokens` and `packages/api-client` are private and unversioned: they ship only inside the services that consume them.

## Consequences

- mobile's Expo config reads its version from `mobile/package.json`, so Changesets is the single source of the app version.
- Every service is built with `version` and `commit` from the `VERSION` and `GIT_SHA` build args. Its public health check reveals what an operator needs and nothing an attacker could use; paths sit under the service's own prefix (backend's is `/api`):
  - `GET /health` is liveness. It answers 200 with `{version, environment}` while the process runs and checks no dependency, so an outage elsewhere never gets healthy containers restarted.
  - `GET /health/ready` is readiness. It answers 200 or 503 from the service's hard dependencies without naming them, so a proxy or orchestrator stops sending traffic without restarting anything.
  - `environment` is the service's `APP_ENV`: `dev`, `staging` or `production`. The commit appears only in the startup log line.
- A gRPC service reports its health through the standard `grpc.health.v1.Health` protocol rather than a custom RPC.
