# Forward-only migrations

Migrations are plain SQL files run by golang-migrate, embedded in the `migrate` binary that ships in backend's image and run as a one-shot step before backend starts. Production only ever migrates up. A rollback redeploys the previous image and leaves the schema where it is, so every migration must keep the previous release working: a breaking change — a rename, a drop, a type change, a new `NOT NULL` — is split across two releases, the first adding the new shape beside the old and the second removing the old once the first is no longer a rollback target. A bad migration is fixed by a new migration.

## Considered Options

- **Run the down migrations automatically on rollback** — rejected: a down that drops a column deletes everything written to it since the deploy, and the previous image cannot use a schema it never knew anyway.

## Consequences

- Down migrations exist for local development only; the `migrate` binary refuses them outside `APP_ENV=dev`.
- golang-migrate neither wraps a migration in a transaction nor applies one older than the database's current version, so backend adds guards against both and against a rollback that finds the database ahead of the image. They are listed in [`backend/CLAUDE.md`](../../CLAUDE.md).
- Files live flat in `backend/migrations/`, named by timestamp rather than sequence number, because branches built in parallel would collide on the next number.
