# One PostgreSQL schema per module

Each module owns one PostgreSQL schema named after it (`projects.project`), created by the module's first migration, and its SQL reads and writes only that schema. Another module's data comes through that module's `public.go`, composed in Go as [0001](./0001-modular-monolith-module-layout.md) sets out. Ownership is then visible in the database itself rather than only in review, and a module split into a service takes its whole schema with it.

## Considered Options

- **One shared `public` schema, ownership kept by rule** — rejected: the rule holds only as long as review catches every breach, and the database says nothing about which module owns a table.

## Consequences

- Every table name in SQL carries its schema.
- Foreign keys across schemas are allowed, as integrity constraints only: dropping one when a module leaves is trivial, while un-joining a query is not.
- Reports that span modules need their own read path, composed in Go, rather than a join.
