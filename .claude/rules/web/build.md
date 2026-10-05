---
paths:
  - "web/Dockerfile"
  - "web/next.config.ts"
  - "web/turbo.json"
  - "web/package.json"
  - ".dockerignore"
---

# web: the image

A bare path is under `web/`; `compose.yaml` and `.dockerignore` are the root's. Links are relative to
this file.

- `Dockerfile` builds with the repository root as its context, under the root
  [`.dockerignore`](../../../.dockerignore).
- The `dev` target is `node:24.21.0-bookworm-slim`, pulled through the public ECR mirror
  (`public.ecr.aws/docker/library/`) and pinned by digest. It copies only the root manifests, web's,
  design-tokens', api-client's and i18n's, runs `npm ci` from the root lock under `/repo`, then runs
  Next's binary with `dev` from `/repo/web`. Compose mounts the source.
- The image is built natively per platform and never cross-built: Next.js and Tailwind ship native
  binaries per platform, and file events fail under emulation.
- `next.config.ts` lists `traefik` in `allowedDevOrigins`, because the e2e browser reaches web at
  `http://traefik:4008` and `next dev` refuses `/_next/*`, the HMR socket included, from any origin
  it does not list. It affects `next dev` only, never the standalone server, and names a hostname
  only the Compose network resolves.
- The production target is the last stage, so the default. Its builder stages use the same pinned
  `node:24.21.0-bookworm-slim`:
  - `prune` runs `turbo prune web --docker`, with turbo's version read from the root `package.json`'s
    `devDependencies.turbo`, never restated;
  - `install` runs `npm ci` against the pruned lock file only, never `npm install`;
  - `build` runs `turbo run build --filter=web`. [`turbo.json`](../../../web/turbo.json) declares `.next/**`,
    without `cache/` and `dev/`, as the outputs a cache hit restores.
- The runner is `gcr.io/distroless/nodejs24-debian12:nonroot`, pinned by digest: Debian 12 on both
  sides, so native packages traced into the standalone server run. It never uses `:debug` or
  overrides `USER`.
- The runner holds exactly three root-owned copies from the build stage: the standalone folder at `/`,
  and `.next/static` and `public` beside `/web/server.js`. No other `COPY`, no `RUN`, no `--chown`.
- Its ENV is exactly `HOSTNAME=0.0.0.0`, `PORT=3000`, `KEEP_ALIVE_TIMEOUT=95000`, `VERSION` and
  `GIT_SHA`. `VERSION` and `GIT_SHA` are build args defaulting to `dev`, with no turbo `env` entry,
  since nothing in `next build` reads them. No stage sets `APP_ENV`, a `NEXT_PUBLIC_*` variable or a
  secret: one image is promoted across environments. 95 000 ms outlasts Traefik's 90 s idle timeout.
- The `HEALTHCHECK` runs `/nodejs/bin/node -e` with `compose.yaml`'s `web` probe script, fetching
  `/health/ready`: distroless has no shell, and `/nodejs/bin` is not on its `PATH`. The zero-downtime
  deploy waits on it ([ADR 0005](../../../docs/adr/0005-deployment-topology.md)).
- The command is `/web/server.js` under the image's own node entrypoint, with no npm, `next start` or
  wrapper, so `SIGTERM` reaches Next.js's handler, which drains in-flight requests.
- The prune stage copies the whole context into a builder layer, so the root `.dockerignore` keeps
  out secrets, env files and gitignored local state. Never add a `Dockerfile.dockerignore`: it
  replaces the root rules.
- A base bump changes tag and digest together, the digest read from
  `docker buildx imagetools inspect <image:tag>`.
