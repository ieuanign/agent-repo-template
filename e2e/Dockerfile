FROM mcr.microsoft.com/playwright:v1.63.0-noble
WORKDIR /repo
# Manifests only: Compose mounts the specs and tokens. Without the tokens manifest npm ci skips its link.
COPY package.json package-lock.json ./
COPY e2e/package.json e2e/
COPY packages/design-tokens/package.json packages/design-tokens/
RUN --mount=type=cache,target=/root/.npm npm ci --workspace e2e
COPY e2e/playwright.config.ts e2e/
WORKDIR /repo/e2e
# Playwright's binary directly: npm would swallow the SIGTERM that init forwards.
CMD ["/repo/node_modules/.bin/playwright", "test"]
