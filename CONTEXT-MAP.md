# Context Map

Each `CONTEXT.md` below is added when its service lands.

## Contexts

- [Backend](./backend/CONTEXT.md): the Go API and sole owner of PostgreSQL.
- [AI](./ai/CONTEXT.md): the stateless, internal-only Python AI service.
- [Web](./web/CONTEXT.md): the Next.js frontend.
- [Mobile](./mobile/CONTEXT.md): the Expo / React Native app.
- [Design tokens](./packages/design-tokens/CONTEXT.md): the design system shared by web and mobile.

## Relationships

- **Backend → AI**: backend calls ai over gRPC (`ai.v1`).
- **Web / Mobile → Backend**: HTTP `/api/*` through Traefik, via `packages/api-client`.
- **Web / Mobile → Design tokens**: both consume `packages/design-tokens` for styling.
