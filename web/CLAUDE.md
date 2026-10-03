# web conventions

The Next.js web app, server-rendered with the App Router, behind Traefik at `/`
([ADR 0004](../docs/adr/0004-compose-topology.md)). What it has and how to run it is in
[`README.md`](./README.md).

## Libraries (hard rules)

| Concern       | Use                                                                            | Never                                     |
| ------------- | ------------------------------------------------------------------------------ | ----------------------------------------- |
| Framework     | Next.js with the App Router                                                    | the Pages Router                          |
| Styling       | Tailwind CSS v4 through PostCSS, with `@agent-repo-template/design-tokens`     | colours outside the tokens                |
| Components    | shadcn/ui on Radix, copied into `src/components/`                              |                                           |
| Animation     | motion through `LazyMotion` and `m.*`; `tw-animate-css` in the kit             | the full `motion.*`                       |
| Icons         | Tabler, through the one `Icon` atom                                            |                                           |
| Translation   | next-intl, without its locale routing                                          | a locale URL segment; i18next             |
| Fonts         | `next/font`, with Inter self-hosted at build                                   | a font CDN at runtime                     |
| Lint          | oxlint                                                                         | ESLint; `next lint`                       |
| Format        | Prettier, one root config                                                      | Biome                                     |
| Types         | `tsc`, with no emit                                                            |                                           |
| Unit tests    | Jest through `next/jest`, React Testing Library, `@testing-library/user-event` | Vitest; `act` with `fireEvent` by default |
| End-to-end    | Playwright                                                                     | Cypress                                   |
| Configuration | the hand-written env helper                                                    | a validation library                      |
| Logging       | the console, through one reporting function                                    | a logging library                         |

## Layout

```
src/instrumentation.ts  # register: logs the startup line and starts the backend report
src/app/                # routing only: layout, pages, metadata, global CSS and the theme reset
src/ui/<View>/          # one route's view, mirroring src/app
src/core/               # constants, env and utilities
src/core/http.ts        # the one place a backend client is built
src/core/apiUrls.ts     # API_URLS: every backend path web calls
src/components/         # the kit: atoms and molecules as flat files, organisms in folders
src/lib/utils.ts        # cn, the kit's class-name merge
src/lib/i18n/           # next-intl's request configuration and the locale resolver
src/mocks/              # the MSW server and its default handlers, for Jest
messages/               # web's own id and en catalogues; shared text comes from @agent-repo-template/i18n
```

- `src/app/` holds routing only. A route file renders its view and sets metadata; the markup lives in
  the view.
- One view per route lives in `src/ui/<View>/`, mirroring the routes.
- `src/core/` holds constants, env and utilities; `src/lib/` holds third-party setup.
- `src/components/`, `src/lib/`, `src/context/` and `src/hooks/` are created when they first hold
  something, never ahead of it.
- Imports are absolute: `@/` for `src/`, `@messages/` for the catalogues. Never a relative path, not
  even to a file beside the importer.
- There is no domain-module layer. Nothing lives under `src/app/api/`: Traefik sends `/api` to
  backend, so a web route there is unreachable.

## Components

Atomic design. A component is as small as it can usefully be, and its **tier** decides where it lives:

| Tier         | What it is                                                       | Where                    |
| ------------ | ---------------------------------------------------------------- | ------------------------ |
| **Atom**     | One element, no domain knowledge — a button, an input, a spinner | `src/components/`        |
| **Molecule** | A few atoms, one job — a field with its label and error, a toast | `src/components/`        |
| **Organism** | A composed section — a header, a project summary panel           | `src/components/<Name>/` |
| **View**     | One route's whole view, assembling the above                     | `src/ui/<View>/`         |

**A component starts local and moves on its second caller.** What one view renders lives in that
view's folder; it earns a place in `src/components/` when a second view wants it.

**An atom knows nothing about the domain.** Domain vocabulary enters at the organism tier.

**Component files are PascalCase** (`Button.tsx`), and every import names the file in the same case:
the production image builds on Linux, which is case-sensitive. A folder's entry is `index.tsx`.
oxlint enforces it under `src/components/` and `src/ui/`.

**A file copied from the kit is edited before it is committed.** Every colour and radius class names
a token, the file goes where its atomic tier puts it, and a lowercase import the copy carries
(`@/components/button`) is fixed to the PascalCase file by hand.

**Animation is the exception, and changes only `transform` and `opacity`.** motion comes in through
`LazyMotion` with `domAnimation` and animates `m.*` elements; `tw-animate-css` drives the kit's own
opening and closing.

**Icons come through one `Icon` atom with a fixed stroke width**, the one place icons are imported,
with named imports from `@tabler/icons-react`.

**Props are camelCase.** An API field's name stops at the destructure that reads it.

**`"use client"` goes on the leaf that needs state**, never the section containing it — everything
below the directive is client code, so hoisting it ships the whole subtree.

## File layout within a folder

The component file builds the component. Everything else has a home beside it:

```
src/ui/Home/
  index.tsx          # the component and its markup — the route renders this
  constants.ts       # literals, lookup tables, copy, tuning values
  utils.ts           # pure functions — formatting, parsing, deriving
  *.test.ts(x)       # colocated
```

**Constants live in `constants.ts`**; one shared across views moves to `src/core/`. **Pure functions
live in `utils.ts`**, which is also what lets them be tested without rendering.

## React practice

**Waterfalls — the first thing to fix, ahead of any `memo` tuning.**

- Independent requests run together (`Promise.all`).
- Cheap synchronous checks come before the `await` they might make unnecessary, and the `await`
  itself moves into the branch that needs it.
- `Suspense` boundaries go around slow sections so the rest of the page streams.

**Server**

- `React.cache()` dedupes a loader within one request.
- Module scope is shared across requests, so request-scoped data held there leaks between users.
  Keep it in the request.
- Send the client the fields it renders, not the row the server read.

**Bundle**

- Import the exact module, never a barrel re-export. A package Next's `optimizePackageImports`
  rewrites to the exact module at build, such as `@tabler/icons-react`, takes named imports.
- Heavy client components come in through `next/dynamic`.
- Import paths stay statically analysable — a computed path bundles the whole directory.

**Re-renders**

- Subscribe to the narrowest thing: a derived boolean over the raw object, and nothing for a value
  only an event handler reads.
- Derived values compute during render; an effect that syncs one into state renders twice.
- Effect dependencies are primitives, not freshly built objects.
- A costly initial value takes the callback form: `useState(() => parse(raw))`.
- `memo` answers a measured problem. A component taking only primitive props does not need it.
- `useTransition` drives pending UI in preference to a hand-managed flag.

**Rendering**

- Conditional JSX uses a ternary — `&&` renders a stray `0` when the left side is a falsy number.

## Tokens and spacing

- Only the tokens' colour, radius and font classes exist. `src/app/theme-reset.css` clears Tailwind's
  `--color-*`, `--radius-*` and `--font-*` and sits between `tailwindcss` and the tokens in
  `src/app/globals.css`, so a class naming anything else generates no CSS and no error.
- The shared tokens file never holds a reset, because mobile compiles it
  ([design-tokens README](../packages/design-tokens/README.md)). Resets live only in
  `theme-reset.css`.
- The colours are the kits' 19 names, each with a light and a dark value: `background`,
  `foreground`, `card`, `card-foreground`, `popover`, `popover-foreground`, `primary`,
  `primary-foreground`, `secondary`, `secondary-foreground`, `muted`, `muted-foreground`, `accent`,
  `accent-foreground`, `destructive`, `destructive-foreground`, `border`, `input` and `ring`.
- `globals.css` declares the `light` and `dark` custom variants. A `light` or `dark` class on an
  element or an ancestor picks that theme; with neither, the device's colour scheme does, so the
  first paint follows the device with no script.
- The base layer in `globals.css` paints `body` with `bg-background text-foreground` and gives every
  element `border-border` and an outline from `ring`.
- The radius has five steps: `rounded-sm`, `rounded-md`, `rounded-lg`, `rounded-xl` and
  `rounded-4xl`.
- No hex colour literal appears under `web/`, in code, CSS or docs.
- Every numeric spacing class is its number times the 4px base, so `p-4` is 16px. Tailwind's own
  spacing is not reset; the tokens' base replaces it.
- Named spacing steps take role names such as `gutter` or `section`, never Tailwind's size words
  (`xs`, `sm`, `md`, `lg`, `xl` and the rest), because `w-*`, `min-w-*`, `max-w-*` and `basis-*`
  read spacing before container sizes.
- Inter comes through `next/font/google`, self-hosted at build. `globals.css` points `--font-sans` at
  next/font's variable in an `@theme inline` block after the tokens import.

## Configuration and startup

- Settings come from the environment when the container starts: Compose's `environment:` in dev, and
  `ops/env/<environment>.env` in staging and production
  ([ADR 0005](../docs/adr/0005-deployment-topology.md)).
- They are read through `withDefault` in `src/core/env.ts` at the moment they are needed, never held
  in module-level constants or in `next.config`'s `env`. Nothing is baked in per environment at build.
  An empty value counts as unset.
- There is no `NEXT_PUBLIC_*` variable and no boot check.

| Variable       | Notes                                                                       |
| -------------- | --------------------------------------------------------------------------- |
| `APP_ENV`      | `dev`, `staging` or `production`; default `dev`                             |
| `VERSION`      | Default `dev`                                                               |
| `GIT_SHA`      | Default `dev`; appears only in the startup line                             |
| `API_BASE_URL` | backend's base URL for server-side calls; default `http://traefik:4008/api` |

- `register` in `src/instrumentation.ts` logs one JSON line per server start, with the keys backend
  and ai log: `service` (`web`), `version`, `commit` and `environment`. It is the only place the
  commit appears.
- `register` then starts the backend report in `src/core/backendReport.ts` without awaiting it. It
  calls backend's liveness route up to 11 times, each try abandoned after 10 s, waiting 2, 4, 8, 16,
  32, then 60 s five times between tries (about six minutes), because backend starts after migrate
  and seed. Any try without a success envelope — a network error, the abandonment, a failure
  envelope, or Traefik's plain-text answer — fails silently and is retried. The first success logs
  one `backend reachable` JSON line with backend's `version` and `environment`; the 11th failure
  reports one warning through `reportError`. It never rejects, and never blocks boot or readiness.
- Nothing is held in process memory between requests — no module-level mutable state or cache —
  because two copies run side by side during a deploy.

## Build

- `Dockerfile` builds with the repository root as its context, under the root
  [`.dockerignore`](../.dockerignore).
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
  - `build` runs `turbo run build --filter=web`. [`turbo.json`](./turbo.json) declares `.next/**`,
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
  deploy waits on it ([ADR 0005](../docs/adr/0005-deployment-topology.md)).
- The command is `/web/server.js` under the image's own node entrypoint, with no npm, `next start` or
  wrapper, so `SIGTERM` reaches Next.js's handler, which drains in-flight requests.
- The prune stage copies the whole context into a builder layer, so the root `.dockerignore` keeps
  out secrets, env files and gitignored local state. Never add a `Dockerfile.dockerignore`: it
  replaces the root rules.
- A base bump changes tag and digest together, the digest read from
  `docker buildx imagetools inspect <image:tag>`.

## Tests and lint

- `npm run lint -w web` runs `scripts/lint.sh`: oxlint, then `prettier --check web` from the
  repository root (Prettier reads its ignore files only from its working directory), then
  `next typegen` and `tsc --noEmit`. It writes only gitignored files.
- Fix (`lint:fix`): `scripts/lint-fix.sh` runs `oxlint --fix`, then `prettier --write web` from the
  repository root. The pre-commit hook runs it before its gate.
- Jest runs through `next/jest`, with `node` as the default environment; a component test file opts
  into jsdom itself.
- Tests are colocated as `*.test.ts(x)`.
- `jest.setup.ts` starts one MSW server from `src/mocks/server.ts` with `onUnhandledRequest: "error"`,
  resets its handlers before each test and closes it after the file. It starts only where the Fetch
  globals exist, the Node environment; jsdom files run without it. `jest.config.js` adds msw and the
  ES-module-only packages it loads to the transform allowlist beside next-intl's.
- Interaction goes through `userEvent`; `act` with `fireEvent` only when `userEvent` cannot express
  the case.
- Async Server Components are not unit-tested.

## Calling backend

- Every call goes through `http()` in `src/core/http.ts` with a path from `API_URLS` in
  `src/core/apiUrls.ts`. No other file builds a client or writes a backend path as a
  string; `API_URLS` is checked against the client's generated paths with `satisfies`.
- Server-side calls go through Traefik (`API_BASE_URL`), never `backend:8080`: during a rollout
  Compose DNS also answers with an unhealthy container, and Traefik routes only to healthy ones.
- The client is built per call, so the base URL is read at call time; never a module-level client.
- Client components call through TanStack Query around the same client, with the same-origin `/api`
  base. Both are added with the first client component that needs backend, never ahead of it.
- Never a `NEXT_PUBLIC_*` setting for backend's URL: one image is promoted across environments.

## Reading a response

- backend answers with its envelope, `{meta, status, code, message, payload}`. On success read
  `data.payload`; on failure read `error.code`, which is never null there.
- Nothing rewrites responses: the client has no middleware and unwraps nothing, so `{data, error,
response}` arrive as backend sent them. `204` and `304` have no body.

## Showing an error

- The text a user sees is chosen by `code` from an `errorCode` namespace, named after the OpenAPI
  enum its keys come from. A code with no text falls back to the text of the global kind its HTTP
  status maps to, then to one generic line.
- `message` is for developers: it goes only to `reportError`, never onto a page.
- A Server Component handles the failures its page expects and throws only unexpected ones, because
  production reduces a thrown error to a digest.
- The `errorCode` namespace and the fallback helper arrive with the first page that shows a backend
  error. `errorCode` lives in `@agent-repo-template/i18n` from its first use, because both apps show it.

## Testing

- MSW mocks the network; never a hand-mocked `fetch` or a mocked client. A test overrides a default
  with `server.use`.
- Handlers answer with backend's full envelope, `meta` included, and take their paths from
  `API_URLS`.

## Runtime

- `GET /health`: liveness, always 200 with `{version, environment}`, no dependency checked. `version`
  is `VERSION` and `environment` is `APP_ENV`, each through `src/core/env.ts`, so each falls back to
  `dev` exactly as in the startup line. The body holds nothing else: the route is public, and the
  commit appears only in the startup line.
- `GET /health/ready`: readiness, 200 or 503 from web's hard dependencies. It has none yet, so it
  always answers 200. It never depends on backend or any other service: a backend outage would pull
  web out of Traefik, and users would see Traefik's bare 503 instead of web's own error page.
- Both are Route Handlers that `await connection()` before reading anything, so the values are read
  per request and never frozen at build or import.
- Both answer plain JSON (`application/json`), never backend's `{meta, status, code, message, payload}`
  envelope: web's health is not an API.

## Language

- web renders in Bahasa Indonesia (`id`) by default and in English (`en`) when asked. Before sign-in
  the locale is, in order:
  1. the `locale` cookie, when its value is exactly `id` or `en`; any other value is ignored;
  2. otherwise the best match in `Accept-Language`, by q value, on the primary language subtag;
  3. otherwise `id`.
- The locale is worked out on the server, per request, in `src/lib/i18n/request.ts`, so the first
  render is already in the right language. It is never kept in local storage: the server cannot read
  it, so the page would render in one language and then flip. It never comes from an IP lookup: an
  IP address is personal data under UU PDP, and where someone is says nothing about what they read.
- No locale in the URL. There is no `[locale]` segment and no proxy or middleware; next-intl runs
  without its routing, and `/` serves both languages.
- Every user-facing string comes from a catalogue, one top-level namespace per view, named after its
  folder under `src/ui/` in camelCase (`ui/NotFound` → `notFound`). Namespaces and keys are both
  camelCase at every level. web's own namespaces live in `messages/id.json` and `messages/en.json`,
  each imported by a literal `@messages/` path; shared namespaces come from
  [`@agent-repo-template/i18n`](../packages/i18n/README.md). `src/lib/i18n/request.ts` merges the two per
  language as `{...shared, ...own}`. Every key is in both languages; `tsc` fails when an `en`
  catalogue lacks a key its `id` has, in web and in the package, and when a web namespace has the
  same name as a shared one.
- Identical text is shared; text that differs stays in its app. A whole namespace starts in its app
  and moves to `@agent-repo-template/i18n` once the other app needs the same text; a namespace is never
  split, and moving one changes no call site. A shared namespace that loses one of its two apps moves
  back into the app still reading it, or is deleted if neither does, in the change that drops the
  last call.
- The namespace is always a string literal: `useTranslations("errorCode")`, or a sub-namespace such
  as `"errorCode.project"`. The key may be dynamic (`t(error.code)`). A root translator with a built
  path is never used. Review enforces it.
- The root layout's `NextIntlClientProvider` receives only the namespaces client components read
  (`error` today), through `pick` in `src/lib/i18n/utils.ts`. Server components read the whole
  catalogue. A new client component adds its namespace to the pick list.
- `<html lang>` is the resolved locale, read through next-intl's `getLocale()` in the root layout.
- Values are formatted through display presets, use-intl `formats` that live in
  `@agent-repo-template/i18n` and are written with the first screen on either app that formats a value:

  | Kind          | `id`           | `en`           | Pinned options                                                           |
  | ------------- | -------------- | -------------- | ------------------------------------------------------------------------ |
  | Rupiah        | `Rp 1.234.567` | `Rp 1,234,567` | `currency: "IDR"`, `currencyDisplay: "narrowSymbol"`, no fraction digits |
  | Date          | `2 Okt 2026`   | `2 Oct 2026`   | medium date style                                                        |
  | Time          | `10.04`        | `10:04`        | `hourCycle: "h23"`                                                       |
  | Other numbers | `1.234`        | `1,234`        | grouping by language                                                     |

- After sign-in the rules belong to whichever module owns the signed-in user, which writes the
  cookie. Nothing here writes the cookie, and there is no language switcher.

## Errors

However a page fails, the user stays on a page of web's own. Three files at the root of `src/app/`
cover it, each rendering its view from `src/ui/`:

| File               | Covers                                                | View             |
| ------------------ | ----------------------------------------------------- | ---------------- |
| `error.tsx`        | an error thrown below the root layout                 | `ui/Error`       |
| `global-error.tsx` | an error in the root layout itself                    | `ui/GlobalError` |
| `not-found.tsx`    | `notFound()` anywhere, and every URL no route matches | `ui/NotFound`    |

- There is no `global-not-found.tsx`. It is experimental in Next.js 16.3 and needed only with several
  root layouts or a dynamic root segment
  ([Next.js `not-found`](https://nextjs.org/docs/app/api-reference/file-conventions/not-found)); the
  root `not-found.tsx` already renders inside the root layout, in the request's language.
- `global-error.tsx` replaces the root layout, so it renders its own `<html lang="id">` and `<body>`,
  imports `globals.css`, and sets its title with React's `<title>` element. Its copy lives in
  `ui/GlobalError/constants.ts` in both languages, Bahasa Indonesia first, because the catalogues and
  the chosen language come from the layout that failed.
- Both retry buttons call the `reset` Next.js passes
  ([Next.js `error`](https://nextjs.org/docs/app/api-reference/file-conventions/error)).
- These pages render fixed copy only: never the error's message, stack or digest, the URL or request
  data. Next.js redacts only Server Component errors in production.
- `reportError` in `src/core/reporting.ts` is the one place a caught error leaves the app.
  `error.tsx` and `global-error.tsx` call it from an effect keyed on the error. It logs with
  `console.error`; an error tracker's capture call goes there, and nowhere else.
- Component tests opt into jsdom per file with the `@jest-environment jsdom` docblock, drive
  interaction through `userEvent`, and wrap a translated component in `NextIntlClientProvider` with
  the real `id` catalogue holding its namespace, web's `messages/id.json` or the package's
  `messages.id`, never a mock of next-intl. `jest.config.js` exempts next-intl and the ES-module-only
  packages it loads from next/jest's `transformIgnorePatterns`.
- An unmatched URL and `notFound()` both answer 404. There is no root `loading.tsx`: a loading
  boundary makes every page below it stream a 200 before it runs, so `notFound()` there would show
  the 404 page with a `noindex` tag but answer 200
  ([Next.js `loading`, Status Codes](https://nextjs.org/docs/app/api-reference/file-conventions/loading#status-codes)).
  A segment adds its own `loading.tsx` only where it awaits slow data and never calls `notFound()`.

<!-- BEGIN:nextjs-agent-rules -->

# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` (resolved from this file's directory; in monorepos the `next` package may not be visible from the repo root) before writing any code. Heed deprecation notices.

This block is written and re-added by `next dev` — verify at `node_modules/next/dist/server/lib/generate-agent-files.js`. Removing it from a diff only re-creates the uncommitted change; committing it with your work keeps the tree clean.

<!-- END:nextjs-agent-rules -->
