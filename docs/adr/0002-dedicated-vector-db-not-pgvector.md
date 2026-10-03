# Dedicated vector DB, not pgvector

ai is stateless and backend is the only PostgreSQL client. When RAG arrives, ai gets its own dedicated vector database rather than reaching into backend's PostgreSQL, so the database keeps a single owner and ai stays independently deployable. Which vector database is deferred until RAG is actually built.

## Considered Options

- **pgvector in backend's PostgreSQL** — rejected: it would make ai a second PostgreSQL client and couple its storage to backend's schema and migrations.
