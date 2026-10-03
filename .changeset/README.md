# Versioning with Changesets

Each service carries its own version in its `package.json`, recorded through Changesets. A change to
one service bumps only that service; there are no fixed or linked groups.

## What is versioned

- **Services** (`backend`, `ai`, and `web`, `mobile` once they exist) are versioned **by not being
  private**. Any workspace package without `"private": true` gets a version.
- **Shared packages** under `packages/` are excluded by `"private": true` and nothing else. With
  `privatePackages: { "version": false, "tag": false }` in `config.json`, marking a package private is
  the whole opt-out; `ignore`, `fixed` and `linked` stay empty.

### Why services carry a `prepublishOnly` guard

A versioned package cannot be private, so npm would otherwise accept `npm publish` for it. Each
service's `scripts.prepublishOnly` exits non-zero; npm runs it before packing (also under
`--dry-run`), so `npm publish` and `changeset publish` both abort.

### Why `bumpVersionsWithWorkspaceProtocolOnly` is on

Changesets rejects a tree where a versioned package depends on a skipped (private) one, which is
exactly a service depending on a shared package. With this flag only `workspace:` ranges couple
versions, and npm has no `workspace:` protocol, so no internal dependency is version-coupled: the
tree stays valid and bumping one service never cascades to another.

## Recording and cutting versions

Run inside devbox (`devbox shell` or `devbox run -- …`).

1. Record a change, alongside the code it describes:
   - `npx changeset add --patch|--minor|--major <name> -m "…"`, or
   - `npx changeset` for the interactive prompt.
2. Cut versions: `npx changeset version` consumes the pending changesets, bumps each named
   `package.json` and writes its `CHANGELOG.md`.
3. `npm install --package-lock-only` — the lockfile records workspace versions.
4. Commit the manifests, changelogs and lockfile together.

`npx changeset status` shows what is pending.

`changeset publish` is never run. Docker images are the release artifact: a service's
`package.json` version becomes the `VERSION` build arg its image is built with.

## Joining: web and mobile

When a service skeleton lands:

1. Give it a `package.json` with `name` and `version`, listed in the root `workspaces`.
2. Delete the `"private": true` the scaffold adds.
3. Add the same `prepublishOnly` guard as `backend/package.json`.
4. Depend on shared packages with `"*"`, never a pinned version.
5. Change nothing in `.changeset/config.json`.

Mobile additionally uses a dynamic `app.config.ts` that reads `version` from `mobile/package.json`,
so the app version has one source.
