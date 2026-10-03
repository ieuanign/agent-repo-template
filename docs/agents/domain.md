# Domain Docs

How the engineering skills should consume this repo's domain documentation when exploring the codebase.

## Before exploring, read these

- **`CONTEXT.md`** at the repo root, or
- **`CONTEXT-MAP.md`** at the repo root if it exists: it points at one `CONTEXT.md` per context. Read each one relevant to the topic.
- **`docs/adr/`**: read ADRs that touch the area you're about to work in. This repo is multi-context, so also check `<service>/docs/adr/` (e.g. `backend/docs/adr/`) for service-scoped decisions.

If any of these files don't exist, **proceed silently**. Don't flag their absence; don't suggest creating them upfront. The `/domain-modeling` skill (reached via `/grill-with-docs` and `/improve-codebase-architecture`) creates them lazily when terms or decisions actually get resolved.

## File structure

Single-context repo (most repos):

```
/
├── CONTEXT.md
├── docs/adr/
│   ├── 0001-event-sourced-orders.md
│   └── 0002-postgres-for-write-model.md
└── src/
```

Multi-context repo (presence of `CONTEXT-MAP.md` at the root) — **this repo**. Each service workspace is its own context; there is no shared `src/` root:

```
/
├── CONTEXT-MAP.md
├── docs/adr/                          ← system-wide decisions
├── ai/
│   ├── CONTEXT.md
│   └── docs/adr/                      ← context-specific decisions
├── backend/
│   ├── CONTEXT.md
│   └── docs/adr/
├── web/
│   ├── CONTEXT.md
│   └── docs/adr/
├── mobile/
│   ├── CONTEXT.md
│   └── docs/adr/
└── packages/
    └── design-tokens/
        ├── CONTEXT.md
        └── docs/adr/
```

The service directories are still being scaffolded; add a context's `CONTEXT.md` when that service lands, and register it in `CONTEXT-MAP.md`.

## Use the glossary's vocabulary

When your output names a domain concept (in an issue title, a refactor proposal, a hypothesis, a test name), use the term as defined in `CONTEXT.md`. Don't drift to synonyms the glossary explicitly avoids.

If the concept you need isn't in the glossary yet, that's a signal: either you're inventing language the project doesn't use (reconsider) or there's a real gap (note it for `/domain-modeling`).

## Flag ADR conflicts

If your output contradicts an existing ADR, surface it explicitly rather than silently overriding:

> _Contradicts ADR-0007 (event-sourced orders), but worth reopening because…_
