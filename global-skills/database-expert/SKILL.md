# Skill: Database Expert

**Description**: Expert in Postgres 15 schema management and SQL optimization.

## Expertise

- Writing atomic, reversible migrations.
- Indexing strategies for patient search.
- Debugging connection pools and timeouts.

## Protocol

- Always use `internal/ports/repository.go` as the interface.
- Never write raw SQL in telegram handlers; always use the Repository adapter.
- Verify migrations against the `docker-compose.test-override.yml` environment first.
