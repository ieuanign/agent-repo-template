# agent-repo-template

A full-stack monorepo boilerplate that a coding agent can work in from the first clone. Each service is a modular monolith in its own workspace; Turborepo runs tasks across them.

What makes it ready for an agent:

- **`CLAUDE.md` and `.claude/rules/`** hold the working rules: surgical changes, stacked pull requests, the pre-commit hook, scratch files, worktree removal, code comments and review. Where a rule is long, `.claude/rules/` holds its instruction and `.claude/reference/` the rest.
- **`.claude/settings.json`** refuses the usual forms of a forced worktree removal, and declares the skill plugins, so Claude Code offers to install them when you trust the folder: [mattpocock-skills](https://github.com/mattpocock/skills) and [ieuanign-skills](https://github.com/ieuanign/skills), which provides `/dev-loop`, the issue-to-pull-request pipeline.
- **`docs/agents/`** answers what those skills ask a repository: where issues live, the label strings, the branch and pull request formats, and the setup and full-suite commands.
- **`make check`** is the one command that says a change is done, and the pre-commit hook runs the same lint and tests on what a commit touches.

## Start a project from this template

1. On GitHub, choose **Use this template**, then clone the new repository.
2. Install [devbox](https://www.jetify.com/devbox) and Docker, then run `devbox shell`. It provides every other tool and installs the pre-commit hook.
3. Inside devbox, run `gh auth login`. devbox keeps gh's login inside the repository, apart from your global one.
4. Create your age key yourself, as Section "Getting access", step 1, of [`docs/secrets.md`](docs/secrets.md) says.
5. Run `npm ci`, then `make bootstrap`. It creates the issue labels the skills use, installs the `gh stack` extension, adds your key as the first SOPS recipient and encrypts generated local database credentials. Commit `.sops.yaml` and `secrets/postgres.sops.env`.
6. Run `make check`. When it passes, open Claude Code in the repository, trust the folder and install the plugins it offers. File an issue and run `/dev-loop <issue number>`.

To rename the project, replace the template's names, then regenerate what embeds them:

```sh
git grep -lz 'ieuanign/agent-repo-template' | xargs -0 perl -pi -e 's#ieuanign/agent-repo-template#<owner>/<name>#g'
git grep -lz 'agent-repo-template' | xargs -0 perl -pi -e 's#agent-repo-template#<name>#g'
npm install && npm run generate -w proto
```

The home page's heading is in `packages/i18n/messages/`.

## Services

| Service   | Path       | Stack                                                                                  |
| --------- | ---------- | -------------------------------------------------------------------------------------- |
| AI        | `ai/`      | Python 3.14 + uv, gRPC, modular monolith                                               |
| Backend   | `backend/` | Go 1.27 (pinned in `devbox.json`), gin, sqlx, oapi-codegen, modular monolith           |
| Frontend  | `web/`     | Next.js 16.x, Tailwind v4, npm                                                         |
| Mobile    | `mobile/`  | Expo SDK 58 / React Native 0.88, Uniwind, TanStack Query, zustand, npm                 |

## Topology

Traefik fronts one host: `/api/*` routes to backend and `/` to web. Each route is labels on its own Compose service, and exposure is opt-in, so ai has no route: it is internal only, reached by backend over gRPC as `ai:50051` on the Compose network. backend is the only PostgreSQL client. mobile runs natively on the host inside `devbox shell`, never in Compose.

## Layout

```
.
├── ai/                  # Python modular monolith
├── backend/             # Go modular monolith
├── web/                 # Next.js frontend
├── mobile/              # React Native app
├── packages/
│   ├── api-client/      # typed backend client generated from its OpenAPI spec
│   ├── design-tokens/   # shared design system tokens (Tailwind + Uniwind)
│   └── i18n/            # locales and catalogue text shared by web and mobile
├── proto/               # gRPC contract between backend and ai (buf)
├── e2e/                 # Playwright browser suite, run in its own container
├── e2e-mobile/          # Maestro suite, run on the host
├── docs/
│   └── adr/             # system-wide architecture decision records
├── secrets/             # SOPS-encrypted env files (*.sops.env)
├── compose.yaml         # local stack: PostgreSQL, services, Traefik
├── Makefile             # runtime verbs
├── turbo.json           # task pipeline
├── .sops.yaml           # SOPS encryption rules
├── CONTEXT-MAP.md       # map of the per-service domain contexts
└── devbox.json          # toolchain
```

## Tooling

| Tool                   | Purpose                                                                        |
| ---------------------- | ------------------------------------------------------------------------------ |
| devbox                 | Pins the toolchain; everything runs inside `devbox shell`                      |
| Turborepo              | Runs `build`, `test`, `lint` across workspaces with caching                    |
| Changesets             | Versions backend, ai, web and mobile independently                             |
| buf                    | Lints, breaking-checks and generates the `proto/` contract into Go and Python  |
| squawk                 | Lints backend's migrations for changes that break the previous release         |
| SOPS + age             | Encrypts secrets committed under `secrets/`                                    |
| Docker Compose         | Local stack: PostgreSQL, Traefik, backend, ai, web with hot reload             |
| Traefik                | Edge proxy in front of the services, routed by their Compose labels            |
| Dockerfile             | One per service, for its production image                                      |
| Tailwind / Uniwind     | Shared design system across web and mobile via `packages/design-tokens`        |
| Makefile               | Entry point for the runtime verbs                                              |
| PostgreSQL             | Primary database, owned by backend                                             |
| Graphify               | Knowledge graph of the codebase for navigation — deferred                      |

## Getting started

Prerequisites: devbox and Docker on the host; `devbox shell` provides the rest of the toolchain. mobile also needs Xcode or Android Studio, as [`mobile/README.md`](mobile/README.md) says.

```sh
devbox shell
make help   # list every make target
make dev    # start the local stack
make down   # stop it
make check  # lint and test everything
make e2e    # run the browser suite against the running stack
make e2e-mobile platform=ios|android  # run the Maestro suite on a booted Simulator or Emulator
```

devbox installs a pre-commit hook that fixes and checks the workspaces a commit touches: it runs `lint:fix` and stops the commit if that changed a file, leaving the fix unstaged for review, then runs `lint` and `test`; a failure stops the commit.

## Docs

System-wide decisions are recorded in `docs/adr/`; service-scoped ones in `<service>/docs/adr/`. Each service carries a `README.md` for its role and stack and a `CLAUDE.md` for its hard rules; the ones tied to particular files are path-scoped rules under `.claude/rules/`. `CONTEXT-MAP.md` lists the domain contexts and how they relate.

## Maintaining the template

Work on the template itself the way any project from it works, with one exception: `make bootstrap` writes your recipient into `.sops.yaml` and creates `secrets/postgres.sops.env`, and neither is ever committed here. A project made from the template must start with no recipients, so its first `make bootstrap` adds its own.
