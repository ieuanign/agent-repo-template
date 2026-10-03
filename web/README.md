# web

The Next.js web app, server-rendered with the App Router. Traefik routes `/` to it, beside backend
at `/api` ([ADR 0004](../docs/adr/0004-compose-topology.md)).

## Stack

| Concern     | Choice                                                                                                               |
| ----------- | -------------------------------------------------------------------------------------------------------------------- |
| Framework   | Next.js 16.3.6 with the App Router, on React 19.3                                                                    |
| Runtime     | Node.js 24                                                                                                           |
| Styling     | Tailwind CSS 4.3.3 through PostCSS, with [`@agent-repo-template/design-tokens`](../packages/design-tokens/README.md) |
| Fonts       | Inter through `next/font`, self-hosted at build                                                                      |
| Translation | next-intl 4.14.7, without its locale routing                                                                         |
| API client  | [`@agent-repo-template/api-client`](../packages/api-client/README.md), openapi-fetch 0.17.0 over backend's spec      |
| Config      | a hand-written env helper, from environment variables only                                                           |
| Logging     | the console                                                                                                          |
| Lint        | oxlint, Prettier and `tsc` with no emit                                                                              |
| Tests       | Jest through `next/jest`                                                                                             |
| Mocking     | MSW 2.15.0, answering with backend's full envelope                                                                   |

## Project structure

```
web/
├── Dockerfile              # the dev target and the production image, built from the repository root
├── messages/               # web's own id and en catalogues; shared text comes from @agent-repo-template/i18n
├── jest.setup.ts           # starts the MSW server for Jest
├── scripts/                # lint steps run by `turbo run lint`
└── src/
    ├── instrumentation.ts  # logs the startup line and starts the backend report
    ├── app/                # routes only: the root layout, pages, global CSS and the theme reset
    ├── ui/<View>/          # one view per route, mirroring app/
    ├── core/               # constants, env and utilities, the backend client and API_URLS
    └── mocks/              # the MSW server and its default handlers
```

`components/`, `lib/`, `context/` and `hooks/` are created under `src/` when they first hold something.

The rules behind this layout are in [`CLAUDE.md`](./CLAUDE.md).

## Commands

Run from the repository root inside `devbox shell`.

| Command                   | Does                                                                |
| ------------------------- | ------------------------------------------------------------------- |
| `make dev`                | Starts the stack, with web running `next dev` behind Traefik at `/` |
| `make down`               | Stops the stack and removes web's anonymous `node_modules` volume   |
| `make check`              | Lints and tests every workspace, then smoke-tests the stack         |
| `make e2e`                | Runs the browser suite against the stack `make dev` started         |
| `npm run lint -w web`     | oxlint, the Prettier check, then `next typegen` and `tsc --noEmit`  |
| `npm run lint:fix -w web` | oxlint's autofix, then Prettier's write mode                        |
| `npm test -w web`         | Runs Jest                                                           |

## Production image

Nothing in `make dev` builds it. Build and run it by hand from the repository root, which is the
build context:

```sh
docker build -f web/Dockerfile \
  --build-arg VERSION=<version> --build-arg GIT_SHA=$(git rev-parse --short HEAD) \
  -t <tag> .
docker run --init -e APP_ENV=<environment> -p 127.0.0.1:3000:3000 <tag>
```

- The production image is the Dockerfile's default target, so no `--target` is needed.
- `VERSION` and `GIT_SHA` default to `dev`. `APP_ENV` comes with the container, never the build, so
  one image is promoted across environments.
- The server listens on port 3000 and answers `GET /health` once started. The image's
  `HEALTHCHECK` turns the container `healthy` when `GET /health/ready` answers 200.
- `docker stop` sends `SIGTERM`, which Next.js handles by draining in-flight requests.

## Hot reload

`make dev` runs `next dev` in Compose, with `web/` and `packages/` bind-mounted into the container.
Native file events reach it through the bind mounts, so no polling and no Compose Watch are needed. Through
`http://localhost:4008`, an edit to a view showed in about 0.2 s and an edit to a token in
`packages/design-tokens/theme.css` in about 0.3 s, with no image rebuild, and `/_next/hmr` upgrades
to a WebSocket through Traefik.

Each `make dev` gives web a fresh anonymous `node_modules` volume, which `make down` removes. Running
`make dev` again without `make down` in between leaves the previous one dangling;
`docker volume prune` reclaims it.

## End-to-end tests

The browser suite in [`e2e/`](../e2e/) checks what only a browser sees on the home page:

- under both the light and the dark colour scheme, the heading is coloured with that theme's
  `primary`, `body` is painted with its `background`, and `main` is padded 16px on every side;
- the language order: Bahasa Indonesia with no cookie and no `Accept-Language`, the header when
  there is no cookie, and the `locale` cookie over the header;
- an unknown path answers 404 with the localised not-found page;
- the HMR websocket connects through Traefik and receives a frame.

Run `make dev`, wait until `http://localhost:4008/` answers, then run `make e2e`.

- The suite runs in its own Compose service, built on Playwright's official image, and reaches web
  through Traefik at `http://traefik:4008`, as a browser does. No browser is installed on the host.
- It is not part of `make check`, because a browser suite is slow.
- The specs and `packages/design-tokens` are mounted into the container, so editing a spec or a
  token needs no rebuild.
- After changing e2e's dependencies, `playwright.config.ts` or its Dockerfile, or after switching
  worktrees, rebuild with `docker compose --profile e2e build e2e`. The image is shared across
  worktrees.
- `@playwright/test` in `e2e/package.json` and the `mcr.microsoft.com/playwright` tag in
  `e2e/Dockerfile` name the same version and only change together: the image's browsers are built
  for that one release.
- `next.config.ts` lists `traefik` in `allowedDevOrigins`, because `next dev` refuses `/_next/*`
  requests, the HMR upgrade included, from an origin it does not list, and the suite's browser
  sends `http://traefik:4008`. It has no effect on the production image.

## Endpoints

| Path                | Answers                                                  |
| ------------------- | -------------------------------------------------------- |
| `GET /`             | The home page                                            |
| `GET /health`       | 200 with `{version, environment}` while the process runs |
| `GET /health/ready` | 200; web has no hard dependency yet                      |
