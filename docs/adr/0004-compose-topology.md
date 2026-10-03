# Compose topology

Local development runs PostgreSQL, Traefik, backend, ai and web in Docker Compose, each service with hot reload. Traefik is the single edge on one host: `/api/*` routes to backend and `/` routes to web, each route declared as labels on its own service, so a service brings its route with it. ai has no route; it is internal only and reached as `ai:50051` on the Compose network. mobile does not run in Compose: it runs natively on the host inside `devbox shell`, because the simulator and device tooling live on the host.

## Considered Options

- **Caddy** — the first choice, replaced before any service ran behind it. The zero-downtime deploy in [0005](./0005-deployment-topology.md) starts a new container beside the old one, so the proxy has to route only to healthy containers as they come and go. Traefik discovers containers from their labels and does; the deploy tool documents Traefik and nginx-proxy, not Caddy.
- **An API gateway such as Kong in front of backend** — rejected for now. Traefik already routes and limits requests per IP with its `RateLimit` middleware; authentication and per-user limits stay in backend, which needs the caller's identity to authorise anyway; API metrics come from Prometheus. Revisit when partner systems call the API directly and need their own keys, quotas and a developer portal.

## Consequences

- Clients reach the API at `http://host.docker.internal:4008/api`; mobile overrides it with `EXPO_PUBLIC_API_URL`.
- Traefik forwards `/api` unstripped: backend owns its prefix, so every URL is the same inside and outside the proxy.
