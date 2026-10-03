FROM public.ecr.aws/docker/library/node:24.21.0-bookworm-slim@sha256:0e0ff40c39bc087845bfb27465a0df4ea419520094bc35842ff83dd8cbe6f9b6 AS dev
WORKDIR /repo
# Manifests only: Compose mounts the source. Without a workspace's manifest npm ci skips its link.
COPY package.json package-lock.json ./
COPY web/package.json web/
COPY packages/api-client/package.json packages/api-client/
COPY packages/design-tokens/package.json packages/design-tokens/
COPY packages/i18n/package.json packages/i18n/
RUN --mount=type=cache,target=/root/.npm npm ci
WORKDIR /repo/web
# Next's binary directly: npm would swallow the SIGTERM that init forwards.
CMD ["/repo/node_modules/.bin/next", "dev"]

FROM public.ecr.aws/docker/library/node:24.21.0-bookworm-slim@sha256:0e0ff40c39bc087845bfb27465a0df4ea419520094bc35842ff83dd8cbe6f9b6 AS prune
WORKDIR /repo
COPY package.json ./
# Read from the root manifest so the prune runs the turbo the pruned lock file installs.
RUN npm install -g "turbo@$(node -p "require('./package.json').devDependencies.turbo")"
COPY . .
# Outside the copied tree, so a stray host out/ cannot mix in.
RUN turbo prune web --docker --out-dir /pruned

FROM public.ecr.aws/docker/library/node:24.21.0-bookworm-slim@sha256:0e0ff40c39bc087845bfb27465a0df4ea419520094bc35842ff83dd8cbe6f9b6 AS install
WORKDIR /repo
COPY --from=prune /pruned/json/ ./
COPY --from=prune /pruned/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci

FROM install AS build
COPY --from=prune /pruned/full/ ./
ARG VERSION=dev
ARG GIT_SHA=dev
# public/ may not exist, and the runner's COPY needs a source.
RUN mkdir -p web/public && node_modules/.bin/turbo run build --filter=web

FROM gcr.io/distroless/nodejs24-debian12:nonroot@sha256:14d42e2511532589a7c7e01a753667a74fcc96266e137e8125006b87b0c32d0a
COPY --from=build /repo/web/.next/standalone/ /
COPY --from=build /repo/web/.next/static /web/.next/static
COPY --from=build /repo/web/public /web/public
ARG VERSION=dev
ARG GIT_SHA=dev
# Keep-alive outlasts Traefik's 90 s idle timeout, so Traefik never reuses a closed connection.
ENV HOSTNAME=0.0.0.0 \
    PORT=3000 \
    KEEP_ALIVE_TIMEOUT=95000 \
    VERSION=$VERSION \
    GIT_SHA=$GIT_SHA
# Full node path: distroless has no shell and /nodejs/bin is not on PATH.
HEALTHCHECK --interval=5s --timeout=5s --start-period=30s --start-interval=1s --retries=3 \
  CMD ["/nodejs/bin/node", "-e", "fetch('http://127.0.0.1:3000/health/ready').then((r) => process.exit(r.status === 200 ? 0 : 1), () => process.exit(1))"]
# The image's node entrypoint runs the server directly, so SIGTERM reaches Next's drain.
CMD ["/web/server.js"]
