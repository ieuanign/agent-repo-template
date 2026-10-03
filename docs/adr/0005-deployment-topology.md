# Deployment topology

Service images are built in CI and pushed to a registry; the VPS pulls them and never holds a repository checkout. JS images are built from `turbo prune <service> --docker` together with a root `.dockerignore`, so mobile never enters a build context. mobile ships through the app stores with `EXPO_PUBLIC_API_URL` pointing at the VPS domain. The Compose layout follows [0004](./0004-compose-topology.md).

Staging and production share one `compose.prod.yaml`, which pulls each image by tag and never builds or mounts source; `compose.yaml` stays the dev stack. Only an environment's files differ: plain settings in `ops/env/<environment>.env`, and SOPS-encrypted secrets under `secrets/<environment>/`, where production's rule admits only the maintainers and the production deploy key. A deploy is a short CI job, or `make prod ENV=<environment>` from a maintainer's machine, that runs Compose against the VPS's Docker over SSH. Between deploys nothing is attached: the containers run under Docker's restart policy.

Deploys have no downtime. A service's migrations run first; then [docker-rollout](https://github.com/wowu/docker-rollout) starts the new container beside the old one, waits for its health check, and removes the old one, while Traefik routes only to healthy containers. A rollback is the same rollout with the previous tag and never touches the database schema, which is why every migration must keep the previous release working.

On the VPS, PostgreSQL runs as a container in `compose.prod.yaml`, with nightly backups that leave the machine and a restore that has been tested — without them the VPS disk is the only copy. A move to ECS takes RDS and rolling deployments instead, and that move is where Terraform arrives; until then the VPS's handful of resources are created once by hand.

## Considered Options

- **Clone the repository and build on the VPS** — rejected: it puts source, build tooling and build load on the production host, and what runs there is no longer the exact image CI built.
- **Plain `docker compose up -d`** — rejected: it stops the old container before the new one starts, a few seconds of errors on every deploy.
- **Blue-green switched at the proxy** — rejected: a deploy script flipping upstreams between two copies of each service, where docker-rollout is one command.
- **Kamal** — rejected: zero-downtime through its own proxy, but production would leave Compose for Kamal's configuration and stop matching dev.

## Consequences

- `secrets/` gains per-environment folders when staging first exists; until then its files are dev's.
- Each service needs a container health check, graceful shutdown on `SIGTERM`, and no state held in process memory, because two copies of it run side by side during every deploy.
